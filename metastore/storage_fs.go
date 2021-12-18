package metastore

import (
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/mosuka/phalanx/errors"
	"github.com/mosuka/phalanx/util"
	"go.uber.org/zap"
)

type FileSystemStorage struct {
	path         string
	logger       *zap.Logger
	fsWatcher    *fsnotify.Watcher
	stopWatching chan bool
	events       chan StorageEvent
}

func NewFileSystemStorageWithUri(uri string, logger *zap.Logger) (*FileSystemStorage, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	if u.Scheme != SchemeType_name[SchemeTypeFile] {
		return nil, errors.ErrInvalidUri
	}

	path := u.Path

	return NewFileSystemStorageWithPath(path, logger)
}

func NewFileSystemStorageWithPath(path string, logger *zap.Logger) (*FileSystemStorage, error) {
	fileLogger := logger.Named("file_system")

	if !util.FileExists(path) {
		if err := os.MkdirAll(path, 0700); err != nil {
			fileLogger.Error("failed to create directory", zap.Error(err), zap.String("path", path))
			return nil, err
		}
	}

	return &FileSystemStorage{
		path:         path,
		logger:       fileLogger,
		fsWatcher:    nil,
		stopWatching: make(chan bool),
		events:       make(chan StorageEvent, 10),
	}, nil
}

func (m *FileSystemStorage) makeStorageEvent(event *fsnotify.Event) (*StorageEvent, error) {
	// Load metadata.
	var value []byte
	var err error
	if util.FileExists(event.Name) {
		value, err = ioutil.ReadFile(event.Name)
		if err != nil {
			m.logger.Error("failed to read file", zap.Error(err), zap.String("path", event.Name))
		}
	}

	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		return &StorageEvent{
			Type:  StorageEventTypePut,
			Path:  event.Name,
			Value: value,
		}, nil
	case event.Op&fsnotify.Write == fsnotify.Write:
		return &StorageEvent{
			Type:  StorageEventTypePut,
			Path:  event.Name,
			Value: value,
		}, nil
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		return &StorageEvent{
			Type:  StorageEventTypeDelete,
			Path:  event.Name,
			Value: value,
		}, nil
	case event.Op&fsnotify.Rename == fsnotify.Rename:
		return &StorageEvent{
			Type:  StorageEventTypePut,
			Path:  event.Name,
			Value: value,
		}, nil
	case event.Op&fsnotify.Chmod == fsnotify.Chmod:
		return &StorageEvent{
			Type:  StorageEventTypePut,
			Path:  event.Name,
			Value: value,
		}, nil
	default:
		return nil, errors.ErrUnsupportedMetastoreEvent
	}
}

func (m *FileSystemStorage) Get(path string) ([]byte, error) {
	fullPath := filepath.Join(m.path, path)
	if !util.FileExists(fullPath) {
		return nil, errors.ErrIndexMetadataDoesNotExist
	}

	content, err := ioutil.ReadFile(fullPath)
	if err != nil {
		m.logger.Error("failed to read file", zap.Error(err), zap.String("path", fullPath))
		return nil, err
	}

	return content, nil
}

func (m *FileSystemStorage) List(prefix string) ([]string, error) {
	prefixPath := filepath.Join(m.path, prefix)
	paths := make([]string, 0)
	err := filepath.Walk(prefixPath, func(path string, info os.FileInfo, err error) error {
		if path != prefixPath {
			// Remove prefixPath.
			// E.g. /tmp/phalanx179449480/hello.txt to /hello.txt
			paths = append(paths, path[len(prefixPath):])
		}

		return nil
	})
	if err != nil {
		m.logger.Error("failed to list files", zap.Error(err), zap.String("prefix", prefixPath))
		return nil, err
	}

	return paths, nil
}

func (m *FileSystemStorage) Put(path string, content []byte) error {
	fullPath := filepath.Join(m.path, path)

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		m.logger.Error("failed to create directory", zap.Error(err), zap.String("path", dir))
		return err
	}

	if err := m.fsWatcher.Add(dir); err != nil {
		m.logger.Error("failed to start watching the file or directory", zap.Error(err), zap.String("path", m.path))
		return err
	}

	if err := ioutil.WriteFile(fullPath, content, 0600); err != nil {
		m.logger.Error("failed to write file", zap.Error(err), zap.String("path", fullPath))
		return err
	}

	return nil
}

func (m *FileSystemStorage) Delete(path string) error {
	fullPath := filepath.Join(m.path, path)
	if err := os.Remove(fullPath); err != nil {
		m.logger.Error("failed to remove file", zap.Error(err), zap.String("path", fullPath))
		return err
	}

	return nil
}

func (m *FileSystemStorage) Exists(path string) (bool, error) {
	fullPath := filepath.Join(m.path, path)

	return util.FileExists(fullPath), nil
}

func (m *FileSystemStorage) Start() error {
	if m.fsWatcher != nil {
		return nil
	}

	if watcher, err := fsnotify.NewWatcher(); err != nil {
		m.logger.Error("failed to create file system watcher", zap.Error(err))
		return err
	} else {
		m.fsWatcher = watcher
	}

	if err := m.fsWatcher.Add(m.path); err != nil {
		m.logger.Error("failed to start watching the file or directory", zap.Error(err), zap.String("path", m.path))
		return err
	}

	go func() {
		for {
			select {
			case cancel := <-m.stopWatching:
				// check
				if cancel {
					return
				}
			case event, ok := <-m.fsWatcher.Events:
				if !ok {
					m.logger.Error("failed to receive event", zap.String("path", m.path))
					continue
				}
				// m.logger.Info("received file system event", zap.String("operation", event.Op.String()), zap.String("name", event.Name))
				metastoreEvent, err := m.makeStorageEvent(&event)
				if err != nil {
					m.logger.Error("failed to convert event", zap.Error(err), zap.Any("event", event))
					continue
				}
				// m.logger.Info("received metastore storage event", zap.String("path", metastoreEvent.Path), zap.String("type", StorageEventType_name[metastoreEvent.Type]))
				m.events <- *metastoreEvent
			case err, ok := <-m.fsWatcher.Errors:
				if !ok {
					m.logger.Error("failed to receive error", zap.String("path", m.path))
					continue
				}
				m.logger.Error("received file system error event", zap.Error(err), zap.String("path", m.path))
			}
		}
	}()

	return nil
}

func (m *FileSystemStorage) Stop() error {
	m.stopWatching <- true

	if err := m.fsWatcher.Close(); err != nil {
		m.logger.Error("failed to close file system watcher", zap.Error(err))
		return err
	}

	return nil
}

func (m *FileSystemStorage) Events() chan StorageEvent {
	return m.events
}
