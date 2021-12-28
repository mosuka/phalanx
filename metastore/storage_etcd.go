package metastore

import (
	"context"
	"net/url"
	"path/filepath"
	"time"

	"github.com/mosuka/phalanx/clients"
	"github.com/mosuka/phalanx/errors"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type EtcdStorage struct {
	client         *clientv3.Client
	kv             clientv3.KV
	root           string
	logger         *zap.Logger
	ctx            context.Context
	requestTimeout time.Duration
}

func NewEtcdStorageWithUri(uri string, logger *zap.Logger) (*EtcdStorage, error) {
	metastorelogger := logger.Named("etcd")

	client, err := clients.NewEtcdClientWithUri(uri)
	if err != nil {
		metastorelogger.Error(err.Error(), zap.String("uri", uri))
		return nil, err
	}

	u, err := url.Parse(uri)
	if err != nil {
		metastorelogger.Error(err.Error(), zap.String("uri", uri))
		return nil, err
	}

	if u.Scheme != SchemeType_name[SchemeTypeEtcd] {
		err := errors.ErrInvalidUri
		metastorelogger.Error(err.Error(), zap.String("scheme", u.Scheme))
		return nil, err
	}

	root := filepath.ToSlash(filepath.Join(string(filepath.Separator), u.Host, u.Path))

	return &EtcdStorage{
		client:         client,
		kv:             clientv3.NewKV(client),
		root:           root,
		logger:         metastorelogger,
		ctx:            context.Background(),
		requestTimeout: 3 * time.Second,
	}, nil
}

// Replace the path separator with '/'.
func (m *EtcdStorage) makePath(path string) string {
	return filepath.ToSlash(filepath.Join(filepath.ToSlash(m.root), filepath.ToSlash(path)))
}

func (m *EtcdStorage) Get(path string) ([]byte, error) {
	fullPath := m.makePath(path)

	ctx, cancel := context.WithTimeout(m.ctx, m.requestTimeout)
	defer cancel()

	resp, err := m.kv.Get(ctx, fullPath)
	if err != nil {
		m.logger.Error(err.Error(), zap.String("key", fullPath))
		return nil, err
	}

	if resp.Count > 0 {
		return resp.Kvs[0].Value, nil
	} else {
		return []byte{}, nil
	}
}

func (m *EtcdStorage) List(prefix string) ([]string, error) {
	prefixPath := m.makePath(prefix)

	ctx, cancel := context.WithTimeout(m.ctx, m.requestTimeout)
	defer cancel()

	opts := []clientv3.OpOption{
		clientv3.WithPrefix(),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend),
	}

	resp, err := m.kv.Get(ctx, prefixPath, opts...)
	if err != nil {
		m.logger.Error(err.Error(), zap.String("key", prefixPath), zap.Any("opts", opts))
		return nil, err
	}

	paths := make([]string, 0)
	for _, kv := range resp.Kvs {
		// Remove prefixPath.
		// E.g. /tmp/phalanx179449480/hello.txt to /hello.txt
		path := string(kv.Key)
		paths = append(paths, path[len(prefixPath):])
	}

	return paths, nil
}

func (m *EtcdStorage) Put(path string, content []byte) error {
	fullPath := m.makePath(path)

	ctx, cancel := context.WithTimeout(m.ctx, m.requestTimeout)
	defer cancel()

	if _, err := m.kv.Put(ctx, fullPath, string(content)); err != nil {
		m.logger.Error(err.Error(), zap.String("key", fullPath))
		return err
	}

	return nil
}

func (m *EtcdStorage) Delete(path string) error {
	fullPath := m.makePath(path)

	ctx, cancel := context.WithTimeout(m.ctx, m.requestTimeout)
	defer cancel()

	if _, err := m.kv.Delete(ctx, fullPath); err != nil {
		m.logger.Error(err.Error(), zap.String("key", fullPath))
		return err
	}

	return nil
}

func (m *EtcdStorage) Exists(path string) (bool, error) {
	fullPath := m.makePath(path)

	ctx, cancel := context.WithTimeout(m.ctx, m.requestTimeout)
	defer cancel()

	resp, err := m.kv.Get(ctx, fullPath)
	if err != nil {
		m.logger.Error(err.Error(), zap.String("key", fullPath))
		return false, err
	}

	if resp.Count > 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func (m *EtcdStorage) Close() error {
	if err := m.client.Close(); err != nil {
		m.logger.Error(err.Error())
		return err
	}

	return nil
}
