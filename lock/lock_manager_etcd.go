package lock

import (
	"context"
	"net/url"
	"path/filepath"
	"time"

	"github.com/mosuka/phalanx/clients"
	"github.com/mosuka/phalanx/errors"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"go.uber.org/zap"
)

type EtcdLockManager struct {
	client *clientv3.Client
	path   string
	logger *zap.Logger
	ctx    context.Context
	mutex  *concurrency.Mutex
}

func NewEtcdLockManagerWithUri(uri string, logger *zap.Logger) (*EtcdLockManager, error) {
	lockManagerLogger := logger.Named("etcd")

	client, err := clients.NewEtcdClientWithUri(uri)
	if err != nil {
		lockManagerLogger.Error(err.Error(), zap.String("uri", uri))
		return nil, err
	}

	u, err := url.Parse(uri)
	if err != nil {
		lockManagerLogger.Error(err.Error(), zap.String("uri", uri))
		return nil, err
	}
	if u.Scheme != SchemeType_name[SchemeTypeEtcd] {
		err := errors.ErrInvalidUri
		lockManagerLogger.Error(err.Error(), zap.String("uri", uri))
		return nil, err
	}

	return &EtcdLockManager{
		client: client,
		path:   filepath.Join("/", u.Host, u.Path),
		logger: lockManagerLogger,
		ctx:    context.Background(),
		mutex:  nil,
	}, nil
}

func (m *EtcdLockManager) Lock() (int64, error) {
	if m.mutex == nil {
		var err error
		session, err := concurrency.NewSession(m.client) // without TTL
		if err != nil {
			m.logger.Error(err.Error())
			return 0, err
		}
		m.mutex = concurrency.NewMutex(session, m.path)
	}

	requestTimeout := 3 * time.Second
	ctx, cancel := context.WithTimeout(m.ctx, requestTimeout)
	defer cancel()

	if err := m.mutex.Lock(ctx); err != nil {
		m.logger.Error(err.Error())
		return 0, err
	}

	return m.mutex.Header().Revision, nil
}

func (m *EtcdLockManager) Unlock() error {
	if m.mutex == nil {
		err := errors.ErrLockDoesNotExists
		m.logger.Error(err.Error())
		return err
	}

	requestTimeout := 3 * time.Second
	ctx, cancel := context.WithTimeout(m.ctx, requestTimeout)
	defer cancel()

	if err := m.mutex.Unlock(ctx); err != nil {
		m.logger.Error(err.Error())
		return err
	}

	return nil
}
