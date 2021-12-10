package directory

import (
	"context"
	"net/url"
	"os"
	"time"

	"github.com/blugelabs/bluge"
	minio "github.com/minio/minio-go/v7"
	"github.com/mosuka/phalanx/clients"
	"github.com/mosuka/phalanx/errors"
	"github.com/mosuka/phalanx/lock"
	"github.com/mosuka/phalanx/util"
	"go.uber.org/zap"
)

type SchemeType int

const (
	SchemeTypeUnknown SchemeType = iota
	SchemeTypeMem
	SchemeTypeFile
	SchemeTypeMinio
)

// Enum value maps for SchemeType.
var (
	SchemeType_name = map[SchemeType]string{
		SchemeTypeUnknown: "unknown",
		SchemeTypeMem:     "mem",
		SchemeTypeFile:    "file",
		SchemeTypeMinio:   "minio",
	}
	SchemeType_value = map[string]SchemeType{
		"unknown": SchemeTypeUnknown,
		"mem":     SchemeTypeMem,
		"file":    SchemeTypeFile,
		"minio":   SchemeTypeMinio,
	}
)

func NewIndexConfigWithUri(uri string, lockManager lock.LockManager, logger *zap.Logger) (bluge.Config, error) {
	directoryLogger := logger.Named("directory")

	u, err := url.Parse(uri)
	if err != nil {
		return bluge.Config{}, err
	}

	switch u.Scheme {
	case SchemeType_name[SchemeTypeMem]:
		return InMemoryIndexConfig(uri, directoryLogger), nil
	case SchemeType_name[SchemeTypeFile]:
		return FileSystemIndexConfig(uri, lockManager, directoryLogger), nil
	case SchemeType_name[SchemeTypeMinio]:
		return MinioIndexConfig(uri, lockManager, directoryLogger), nil
	default:
		err := errors.ErrUnsupportedDirectoryType
		directoryLogger.Error("unknown directory type", zap.Error(err), zap.String("scheme", u.Scheme))
		return bluge.Config{}, errors.ErrUnsupportedDirectoryType
	}
}

func DirectoryExists(uri string) (bool, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return false, err
	}

	switch u.Scheme {
	case SchemeType_name[SchemeTypeMem]:
		// TODO: Remove index from memory.
		return true, nil
	case SchemeType_name[SchemeTypeFile]:
		return util.FileExists(u.Path), nil
	case SchemeType_name[SchemeTypeMinio]:
		client, err := clients.NewMinioClientWithUri(uri)
		if err != nil {
			return false, err
		}

		bucket := u.Host
		path := u.Path

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		opts := minio.ListObjectsOptions{
			Prefix:    path,
			Recursive: true,
		}

		for object := range client.ListObjects(ctx, bucket, opts) {
			if object.Err != nil {
				return false, object.Err
			}

			// If at least one object is found,
			// it means that the directory exists and returns true.
			if object.Key != "" {
				return true, nil
			}
		}

		// If the object is not found, the directory does not exist and false is returned.
		return false, nil
	default:
		err := errors.ErrUnsupportedDirectoryType
		return false, err
	}
}

func DeleteDirectory(uri string) error {
	u, err := url.Parse(uri)
	if err != nil {
		return err
	}

	switch u.Scheme {
	case SchemeType_name[SchemeTypeMem]:
		// TODO: Remove index from memory.
		return nil
	case SchemeType_name[SchemeTypeFile]:
		return os.RemoveAll(u.Path)
	case SchemeType_name[SchemeTypeMinio]:
		client, err := clients.NewMinioClientWithUri(uri)
		if err != nil {
			return err
		}

		bucket := u.Host
		path := u.Path

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		objectsChan := make(chan minio.ObjectInfo)

		// Send object info that are needed to be removed to objectsCh
		go func() {
			defer close(objectsChan)

			opts := minio.ListObjectsOptions{
				Prefix:    path,
				Recursive: true,
			}
			for object := range client.ListObjects(ctx, bucket, opts) {
				objectsChan <- object
			}
		}()

		opts := minio.RemoveObjectsOptions{
			GovernanceBypass: true,
		}

		for removeObjErr := range client.RemoveObjects(ctx, bucket, objectsChan, opts) {
			if removeObjErr.Err != nil {
				return removeObjErr.Err
			}
		}

		return nil
	default:
		return errors.ErrUnsupportedDirectoryType
	}
}
