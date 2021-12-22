package metastore

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

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
// e.g. wikipedia_en -> wikipedia_en/shard-pb0g8d8hmvcg9hvaiol3.json
func makeShardMetadataPath(indexName string, shardName string) string {
	return filepath.Join(indexName, fmt.Sprintf("%s.json", shardName))
}

type Metastore struct {
	storage          Storage
	indexMetadataMap map[string]*IndexMetadata
	ringMap          map[string]*rendezvous.Ring
	events           chan MetastoreEvent
	stopWatching     chan bool
	logger           *zap.Logger
	mutex            sync.RWMutex
}

func NewMetastore(uri string, logger *zap.Logger) (*Metastore, error) {
	metastoreLogger := logger.Named("metastore")

	storage, err := NewStorageWithUri(uri, metastoreLogger)
	if err != nil {
		metastoreLogger.Error(err.Error(), zap.String("uri", uri))
		return nil, err
	}

	paths, err := storage.List(string(filepath.Separator))
	if err != nil {
		metastoreLogger.Error(err.Error(), zap.String("prefix", string(filepath.Separator)))
		return nil, err
	}

	indexMetadataMap := make(map[string]*IndexMetadata)
	ringMap := make(map[string]*rendezvous.Ring)
	for _, path := range paths {
		fileName := filepath.Base(path)
		if fileName == "index.json" {
			value, err := storage.Get(path)
			if err != nil {
				metastoreLogger.Error(err.Error(), zap.String("path", path))
				return nil, err
			}

			indexMetadata, err := NewIndexMetadataWithBytes(value)
			if err != nil {
				metastoreLogger.Error(err.Error())
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
				metastoreLogger.Error(err.Error(), zap.String("path", path))
				return nil, err
			}

			shardMetadata, err := NewShardMetadataWithBytes(value)
			if err != nil {
				metastoreLogger.Error(err.Error())
				return nil, err
			}

			indexName := filepath.Base(filepath.Dir(path))
			shardName := strings.TrimSuffix(filepath.Base(path), ".json")

			// Update shard metadata
			if indexMetadata, ok := indexMetadataMap[indexName]; ok {
				indexMetadata.SetShardMetadata(shardName, shardMetadata)
			} else {
				metastoreLogger.Warn("index metadata do not found", zap.String("index_name", indexName))
			}

			// Add new hash ring item for shard.
			if hashRing, ok := ringMap[indexName]; ok {
				hashRing.AddWithWeight(shardName, 1.0)
			} else {
				metastoreLogger.Warn("hash ring do not found", zap.String("index_name", indexName))
			}
		}
	}

	return &Metastore{
		storage:          storage,
		indexMetadataMap: indexMetadataMap,
		ringMap:          ringMap,
		events:           make(chan MetastoreEvent, 10),
		stopWatching:     make(chan bool),
		logger:           metastoreLogger,
	}, nil
}

