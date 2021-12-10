package membership

import (
	"encoding/json"
)

type NodeRole int

const (
	NodeRoleUnknown NodeRole = iota
	NodeRoleIndexer
	NodeRoleSearcher
)

// Enum value maps for NodeRole.
var (
	NodeRole_name = map[NodeRole]string{
		NodeRoleUnknown:  "unknown",
		NodeRoleIndexer:  "indexer",
		NodeRoleSearcher: "searcher",
	}
	NodeRole_value = map[string]NodeRole{
		"unknown":  NodeRoleUnknown,
		"indexer":  NodeRoleIndexer,
		"searcher": NodeRoleSearcher,
	}
)

type NodeMetadata struct {
	GrpcPort int        `json:"grpc_port"`
	HttpPort int        `json:"http_port"`
	Roles    []NodeRole `json:"roles"`
}

func NewNodeMetadata() *NodeMetadata {
	return &NodeMetadata{
		GrpcPort: 0,
		HttpPort: 0,
		Roles:    []NodeRole{},
	}
}

func NewNodeMetadataWithBytes(data []byte) (*NodeMetadata, error) {
	metadata := NewNodeMetadata()
	if err := json.Unmarshal(data, metadata); err != nil {
		return nil, err
	}

	return metadata, nil
}

func (m *NodeMetadata) Bytes() ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (m *NodeMetadata) IsIndexer() bool {
	for _, role := range m.Roles {
		if role == NodeRoleIndexer {
			return true
		}
	}

	return false
}

func (m *NodeMetadata) IsSearcher() bool {
	for _, role := range m.Roles {
		if role == NodeRoleSearcher {
			return true
		}
	}

	return false
}
