package lock

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

type LockManager interface {
	Lock() (int64, error)
	Unlock() error
}

func NewLockManagerWithUri(uri string, logger *zap.Logger) (LockManager, error) {
	lockManagerLogger := logger.Named("lock_manager")

	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case SchemeType_name[SchemeTypeFile]:
		return NewFileSystemLockManagerWithUri(uri, logger)
	case SchemeType_name[SchemeTypeEtcd]:
		return NewEtcdLockManagerWithUri(uri, logger)
	default:
		err := errors.ErrUnsupportedLockManagerType
		lockManagerLogger.Error("unknown lock manager type", zap.Error(err), zap.String("scheme", u.Scheme))
		return nil, err
	}
}