func (m *Metastore) handleStorageEvent(event StorageEvent) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	switch event.Type {
	case StorageEventTypePut:
		fileName := filepath.Base(event.Path)
		if fileName == "index.json" {
			indexName := filepath.Base(filepath.Dir(event.Path))

			// Update index metadata
			indexMetadata, err := NewIndexMetadataWithBytes(event.Value)
			if err != nil {
				m.logger.Error(err.Error())
				return err
			}
			if _, ok := m.indexMetadataMap[indexName]; ok {
				m.logger.Warn("overwrite index metadata", zap.Error(err), zap.String("index_name", indexName))
			}
			m.indexMetadataMap[indexName] = indexMetadata

			// Add new hash ring for index
			if _, ok := m.ringMap[indexName]; !ok {
				m.ringMap[indexName] = rendezvous.New()
			}

			// Send event
			m.events <- MetastoreEvent{
				Type:  MetastoreEventTypePutIndex,
				Index: indexName,
			}
		} else if strings.HasPrefix(fileName, shardNamePrefix) && strings.HasSuffix(fileName, ".json") {
			indexName := filepath.Base(filepath.Dir(event.Path))
			shardName := strings.TrimSuffix(filepath.Base(event.Path), ".json")

			// Update shard metadata
			shardMetadata, err := NewShardMetadataWithBytes(event.Value)
			if err != nil {
				m.logger.Error(err.Error())
				return err
			}
			if indexMetadata, ok := m.indexMetadataMap[indexName]; ok {
				if indexMetadata.ShardMetadataExists(shardName) {
					m.logger.Warn("overwrite shard metadata", zap.Error(err), zap.String("index_name", indexName), zap.String("shard_name", shardName))
				}
				indexMetadata.SetShardMetadata(shardName, shardMetadata)
			} else {
				err := errors.ErrIndexMetadataDoesNotExist
				m.logger.Warn(err.Error(), zap.String("index_name", indexName))
			}

			// Add new hash ring item for shard
			if hashRing, ok := m.ringMap[indexName]; ok {
				hashRing.AddWithWeight(shardName, 1.0)
			} else {
				m.logger.Warn("hash ring does not found", zap.String("index_name", indexName))
			}

			// Send event
			m.events <- MetastoreEvent{
				Type:  MetastoreEventTypePutShard,
				Index: indexName,
				Shard: shardName,
			}
		}
	case StorageEventTypeDelete:
		fileName := filepath.Base(event.Path)
		if fileName == "index.json" {
			indexName := filepath.Base(filepath.Dir(event.Path))

			// Delete index metadata
			if _, ok := m.indexMetadataMap[indexName]; ok {
				delete(m.indexMetadataMap, indexName)
			} else {
				err := errors.ErrIndexMetadataDoesNotExist
				m.logger.Warn(err.Error(), zap.String("index_name", indexName))
			}

			// Delete hash ring for index
			if _, ok := m.ringMap[indexName]; ok {
				delete(m.ringMap, indexName)
			} else {
				m.logger.Warn("hash ring does not found", zap.String("index_name", indexName))
			}

			// Send event
			m.events <- MetastoreEvent{
				Type:  MetastoreEventTypeDeleteIndex,
				Index: indexName,
			}
		} else if strings.HasPrefix(fileName, shardNamePrefix) && strings.HasSuffix(fileName, ".json") {
			indexName := filepath.Base(filepath.Dir(event.Path))
			shardName := strings.TrimSuffix(filepath.Base(event.Path), ".json")

			// Delete shard metadata
			if indexMetadata, ok := m.indexMetadataMap[indexName]; ok {
				indexMetadata.DeleteShardMetadata(shardName)
			} else {
				err := errors.ErrIndexMetadataDoesNotExist
				m.logger.Warn(err.Error(), zap.String("index_name", indexName))
			}

			// Delete hash ring item for shard
			if hashRing, ok := m.ringMap[indexName]; ok {
				hashRing.Remove(shardName)
			} else {
				m.logger.Warn("hash ring does not found", zap.String("index_name", indexName))
			}

			// Send event
			m.events <- MetastoreEvent{
				Type:  MetastoreEventTypeDeleteShard,
				Index: indexName,
				Shard: shardName,
			}
		}
	}

	return nil
}

func (m *Metastore) Start() error {
	// if err := m.storage.Start(); err != nil {
	// 	m.logger.Error(err.Error())
	// 	return err
	// }

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
					m.logger.Warn(err.Error(), zap.Any("event", event))
					continue
				}
			}
		}
	}()

	return nil
}

func (m *Metastore) Stop() error {
	if err := m.storage.Close(); err != nil {
		m.logger.Error(err.Error())
		return err
	}

	m.stopWatching <- true

	return nil
}

func (m *Metastore) Events() chan MetastoreEvent {
	return m.events
}

func (m *Metastore) IndexMetadataExists(indexName string) bool {
	_, ok := m.indexMetadataMap[indexName]
	return ok
}

func (m *Metastore) GetIndexNames() []string {
	var indexNames []string

	for indexName := range m.ringMap {
		indexNames = append(indexNames, indexName)
	}

	return indexNames
}

