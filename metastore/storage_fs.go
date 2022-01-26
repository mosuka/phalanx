package metastore

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gofrs/flock"
	"github.com/mosuka/phalanx/errors"
	"github.com/mosuka/phalanx/util"
	"go.uber.org/zap"
)

const (
	lockFileSuffix = ".lock"
)

type FileSystemStorage struct {
	path         string
	logger       *zap.Logger
	mutex        sync.RWMutex
	stopWatching chan bool
	watchFileSet map[string][]byte
	events       chan StorageEvent
	ticker       *time.Ticker
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

	watchSet := make(map[string]bool)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	if err := watcher.Add(path); err != nil {
		fileLogger.Error(err.Error())
		return nil, err
	}
	watchSet[path] = true

	fsStorage := &FileSystemStorage{
		path:         path,
		logger:       fileLogger,
		stopWatching: make(chan bool),
		watchFileSet: make(map[string][]byte),
		events:       make(chan StorageEvent, storageEventSize),
		ticker:       time.NewTicker(time.Millisecond * 200),
	}

	fsStorage.watch()

	return fsStorage, nil
}

func (m *FileSystemStorage) watch() error {
	// Watch file system event.
	go func() {
		for {
			select {
			case cancel := <-m.stopWatching:
				// check
				if cancel {
					return
				}
			case <-m.ticker.C:
				files, err := m.listFiles()
				if err != nil {
					m.logger.Warn(err.Error())
					continue
				}
				sort.Strings(files)

				for _, file := range files {
					data, err := ioutil.ReadFile(file)
					if err != nil {
						m.logger.Warn(err.Error(), zap.String("path", file))
						continue
					}

					if _, ok := m.watchFileSet[file]; !ok {
						m.logger.Info("file added", zap.String("path", file))
						m.watchFileSet[file] = data

						m.events <- StorageEvent{
							Type:  StorageEventTypePut,
							Path:  file,
							Value: data,
						}
						m.logger.Info("sent storage event", zap.String("type", "put"), zap.String("path", file))
					} else {
						if !bytes.Equal(m.watchFileSet[file], data) {
							m.logger.Info("file changed", zap.String("path", file))
							m.watchFileSet[file] = data

							m.events <- StorageEvent{
								Type:  StorageEventTypePut,
								Path:  file,
								Value: data,
							}
							m.logger.Info("sent storage event", zap.String("type", "put"), zap.String("path", file))
						}
					}
				}

				for file := range m.watchFileSet {
					exists := false
					for _, file2 := range files {
						if file == file2 {
							exists = true
							break
						}
					}
					if !exists {
						m.logger.Info("file removed", zap.String("path", file))
						delete(m.watchFileSet, file)

						m.events <- StorageEvent{
							Type: StorageEventTypeDelete,
							Path: file,
						}
						m.logger.Info("sent storage event", zap.String("type", "delete"), zap.String("path", file))
					}
				}
			}
		}
	}()

	return nil
}

func (m *FileSystemStorage) listFiles() ([]string, error) {
	var files []string
	err := filepath.Walk(m.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && !strings.HasSuffix(path, lockFileSuffix) {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// Replace the path separator with '/'.
func (m *FileSystemStorage) makePath(path string) string {
	return filepath.FromSlash(filepath.Join(filepath.FromSlash(m.path), filepath.FromSlash(path)))
}

func (m *FileSystemStorage) Get(path string) ([]byte, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	fullPath := m.makePath(path)

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

	prefixPath := m.makePath(prefix)

	paths := make([]string, 0)
	err := filepath.Walk(prefixPath, func(path string, Debug os.FileInfo, err error) error {
		if path != prefixPath {
			// exclude .lock files
			if !strings.HasSuffix(path, lockFileSuffix) {
				// Remove prefixPath.
				// E.g. /tmp/phalanx179449480/hello.txt to /hello.txt
				paths = append(paths, path[len(prefixPath):])
			}
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

	if strings.HasSuffix(path, lockFileSuffix) {
		return fmt.Errorf("cannot put lock file directory: %s", path)
	}

	fullPath := m.makePath(path)

	// Create directory.
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		m.logger.Error(err.Error(), zap.String("path", dir))
		return err
	}

	lock := flock.New(fmt.Sprintf("%s%s", fullPath, lockFileSuffix))
	defer lock.Unlock()

	if err := lock.Lock(); err != nil {
		m.logger.Error(err.Error(), zap.String("path", fullPath))
		return err
	}

	// Write file.
	m.logger.Info("write file", zap.String("path", fullPath))
	if err := ioutil.WriteFile(fullPath, content, 0600); err != nil {
		m.logger.Error(err.Error(), zap.String("path", fullPath))
		return err
	}

	return nil
}

func (m *FileSystemStorage) Delete(path string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	fullPath := m.makePath(path)

	lock := flock.New(fmt.Sprintf("%s%s", fullPath, lockFileSuffix))
	defer lock.Unlock()

	if err := lock.Lock(); err != nil {
		m.logger.Error(err.Error(), zap.String("path", fullPath))
		return err
	}

	// Remove file.
	if err := os.Remove(fullPath); err != nil {
		m.logger.Error(err.Error(), zap.String("path", fullPath))
		return err
	}

	return nil
}

func (m *FileSystemStorage) Exists(path string) (bool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	fullPath := m.makePath(path)

	return util.FileExists(fullPath), nil
}

func (m *FileSystemStorage) Events() <-chan StorageEvent {
	return m.events
}

func (m *FileSystemStorage) Close() error {
	m.ticker.Stop()

	m.stopWatching <- true

	return nil
}
