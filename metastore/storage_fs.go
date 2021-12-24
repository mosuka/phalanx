package metastore

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/mosuka/phalanx/errors"
	"github.com/mosuka/phalanx/util"
	"go.uber.org/zap"
)

type FileSystemStorage struct {
	path        string
	logger      *zap.Logger
	fsWatcher   *fsnotify.Watcher
	stopWatcher chan bool
	events      chan StorageEvent
	watchSet    map[string]bool
	mutex       sync.RWMutex
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

	watchSet := make(map[string]bool)

	// Add the root path of the metastore to the watch list.
	watchSet[path] = true
	if err := fsWatcher.Add(path); err != nil {
		return nil, err
	}

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

				storageEvent := &StorageEvent{
					Type:  StorageEventTypeUnknown,
					Path:  event.Name,
					Value: []byte{},
				}

				switch {
				case event.Op&fsnotify.Create == fsnotify.Create:
					storageEvent.Type = StorageEventTypePut
				case event.Op&fsnotify.Write == fsnotify.Write:
					storageEvent.Type = StorageEventTypePut
				case event.Op&fsnotify.Remove == fsnotify.Remove:
					storageEvent.Type = StorageEventTypeDelete
				case event.Op&fsnotify.Rename == fsnotify.Rename:
					// TODO
					continue
				case event.Op&fsnotify.Chmod == fsnotify.Chmod:
					// ignore
					continue
				default:
					err := errors.ErrUnsupportedMetastoreEvent
					logger.Warn(err.Error())
					continue
				}

				events <- *storageEvent
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
		watchSet:    watchSet,
	}

	return fsStorage, nil
}

func (m *FileSystemStorage) Get(path string) ([]byte, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

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
	m.mutex.RLock()
	defer m.mutex.RUnlock()

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
	m.mutex.Lock()
	defer m.mutex.Unlock()

	fullPath := filepath.Join(m.path, path)

	// Create directory.
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		m.logger.Error(err.Error(), zap.String("path", dir))
		return err
	}

	// Write file.
	if err := ioutil.WriteFile(fullPath, content, 0600); err != nil {
		m.logger.Error(err.Error(), zap.String("path", fullPath))
		return err
	}

	// Add the created path to the watch list.
	// Do not add m.path to watchSet twice.
	watchDir := filepath.Dir(fullPath)
	if _, ok := m.watchSet[watchDir]; !ok && watchDir != m.path {
		m.logger.Info("add to watch list", zap.String("path", watchDir))
		if err := m.fsWatcher.Add(watchDir); err != nil {
			m.logger.Warn(err.Error())
			return err
		}
		m.watchSet[watchDir] = true
	}

	return nil
}

func (m *FileSystemStorage) Delete(path string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	fullPath := filepath.Join(m.path, path)

	// Remove file.
	if err := os.Remove(fullPath); err != nil {
		m.logger.Error(err.Error(), zap.String("path", fullPath))
		return err
	}

	// Remove the removed path from the watch list.
	// Do not remove m.path from watchSet.
	watchDir := filepath.Dir(fullPath)
	if _, ok := m.watchSet[watchDir]; ok && watchDir != m.path {
		m.logger.Info("remove to watch list", zap.String("path", watchDir))
		if err := m.fsWatcher.Remove(watchDir); err != nil {
			m.logger.Warn(err.Error())
			return err
		}
		delete(m.watchSet, watchDir)
	}

	return nil
}

func (m *FileSystemStorage) Exists(path string) (bool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

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
