package metastore

import (
	"encoding/json"

	"github.com/mosuka/phalanx/analysis/analyzer"
	"github.com/mosuka/phalanx/mapping"
	cmap "github.com/orcaman/concurrent-map"
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

func (m *ShardMetadata) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

type IndexMetadata struct {
	IndexName           string                   `json:"index_name"`
	IndexUri            string                   `json:"index_uri"`
	IndexLockUri        string                   `json:"index_lock_uri"`
	IndexMapping        mapping.IndexMapping     `json:"index_mapping"`
	IndexMappingVersion int64                    `json:"index_mapping_version"`
	DefaultSearchField  string                   `json:"default_search_field"`
	DefaultAnalyzer     analyzer.AnalyzerSetting `json:"default_analyzer"`
	shardMetadataMap    cmap.ConcurrentMap       `json:"-"`
}

func NewIndexMetadata() *IndexMetadata {
	return &IndexMetadata{
		shardMetadataMap: cmap.New(),
	}
}

func NewIndexMetadataWithBytes(bytes []byte) (*IndexMetadata, error) {
	indexMetadata := NewIndexMetadata()
	err := json.Unmarshal(bytes, indexMetadata)
	if err != nil {
		return nil, err
	}
	indexMetadata.shardMetadataMap = cmap.New()

	return indexMetadata, nil
}

func (m *IndexMetadata) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func (m *IndexMetadata) ShardMetadataIter() <-chan cmap.Tuple {
	return m.shardMetadataMap.IterBuffered()
}

func (m *IndexMetadata) NumShards() int {
	return m.shardMetadataMap.Count()
}

func (m *IndexMetadata) GetShardMetadata(shardName string) *ShardMetadata {
	if tmpShardMetadata, ok := m.shardMetadataMap.Get(shardName); ok {
		if shardMetadata, ok := tmpShardMetadata.(*ShardMetadata); ok {
			return shardMetadata
		} else {
			return nil
		}
	} else {
		return nil
	}
}

func (m *IndexMetadata) SetShardMetadata(shardName string, shardMetadata *ShardMetadata) {
	m.shardMetadataMap.Set(shardName, shardMetadata)
}

func (m *IndexMetadata) DeleteShardMetadata(shardName string) {
	m.shardMetadataMap.Remove(shardName)
}
