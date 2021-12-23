package metastore

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/mosuka/phalanx/errors"
	"github.com/mosuka/phalanx/util"
	"go.uber.org/zap"
)

func makeFileSystemStorageEvent(event *fsnotify.Event, logger *zap.Logger) (*StorageEvent, error) {
	// Load metadata.
	var value []byte
	var err error
	if util.FileExists(event.Name) {
		value, err = ioutil.ReadFile(event.Name)
		if err != nil {
			logger.Warn(err.Error())
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
			Value: []byte{},
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

type FileSystemStorage struct {
	path        string
	logger      *zap.Logger
	fsWatcher   *fsnotify.Watcher
	stopWatcher chan bool
	events      chan StorageEvent
	watchSet    map[string]bool
}

func NewFileSystemStorageWithUri(uri string, logger *zap.Logger) (*FileSystemStorage, error) {
	u, err := url.Parse(uri)
	if err != nil {
		logger.Error(err.Error(), zap.String("uri", uri))
		return nil, err
	}

	if u.Scheme != SchemeType_name[SchemeTypeFile] {
		err := errors.ErrInvalidUri
		logger.Error(err.Error(), zap.String("uri", uri))
		return nil, err
	}

	path := u.Path

	return NewFileSystemStorageWithPath(path, logger)
}

func NewFileSystemStorageWithPath(path string, logger *zap.Logger) (*FileSystemStorage, error) {
	fileLogger := logger.Named("file_system")

	if !util.FileExists(path) {
		if err := os.MkdirAll(path, 0700); err != nil {
			fileLogger.Error(err.Error(), zap.String("path", path))
			return nil, err
		}
	}

	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	stopWatcher := make(chan bool)
	events := make(chan StorageEvent, 10)

	// if err := addWatchList(fsWatcher, path, fileLogger); err != nil {
	// 	fileLogger.Error(err.Error(), zap.String("path", path))
	// 	return nil, err
	// }

	// Start file system watcher
	go func(fsWatcher *fsnotify.Watcher, stopWatcher chan bool, event chan StorageEvent, logger *zap.Logger) {
		for {
			select {
			case cancel := <-stopWatcher:
				// check
				if cancel {
					return
				}
			case event, ok := <-fsWatcher.Events:
				if !ok {
					err := fmt.Errorf("failed to receive event")
					logger.Warn(err.Error())
					continue
				}

				logger.Info("received file system event", zap.Any("event", event))
				// if event.Op == fsnotify.Create || event.Op == fsnotify.Chmod || event.Op == fsnotify.Rename {
				// 	logger.Debug("event ignored", zap.Any("event", event))
				// 	continue
				// }

				metastoreEvent, err := makeFileSystemStorageEvent(&event, logger)
				if err != nil {
					logger.Warn(err.Error(), zap.Any("event", event))
					continue
				}

				events <- *metastoreEvent
			case err, ok := <-fsWatcher.Errors:
				if !ok {
					err := fmt.Errorf("failed to receive error")
					logger.Warn(err.Error())
					continue
				}
				logger.Warn(err.Error())
			}
		}
	}(fsWatcher, stopWatcher, events, fileLogger)

	fsStorage := &FileSystemStorage{
		path:        path,
		logger:      fileLogger,
		fsWatcher:   fsWatcher,
		stopWatcher: stopWatcher,
		events:      events,
		watchSet:    make(map[string]bool),
	}

	fsStorage.addWatchList(path)

	return fsStorage, nil
}

func (m *FileSystemStorage) Get(path string) ([]byte, error) {
	fullPath := filepath.Join(m.path, path)
	if !util.FileExists(fullPath) {
		err := errors.ErrIndexMetadataDoesNotExist
		m.logger.Error(err.Error(), zap.String("path", fullPath))
		return nil, err
	}

	content, err := ioutil.ReadFile(fullPath)
	if err != nil {
		m.logger.Error(err.Error(), zap.String("path", fullPath))
		return nil, err
	}

	return content, nil
}

func (m *FileSystemStorage) List(prefix string) ([]string, error) {
	prefixPath := filepath.Join(m.path, prefix)
	paths := make([]string, 0)
	err := filepath.Walk(prefixPath, func(path string, Debug os.FileInfo, err error) error {
		if path != prefixPath {
			// Remove prefixPath.
			// E.g. /tmp/phalanx179449480/hello.txt to /hello.txt
			paths = append(paths, path[len(prefixPath):])
		}

		return nil
	})
	if err != nil {
		m.logger.Error(err.Error(), zap.String("prefix", prefixPath))
		return nil, err
	}

	return paths, nil
}

func (m *FileSystemStorage) Put(path string, content []byte) error {
	fullPath := filepath.Join(m.path, path)

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		m.logger.Error(err.Error(), zap.String("path", dir))
		return err
	}

	if err := m.addWatchList(dir); err != nil {
		m.logger.Error(err.Error(), zap.String("path", dir))
		return err
	}

	if err := ioutil.WriteFile(fullPath, content, 0600); err != nil {
		m.logger.Error(err.Error(), zap.String("path", fullPath))
		return err
	}

	if err := m.addWatchList(fullPath); err != nil {
		m.logger.Error(err.Error(), zap.String("path", fullPath))
		return err
	}

	return nil
}

func (m *FileSystemStorage) Delete(path string) error {
	fullPath := filepath.Join(m.path, path)
	m.logger.Info("delete file", zap.String("path", fullPath))

	if err := m.removeWatchList(fullPath); err != nil {
		m.logger.Error(err.Error(), zap.String("path", fullPath))
		return err
	}

	if err := os.Remove(fullPath); err != nil {
		m.logger.Error(err.Error(), zap.String("path", fullPath))
		return err
	}

	return nil
}

func (m *FileSystemStorage) Exists(path string) (bool, error) {
	fullPath := filepath.Join(m.path, path)

	return util.FileExists(fullPath), nil
}

func (m *FileSystemStorage) Close() error {
	m.stopWatcher <- true

	if err := m.fsWatcher.Close(); err != nil {
		m.logger.Error(err.Error())
		return err
	}

	return nil
}

func (m *FileSystemStorage) Events() chan StorageEvent {
	return m.events
}

func (m *FileSystemStorage) addWatchList(watchPath string) error {
	if err := filepath.Walk(watchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		m.logger.Info("add to watch list", zap.String("path", path))
		return m.fsWatcher.Add(path)
	}); err != nil {
		return err
	}

	return nil
}

func (m *FileSystemStorage) removeWatchList(watchPath string) error {
	if err := filepath.Walk(watchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		m.logger.Info("remove from watch list", zap.String("path", path))
		return m.fsWatcher.Remove(path)
	}); err != nil {
		return err
	}

	return nil
}
