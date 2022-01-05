package server

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

		docs := make([]map[string]interface{}, 0)
		for _, doc := range value.Documents {
			var fields map[string]interface{}
			if err := json.Unmarshal(doc.Fields, &fields); err != nil {
				return nil, err
			}

			docMap := map[string]interface{}{
				"id":        doc.Id,
				"score":     doc.Score,
				"timestamp": doc.Timestamp,
				"fields":    fields,
			}

			docs = append(docs, docMap)
		}
		resp["documents"] = docs

		resp["aggregations"] = make(map[string]interface{})
		for aggName, aggResp := range value.Aggregations {
			values := make(map[string]float64)
			for name, count := range aggResp.Buckets {
				values[name] = count
			}
			resp["aggregations"].(map[string]interface{})[aggName] = values
		}

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

		if lockUri, ok := m["lock_uri"].(string); ok {
			value.LockUri = lockUri
		}

		if indexMapping, ok := m["index_mapping"].(map[string]interface{}); ok {
			indexMappingBytes, err := json.Marshal(indexMapping)
			if err != nil {
				return err
			}
			value.IndexMapping = indexMappingBytes
		}

		numShards, ok := m["num_shards"].(float64)
		if ok {
			value.NumShards = uint32(numShards)
		} else {
			value.NumShards = 1
		}

		if defaultSearchField, ok := m["default_search_field"].(string); ok {
			value.DefaultSearchField = defaultSearchField
		}

		if defaultAnalyzer, ok := m["default_analyzer"].(map[string]interface{}); ok {
			defaultAnalyuzerBytes, err := json.Marshal(defaultAnalyzer)
			if err != nil {
				return err
			}
			value.DefaultAnalyzer = defaultAnalyuzerBytes
		}

		return nil
	case *proto.SearchRequest:
		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			return err
		}

		if indexName, ok := m["index_name"].(string); ok {
			value.IndexName = indexName
		}

		if query, ok := m["query"].(string); ok {
			value.Query = query
		}

		if boost, ok := m["boost"].(float64); ok {
			value.Boost = boost
		}

		if start, ok := m["start"].(float64); ok {
			value.Start = int32(start)
		}

		if num, ok := m["num"].(float64); ok {
			value.Num = int32(num)
		}

		if sortBy, ok := m["sort_by"].(string); ok {
			value.SortBy = sortBy
		}

		if fields, ok := m["fields"].([]interface{}); ok {
			value.Fields = make([]string, len(fields))
			for i, fieldValue := range fields {
				field, ok := fieldValue.(string)
				if !ok {
					return fmt.Errorf("fields option has unexpected data: %v", fieldValue)
				}
				value.Fields[i] = field
			}
		}

		if aggregations, ok := m["aggregations"].(map[string]interface{}); ok {
			value.Aggregations = make(map[string]*proto.AggregationRequest)
			for name, aggregation := range aggregations {
				if agg, ok := aggregation.(map[string]interface{}); ok {
					aggType, ok := agg["type"].(string)
					if !ok {
						return fmt.Errorf("aggregation type is not a string: %v", agg["type"])
					}
					aggOpts, ok := agg["options"].(map[string]interface{})
					if !ok {
						return fmt.Errorf("aggregation options is not a map: %v", agg["params"])
					}
					aggOptsBytes, err := json.Marshal(aggOpts)
					if err != nil {
						return err
					}
					value.Aggregations[name] = &proto.AggregationRequest{
						Type:    aggType,
						Options: aggOptsBytes,
					}
				}
			}
		}

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
