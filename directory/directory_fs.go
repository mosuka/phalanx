package directory

import (
	"net/url"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/index"
	"github.com/mosuka/phalanx/errors"
	"github.com/mosuka/phalanx/lock"
	"go.uber.org/zap"
)

func FileSystemIndexConfig(uri string, lockManager lock.LockManager, logger *zap.Logger) bluge.Config {
	return bluge.DefaultConfigWithDirectory(func() index.Directory {
		return NewFileSystemDirectoryWithUri(uri, lockManager, logger)
	})
}

type FileSystemDirectory struct {
	*index.FileSystemDirectory
	path        string
	lockManager lock.LockManager
	logger      *zap.Logger
}

func NewFileSystemDirectoryWithUri(uri string, lockManager lock.LockManager, logger *zap.Logger) *FileSystemDirectory {
	fileSystemLogger := logger.Named("file_system")

	// Parse URI.
	u, err := url.Parse(uri)
	if err != nil {
		fileSystemLogger.Error(err.Error(), zap.String("uri", uri))
		return nil
	}

	if u.Scheme != SchemeType_name[SchemeTypeFile] {
		err := errors.ErrInvalidUri
		fileSystemLogger.Error(err.Error(), zap.String("uri", uri))
		return nil
	}

	path := u.Path

	parent := index.NewFileSystemDirectory(path)

	return &FileSystemDirectory{
		FileSystemDirectory: parent,
		path:                path,
		lockManager:         lockManager,
		logger:              logger,
	}
}

func (d *FileSystemDirectory) Lock() error {
	// 1. create lockmanager
	// 2. lock

	if _, err := d.lockManager.Lock(); err != nil {
		d.logger.Error(err.Error(), zap.String("path", d.path))
		return err
	}

	return nil
}

func (d *FileSystemDirectory) Unlock() error {
	// 1. unlock
	// 2. remove lockmanager
	if err := d.lockManager.Unlock(); err != nil {
		d.logger.Error(err.Error(), zap.String("path", d.path))
		return err
	}

	return nil
}
