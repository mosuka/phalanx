package metastore

import (
	"context"
	"encoding/base64"
	"errors"
	"net/url"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/mosuka/phalanx/clients"
	phalanxerrors "github.com/mosuka/phalanx/errors"
	"go.uber.org/zap"
	// "github.com/aws/aws-sdk-go-v2/service/dynamodbstreams"
)

const (
	partitionKeyName = "pk"
	sortKeyName      = "sk"
	partitionValue   = "metadata"
)

var (
	// ErrRecordNotFound returned when a get returns no results
	ErrRecordNotFound = errors.New("record not found")

	// ErrDuplicateRecord returned when a conflict occurs putting a record due to duplicate
	ErrDuplicateRecord = errors.New("record already exists")
)

type kv struct {
	Partition string `dynamodbav:"pk"`
	Path      string `dynamodbav:"sk"`
	Version   string `dynamodbav:"version"`
	Value     string `dynamodbav:"value"`
}

type DynamodbStorage struct {
	client *dynamodb.Client
	// stream         *dynamodbstreams.Client
	tableName      string
	root           string
	logger         *zap.Logger
	ctx            context.Context
	requestTimeout time.Duration
	stopWatching   chan bool
	events         chan StorageEvent
}

func NewDynamodbStorageWithUri(uri string, logger *zap.Logger) (*DynamodbStorage, error) {
	metastorelogger := logger.Named("dynamodb")

	client, err := clients.NewDynamoDBClientWithUri(uri)
	if err != nil {
		logger.Error(err.Error(), zap.String("uri", uri))
		return nil, err
	}

	// Parse URI.
	u, err := url.Parse(uri)
	if err != nil {
		logger.Error(err.Error(), zap.String("uri", uri))
		return nil, err
	}
	if u.Scheme != SchemeType_name[SchemeTypeDynamodb] {
		err := phalanxerrors.ErrInvalidUri
		logger.Error(err.Error(), zap.String("uri", uri))
		return nil, err
	}

	// stream, err := clients.NewDynamoDBStreamsClientWithUri(uri)
	// if err != nil {
	// 	logger.Error(err.Error(), zap.String("uri", uri))
	// 	return nil, err
	// }

	root := u.Path
	if root == "" {
		root = "/"
	}

	dynamodbStorage := &DynamodbStorage{
		client: client,
		// stream:         stream,
		tableName:      u.Host,
		root:           root,
		logger:         metastorelogger,
		ctx:            context.Background(),
		requestTimeout: 3 * time.Second,
		stopWatching:   make(chan bool),
		events:         make(chan StorageEvent, storageEventSize),
	}

	if err := dynamodbStorage.createTable(); err != nil {
		return nil, err
	}

	dynamodbStorage.watch()

	return dynamodbStorage, nil
}

func (m *DynamodbStorage) watch() error {
	// Watch file system event.
	go func() {
		for {
			select {
			case cancel := <-m.stopWatching:
				// check
				if cancel {
					return
				}
				// TODO: implement
				// case DynamoDB events:
				// Catches changes made to the database and sends storage events to the event channel.
			}
		}
	}()

	return nil
}

// Replace the path separator with '/'.
func (m *DynamodbStorage) makePath(path string) string {
	return filepath.ToSlash(filepath.Join(filepath.ToSlash(m.root), filepath.ToSlash(path)))
}

func (m *DynamodbStorage) Get(path string) ([]byte, error) {
	fullPath := m.makePath(path)

	res, err := m.client.GetItem(m.ctx, &dynamodb.GetItemInput{
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

	if res.Item == nil {
		return nil, ErrRecordNotFound
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

	rec := &kv{
		Partition: partitionValue,
		Path:      fullPath,
		Value:     base64.RawURLEncoding.EncodeToString(value),
	}

	attr, err := attributevalue.MarshalMap(rec)
	if err != nil {
		m.logger.Error(err.Error(), zap.String("key", fullPath))
		return err
	}

	// A simple PutItem will suffice. There is no need to worry about conditions.
	_, err = m.client.PutItem(m.ctx, &dynamodb.PutItemInput{
		TableName: aws.String(m.tableName),
		Item:      attr,
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

	res, err := m.client.Query(m.ctx, &dynamodb.QueryInput{
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

	_, err := m.client.DeleteItem(m.ctx, &dynamodb.DeleteItemInput{
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

	res, err := m.client.GetItem(m.ctx, &dynamodb.GetItemInput{
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

func (m *DynamodbStorage) Events() <-chan StorageEvent {
	return m.events
}

func (m *DynamodbStorage) Close() error {
	m.stopWatching <- true

	return nil
}

func (m *DynamodbStorage) createTable() error {
	_, err := m.client.CreateTable(m.ctx, &dynamodb.CreateTableInput{
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
		StreamSpecification: &types.StreamSpecification{
			StreamEnabled:  aws.Bool(true),
			StreamViewType: types.StreamViewTypeNewAndOldImages,
		},
	})
	if err != nil {
		var rne *types.ResourceInUseException
		if errors.As(err, &rne) {
			m.logger.Warn(err.Error(), zap.String("table_name", m.tableName))
			return nil
		}

		return err
	}

	return nil
}
