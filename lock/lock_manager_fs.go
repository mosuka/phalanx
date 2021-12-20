package lock

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
	"github.com/mosuka/phalanx/errors"
	"go.uber.org/zap"
)

const pidFilename = "phalanx.lock"

type FileSystemLockManager struct {
	root     string
	logger   *zap.Logger
	lockFile *flock.Flock
	locked   bool
}

func NewFileSystemLockManagerWithUri(uri string, logger *zap.Logger) (*FileSystemLockManager, error) {
	lockManagerLogger := logger.Named("file_system")

	u, err := url.Parse(uri)
	if err != nil {
		lockManagerLogger.Error(err.Error(), zap.String("uri", uri))
		return nil, err
	}
	if u.Scheme != SchemeType_name[SchemeTypeFile] {
		err := errors.ErrInvalidUri
		lockManagerLogger.Error(err.Error(), zap.String("uri", uri))
		return nil, err
	}

	root := filepath.FromSlash(u.Path)

	return &FileSystemLockManager{
		root:   root,
		logger: lockManagerLogger,
	}, nil
}

func (m *FileSystemLockManager) Lock() (int64, error) {
	fullPath := filepath.Join(m.root, pidFilename)

	lockDir := filepath.Dir(fullPath)
	if err := os.MkdirAll(lockDir, 0700); err != nil {
		m.logger.Error(err.Error(), zap.String("path", lockDir))
		return 0, err
	}

	m.lockFile = flock.New(fullPath)

	requestTimeout := 3 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	var err error
	retryDelay := 500 * time.Millisecond
	m.locked, err = m.lockFile.TryLockContext(ctx, retryDelay)
	if err != nil {
		m.logger.Error(err.Error())
		return 0, err
	}
	if !m.locked {
		err := fmt.Errorf("not locked")
		m.logger.Error(err.Error())
		return 0, err
	}

	return 0, nil
}

func (m *FileSystemLockManager) Unlock() error {
	if !m.locked {
		err := errors.ErrLockDoesNotExists
		m.logger.Error(err.Error())
		return err
	}

	if err := m.lockFile.Unlock(); err != nil {
		m.logger.Error(err.Error())
		return err
	}

	return nil
}
