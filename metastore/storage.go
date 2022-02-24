package metastore

import (
	"context"
	"net/url"

	"github.com/mosuka/phalanx/errors"
	"go.uber.org/zap"
)

const (
	// Evvent size.
	// Cluster events can occur in large numbers at once,
	// so make sure they are large enough.
	storageEventSize = 1024
)

type SchemeType int

const (
	SchemeTypeUnknown SchemeType = iota
	SchemeTypeFile
	SchemeTypeEtcd
	SchemeTypeDynamodb
)

// Enum value maps for SchemeType.
var (
	SchemeType_name = map[SchemeType]string{
		SchemeTypeUnknown:  "unknown",
		SchemeTypeFile:     "file",
		SchemeTypeEtcd:     "etcd",
		SchemeTypeDynamodb: "dynamodb",
	}
	SchemeType_value = map[string]SchemeType{
		"unknown":  SchemeTypeUnknown,
		"file":     SchemeTypeFile,
		"etcd":     SchemeTypeEtcd,
		"dynamodb": SchemeTypeDynamodb,
	}
)

type StorageEventType int

const (
	StorageEventTypeUnknown StorageEventType = iota
	StorageEventTypePut
	StorageEventTypeDelete
)

// Enum value maps for StorageEventType.
var (
	StorageEventType_name = map[StorageEventType]string{
		StorageEventTypeUnknown: "unknown",
		StorageEventTypePut:     "put",
		StorageEventTypeDelete:  "delete",
	}
	StorageEventType_value = map[string]StorageEventType{
		"unknown": StorageEventTypeUnknown,
		"put":     StorageEventTypePut,
		"delete":  StorageEventTypeDelete,
	}
)

type StorageEvent struct {
	Type  StorageEventType
	Path  string
	Value []byte
}

type Storage interface {
	Get(ctx context.Context, key string) ([]byte, error)
	List(ctx context.Context, prefix string) ([]string, error)
	Put(ctx context.Context, key string, value []byte) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Events() <-chan StorageEvent
	Close() error
}

func NewStorageWithUri(uri string, logger *zap.Logger) (Storage, error) {
	metastoreLogger := logger.Named("storage")

	u, err := url.Parse(uri)
	if err != nil {
		metastoreLogger.Error(err.Error(), zap.String("uri", uri))
		return nil, err
	}

	switch u.Scheme {
	case SchemeType_name[SchemeTypeFile]:
		return NewFileSystemStorageWithUri(uri, metastoreLogger)
	case SchemeType_name[SchemeTypeEtcd]:
		return NewEtcdStorageWithUri(uri, metastoreLogger)
	case SchemeType_name[SchemeTypeDynamodb]:
		return NewDynamodbStorageWithUri(uri, metastoreLogger)
	default:
		err := errors.ErrUnsupportedStorageType
		metastoreLogger.Error(err.Error(), zap.String("scheme", u.Scheme))
		return nil, err
	}
}
