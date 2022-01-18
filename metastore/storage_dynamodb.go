package metastore

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go/logging"
	"go.uber.org/zap"
)

const (
	partitionKeyName = "pk"
	sortKeyName      = "sk"
	partitionValue   = "metadata"
)

type kv struct {
	Partition string `dynamodbav:"pk"`
	Path      string `dynamodbav:"sk"`
	Version   string `dynamodbav:"version"`
	Value     string `dynamodbav:"value"`
}

type DynamodbStorage struct {
	dynamoSvc      *dynamodb.Client
	tableName      string
	root           string
	logger         *zap.Logger
	ctx            context.Context
	requestTimeout time.Duration
}

func NewDynamodbStorage(uri string, logger *zap.Logger) (*DynamodbStorage, error) {
	metastorelogger := logger.Named("dynamodb")

	ctx := context.Background()

	u, awsCfg, err := buildAwsCfg(uri)
	if err != nil {
		return nil, err
	}

	if u.Query().Get("aws_sdk_request_logging") == "true" {
		awsCfg.Logger = logging.LoggerFunc(func(classification logging.Classification, format string, v ...interface{}) {
			metastorelogger.Info(fmt.Sprintf(format, v...))
		})

		awsCfg.ClientLogMode = aws.LogRetries | aws.LogRequest
	}

	ds := &DynamodbStorage{
		dynamoSvc:      dynamodb.NewFromConfig(awsCfg),
		tableName:      u.Host,
		root:           u.Path,
		logger:         metastorelogger,
		ctx:            ctx,
		requestTimeout: 3 * time.Second,
	}

	// primarily used for testing locally
	if u.Query().Get("create_table") == "true" {
		err := ds.createTable()
		if err != nil {
			return nil, err
		}
	}

	return ds, nil
}

// Replace the path separator with '/'.
func (m *DynamodbStorage) makePath(path string) string {
	return filepath.ToSlash(filepath.Join(filepath.ToSlash(m.root), filepath.ToSlash(path)))
}

func (m *DynamodbStorage) Get(path string) ([]byte, error) {
	fullPath := m.makePath(path)

	res, err := m.dynamoSvc.GetItem(m.ctx, &dynamodb.GetItemInput{
		TableName: aws.String(m.tableName),
		Key: map[string]types.AttributeValue{
			partitionKeyName: &types.AttributeValueMemberS{
				Value: partitionValue,
			},
			sortKeyName: &types.AttributeValueMemberS{
				Value: fullPath,
			},
		},
		ConsistentRead: aws.Bool(true), // enable consistent reads as we need this for atomic reads
	})
	if err != nil {
		m.logger.Error(err.Error(), zap.String("key", fullPath))
		return nil, err
	}

	m.logger.Info("get record", zap.String("fullPath", fullPath))

	rec := new(kv)
	err = attributevalue.UnmarshalMap(res.Item, rec)
	if err != nil {
		return nil, err
	}

	return base64.RawStdEncoding.DecodeString(rec.Value)
}

func (m *DynamodbStorage) Put(path string, value []byte) error {
	fullPath := m.makePath(path)

	rec := &kv{Partition: partitionValue, Path: fullPath, Value: base64.RawURLEncoding.EncodeToString(value)}

	attr, err := attributevalue.MarshalMap(rec)
	if err != nil {
		m.logger.Error(err.Error(), zap.String("key", fullPath))
		return err
	}

	// this adds a condition which checks if the sort key already exists, if it does
	// this operation will return a condition error.
	existCond := expression.AttributeNotExists(expression.Name(sortKeyName))
	condExpr, err := expression.NewBuilder().WithCondition(existCond).Build()
	if err != nil {
		m.logger.Error(err.Error(), zap.String("key", fullPath))
		return err
	}

	_, err = m.dynamoSvc.PutItem(m.ctx, &dynamodb.PutItemInput{
		TableName:                 aws.String(m.tableName),
		Item:                      attr,
		ConditionExpression:       condExpr.Condition(),
		ExpressionAttributeNames:  condExpr.Names(),
		ExpressionAttributeValues: condExpr.Values(),
	})
	if err != nil {
		m.logger.Error(err.Error(), zap.String("key", fullPath))
		return err
	}

	m.logger.Info("put record", zap.String("fullPath", fullPath))

	return nil
}

func (m *DynamodbStorage) List(prefix string) ([]string, error) {
	prefixPath := m.makePath(prefix)

	keyCond := expression.
		Key(partitionKeyName).Equal(expression.Value(partitionValue)).
		And(expression.Key(sortKeyName).BeginsWith(prefixPath))

	keyExpr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		m.logger.Error(err.Error(), zap.String("key", prefixPath))
		return nil, err
	}

	res, err := m.dynamoSvc.Query(m.ctx, &dynamodb.QueryInput{
		TableName:                 &m.tableName,
		KeyConditionExpression:    keyExpr.KeyCondition(),
		ExpressionAttributeNames:  keyExpr.Names(),
		ExpressionAttributeValues: keyExpr.Values(),
		ConsistentRead:            aws.Bool(true), // enable consistent reads as we need this for atomic reads
	})
	if err != nil {
		m.logger.Error(err.Error(), zap.String("key", prefixPath))
		return nil, err
	}

	paths := make([]string, 0)

	for _, item := range res.Items {
		rec := new(kv)
		err = attributevalue.UnmarshalMap(item, rec)
		if err != nil {
			m.logger.Error(err.Error(), zap.String("key", prefixPath))
			return nil, err
		}

		paths = append(paths, rec.Path[len(prefixPath):])
	}

	return paths, nil
}

func (m *DynamodbStorage) Delete(path string) error {
	fullPath := m.makePath(path)

	_, err := m.dynamoSvc.DeleteItem(m.ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(m.tableName),
		Key: map[string]types.AttributeValue{
			partitionKeyName: &types.AttributeValueMemberS{
				Value: partitionValue,
			},
			sortKeyName: &types.AttributeValueMemberS{
				Value: fullPath,
			},
		},
	})
	if err != nil {
		m.logger.Error(err.Error(), zap.String("key", fullPath))
		return err
	}

	return nil
}

func (m *DynamodbStorage) Exists(path string) (bool, error) {
	fullPath := m.makePath(path)

	res, err := m.dynamoSvc.GetItem(m.ctx, &dynamodb.GetItemInput{
		TableName: aws.String(m.tableName),
		Key: map[string]types.AttributeValue{
			partitionKeyName: &types.AttributeValueMemberS{
				Value: partitionValue,
			},
			sortKeyName: &types.AttributeValueMemberS{
				Value: fullPath,
			},
		},
		ConsistentRead: aws.Bool(true), // enable consistent reads as we need this for atomic reads
	})
	if err != nil {
		m.logger.Error(err.Error(), zap.String("key", fullPath))
		return false, err
	}

	exists := res.Item != nil

	m.logger.Info("check record exists", zap.String("fullPath", fullPath), zap.Bool("exists", exists))

	return exists, nil
}

func (m *DynamodbStorage) Close() error {
	return nil
}

func (m *DynamodbStorage) createTable() error {
	_, err := m.dynamoSvc.CreateTable(m.ctx, &dynamodb.CreateTableInput{
		TableName: &m.tableName,
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String(partitionKeyName),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String(sortKeyName),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String(partitionKeyName),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String(sortKeyName),
				KeyType:       types.KeyTypeRange,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		var rne *types.ResourceInUseException
		if errors.As(err, &rne) {
			return nil
		}

		return err
	}

	return nil
}
