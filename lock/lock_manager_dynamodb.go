package lock

import (
	"context"
	"errors"
	"net/url"
	"time"

	"cirello.io/dynamolock/v2"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/mosuka/phalanx/clients"
	phalanxerrors "github.com/mosuka/phalanx/errors"
	"go.uber.org/zap"
)

type DynamoDBLockManager struct {
	client *dynamolock.Client
	table  string
	key    string
	lock   *dynamolock.Lock
	logger *zap.Logger
	ctx    context.Context
}

func NewDynamoDBLockManagerWithUri(uri string, logger *zap.Logger) (*DynamoDBLockManager, error) {
	lockManagerLogger := logger.Named("dynamodb")

	client, err := clients.NewDynamoDBClientWithUri(uri)
	if err != nil {
		lockManagerLogger.Error(err.Error(), zap.String("uri", uri))
		return nil, err
	}

	u, err := url.Parse(uri)
	if err != nil {
		lockManagerLogger.Error(err.Error(), zap.String("uri", uri))
		return nil, err
	}
	if u.Scheme != SchemeType_name[SchemeTypeDynamoDB] {
		err := phalanxerrors.ErrInvalidUri
		lockManagerLogger.Error(err.Error(), zap.String("uri", uri))
		return nil, err
	}

	table := u.Host

	key := u.Path

	ctx := context.Background()

	lockClient, err := dynamolock.New(client,
		table,
		dynamolock.WithHeartbeatPeriod(1*time.Second),
	)

	// create table
	if _, err := lockClient.CreateTableWithContext(ctx, table); err != nil {
		var riue *types.ResourceInUseException
		if errors.As(err, &riue) {
			lockManagerLogger.Info(err.Error(), zap.String("uri", uri))
		} else {
			lockManagerLogger.Error(err.Error(), zap.String("uri", uri))
			return nil, err
		}
	}

	return &DynamoDBLockManager{
		client: lockClient,
		table:  table,
		key:    key,
		lock:   nil,
		logger: lockManagerLogger,
		ctx:    ctx,
	}, nil
}

func (m *DynamoDBLockManager) Lock() (int64, error) {
	requestTimeout := 3 * time.Second
	ctx, cancel := context.WithTimeout(m.ctx, requestTimeout)
	defer cancel()

	data := []byte("locked")
	lock, err := m.client.AcquireLockWithContext(ctx, m.key, dynamolock.WithData(data), dynamolock.ReplaceData())
	if err != nil {
		m.logger.Error(err.Error(), zap.String("key", m.key))
		return 0, phalanxerrors.ErrLockFailed
	}
	m.lock = lock

	return 0, nil
}

func (m *DynamoDBLockManager) Unlock() error {
	if m.lock == nil {
		err := phalanxerrors.ErrLockDoesNotExists
		m.logger.Error(err.Error())
		return err
	}

	requestTimeout := 3 * time.Second
	ctx, cancel := context.WithTimeout(m.ctx, requestTimeout)
	defer cancel()

	success, err := m.client.ReleaseLockWithContext(ctx, m.lock)
	if err != nil {
		m.logger.Error(err.Error())
		return err
	}
	if !success {
		err := phalanxerrors.ErrLockDoesNotExists
		m.logger.Error(err.Error())
		return err
	}

	return nil
}

func (m *DynamoDBLockManager) Close() error {
	if err := m.client.Close(); err != nil {
		m.logger.Error(err.Error())
		return err
	}

	return nil
}
