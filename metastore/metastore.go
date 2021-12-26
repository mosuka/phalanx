package metastore

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/copier"
	"github.com/mosuka/phalanx/errors"
	"github.com/mosuka/phalanx/mapping"
	"github.com/mosuka/rendezvous"
	"go.uber.org/zap"
)

const (
	shardNamePrefix = "shard-"
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
	indexMetadataMap map[string]*IndexMetadata
	ringMap          map[string]*rendezvous.Ring
	events           chan MetastoreEvent
	logger           *zap.Logger
	mutex            sync.RWMutex
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
	paths, err := storage.List(string(filepath.Separator))
	if err != nil {
		logger.Error(err.Error(), zap.String("prefix", string(filepath.Separator)))
		return nil, err
	}

	indexMetadataMap := make(map[string]*IndexMetadata)
	ringMap := make(map[string]*rendezvous.Ring)
	for _, path := range paths {
		fileName := filepath.Base(path)
		if fileName == "index.json" {
			value, err := storage.Get(path)
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

			indexMetadataMap[indexName] = indexMetadata
			ringMap[indexName] = rendezvous.New()
		}
	}

	for _, path := range paths {
		fileName := filepath.Base(path)
		if strings.HasPrefix(fileName, shardNamePrefix) && strings.HasSuffix(fileName, ".json") {
			value, err := storage.Get(path)
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
			if indexMetadata, ok := indexMetadataMap[indexName]; ok {
				indexMetadata.ShardMetadataMap[shardName] = shardMetadata
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

	return &Metastore{
		storage:          storage,
		indexMetadataMap: indexMetadataMap,
		ringMap:          ringMap,
		events:           make(chan MetastoreEvent, 10),
		logger:           logger,
		mutex:            sync.RWMutex{},
	}, nil
}

func (m *Metastore) Close() error {
	if err := m.storage.Close(); err != nil {
		m.logger.Error(err.Error())
		return err
	}

	return nil
}

func (m *Metastore) Events() chan MetastoreEvent {
	return m.events
}

func (m *Metastore) IndexMetadataExists(indexName string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	_, ok := m.indexMetadataMap[indexName]
	return ok
}

func (m *Metastore) GetIndexNames() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var indexNames []string

	for indexName := range m.indexMetadataMap {
		indexNames = append(indexNames, indexName)
	}

	return indexNames
}

func (m *Metastore) GetShardNames(indexName string) []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var shardNames []string

	indexMetadata, ok := m.indexMetadataMap[indexName]
	if !ok {
		return []string{}
	}
	for shardName := range indexMetadata.ShardMetadataMap {
		shardNames = append(shardNames, shardName)
	}

	return shardNames
}

func (m *Metastore) GetIndexMetadata(indexName string) (*IndexMetadata, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	indexMetadata, ok := m.indexMetadataMap[indexName]
	if !ok {
		err := errors.ErrIndexMetadataDoesNotExist
		m.logger.Error(err.Error(), zap.String("index_name", indexName))
		return nil, err
	}

	return indexMetadata, nil
}

func (m *Metastore) GetShardMetadata(indexName string, shardName string) (*ShardMetadata, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	indexMetadata, ok := m.indexMetadataMap[indexName]
	if !ok {
		err := errors.ErrIndexMetadataDoesNotExist
		m.logger.Error(err.Error(), zap.String("index_name", indexName))
		return nil, err
	}

	shardMetadata, ok := indexMetadata.ShardMetadataMap[shardName]
	if !ok {
		err := errors.ErrShardMetadataDoesNotExist
		m.logger.Error(err.Error(), zap.String("index_name", indexName), zap.String("shard_name", shardName))
		return nil, err
	}

	return shardMetadata, nil
}

func (m *Metastore) SetIndexMetadata(indexName string, indexMetadata *IndexMetadata) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// new index metadata
	m.logger.Info("new index metadata", zap.String("index_name", indexName))

	// Serialize index metadata
	value, err := indexMetadata.Marshal()
	if err != nil {
		m.logger.Error(err.Error())
		return err
	}

	// Put index metadata
	indexMetadataPath := makeIndexMetadataPath(indexName)
	m.logger.Info("put index metadata", zap.String("path", indexMetadataPath))
	if err := m.storage.Put(indexMetadataPath, value); err != nil {
		m.logger.Error(err.Error(), zap.String("path", indexMetadataPath))
		return err
	}

	// Set hash ring
	m.ringMap[indexName] = rendezvous.New()

	// Send put index event
	m.logger.Info("Send metastore event", zap.String("index_name", indexName))
	m.events <- MetastoreEvent{
		Type:  MetastoreEventTypePutIndex,
		Index: indexName,
	}

	for shardName, shardMetadata := range indexMetadata.ShardMetadataMap {
		value, err := shardMetadata.Marshal()
		if err != nil {
			m.logger.Warn(err.Error())
			continue
		}

		// Put shard metadata
		shardMetadataPath := makeShardMetadataPath(indexName, shardName)
		m.logger.Info("put shard metadata", zap.String("path", shardMetadataPath))
		if err := m.storage.Put(shardMetadataPath, value); err != nil {
			m.logger.Warn(err.Error(), zap.String("path", shardMetadataPath))
			continue
		}

		// Add new hash ring item for shard
		if hashRing, ok := m.ringMap[indexName]; ok {
			hashRing.AddWithWeight(shardName, 1.0)
		} else {
			m.logger.Warn("hash ring does not found", zap.String("index_name", indexName))
		}

		// Send put shard event
		m.events <- MetastoreEvent{
			Type:  MetastoreEventTypePutShard,
			Index: indexName,
			Shard: shardName,
		}
	}

	// Update local index metadata map
	m.indexMetadataMap[indexName] = indexMetadata

	return nil
}

