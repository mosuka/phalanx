package metastore

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mosuka/phalanx/errors"
	"github.com/mosuka/phalanx/mapping"
	"github.com/mosuka/rendezvous"
	cmap "github.com/orcaman/concurrent-map"
	"go.uber.org/zap"
)

const (
	shardNamePrefix = "shard-"

	// Evvent size.
	// Cluster events can occur in large numbers at once,
	// so make sure they are large enough.
	metastoreEventSize = 1024
)

type MetastoreEventType int

const (
	MetastoreEventTypeUnknown MetastoreEventType = iota
	MetastoreEventTypePutIndex
	MetastoreEventTypeDeleteIndex
	MetastoreEventTypePutShard
	MetastoreEventTypeDeleteShard
)

// Event value maps for MetastoreEventType.
var (
	MetastoreEventType_name = map[MetastoreEventType]string{
		MetastoreEventTypeUnknown:     "unknown",
		MetastoreEventTypePutIndex:    "put_index",
		MetastoreEventTypeDeleteIndex: "delete_index",
		MetastoreEventTypePutShard:    "put_shard",
		MetastoreEventTypeDeleteShard: "delete_shard",
	}
	MetastoreEventType_value = map[string]MetastoreEventType{
		"unknown":      MetastoreEventTypeUnknown,
		"put_index":    MetastoreEventTypePutIndex,
		"delete_index": MetastoreEventTypeDeleteIndex,
		"put_shard":    MetastoreEventTypePutShard,
		"delete_shard": MetastoreEventTypeDeleteShard,
	}
)

type MetastoreEvent struct {
	Type  MetastoreEventType
	Index string
	Shard string
}

// Make index metadata path
// e.g. wikipedia_en -> wikipedia_en/index.json
func makeIndexMetadataPath(indexName string) string {
	return filepath.Join(indexName, "index.json")
}

// Make shard metadata path
// e.g. wikipedia_en -> wikipedia_en/shardpb0g8d8hmvcg9hvaiol3.json
func makeShardMetadataPath(indexName string, shardName string) string {
	return filepath.Join(indexName, fmt.Sprintf("%s.json", shardName))
}

type Metastore struct {
	storage          Storage
	indexMetadataMap cmap.ConcurrentMap
	ringMap          map[string]*rendezvous.Ring
	events           chan MetastoreEvent
	stopWatching     chan bool
	logger           *zap.Logger
	mutex            sync.RWMutex
	ctx              context.Context
}

func NewMetastoreWithUri(uri string, logger *zap.Logger) (*Metastore, error) {
	storage, err := NewStorageWithUri(uri, logger)
	if err != nil {
		logger.Error(err.Error(), zap.String("uri", uri))
		return nil, err
	}

	return NewMetastore(storage, logger)
}

func NewMetastore(storage Storage, logger *zap.Logger) (*Metastore, error) {
	ctx := context.Background()

	paths, err := storage.List(ctx, string(filepath.Separator))
	if err != nil {
		logger.Error(err.Error(), zap.String("prefix", string(filepath.Separator)))
		return nil, err
	}

	indexMetadataMap := cmap.New()
	ringMap := make(map[string]*rendezvous.Ring)

	for _, path := range paths {
		fileName := filepath.Base(path)
		if fileName == "index.json" {
			value, err := storage.Get(ctx, path)
			if err != nil {
				logger.Error(err.Error(), zap.String("path", path))
				return nil, err
			}

			indexMetadata, err := NewIndexMetadataWithBytes(value)
			if err != nil {
				logger.Error(err.Error())
				return nil, err
			}

			indexName := filepath.Base(filepath.Dir(path))

			indexMetadataMap.Set(indexName, indexMetadata)
			ringMap[indexName] = rendezvous.New()
		}
	}

	for _, path := range paths {
		fileName := filepath.Base(path)
		if strings.HasPrefix(fileName, shardNamePrefix) && strings.HasSuffix(fileName, ".json") {
			value, err := storage.Get(ctx, path)
			if err != nil {
				logger.Error(err.Error(), zap.String("path", path))
				return nil, err
			}

			shardMetadata, err := NewShardMetadataWithBytes(value)
			if err != nil {
				logger.Error(err.Error())
				return nil, err
			}

			indexName := filepath.Base(filepath.Dir(path))
			shardName := strings.TrimSuffix(filepath.Base(path), ".json")

			// Update shard metadata
			if tmpIndexMetadata, ok := indexMetadataMap.Get(indexName); ok {
				if indexMetadata, ok := tmpIndexMetadata.(*IndexMetadata); ok {
					indexMetadata.SetShardMetadata(shardName, shardMetadata)
				} else {
					logger.Warn("type unexpected", zap.String("index_name", indexName))
				}
			} else {
				logger.Warn("index metadata do not found", zap.String("index_name", indexName))
			}

			// Add new hash ring item for shard.
			if hashRing, ok := ringMap[indexName]; ok {
				hashRing.AddWithWeight(shardName, 1.0)
			} else {
				logger.Warn("hash ring do not found", zap.String("index_name", indexName))
			}
		}
	}

	metastore := &Metastore{
		storage:          storage,
		indexMetadataMap: indexMetadataMap,
		ringMap:          ringMap,
		stopWatching:     make(chan bool),
		events:           make(chan MetastoreEvent, metastoreEventSize),
		logger:           logger,
		mutex:            sync.RWMutex{},
		ctx:              ctx,
	}

	metastore.watch()

	return metastore, nil
}

