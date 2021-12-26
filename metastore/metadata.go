package metastore

import (
	"encoding/json"

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

func (m *ShardMetadata) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

type IndexMetadata struct {
	IndexName           string                    `json:"index_name"`
	IndexUri            string                    `json:"index_uri"`
	IndexLockUri        string                    `json:"index_lock_uri"`
	IndexMapping        mapping.IndexMapping      `json:"index_mapping"`
	IndexMappingVersion int64                     `json:"index_mapping_version"`
	DefaultSearchField  string                    `json:"default_search_field"`
	ShardMetadataMap    map[string]*ShardMetadata `json:"-"`
}

func NewIndexMetadataWithBytes(bytes []byte) (*IndexMetadata, error) {
	var indexMetadata IndexMetadata
	err := json.Unmarshal(bytes, &indexMetadata)
	if err != nil {
		return nil, err
	}
	indexMetadata.ShardMetadataMap = make(map[string]*ShardMetadata)

	return &indexMetadata, nil
}

func (m *IndexMetadata) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
