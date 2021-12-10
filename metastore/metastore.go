package metastore

import (
	"net/url"
	"path/filepath"
	"strings"

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

type EventType int

const (
	EventTypeUnknown EventType = iota
	EventTypePut
	EventTypeDelete
)

// Enum value maps for EventType.
var (
	EventType_name = map[EventType]string{
		EventTypeUnknown: "unknown",
		EventTypePut:     "put",
		EventTypeDelete:  "delete",
	}
	EventType_value = map[string]EventType{
		"unknown": EventTypeUnknown,
		"put":     EventTypePut,
		"delete":  EventTypeDelete,
	}
)

type MetastoreEvent struct {
	Type  EventType
	Path  string
	Value []byte
}

type Metastore interface {
	Get(key string) ([]byte, error)
	List(prefix string) ([]string, error)
	Put(key string, value []byte) error
	Delete(key string) error
	Exists(key string) (bool, error)
	Start() error
	Stop() error
	Events() chan MetastoreEvent
}

func NewMetastoreWithUri(uri string, logger *zap.Logger) (Metastore, error) {
	metastoreLogger := logger.Named("metastore")

	u, err := url.Parse(uri)
	if err != nil {
		metastoreLogger.Error("failed to create metastore", zap.Error(err), zap.String("metastore_uri", uri))
		return nil, err
	}

	switch u.Scheme {
	case SchemeType_name[SchemeTypeFile]:
		return NewFileSystemMetastoreWithUri(uri, metastoreLogger)
	case SchemeType_name[SchemeTypeEtcd]:
		return NewEtcdMetastoreWithUri(uri, metastoreLogger)
	default:
		err := errors.ErrUnsupportedMetastoreType
		metastoreLogger.Error("unknown metastore type", zap.Error(err), zap.String("scheme", u.Scheme))
		return nil, errors.ErrUnsupportedMetastoreType
	}
}

func ExtractName(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := base[:strings.LastIndex(base, ext)]

	return name
}