func (m *Metastore) IndexMetadataIter() <-chan cmap.Tuple {
	return m.indexMetadataMap.IterBuffered()
}

func (m *Metastore) indexMetadataExists(indexName string) bool {
	_, ok := m.indexMetadataMap.Get(indexName)
	return ok
}

func (m *Metastore) IndexMetadataExists(indexName string) bool {
	return m.indexMetadataExists(indexName)
}

func (m *Metastore) getIndexMetadata(indexName string) *IndexMetadata {
	if tmpIndexMetadata, ok := m.indexMetadataMap.Get(indexName); ok {
		if indexMetadata, ok := tmpIndexMetadata.(*IndexMetadata); ok {
			return indexMetadata
		} else {
			return nil
		}
	} else {
		return nil
	}
}

func (m *Metastore) GetIndexMetadata(indexName string) *IndexMetadata {
	return m.getIndexMetadata(indexName)
}

func (m *Metastore) setIndexMetadata(indexName string, indexMetadata *IndexMetadata) {
	m.indexMetadataMap.Set(indexName, indexMetadata)
}

func (m *Metastore) SetIndexMetadata(indexName string, indexMetadata *IndexMetadata) error {
	// Serialize index metadata
	value, err := indexMetadata.Marshal()
	if err != nil {
		m.logger.Error(err.Error())
		return err
	}

	// Put index metadata
	indexMetadataPath := makeIndexMetadataPath(indexName)
	m.logger.Info("put index metadata", zap.String("path", indexMetadataPath))
	if err := m.storage.Put(m.ctx, indexMetadataPath, value); err != nil {
		m.logger.Error(err.Error(), zap.String("path", indexMetadataPath))
		return err
	}

	for item := range indexMetadata.ShardMetadataIter() {
		shardName := item.Key

		if shardMetadata, ok := item.Val.(*ShardMetadata); ok {
			// Serialize shard metadata
			value, err := shardMetadata.Marshal()
			if err != nil {
				m.logger.Error(err.Error())
				return err
			}

			// Put shard metadata
			shardMetadataPath := makeShardMetadataPath(indexName, shardName)
			m.logger.Info("put shard metadata", zap.String("path", shardMetadataPath))
			if err := m.storage.Put(m.ctx, shardMetadataPath, value); err != nil {
				m.logger.Warn(err.Error(), zap.String("path", shardMetadataPath))
				continue
			}
		} else {
			m.logger.Warn("shard metadata type error", zap.String("shard_name", shardName))
			continue
		}

	}

	return nil
}

func (m *Metastore) deleteIndexMetadata(indexName string) {
	m.indexMetadataMap.Remove(indexName)
}

