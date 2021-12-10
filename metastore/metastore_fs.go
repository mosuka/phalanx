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

type FileSystemMetastore struct {
	path         string
	logger       *zap.Logger
	fsWatcher    *fsnotify.Watcher
	stopWatching chan bool
	events       chan MetastoreEvent
}

func NewFileSystemMetastoreWithUri(uri string, logger *zap.Logger) (*FileSystemMetastore, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	if u.Scheme != SchemeType_name[SchemeTypeFile] {
		return nil, errors.ErrInvalidUri
	}

	path := u.Path

	return NewFileSystemMetastoreWithPath(path, logger)
}

func NewFileSystemMetastoreWithPath(path string, logger *zap.Logger) (*FileSystemMetastore, error) {
	fileLogger := logger.Named("file_system")

	if !util.FileExists(path) {
		if err := os.MkdirAll(path, 0700); err != nil {
			fileLogger.Error("failed to create directory", zap.Error(err), zap.String("path", path))
			return nil, err
		}
	}

	return &FileSystemMetastore{
		path:         path,
		logger:       fileLogger,
		fsWatcher:    nil,
		stopWatching: make(chan bool),
		events:       make(chan MetastoreEvent, 10),
	}, nil
}

func (m *FileSystemMetastore) convertMetastoreEvent(event *fsnotify.Event) (*MetastoreEvent, error) {
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
		return &MetastoreEvent{
			Type:  EventTypePut,
			Path:  event.Name,
			Value: value,
		}, nil
	case event.Op&fsnotify.Write == fsnotify.Write:
		return &MetastoreEvent{
			Type:  EventTypePut,
			Path:  event.Name,
			Value: value,
		}, nil
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		return &MetastoreEvent{
			Type:  EventTypeDelete,
			Path:  event.Name,
			Value: value,
		}, nil
	case event.Op&fsnotify.Rename == fsnotify.Rename:
		return &MetastoreEvent{
			Type:  EventTypePut,
			Path:  event.Name,
			Value: value,
		}, nil
	case event.Op&fsnotify.Chmod == fsnotify.Chmod:
		return &MetastoreEvent{
			Type:  EventTypePut,
			Path:  event.Name,
			Value: value,
		}, nil
	default:
		return nil, errors.ErrUnsupportedMetastoreEvent
	}
}

func (m *FileSystemMetastore) Get(path string) ([]byte, error) {
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

func (m *FileSystemMetastore) List(prefix string) ([]string, error) {
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

func (m *FileSystemMetastore) Put(path string, content []byte) error {
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

func (m *FileSystemMetastore) Delete(path string) error {
	fullPath := filepath.Join(m.path, path)
	if err := os.Remove(fullPath); err != nil {
		m.logger.Error("failed to remove file", zap.Error(err), zap.String("path", fullPath))
		return err
	}

	return nil
}

func (m *FileSystemMetastore) Exists(path string) (bool, error) {
	fullPath := filepath.Join(m.path, path)

	return util.FileExists(fullPath), nil
}

func (m *FileSystemMetastore) Start() error {
	if m.fsWatcher != nil {
		return nil
	}

	if watcher, err := fsnotify.NewWatcher(); err != nil {
		m.logger.Error("failed to create file system watcher", zap.Error(err))
		return err
	} else {
		m.fsWatcher = watcher
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

				metastoreEvent, err := m.convertMetastoreEvent(&event)
				if err != nil {
					m.logger.Error("failed to convert event", zap.Error(err), zap.Any("event", event))
					continue
				}

				m.logger.Info("received file system event", zap.String("path", metastoreEvent.Path), zap.String("type", EventType_name[metastoreEvent.Type]))

				m.events <- *metastoreEvent
			case err, ok := <-m.fsWatcher.Errors:
				m.logger.Error("received file system error event", zap.Error(err))

				if !ok {
					m.logger.Error("failed to receive error", zap.String("path", m.path))
					continue
				}

				m.logger.Error("receive error", zap.Error(err), zap.String("path", m.path))
			}
		}
	}()

	if err := m.fsWatcher.Add(m.path); err != nil {
		m.logger.Error("failed to start watching the file or directory", zap.Error(err), zap.String("path", m.path))
		return err
	}

	return nil
}

func (m *FileSystemMetastore) Stop() error {
	m.stopWatching <- true

	if err := m.fsWatcher.Close(); err != nil {
		m.logger.Error("failed to close file system watcher", zap.Error(err))
		return err
	}

	return nil
}

func (m *FileSystemMetastore) Events() chan MetastoreEvent {
	return m.events
}
