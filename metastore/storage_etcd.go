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
	requestTimeout time.Duration
	stopWatching   chan bool
	events         chan StorageEvent
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

	etcdStorage := &EtcdStorage{
		client:         client,
		kv:             clientv3.NewKV(client),
		root:           root,
		logger:         metastorelogger,
		requestTimeout: 3 * time.Second,
		stopWatching:   make(chan bool),
		events:         make(chan StorageEvent, storageEventSize),
	}

	etcdStorage.watch(context.Background())

	return etcdStorage, nil
}

func (m *EtcdStorage) watch(ctx context.Context) error {
	// Watch etcd event.
	go func() {
		watchPath := m.root + "/"
		opts := []clientv3.OpOption{
			clientv3.WithFromKey(),
		}
		watchChan := m.client.Watch(ctx, watchPath, opts...)

		for {
			select {
			case cancel := <-m.stopWatching:
				// check
				if cancel {
					return
				}
			case result := <-watchChan:
				for _, event := range result.Events {
					switch {
					case event.Type == clientv3.EventTypePut:
						m.logger.Info("put", zap.String("path", string(event.Kv.Key)))
						m.events <- StorageEvent{
							Type:  StorageEventTypePut,
							Path:  string(event.Kv.Key),
							Value: event.Kv.Value,
						}
					case event.Type == clientv3.EventTypeDelete:
						m.logger.Info("delete", zap.String("path", string(event.Kv.Key)))
						m.events <- StorageEvent{
							Type:  StorageEventTypeDelete,
							Path:  string(event.Kv.Key),
							Value: event.Kv.Value,
						}
					}
				}
			}
		}
	}()

	return nil
}

// Replace the path separator with '/'.
func (m *EtcdStorage) makePath(path string) string {
	return filepath.ToSlash(filepath.Join(filepath.ToSlash(m.root), filepath.ToSlash(path)))
}

func (m *EtcdStorage) Get(ctx context.Context, path string) ([]byte, error) {
	fullPath := m.makePath(path)

	ctx, cancel := context.WithTimeout(ctx, m.requestTimeout)
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

func (m *EtcdStorage) List(ctx context.Context, prefix string) ([]string, error) {
	prefixPath := m.makePath(prefix)

	ctx, cancel := context.WithTimeout(ctx, m.requestTimeout)
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

func (m *EtcdStorage) Put(ctx context.Context, path string, content []byte) error {
	fullPath := m.makePath(path)

	ctx, cancel := context.WithTimeout(ctx, m.requestTimeout)
	defer cancel()

	if _, err := m.kv.Put(ctx, fullPath, string(content)); err != nil {
		m.logger.Error(err.Error(), zap.String("key", fullPath))
		return err
	}

	return nil
}

func (m *EtcdStorage) Delete(ctx context.Context, path string) error {
	fullPath := m.makePath(path)

	ctx, cancel := context.WithTimeout(ctx, m.requestTimeout)
	defer cancel()

	if _, err := m.kv.Delete(ctx, fullPath); err != nil {
		m.logger.Error(err.Error(), zap.String("key", fullPath))
		return err
	}

	return nil
}

func (m *EtcdStorage) Exists(ctx context.Context, path string) (bool, error) {
	fullPath := m.makePath(path)

	ctx, cancel := context.WithTimeout(ctx, m.requestTimeout)
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

func (m *EtcdStorage) Events() <-chan StorageEvent {
	return m.events
}

func (m *EtcdStorage) Close() error {
	m.stopWatching <- true

	if err := m.client.Close(); err != nil {
		m.logger.Error(err.Error())
		return err
	}

	return nil
}