func (m *Metastore) DeleteIndexMetadata(indexName string) error {
	if !m.IndexMetadataExists(indexName) {
		err := errors.ErrIndexMetadataDoesNotExist
		m.logger.Error(err.Error(), zap.String("index_name", indexName))
		return err
	}

	indexMetadata := m.GetIndexMetadata(indexName)

	for item := range indexMetadata.ShardMetadataIter() {
		shardName := item.Key

		// Delete shard metadata
		shardMetadataPath := makeShardMetadataPath(indexName, shardName)
		if err := m.storage.Delete(m.ctx, shardMetadataPath); err != nil {
			m.logger.Warn(err.Error(), zap.String("path", shardMetadataPath))
			continue
		}
	}

	// Delete index metadata
	indexMetadataPath := makeIndexMetadataPath(indexName)
	if err := m.storage.Delete(m.ctx, indexMetadataPath); err != nil {
		m.logger.Warn(err.Error(), zap.String("index_metadata_path", indexMetadataPath))
	}

	// Delete index metadata directory
	indexMetadataDir := filepath.Dir(indexMetadataPath)
	if err := m.storage.Delete(m.ctx, indexMetadataDir); err != nil {
		m.logger.Warn(err.Error(), zap.String("index_metadata_dir", indexMetadataDir))
	}

	return nil
}

func (m *Metastore) handleStorageEvent(event StorageEvent) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	fileName := filepath.Base(event.Path)

	switch event.Type {
	case StorageEventTypePut:
		if fileName == "index.json" {
			indexName := filepath.Base(filepath.Dir(event.Path))

			indexMetadata, err := NewIndexMetadataWithBytes(event.Value)
			if err != nil {
				m.logger.Warn("failed to make index metadata", zap.Error(err), zap.String("path", event.Path))
				return err
			}

			if m.indexMetadataExists(indexName) {
				// Copy existing shard metadata to new index metadata
				existingIndexMetadata := m.getIndexMetadata(indexName)

				for items := range existingIndexMetadata.ShardMetadataIter() {
					shardName := items.Key
					if shardMetadata, ok := items.Val.(*ShardMetadata); ok {
						indexMetadata.SetShardMetadata(shardName, shardMetadata)
					} else {
						m.logger.Warn("shard metadata type error", zap.String("index_name", indexName), zap.String("shard_name", shardName))
						continue
					}
				}
			}
			m.setIndexMetadata(indexName, indexMetadata)

			if _, ok := m.ringMap[indexName]; !ok {
				m.ringMap[indexName] = rendezvous.New()
			}

			// Send put index event
			m.events <- MetastoreEvent{
				Type:  MetastoreEventTypePutIndex,
				Index: indexName,
			}
			m.logger.Info("sent metastore index put event", zap.String("index_name", indexName))
		} else if strings.HasPrefix(fileName, shardNamePrefix) && strings.HasSuffix(fileName, ".json") {
			indexName := filepath.Base(filepath.Dir(event.Path))
			shardName := strings.TrimSuffix(filepath.Base(event.Path), ".json")

			shardMetadata, err := NewShardMetadataWithBytes(event.Value)
			if err != nil {
				err := errors.ErrInvalidShardMetadata
				m.logger.Warn(err.Error(), zap.Error(err), zap.String("index_name", indexName), zap.String("shard_name", shardName))
				return err
			}

			if m.indexMetadataExists(indexName) {
				m.getIndexMetadata(indexName).SetShardMetadata(shardName, shardMetadata)
			} else {
				err := errors.ErrIndexMetadataDoesNotExist
				m.logger.Warn(err.Error(), zap.String("index_name", indexName))
				return err
			}

			if _, ok := m.ringMap[indexName]; ok {
				m.ringMap[indexName].AddWithWeight(shardName, 1.0)
			} else {
				m.logger.Warn("hash ring does not found", zap.String("index_name", indexName))
				return err
			}

			// Send put shard event
			m.events <- MetastoreEvent{
				Type:  MetastoreEventTypePutShard,
				Index: indexName,
				Shard: shardName,
			}
			m.logger.Info("sent metastore shard put event", zap.String("index_name", indexName), zap.String("shard_name", shardName))
		}
	case StorageEventTypeDelete:
		if fileName == "index.json" {
			indexName := filepath.Base(filepath.Dir(event.Path))

			delete(m.ringMap, indexName)

			m.deleteIndexMetadata(indexName)

			// Send delete index event
			m.events <- MetastoreEvent{
				Type:  MetastoreEventTypeDeleteIndex,
				Index: indexName,
			}
			m.logger.Info("sent metastore index delete event", zap.String("index_name", indexName))
		} else if strings.HasPrefix(fileName, shardNamePrefix) && strings.HasSuffix(fileName, ".json") {
			indexName := filepath.Base(filepath.Dir(event.Path))
			shardName := strings.TrimSuffix(filepath.Base(event.Path), ".json")

			if m.indexMetadataExists(indexName) {
				m.getIndexMetadata(indexName).DeleteShardMetadata(shardName)
			} else {
				err := errors.ErrIndexMetadataDoesNotExist
				m.logger.Warn(err.Error(), zap.String("index_name", indexName))
				return err
			}

			if _, ok := m.ringMap[indexName]; ok {
				m.ringMap[indexName].Remove(shardName)
			} else {
				err := errors.ErrIndexMetadataDoesNotExist
				m.logger.Warn(err.Error(), zap.String("index_name", indexName))
				return err
			}

			// Send put shard event
			m.events <- MetastoreEvent{
				Type:  MetastoreEventTypeDeleteShard,
				Index: indexName,
				Shard: shardName,
			}
			m.logger.Info("sent metastore shard delete event", zap.String("index_name", indexName), zap.String("shard_name", shardName))
		}
	}

	return nil
}

