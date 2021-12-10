package index

import (
	"encoding/json"
	"sync"

	"github.com/mosuka/phalanx/mapping"
)

type ShardMetadata struct {
	ShardName    string `json:"shard_name"`
	ShardUri     string `json:"shard_uri"`
	ShardLockUri string `json:"shard_lock_uri"`
	ShardVersion int64  `json:"shard_update_timestamp"`
}

func NewShardMetadataWithBytes(bytes []byte) (*ShardMetadata, error) {
	var shardMetadata ShardMetadata
	err := json.Unmarshal(bytes, &shardMetadata)
	if err != nil {
		return nil, err
	}

	return &shardMetadata, nil
}

type IndexMetadata struct {
	IndexName           string                    `json:"index_name"`
	IndexUri            string                    `json:"index_uri"`
	IndexLockUri        string                    `json:"index_lock_uri"`
	IndexMapping        mapping.IndexMapping      `json:"index_mapping"`
	IndexMappingVersion int64                     `json:"index_mapping_version"`
	DefaultSearchField  string                    `json:"default_search_field"`
	shardMetadataMap    map[string]*ShardMetadata `json:"-"`
	mutex               sync.RWMutex
}

func NewIndexMetadata() *IndexMetadata {
	return &IndexMetadata{
		shardMetadataMap: make(map[string]*ShardMetadata),
	}
}

func NewIndexMetadataWithBytes(bytes []byte) (*IndexMetadata, error) {
	var indexMetadata IndexMetadata
	err := json.Unmarshal(bytes, &indexMetadata)
	if err != nil {
		return nil, err
	}
	indexMetadata.shardMetadataMap = make(map[string]*ShardMetadata)

	return &indexMetadata, nil
}

func (m *IndexMetadata) NumShards() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return len(m.shardMetadataMap)
}

func (m *IndexMetadata) AllShardMetadata() map[string]*ShardMetadata {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.shardMetadataMap
}

func (m *IndexMetadata) ShardMetadataExists(shardName string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	_, ok := m.shardMetadataMap[shardName]

	return ok
}

func (m *IndexMetadata) GetShardMetadata(shardName string) *ShardMetadata {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.shardMetadataMap[shardName]
}

func (m *IndexMetadata) SetShardMetadata(shardName string, shardMetadata *ShardMetadata) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.shardMetadataMap[shardName] = shardMetadata
}

func (m *IndexMetadata) DeleteShardMetadata(shardName string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.shardMetadataMap, shardName)
}

type IndexMetadataMap struct {
	IndexMetadataMap map[string]*IndexMetadata
	mutex            sync.RWMutex
}

func NewIndexMetadataMap() *IndexMetadataMap {
	return &IndexMetadataMap{
		IndexMetadataMap: make(map[string]*IndexMetadata),
	}
}

func (m *IndexMetadataMap) AllIndexMetadata() map[string]*IndexMetadata {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.IndexMetadataMap
}

func (m *IndexMetadataMap) IndexMetadataExists(indexName string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	_, ok := m.IndexMetadataMap[indexName]

	return ok
}

func (m *IndexMetadataMap) GetIndexMetadata(indexName string) *IndexMetadata {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.IndexMetadataMap[indexName]
}

func (m *IndexMetadataMap) SetIndexMetadata(indexName string, indexMetadata *IndexMetadata) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.IndexMetadataMap[indexName] = indexMetadata
}

func (m *IndexMetadataMap) DeleteIndexMetadata(indexName string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.IndexMetadataMap, indexName)
}

func (m *IndexMetadataMap) ShardMetadataExists(indexName string, shardName string) bool {
	if !m.IndexMetadataExists(indexName) {
		return false
	}

	return m.GetIndexMetadata(indexName).ShardMetadataExists(shardName)
}

func (m *IndexMetadataMap) GetShardMetadata(indexName string, shardName string) *ShardMetadata {
	if !m.IndexMetadataExists(indexName) {
		return nil
	}

	return m.GetIndexMetadata(indexName).GetShardMetadata(shardName)
}

func (m *IndexMetadataMap) SetShardMetadata(indexName string, shardName string, shardMetadata *ShardMetadata) {
	if !m.IndexMetadataExists(indexName) {
		return
	}

	m.GetIndexMetadata(indexName).SetShardMetadata(shardName, shardMetadata)
}

func (m *IndexMetadataMap) DeleteShardMetadata(indexName string, shardName string) {
	if !m.IndexMetadataExists(indexName) {
		return
	}

	m.GetIndexMetadata(indexName).DeleteShardMetadata(shardName)
}
