package lock

import (
	"net/url"

	"github.com/mosuka/phalanx/errors"
	"go.uber.org/zap"
)

type SchemeType int

const (
	SchemeTypeUnknown SchemeType = iota
	SchemeTypeEtcd
	SchemeTypeDynamoDB
)

// Enum value maps for SchemeType.
var (
	SchemeType_name = map[SchemeType]string{
		SchemeTypeUnknown:  "unknown",
		SchemeTypeEtcd:     "etcd",
		SchemeTypeDynamoDB: "dynamodb",
	}
	SchemeType_value = map[string]SchemeType{
		"unknown": SchemeTypeUnknown,
		"etcd":    SchemeTypeEtcd,
		"dynamo":  SchemeTypeDynamoDB,
	}
)

type LockManager interface {
	Lock() (int64, error)
	Unlock() error
	Close() error
}

func NewLockManagerWithUri(uri string, logger *zap.Logger) (LockManager, error) {
	lockManagerLogger := logger.Named("lock_manager")

	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case SchemeType_name[SchemeTypeEtcd]:
		return NewEtcdLockManagerWithUri(uri, logger)
	case SchemeType_name[SchemeTypeDynamoDB]:
		return NewDynamoDBLockManagerWithUri(uri, logger)
	default:
		err := errors.ErrUnsupportedLockManagerType
		lockManagerLogger.Error(err.Error(), zap.String("scheme", u.Scheme))
		return nil, err
	}
}