func (m *Metastore) watch() error {
	// Watch storage events.
	go func() {
		for {
			select {
			case cancel := <-m.stopWatching:
				// check
				if cancel {
					return
				}
			case event := <-m.storage.Events():
				if err := m.handleStorageEvent(event); err != nil {
					m.logger.Warn("failed to handle storage event", zap.Error(err))
				}
			}
		}
	}()

	return nil
}

func (m *Metastore) Close() error {
	m.stopWatching <- true

	if err := m.storage.Close(); err != nil {
		m.logger.Error(err.Error())
		return err
	}

	return nil
}

func (m *Metastore) Events() chan MetastoreEvent {
	return m.events
}

func (m *Metastore) TouchShardMetadata(indexName string, shardName string) error {
	indexMetadata := m.getIndexMetadata(indexName)
	if indexMetadata == nil {
		err := errors.ErrIndexMetadataDoesNotExist
		m.logger.Error(err.Error(), zap.String("index_name", indexName))
		return err
	}

	shardMetadata := indexMetadata.GetShardMetadata(shardName)
	if shardMetadata == nil {
		err := errors.ErrShardMetadataDoesNotExist
		m.logger.Error(err.Error(), zap.String("index_name", indexName), zap.String("shard_name", shardName))
		return err
	}

	// Copy new shard metadata
	newShardMetadata := &ShardMetadata{
		ShardName:    shardMetadata.ShardName,
		ShardUri:     shardMetadata.ShardUri,
		ShardLockUri: shardMetadata.ShardLockUri,
		ShardVersion: time.Now().UTC().UnixNano(),
	}

	value, err := newShardMetadata.Marshal()
	if err != nil {
		m.logger.Error(err.Error())
		return err
	}

	// Put new shard metadata
	shardMetadataPath := makeShardMetadataPath(indexName, shardName)
	m.logger.Info("touch shard metadata", zap.String("path", shardMetadataPath))
	if err := m.storage.Put(m.ctx, shardMetadataPath, value); err != nil {
		m.logger.Error(err.Error(), zap.String("path", shardMetadataPath))
		return err
	}

	return nil
}

func (m *Metastore) GetResponsibleShard(indexName string, key string) string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if hashRing, ok := m.ringMap[indexName]; ok {
		return hashRing.Lookup(key)
	} else {
		return ""
	}
}

func (m *Metastore) NumShards(indexName string) int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	indexMetadata := m.getIndexMetadata(indexName)
	if indexMetadata == nil {
		err := errors.ErrIndexMetadataDoesNotExist
		m.logger.Error(err.Error(), zap.String("index_name", indexName))
		return 0
	}

	return indexMetadata.NumShards()
}

func (m *Metastore) GetMapping(indexName string) (mapping.IndexMapping, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	indexMetadata := m.getIndexMetadata(indexName)
	if indexMetadata == nil {
		err := errors.ErrIndexMetadataDoesNotExist
		m.logger.Error(err.Error(), zap.String("index_name", indexName))
		return nil, err
	}

	return indexMetadata.IndexMapping, nil
}