func (m *Metastore) GetShardNames(indexName string) []string {
	if shardsRing, ok := m.ringMap[indexName]; !ok {
		return []string{}
	} else {
		return shardsRing.List()
	}
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

	indexMetadata, err := m.GetIndexMetadata(indexName)
	if err != nil {
		m.logger.Error(err.Error(), zap.String("index_name", indexName))
		return nil, err
	}

	shardMetadata, err := indexMetadata.GetShardMetadata(shardName)
	if err != nil {
		m.logger.Error(err.Error(), zap.String("index_name", indexName), zap.String("shard_name", shardName))
		return nil, err
	}

	return shardMetadata, nil
}

func (m *Metastore) SetIndexMetadata(indexName string, indexMetadata *IndexMetadata) error {
	value, err := indexMetadata.Marshal()
	if err != nil {
		m.logger.Error(err.Error())
		return err
	}

	indexMetadataPath := makeIndexMetadataPath(indexName)
	if err := m.storage.Put(indexMetadataPath, value); err != nil {
		m.logger.Error(err.Error(), zap.String("index_metadata_path", indexMetadataPath))
		return err
	}

	for shardName, shardMetadata := range indexMetadata.ShardMetadataMap {
		if err := m.SetShardMetadata(indexName, shardName, shardMetadata); err != nil {
			m.logger.Error(err.Error(), zap.String("index_name", indexName), zap.String("shard_name", shardName))
			return err
		}
	}

	return nil
}

func (m *Metastore) SetShardMetadata(indexName string, shardName string, shardMetadata *ShardMetadata) error {
	value, err := shardMetadata.Marshal()
	if err != nil {
		m.logger.Error(err.Error())
		return err
	}

	shardMetadataPath := makeShardMetadataPath(indexName, shardName)
	if err := m.storage.Put(shardMetadataPath, value); err != nil {
		m.logger.Error(err.Error(), zap.String("shard_metadata_path", shardMetadataPath))
		return err
	}

	return nil
}

func (m *Metastore) TouchShardMetadata(indexName string, shardName string) error {
	shardMetadata, err := m.GetShardMetadata(indexName, shardName)
	if err != nil {
		m.logger.Error(err.Error(), zap.String("index_name", indexName), zap.String("shard_name", shardName))
		return err
	}

	shardMetadata.ShardVersion = time.Now().UTC().UnixNano()

	if err := m.SetShardMetadata(indexName, shardName, shardMetadata); err != nil {
		m.logger.Error(err.Error(), zap.String("index_name", indexName), zap.String("shard_name", shardName))
		return err
	}

	return nil
}

func (m *Metastore) DeleteIndexMetadata(indexName string) error {
	indexMetadata, err := m.GetIndexMetadata(indexName)
	if err != nil {
		m.logger.Error(err.Error(), zap.String("index_name", indexName))
		return err
	}

	for shardName := range indexMetadata.ShardMetadataMap {
		if err := m.DeleteShardMetadata(indexName, shardName); err != nil {
			m.logger.Warn(err.Error(), zap.String("index_name", indexName), zap.String("shard_name", shardName))
			continue
		}
	}

	indexMetadataPath := makeIndexMetadataPath(indexName)
	if err := m.storage.Delete(indexMetadataPath); err != nil {
		m.logger.Error(err.Error(), zap.String("index_metadata_path", indexMetadataPath))
		return err
	}

	return nil
}

func (m *Metastore) DeleteShardMetadata(indexName string, shardName string) error {
	shardMetadataPath := makeShardMetadataPath(indexName, shardName)
	if err := m.storage.Delete(shardMetadataPath); err != nil {
		m.logger.Error(err.Error(), zap.String("shard_metadata_path", shardMetadataPath))
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
	indexMetadata, err := m.GetIndexMetadata(indexName)
	if err != nil {
		m.logger.Error(err.Error(), zap.String("index_name", indexName))
		return 0
	}

	return indexMetadata.NumShards()
}

func (m *Metastore) GetMapping(indexName string) (mapping.IndexMapping, error) {
	indexMetadata, err := m.GetIndexMetadata(indexName)
	if err != nil {
		m.logger.Error(err.Error(), zap.String("index_name", indexName))
		return nil, err
	}

	return indexMetadata.IndexMapping, nil
}
