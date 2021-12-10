package marshaler

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/mosuka/phalanx/proto"
)

const DefaultContentType = "application/json"

type Marshaler struct{}

func NewMarshaler() *Marshaler {
	return &Marshaler{}
}

func (m *Marshaler) ContentType(v interface{}) string {
	return DefaultContentType
}

func (m *Marshaler) Marshal(v interface{}) ([]byte, error) {
	switch value := v.(type) {
	case *proto.LivenessCheckResponse:
		resp := make(map[string]interface{})
		switch value.State {
		case proto.LivenessState_LIVENESS_STATE_ALIVE:
			resp["state"] = "alive"
		case proto.LivenessState_LIVENESS_STATE_DEAD:
			resp["state"] = "dead"
		default:
			resp["state"] = "unknown"
		}
		return json.Marshal(resp)
	case *proto.ReadinessCheckResponse:
		resp := make(map[string]interface{})
		switch value.State {
		case proto.ReadinessState_READINESS_STATE_READY:
			resp["state"] = "ready"
		case proto.ReadinessState_READINESS_STATE_NOT_READY:
			resp["state"] = "not ready"
		default:
			resp["state"] = "unknown"
		}
		return json.Marshal(resp)
	case *proto.MetricsResponse:
		return value.Metrics, nil
	case *proto.ClusterResponse:
		resp := make(map[string]interface{})
		resp["nodes"] = make(map[string]interface{})
		for name, node := range value.Nodes {
			nodeInfo := make(map[string]interface{})
			nodeInfo["addr"] = node.Addr
			nodeInfo["port"] = node.Port
			nodeMeta := map[string]interface{}{
				"grpc_port": node.Meta.GrpcPort,
				"http_port": node.Meta.HttpPort,
			}
			nodeRoles := make([]string, 0)
			for _, role := range node.Meta.Roles {
				switch role {
				case proto.NodeRole_NODE_ROLE_INDEXER:
					nodeRoles = append(nodeRoles, "indexer")
				case proto.NodeRole_NODE_ROLE_SEARCHER:
					nodeRoles = append(nodeRoles, "searcher")
				default:
					nodeRoles = append(nodeRoles, "unknown")
				}
			}
			var nodeState string
			switch node.State {
			case proto.NodeState_NODE_STATE_ALIVE:
				nodeState = "alive"
			case proto.NodeState_NODE_STATE_DEAD:
				nodeState = "dead"
			case proto.NodeState_NODE_STATE_SUSPECT:
				nodeState = "suspect"
			case proto.NodeState_NODE_STATE_LEFT:
				nodeState = "left"
			default:
				nodeState = "unknown"
			}
			nodeMeta["roles"] = nodeRoles
			nodeInfo["meta"] = nodeMeta
			nodeInfo["state"] = nodeState
			resp["nodes"].(map[string]interface{})[name] = nodeInfo
		}
		resp["indexes"] = make(map[string]interface{})
		for indexName, indexMeta := range value.Indexes {
			indexInfo := make(map[string]interface{})
			indexInfo["index_uri"] = indexMeta.IndexUri
			indexInfo["index_lock_uri"] = indexMeta.IndexLockUri
			indexInfo["shards"] = make(map[string]interface{})
			for shardName, shardMeta := range indexMeta.Shards {
				shardInfo := make(map[string]interface{})
				shardInfo["shard_uri"] = shardMeta.ShardUri
				shardInfo["shard_lock_uri"] = shardMeta.ShardLockUri
				indexInfo["shards"].(map[string]interface{})[shardName] = shardInfo
			}
			resp["indexes"].(map[string]interface{})[indexName] = indexInfo
		}
		var indexerAssignment map[string]map[string]string
		if err := json.Unmarshal(value.IndexerAssignment, &indexerAssignment); err != nil {
			return nil, err
		}
		resp["indexer_assignment"] = indexerAssignment
		var searcherAssignment map[string]map[string][]string
		if err := json.Unmarshal(value.SearcherAssignment, &searcherAssignment); err != nil {
			return nil, err
		}
		resp["searcher_assignment"] = searcherAssignment
		return json.Marshal(resp)
	case *proto.SearchResponse:
		resp := make(map[string]interface{})
		resp["index_name"] = value.IndexName
		resp["hits"] = value.Hits
		docSlice := make([]map[string]interface{}, 0)
		for _, doc := range value.Documents {
			docMap := make(map[string]interface{})
			docMap["id"] = doc.Id
			docMap["score"] = doc.Score
			var fields map[string]interface{}
			if err := json.Unmarshal(doc.Fields, &fields); err != nil {
				return nil, err
			}
			docMap["fields"] = fields
			docSlice = append(docSlice, docMap)
		}
		resp["documents"] = docSlice
		return json.Marshal(resp)
	default:
		return json.Marshal(value)
	}
}

func (m *Marshaler) Unmarshal(data []byte, v interface{}) error {
	switch value := v.(type) {
	case *proto.CreateIndexRequest:
		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			return err
		}
		if indexName, ok := m["index_name"].(string); ok {
			value.IndexName = indexName
		}
		if indexUri, ok := m["index_uri"].(string); ok {
			value.IndexUri = indexUri
		}
		if indexMapping, ok := m["index_mapping"].(map[string]interface{}); ok {
			indexMappingBytes, err := json.Marshal(indexMapping)
			if err != nil {
				return err
			}
			value.IndexMapping = indexMappingBytes
		}
		return nil
	case *proto.DeleteIndexRequest:
		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			return err
		}
		if indexName, ok := m["index_name"].(string); ok {
			value.IndexName = indexName
		}
		return nil
	case *proto.Document:
		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			return err
		}
		id, ok := m["id"].(string)
		if !ok {
			err := fmt.Errorf("document id does not exit or is not a string")
			return err
		}
		fields := m["fields"].(map[string]interface{})
		if !ok {
			err := fmt.Errorf("%s fields do not exist or is not a map[string]interface{}", id)
			return err
		}
		fieldsBytes, err := json.Marshal(fields)
		if err != nil {
			err := fmt.Errorf("%s failed to marshal fields", id)
			return err
		}
		value.Id = id
		value.Fields = fieldsBytes
		return nil
	default:
		return json.Unmarshal(data, value)
	}
}

func (m *Marshaler) NewDecoder(r io.Reader) runtime.Decoder {
	return runtime.DecoderFunc(
		func(v interface{}) error {
			buffer, err := ioutil.ReadAll(r)
			if err != nil {
				return err
			}

			return m.Unmarshal(buffer, v)
		},
	)
}

func (m *Marshaler) NewEncoder(w io.Writer) runtime.Encoder {
	return json.NewEncoder(w)
}

func (m *Marshaler) Delimiter() []byte {
	return []byte("\n")
}