func (m *Metastore) TouchShardMetadata(indexName string, shardName string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	indexMetadata, ok := m.indexMetadataMap[indexName]
	if !ok {
		err := errors.ErrIndexMetadataDoesNotExist
		m.logger.Error(err.Error(), zap.String("index_name", indexName))
		return err
	}

	shardMetadata, ok := indexMetadata.ShardMetadataMap[shardName]
	if !ok {
		err := errors.ErrShardMetadataDoesNotExist
		m.logger.Error(err.Error(), zap.String("index_name", indexName), zap.String("shard_name", shardName))
		return err
	}

	// Copy new shard metadata
	newShardMetadata := &ShardMetadata{}
	copier.Copy(newShardMetadata, shardMetadata)
	newShardMetadata.ShardVersion = time.Now().UTC().UnixNano()

	value, err := newShardMetadata.Marshal()
	if err != nil {
		m.logger.Error(err.Error())
		return err
	}

	// Put new shard metadata
	shardMetadataPath := makeShardMetadataPath(indexName, shardName)
	m.logger.Info("touch shard metadata", zap.String("path", shardMetadataPath))
	if err := m.storage.Put(shardMetadataPath, value); err != nil {
		m.logger.Error(err.Error(), zap.String("path", shardMetadataPath))
		return err
	}

	// Update local shard metadata map
	indexMetadata.ShardMetadataMap[shardName] = newShardMetadata

	// Send metastore event
	m.events <- MetastoreEvent{
		Type:  MetastoreEventTypePutShard,
		Index: indexName,
		Shard: shardName,
	}

	return nil
}

func (m *Metastore) DeleteIndexMetadata(indexName string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	indexMetadata, ok := m.indexMetadataMap[indexName]
	if !ok {
		err := errors.ErrIndexMetadataDoesNotExist
		m.logger.Error(err.Error(), zap.String("index_name", indexName))
		return err
	}

	for shardName := range indexMetadata.ShardMetadataMap {
		// Delete hash ring item for shard
		if hashRing, ok := m.ringMap[indexName]; ok {
			hashRing.Remove(shardName)
		} else {
			m.logger.Warn("hash ring does not found", zap.String("index_name", indexName))
		}

		// Delete shard metadata
		shardMetadataPath := makeShardMetadataPath(indexName, shardName)
		if err := m.storage.Delete(shardMetadataPath); err != nil {
			m.logger.Warn(err.Error(), zap.String("path", shardMetadataPath))
			continue
		}

		// Send event
		m.events <- MetastoreEvent{
			Type:  MetastoreEventTypeDeleteShard,
			Index: indexName,
			Shard: shardName,
		}
	}

	// Delete hash ring for index
	delete(m.ringMap, indexName)

	// Update local index metadata
	delete(m.indexMetadataMap, indexName)

	// Delete index metadata
	indexMetadataPath := makeIndexMetadataPath(indexName)
	if err := m.storage.Delete(indexMetadataPath); err != nil {
		m.logger.Warn(err.Error(), zap.String("index_metadata_path", indexMetadataPath))
	}

	// Delete index metadata directory
	indexMetadataDir := filepath.Dir(indexMetadataPath)
	if err := m.storage.Delete(indexMetadataDir); err != nil {
		m.logger.Warn(err.Error(), zap.String("index_metadata_dir", indexMetadataDir))
	}

	// Send event
	m.events <- MetastoreEvent{
		Type:  MetastoreEventTypeDeleteIndex,
		Index: indexName,
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

	indexMetadata, ok := m.indexMetadataMap[indexName]
	if !ok {
		err := errors.ErrIndexMetadataDoesNotExist
		m.logger.Error(err.Error(), zap.String("index_name", indexName))
		return 0
	}

	return len(indexMetadata.ShardMetadataMap)
}

func (m *Metastore) GetMapping(indexName string) (mapping.IndexMapping, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	indexMetadata, ok := m.indexMetadataMap[indexName]
	if !ok {
		err := errors.ErrIndexMetadataDoesNotExist
		m.logger.Error(err.Error(), zap.String("index_name", indexName))
		return nil, err
	}

	return indexMetadata.IndexMapping, nil
}
