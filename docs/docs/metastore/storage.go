package metastore

import (
	"net/url"

	"github.com/mosuka/phalanx/errors"
	"go.uber.org/zap"
)

type SchemeType int

const (
	SchemeTypeUnknown SchemeType = iota
	SchemeTypeFile
	SchemeTypeEtcd
)

// Enum value maps for SchemeType.
var (
	SchemeType_name = map[SchemeType]string{
		SchemeTypeUnknown: "unknown",
		SchemeTypeFile:    "file",
		SchemeTypeEtcd:    "etcd",
	}
	SchemeType_value = map[string]SchemeType{
		"unknown": SchemeTypeUnknown,
		"file":    SchemeTypeFile,
		"etcd":    SchemeTypeEtcd,
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
	Get(key string) ([]byte, error)
	List(prefix string) ([]string, error)
	Put(key string, value []byte) error
	Delete(key string) error
	Exists(key string) (bool, error)
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
	default:
		err := errors.ErrUnsupportedStorageType
		metastoreLogger.Error(err.Error(), zap.String("scheme", u.Scheme))
		return nil, err
	}
}
