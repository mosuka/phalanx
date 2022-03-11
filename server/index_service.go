package server

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/numeric/geo"
	"github.com/jinzhu/copier"
	"github.com/mosuka/phalanx/analysis/analyzer"
	phalanxclients "github.com/mosuka/phalanx/clients"
	phalanxcluster "github.com/mosuka/phalanx/cluster"
	"github.com/mosuka/phalanx/directory"
	"github.com/mosuka/phalanx/errors"
	"github.com/mosuka/phalanx/index"
	"github.com/mosuka/phalanx/mapping"
	phalanxmetastore "github.com/mosuka/phalanx/metastore"
	"github.com/mosuka/phalanx/proto"
	phalanxaggregations "github.com/mosuka/phalanx/search/aggregations"
	phalanxhighlight "github.com/mosuka/phalanx/search/highlight"
	phalanxqueries "github.com/mosuka/phalanx/search/queries"
	"github.com/mosuka/phalanx/util/wildcard"
	"github.com/thanhpk/randstr"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	shardNamePrefix = "shard-"
)

func generateShardName() string {
	return fmt.Sprintf("%s%s", shardNamePrefix, randstr.String(8))
}

func shuffleNodes(nodeNames []string) {
	n := len(nodeNames)
	for i := n - 1; i >= 0; i-- {
		j := rand.Intn(i + 1)
		nodeNames[i], nodeNames[j] = nodeNames[j], nodeNames[i]
	}
}

type IndexService struct {
	cluster            *phalanxcluster.Cluster
	metastore          *phalanxmetastore.Metastore
	certificateFile    string
	commonName         string
	logger             *zap.Logger
	indexWriters       *index.IndexWriters
	indexReaders       *index.IndexReaders
	stopWatching       chan bool
	indexerAssignment  map[string]map[string]string
	searcherAssignment map[string]map[string][]string
	clients            map[string]*phalanxclients.GRPCIndexClient
	mutex              sync.RWMutex
}

func NewIndexService(cluster *phalanxcluster.Cluster, metastore *phalanxmetastore.Metastore, certificateFile string, commonName string, logger *zap.Logger) (*IndexService, error) {
	managerLogger := logger.Named("manager")

	return &IndexService{
		cluster:            cluster,
		metastore:          metastore,
		certificateFile:    certificateFile,
		commonName:         commonName,
		logger:             logger,
		indexWriters:       index.NewIndexWriters(managerLogger),
		indexReaders:       index.NewIndexReaders(managerLogger),
		stopWatching:       make(chan bool),
		indexerAssignment:  map[string]map[string]string{},
		searcherAssignment: map[string]map[string][]string{},
		clients:            map[string]*phalanxclients.GRPCIndexClient{},
		mutex:              sync.RWMutex{},
	}, nil
}

func (s *IndexService) Start() error {
	// Watch metastore events and cluster events.
	go func() {
		for {
			select {
			case cancel := <-s.stopWatching:
				// check
				if cancel {
					return
				}
			case event := <-s.metastore.Events():
				switch event.Type {
				case phalanxmetastore.MetastoreEventTypePutIndex:
					s.logger.Info("index metadata has been created", zap.Any("metastore_event", event))
					// NOP
				case phalanxmetastore.MetastoreEventTypeDeleteIndex:
					s.logger.Info("index metadata has been deleted", zap.Any("metastore_event", event))
					// NOP
				case phalanxmetastore.MetastoreEventTypePutShard:
					s.logger.Info("shard metadata has been created", zap.Any("metastore_event", event))
					s.assignShardsToNode()
				case phalanxmetastore.MetastoreEventTypeDeleteShard:
					s.logger.Info("shard metadata has been deleted", zap.Any("metastore_event", event))
					s.assignShardsToNode()
				}
			case event := <-s.cluster.ClusterEvents():
				switch event.NodeEvent.Type {
				case phalanxcluster.NodeEventTypeJoin:
					s.logger.Info("node has been joined", zap.Any("cluster_event", event))
					if s.cluster.IsSeedNode() || event.NodeEvent.NodeName != s.cluster.LocalNodeName() {
						s.assignShardsToNode()
					}
				case phalanxcluster.NodeEventTypeUpdate:
					s.logger.Info("node has been updated", zap.Any("cluster_event", event))
					s.assignShardsToNode()
				case phalanxcluster.NodeEventTypeLeave:
					s.logger.Info("node has been left", zap.Any("cluster_event", event))
					s.assignShardsToNode()
				}
			}
		}
	}()

	return nil
}

func (s *IndexService) Stop() error {
	s.stopWatching <- true

	// Close all index writers.
	if err := s.indexWriters.CloseAll(); err != nil {
		s.logger.Warn(err.Error())
	}

	// Close all index readers.
	if err := s.indexReaders.CloseAll(); err != nil {
		s.logger.Warn(err.Error())
	}

	// Close all index clients.
	for address, indexClient := range s.clients {
		if err := indexClient.Close(); err != nil {
			s.logger.Warn(err.Error(), zap.String("address", address))
		}
	}

	return nil
}

func (s *IndexService) openAndCloseWriters() {
	// Open the index writers for assigned shards.
	s.logger.Info("opening index writers")
	for assignedIndexName, shardAssignment := range s.indexerAssignment {
		for assignedShardName, assignedNodeName := range shardAssignment {
			isAssigned := assignedNodeName == s.cluster.LocalNodeName()
			if isAssigned {
				if !s.indexWriters.Contains(assignedIndexName, assignedShardName) {
					indexMetadata := s.metastore.GetIndexMetadata(assignedIndexName)
					if indexMetadata == nil {
						err := errors.ErrIndexMetadataDoesNotExist
						s.logger.Warn(err.Error(), zap.String("index_name", assignedIndexName))
						continue
					}

					shardMetadata := indexMetadata.GetShardMetadata(assignedShardName)
					if shardMetadata == nil {
						err := errors.ErrShardMetadataDoesNotExist
						s.logger.Warn(err.Error(), zap.String("index_name", assignedIndexName), zap.String("shard_name", assignedShardName))
						continue
					}

					if err := s.indexWriters.Open(assignedIndexName, assignedShardName, indexMetadata, shardMetadata); err != nil {
						s.logger.Warn(err.Error(), zap.String("index_name", assignedIndexName), zap.String("shard_name", assignedShardName))
						continue
					}
					s.logger.Info("opened the index writer for the assigned shard", zap.String("index_name", assignedIndexName), zap.String("shard_name", assignedShardName))
				}
			} else {
				if s.indexWriters.Contains(assignedIndexName, assignedShardName) {
					// Close the index writer for the shard assigned to the other node.
					if err := s.indexWriters.Close(assignedIndexName, assignedShardName); err != nil {
						s.logger.Warn(err.Error(), zap.String("index_name", assignedIndexName), zap.String("shard_name", assignedShardName))
						continue
					}
					s.logger.Info("closed the index writer for a shard assigned to another node", zap.String("index_name", assignedIndexName), zap.String("shard_name", assignedShardName))
				}
			}
		}
	}

	// Close the index writers for unassigned shards.
	for _, openedIndexName := range s.indexWriters.Indexes() {
		for _, openedShardName := range s.indexWriters.Shards(openedIndexName) {
			assignedNodeName, ok := s.indexerAssignment[openedIndexName][openedShardName]
			if !ok {
				// Close the index writer for shard that doesn't already exist.
				if err := s.indexWriters.Close(openedIndexName, openedShardName); err != nil {
					s.logger.Warn(err.Error(), zap.String("index_name", openedIndexName), zap.String("shard_name", openedShardName))
					continue
				}
				s.logger.Info("closed the index writer for the non-existent shard", zap.String("index_name", openedIndexName), zap.String("shard_name", openedShardName))
			}

			isAssigned := assignedNodeName == s.cluster.LocalNodeName()
			if !isAssigned {
				// Close the index writer for the shard assigned to the other node.
				if err := s.indexWriters.Close(openedIndexName, openedShardName); err != nil {
					s.logger.Warn(err.Error(), zap.String("index_name", openedIndexName), zap.String("shard_name", openedShardName))
					continue
				}
				s.logger.Info("closed the index writer for a shard assigned to another node", zap.String("index_name", openedShardName), zap.String("shard_name", openedShardName))
			}
		}
	}
}

func (s *IndexService) openAndCloseSearchers() {
	// open searchers
	for assignedIndexName, shardAssignment := range s.searcherAssignment {
		for assignedShardName, assignedNodeNames := range shardAssignment {
			isAssigned := false
			for _, assignedNodeName := range assignedNodeNames {
				if assignedNodeName == s.cluster.LocalNodeName() {
					isAssigned = true
					break
				}
			}
			if isAssigned {
				indexMetadata := s.metastore.GetIndexMetadata(assignedIndexName)
				if indexMetadata == nil {
					err := errors.ErrIndexMetadataDoesNotExist
					s.logger.Warn(err.Error(), zap.String("index_name", assignedIndexName))
					continue
				}

				shardMetadata := indexMetadata.GetShardMetadata(assignedShardName)
				if shardMetadata == nil {
					err := errors.ErrShardMetadataDoesNotExist
					s.logger.Warn(err.Error(), zap.String("index_name", assignedIndexName), zap.String("shard_name", assignedShardName))
					continue
				}

				if s.indexReaders.Contains(assignedIndexName, assignedShardName) {
					if s.indexReaders.Version(assignedIndexName, assignedShardName) != shardMetadata.ShardVersion {
						// Reopen the index reader for the assigned shard.
						if err := s.indexReaders.Reopen(assignedIndexName, assignedShardName, indexMetadata, shardMetadata); err != nil {
							s.logger.Warn(err.Error(), zap.String("index_name", assignedIndexName), zap.String("shard_name", assignedShardName))
							continue
						}
						s.logger.Info("reopen the existing index reader due to the index has been updated", zap.String("index_name", assignedIndexName), zap.String("shard_name", assignedShardName))
					}
				} else {
					// TODO: check the shard index existence before opening
					// Open the index reader for the assigned shard.
					if err := s.indexReaders.Open(assignedIndexName, assignedShardName, indexMetadata, shardMetadata); err != nil {
						s.logger.Warn(err.Error(), zap.String("index_name", assignedIndexName), zap.String("shard_name", assignedShardName))
						continue
					}
					s.logger.Info("opened the index reader for the assigned shard", zap.String("index_name", assignedIndexName), zap.String("shard_name", assignedShardName))
				}
			} else {
				// Close the index reader for the shard assigned to the other node.
				if s.indexReaders.Contains(assignedIndexName, assignedShardName) {
					// TODO: check the shard index existence before closing
					if err := s.indexReaders.Close(assignedIndexName, assignedShardName); err != nil {
						s.logger.Warn(err.Error(), zap.String("index_name", assignedIndexName), zap.String("shard_name", assignedShardName))
						continue
					}
					s.logger.Info("closed the index reader for a shard assigned to another node", zap.String("index_name", assignedIndexName), zap.String("shard_name", assignedShardName))
				}
			}
		}
	}

	// Close the index readers for unassigned shards.
	for _, indexName := range s.indexReaders.Indexes() {
		for _, shardName := range s.indexReaders.Shards(indexName) {
			assignedNodenames, ok := s.searcherAssignment[indexName][shardName]
			if !ok {
				// Close the index reader for shard that doesn't already exist.
				if err := s.indexReaders.Close(indexName, shardName); err != nil {
					s.logger.Warn(err.Error(), zap.String("index_name", indexName), zap.String("shard_name", shardName))
				}
				s.logger.Info("closed the index reader for the non-existent shard", zap.String("index_name", indexName), zap.String("shard_name", shardName))
				continue
			}

			isAssigned := false
			for _, assignedNodeName := range assignedNodenames {
				if assignedNodeName == s.cluster.LocalNodeName() {
					isAssigned = true
					break
				}
			}
			if !isAssigned {
				// Close the index reader for the shard assigned to the other node.
				if err := s.indexReaders.Close(indexName, shardName); err != nil {
					s.logger.Warn(err.Error(), zap.String("index_name", indexName), zap.String("shard_name", shardName))
				}
				s.logger.Info("closed the index reader for a shard assigned to another node", zap.String("index_name", indexName), zap.String("shard_name", shardName))
				continue
			}
		}
	}
}

func (s *IndexService) openAndCloseClients() {
	// open clients
	for _, nodeName := range s.cluster.Nodes() {
		if nodeName != s.cluster.LocalNodeName() {
			metadata, err := s.cluster.NodeMetadata(nodeName)
			if err != nil {
				s.logger.Warn(err.Error(), zap.String("node_name", nodeName))
				continue
			}

			nodeAddr, err := s.cluster.NodeAddress(nodeName)
			if err != nil {
				s.logger.Warn(err.Error(), zap.String("node_name", nodeName))
				continue
			}

			grpcAddress := fmt.Sprintf("%s:%d", nodeAddr, metadata.GrpcPort)
			if _, ok := s.clients[grpcAddress]; !ok {
				client, err := phalanxclients.NewGRPCIndexClientWithTLS(grpcAddress, s.certificateFile, s.commonName)
				if err != nil {
					s.logger.Warn(err.Error(), zap.String("grpc_address", grpcAddress), zap.String("certificate_file", s.certificateFile), zap.String("common_name", s.commonName))
					continue
				}
				s.mutex.Lock()
				s.clients[grpcAddress] = client
				s.mutex.Unlock()
			}
		}
	}

	// close clients
	for address, client := range s.clients {
		notFound := true
		for _, nodeName := range s.cluster.Nodes() {
			metadata, err := s.cluster.NodeMetadata(nodeName)
			if err != nil {
				s.logger.Warn(err.Error(), zap.String("node_name", nodeName))
				continue
			}

			nodeAddr, err := s.cluster.NodeAddress(nodeName)
			if err != nil {
				s.logger.Warn(err.Error(), zap.String("node_name", nodeName))
				continue
			}

			grpcAddress := fmt.Sprintf("%s:%d", nodeAddr, metadata.GrpcPort)
			if address == grpcAddress {
				notFound = false
				break
			}
		}
		if notFound {
			if err := client.Close(); err != nil {
				s.logger.Warn(err.Error(), zap.String("address", client.Address()))
			}
			if _, ok := s.clients[address]; ok {
				s.mutex.Lock()
				delete(s.clients, address)
				s.mutex.Unlock()
			}
		}
	}
}

func (s *IndexService) assignShardsToNode() error {
	searchReplicationFactor := 3
	indexerAssignment := make(map[string]map[string]string)    // index/shard/node
	searcherAssignment := make(map[string]map[string][]string) // index/shard/nodes

	// Assign shards to indexers and searchers.
	for item := range s.metastore.IndexMetadataIter() {
		indexName := item.Key

		if indexMetadata, ok := item.Val.(*phalanxmetastore.IndexMetadata); ok {
			for shardItem := range indexMetadata.ShardMetadataIter() {
				shardName := shardItem.Key

				// Assign indexer.
				if _, ok := indexerAssignment[indexName]; !ok {
					indexerAssignment[indexName] = make(map[string]string)
				}
				indexerAssignment[indexName][shardName] = s.cluster.LookupIndexer(shardName)

				// Assign searchers.
				if _, ok := searcherAssignment[indexName]; !ok {
					searcherAssignment[indexName] = make(map[string][]string)
				}
				searcherAssignment[indexName][shardName] = s.cluster.LookupSearchers(shardName, searchReplicationFactor)
			}
		} else {
			s.logger.Warn("index metadata type error", zap.Any("index_name", indexName))
		}
	}

	s.indexerAssignment = indexerAssignment
	s.searcherAssignment = searcherAssignment

	// fmt.Println("indexer:")
	// for indexName, shards := range s.indexerAssignment {
	// 	fmt.Println("  index:", indexName)
	// 	for shardName, nodeName := range shards {
	// 		fmt.Println("    shard:", shardName)
	// 		fmt.Println("      node:", nodeName)
	// 	}
	// }
	// fmt.Println("searcher:")
	// for indexName, shards := range s.searcherAssignment {
	// 	fmt.Println("  index:", indexName)
	// 	for shardName, nodes := range shards {
	// 		fmt.Println("    shard:", shardName)
	// 		for _, nodeName := range nodes {
	// 			fmt.Println("      node:", nodeName)
	// 		}
	// 	}
	// }

	s.openAndCloseWriters()
	s.openAndCloseSearchers()
	s.openAndCloseClients()

	return nil
}

func (s *IndexService) Cluster(ctx context.Context, req *proto.ClusterRequest) (*proto.ClusterResponse, error) {
	resp := &proto.ClusterResponse{}

	resp.Nodes = make(map[string]*proto.Node)
	for _, member := range s.cluster.Nodes() {
		// Deserialize the node metadata
		nodeMetadata, err := s.cluster.NodeMetadata(member)
		if err != nil {
			nodeMetadata = &phalanxcluster.NodeMetadata{}
		}

		nodeRoles := make([]proto.NodeRole, 0)
		for _, role := range nodeMetadata.Roles {
			switch role {
			case phalanxcluster.NodeRoleIndexer:
				nodeRoles = append(nodeRoles, proto.NodeRole_NODE_ROLE_INDEXER)
			case phalanxcluster.NodeRoleSearcher:
				nodeRoles = append(nodeRoles, proto.NodeRole_NODE_ROLE_SEARCHER)
			default:
				nodeRoles = append(nodeRoles, proto.NodeRole_NODE_ROLE_UNKNOWN)
			}
		}

		nodeState, err := s.cluster.NodeState(member)
		if err != nil {
			return nil, err
		}

		state := proto.NodeState_NODE_STATE_UNKNOWN
		switch nodeState {
		case phalanxcluster.NodeStateAlive:
			state = proto.NodeState_NODE_STATE_ALIVE
		case phalanxcluster.NodeStateSuspect:
			state = proto.NodeState_NODE_STATE_SUSPECT
		case phalanxcluster.NodeStateDead:
			state = proto.NodeState_NODE_STATE_DEAD
		case phalanxcluster.NodeStateLeft:
			state = proto.NodeState_NODE_STATE_LEFT
		}

		nodeAddr, err := s.cluster.NodeAddress(member)
		if err != nil {
			return nil, err
		}
		nodePort, err := s.cluster.NodePort(member)
		if err != nil {
			return nil, err
		}

		node := &proto.Node{
			Addr: nodeAddr,
			Port: uint32(nodePort),
			Meta: &proto.NodeMeta{
				GrpcPort: uint32(nodeMetadata.GrpcPort),
				HttpPort: uint32(nodeMetadata.HttpPort),
				Roles:    nodeRoles,
			},
			State: state,
		}

		resp.Nodes[member] = node
	}

	resp.Indexes = make(map[string]*proto.IndexMetadata)
	for indexMetadataItem := range s.metastore.IndexMetadataIter() {
		indexName := indexMetadataItem.Key
		if indexMetadata, ok := indexMetadataItem.Val.(*phalanxmetastore.IndexMetadata); ok {
			indexMappingBytes, err := indexMetadata.IndexMapping.Marshal()
			if err != nil {
				s.logger.Warn(err.Error(), zap.String("index_name", indexName))
				indexMappingBytes = []byte("{}") // Set empty JSON
			}

			indexMeta := &proto.IndexMetadata{
				IndexUri:     indexMetadata.IndexUri,
				IndexLockUri: indexMetadata.IndexLockUri,
				IndexMapping: indexMappingBytes,
				Shards:       make(map[string]*proto.ShardMetadata),
			}

			for shardMetadataItem := range indexMetadata.ShardMetadataIter() {
				shardName := shardMetadataItem.Key
				if shardMetadata, ok := shardMetadataItem.Val.(*phalanxmetastore.ShardMetadata); ok {
					indexMeta.Shards[shardName] = &proto.ShardMetadata{
						ShardUri:     shardMetadata.ShardUri,
						ShardLockUri: shardMetadata.ShardLockUri,
					}
				} else {
					s.logger.Warn("shard metadata type error", zap.Any("index_name", indexName), zap.Any("shard_name", shardName))
				}
			}
		} else {
			s.logger.Warn("index metadata type error", zap.Any("index_name", indexName))
		}
	}

	// indexer assignment
	indexerAssingmentBytes, err := json.Marshal(s.indexerAssignment)
	if err != nil {
		return nil, err
	}
	resp.IndexerAssignment = indexerAssingmentBytes

	// searcher assignment
	searcherAssingmentBytes, err := json.Marshal(s.searcherAssignment)
	if err != nil {
		return nil, err
	}
	resp.SearcherAssignment = searcherAssingmentBytes

	return resp, nil
}

func (s *IndexService) CreateIndex(ctx context.Context, req *proto.CreateIndexRequest) (*proto.CreateIndexResponse, error) {
	// Check if the index has already been opened.
	if s.metastore.IndexMetadataExists(req.IndexName) {
		err := errors.ErrIndexMetadataAlreadyExists
		s.logger.Error(err.Error(), zap.String("index_name", req.IndexName))
		return nil, err
	}

	// Load the index mapping.
	var indexMapping mapping.IndexMapping
	if len(req.IndexMapping) > 0 {
		if err := json.Unmarshal(req.IndexMapping, &indexMapping); err != nil {
			s.logger.Error(err.Error())
			return nil, err
		}
	}

	// Load the defalt analyzer.
	var defaultAnalyzer analyzer.AnalyzerSetting
	if len(req.DefaultAnalyzer) > 0 {
		if err := json.Unmarshal(req.DefaultAnalyzer, &defaultAnalyzer); err != nil {
			s.logger.Error(err.Error())
			return nil, err
		}
	}

	// Make the index metadata.
	indexMetadata := phalanxmetastore.NewIndexMetadata()
	indexMetadata.IndexName = req.IndexName
	indexMetadata.IndexUri = req.IndexUri
	indexMetadata.IndexLockUri = req.LockUri
	indexMetadata.IndexMapping = indexMapping
	indexMetadata.IndexMappingVersion = time.Now().UTC().UnixNano()
	indexMetadata.DefaultSearchField = req.DefaultSearchField
	indexMetadata.DefaultAnalyzer = defaultAnalyzer

	// Make shards
	numShards := req.NumShards
	if numShards == 0 {
		numShards = 1
	}
	for i := uint32(0); i < numShards; i++ {
		// Make the shard metadata.
		shardName := generateShardName()

		// If the index lock is omitted, the shard lock is also omitted.
		shardLockUri := ""
		if req.LockUri != "" {
			// Parse lock URI.
			lu, err := url.Parse(req.LockUri)
			if err != nil {
				return nil, err
			}
			lu.Path = path.Join(lu.Path, shardName)

			shardLockUri = lu.String()
		}

		// Parse index URI.
		iu, err := url.Parse(req.IndexUri)
		if err != nil {
			return nil, err
		}
		iu.Path = path.Join(iu.Path, shardName)

		shardMetadata := &phalanxmetastore.ShardMetadata{
			ShardName:    shardName,
			ShardUri:     iu.String(),
			ShardLockUri: shardLockUri,
		}

		// Set the shard metadata to the index metadata.
		indexMetadata.SetShardMetadata(shardName, shardMetadata)
	}

	if err := s.metastore.SetIndexMetadata(req.IndexName, indexMetadata); err != nil {
		s.logger.Error(err.Error(), zap.String("index_name", req.IndexName))
		return nil, err
	}

	return &proto.CreateIndexResponse{}, nil
}

func (s *IndexService) DeleteIndex(ctx context.Context, req *proto.DeleteIndexRequest) (*proto.DeleteIndexResponse, error) {
	if !s.metastore.IndexMetadataExists(req.IndexName) {
		err := errors.ErrIndexMetadataDoesNotExist
		s.logger.Error(err.Error(), zap.String("index_name", req.IndexName))
		return nil, err
	}

	indexMetadata := s.metastore.GetIndexMetadata(req.IndexName)
	if indexMetadata == nil {
		err := errors.ErrIndexMetadataDoesNotExist
		s.logger.Error(err.Error(), zap.String("index_name", req.IndexName))
		return nil, err
	}

	for shardMetadataItem := range indexMetadata.ShardMetadataIter() {
		if shardMetadata, ok := shardMetadataItem.Val.(*phalanxmetastore.ShardMetadata); ok {
			// Check shard directory existence.
			if exists, err := directory.DirectoryExists(shardMetadata.ShardUri); err != nil {
				s.logger.Warn(err.Error(), zap.String("shard_uri", shardMetadata.ShardUri))
			} else {
				if exists {
					// Delete shard directory.
					if err := directory.DeleteDirectory(shardMetadata.ShardUri); err != nil {
						s.logger.Warn(err.Error(), zap.String("shard_uri", shardMetadata.ShardUri))
					}
				} else {
					err := errors.ErrIndexDirectoryDoesNotExist
					s.logger.Warn(err.Error(), zap.String("shard_uri", shardMetadata.ShardUri))
				}
			}
		} else {
			s.logger.Warn("shard metadata type error", zap.Any("index_name", req.IndexName), zap.Any("shard_name", shardMetadataItem.Key))
		}
	}

	// Delete index metadata.
	if err := s.metastore.DeleteIndexMetadata(req.IndexName); err != nil {
		s.logger.Error(err.Error(), zap.String("index_name", req.IndexName))
		return nil, err
	}

	return &proto.DeleteIndexResponse{}, nil
}

func (s *IndexService) AddDocuments(ctx context.Context, req *proto.AddDocumentsRequest) (*proto.AddDocumentsResponse, error) {
	if !s.metastore.IndexMetadataExists(req.IndexName) {
		err := errors.ErrIndexMetadataDoesNotExist
		s.logger.Error(err.Error(), zap.String("index_name", req.IndexName))
		return nil, err
	}

	isRootRequest := req.ShardName == ""

	// Assign shards to nodes.
	assignedNodes := make(map[string][]string)
	if isRootRequest {
		for shardName, nodeName := range s.indexerAssignment[req.IndexName] {
			if _, ok := assignedNodes[nodeName]; !ok {
				assignedNodes[nodeName] = []string{}
			}
			assignedNodes[nodeName] = append(assignedNodes[nodeName], shardName)
		}
	} else {
		assignedNodes[s.cluster.LocalNodeName()] = []string{req.ShardName}
	}

	// Assign documents.
	addDocumentsRequests := make(map[string]*proto.AddDocumentsRequest)
	if isRootRequest {
		for _, doc := range req.Documents {
			shardName := s.metastore.GetResponsibleShard(req.IndexName, doc.Id)
			if _, ok := addDocumentsRequests[shardName]; !ok {
				addDocumentsRequests[shardName] = &proto.AddDocumentsRequest{
					IndexName: req.IndexName,
					ShardName: shardName,
					Documents: make([]*proto.Document, 0),
				}
			}
			addDocumentsRequests[shardName].Documents = append(addDocumentsRequests[shardName].Documents, doc)
		}
	} else {
		addDocumentsRequests[req.ShardName] = req
	}

	type addDocumentsResponse struct {
		indexName string
		shardName string
		err       error
	}

	responsesChan := make(chan addDocumentsResponse, s.metastore.NumShards(req.IndexName))

	baseCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	eg, ctx := errgroup.WithContext(baseCtx)

	for nodeName, shardNames := range assignedNodes {
		for _, shardName := range shardNames {
			request, ok := addDocumentsRequests[shardName]
			if !ok {
				err := fmt.Errorf("failed to get add documents request from add documents requests map")
				s.logger.Warn(err.Error(), zap.String("index_name", req.IndexName), zap.String("shard_name", shardName))
				continue
			}

			nodeName := nodeName

			eg.Go(func() error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					if nodeName == s.cluster.LocalNodeName() {
						s.logger.Debug("adding documents", zap.String("node_name", nodeName), zap.String("index_name", request.IndexName), zap.String("shard_name", request.ShardName))

						// Make batch.
						batch := bluge.NewBatch()
						for _, doc := range request.Documents {
							// Get mapping.
							indexMapping, err := s.metastore.GetMapping(request.IndexName)
							if err != nil {
								s.logger.Error(err.Error(), zap.String("index_name", request.IndexName))
								return err
							}

							// Create bluge document.
							blugeDoc, err := indexMapping.MakeDocument(doc)
							if err != nil {
								s.logger.Error(err.Error(), zap.String("index_name", request.IndexName), zap.String("shard_name", request.ShardName), zap.Any("doc", doc))
								return err
							}

							batch.Update(blugeDoc.ID(), blugeDoc)
						}

						// Get index writer.
						writer, err := s.indexWriters.Get(request.IndexName, request.ShardName)
						if err != nil {
							s.logger.Error(err.Error(), zap.String("index_name", request.IndexName), zap.String("shard_name", request.ShardName))
							responsesChan <- addDocumentsResponse{
								indexName: request.IndexName,
								shardName: request.ShardName,
								err:       err,
							}
							return err
						}

						// Execute the batch.
						if err := writer.Batch(batch); err != nil {
							s.logger.Error(err.Error(), zap.String("index_name", request.IndexName), zap.String("shard_name", request.ShardName))
							responsesChan <- addDocumentsResponse{
								indexName: request.IndexName,
								shardName: request.ShardName,
								err:       err,
							}
							return err
						}

						// Update shard version in shard metadata.
						if err := s.metastore.TouchShardMetadata(request.IndexName, request.ShardName); err != nil {
							s.logger.Error(err.Error(), zap.String("index_name", request.IndexName), zap.String("shard_name", request.ShardName))
							responsesChan <- addDocumentsResponse{
								indexName: request.IndexName,
								shardName: request.ShardName,
								err:       err,
							}
							return err
						}
					} else {
						metadata, err := s.cluster.NodeMetadata(nodeName)
						if err != nil {
							s.logger.Error(err.Error(), zap.String("node_name", nodeName))
							responsesChan <- addDocumentsResponse{
								indexName: request.IndexName,
								shardName: request.ShardName,
								err:       err,
							}
							return err
						}

						nodeAddr, err := s.cluster.NodeAddress(nodeName)
						if err != nil {
							s.logger.Error(err.Error(), zap.String("node_name", nodeName))
							responsesChan <- addDocumentsResponse{
								indexName: request.IndexName,
								shardName: request.ShardName,
								err:       err,
							}
							return err
						}

						grpcAddress := fmt.Sprintf("%s:%d", nodeAddr, metadata.GrpcPort)
						client, ok := s.clients[grpcAddress]
						if !ok {
							err := errors.ErrNodeDoesNotFound
							s.logger.Error(err.Error(), zap.String("node_name", nodeName), zap.String("grpc_address", grpcAddress))
							responsesChan <- addDocumentsResponse{
								indexName: request.IndexName,
								shardName: request.ShardName,
								err:       err,
							}
							return err
						}

						_, err = client.AddDocuments(ctx, request)
						if err != nil {
							s.logger.Error(err.Error(), zap.String("node_name", nodeName), zap.String("address", grpcAddress), zap.String("index_name", request.IndexName), zap.String("shard_name", request.ShardName))
							responsesChan <- addDocumentsResponse{
								indexName: request.IndexName,
								shardName: request.ShardName,
								err:       err,
							}
							return err
						}
					}

					// Update successfull.
					responsesChan <- addDocumentsResponse{
						indexName: request.IndexName,
						shardName: request.ShardName,
						err:       nil,
					}
					return nil
				}

			})
		}
	}

	if err := eg.Wait(); err != nil {
		s.logger.Error(err.Error(), zap.String("index_name", req.IndexName))
		return nil, err
	}
	close(responsesChan)

	for response := range responsesChan {
		if response.err != nil {
			s.logger.Error(response.err.Error(), zap.String("index_name", response.indexName), zap.String("shard_name", response.shardName))
			return nil, response.err
		}
	}

	return &proto.AddDocumentsResponse{}, nil
}

func (s *IndexService) DeleteDocuments(ctx context.Context, req *proto.DeleteDocumentsRequest) (*proto.DeleteDocumentsResponse, error) {
	if !s.metastore.IndexMetadataExists(req.IndexName) {
		err := errors.ErrIndexMetadataDoesNotExist
		s.logger.Error(err.Error(), zap.String("index_name", req.IndexName))
		return nil, err
	}

	isRootRequest := req.ShardName == ""

	// Assign shards to nodes.
	assignedNodes := make(map[string][]string)
	if isRootRequest {
		for shardName, nodeName := range s.indexerAssignment[req.IndexName] {
			if _, ok := assignedNodes[nodeName]; !ok {
				assignedNodes[nodeName] = []string{}
			}
			assignedNodes[nodeName] = append(assignedNodes[nodeName], shardName)
		}
	} else {
		assignedNodes[s.cluster.LocalNodeName()] = []string{req.ShardName}
	}

	type deleteDocumentsResponse struct {
		indexName string
		shardName string
		err       error
	}
	responsesChan := make(chan deleteDocumentsResponse, s.metastore.NumShards(req.IndexName))

	baseCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	eg, ctx := errgroup.WithContext(baseCtx)

	for nodeName, shardNames := range assignedNodes {
		for _, shardName := range shardNames {
			request := &proto.DeleteDocumentsRequest{}
			copier.Copy(request, req)
			request.ShardName = shardName

			nodeName := nodeName

			eg.Go(func() error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					if nodeName == s.cluster.LocalNodeName() {
						s.logger.Debug("deleting documents", zap.String("node_name", nodeName), zap.String("index_name", request.IndexName), zap.String("shard_name", request.ShardName))

						batch := bluge.NewBatch()
						for _, id := range request.Ids {
							// Add a document ID for deletion to the batch.
							batch.Delete(bluge.Identifier(id))
						}

						// Get index writer.
						writer, err := s.indexWriters.Get(request.IndexName, request.ShardName)
						if err != nil {
							s.logger.Error(err.Error(), zap.String("index_name", request.IndexName), zap.String("shard_name", request.ShardName))
							responsesChan <- deleteDocumentsResponse{
								indexName: request.IndexName,
								shardName: request.ShardName,
								err:       err,
							}
							return nil
						}

						// Execute the batch.
						if err := writer.Batch(batch); err != nil {
							s.logger.Error(err.Error(), zap.String("index_name", request.IndexName), zap.String("shard_name", request.ShardName))
							responsesChan <- deleteDocumentsResponse{
								indexName: request.IndexName,
								shardName: request.ShardName,
								err:       err,
							}
							return nil
						}

						// Update shard version in shard metadata.
						if err := s.metastore.TouchShardMetadata(request.IndexName, request.ShardName); err != nil {
							s.logger.Error(err.Error(), zap.String("index_name", request.IndexName), zap.String("shard_name", request.ShardName))
							responsesChan <- deleteDocumentsResponse{
								indexName: request.IndexName,
								shardName: request.ShardName,
								err:       err,
							}
							return nil
						}
					} else {
						metadata, err := s.cluster.NodeMetadata(nodeName)
						if err != nil {
							s.logger.Error(err.Error(), zap.String("node_name", nodeName))
							responsesChan <- deleteDocumentsResponse{
								indexName: request.IndexName,
								shardName: request.ShardName,
								err:       err,
							}
							return err
						}

						nodeAddr, err := s.cluster.NodeAddress(nodeName)
						if err != nil {
							s.logger.Error(err.Error(), zap.String("node_name", nodeName))
							responsesChan <- deleteDocumentsResponse{
								indexName: request.IndexName,
								shardName: request.ShardName,
								err:       err,
							}
							return err
						}
						grpcAddress := fmt.Sprintf("%s:%d", nodeAddr, metadata.GrpcPort)
						client, ok := s.clients[grpcAddress]
						if !ok {
							err := errors.ErrNodeDoesNotFound
							s.logger.Error(err.Error(), zap.String("node_name", nodeName), zap.String("grpc_address", grpcAddress))
							responsesChan <- deleteDocumentsResponse{
								indexName: request.IndexName,
								shardName: request.ShardName,
								err:       err,
							}
							return err
						}

						_, err = client.DeleteDocuments(ctx, request)
						if err != nil {
							s.logger.Error(err.Error(), zap.String("node_name", nodeName), zap.String("address", grpcAddress), zap.String("index_name", request.IndexName), zap.String("shard_name", request.ShardName))
							responsesChan <- deleteDocumentsResponse{
								indexName: request.IndexName,
								shardName: request.ShardName,
								err:       err,
							}
							return err
						}
					}

					// Delete successfull.
					responsesChan <- deleteDocumentsResponse{
						indexName: request.IndexName,
						shardName: request.ShardName,
						err:       nil,
					}
					return nil
				}
			})
		}
	}

	if err := eg.Wait(); err != nil {
		s.logger.Error(err.Error(), zap.String("index_name", req.IndexName))
		return nil, err
	}
	close(responsesChan)

	for response := range responsesChan {
		if response.err != nil {
			s.logger.Error(response.err.Error(), zap.String("index_name", response.indexName), zap.String("shard_name", response.shardName))
		}
	}

	return &proto.DeleteDocumentsResponse{}, nil
}

func (s *IndexService) Search(ctx context.Context, req *proto.SearchRequest) (*proto.SearchResponse, error) {
	if !s.metastore.IndexMetadataExists(req.IndexName) {
		err := errors.ErrIndexMetadataDoesNotExist
		s.logger.Error(err.Error(), zap.String("index_name", req.IndexName))
		return nil, err
	}

	isRootRequest := len(req.ShardNames) == 0

	assignedNodes := make(map[string][]string)
	if isRootRequest {
		for shardName, nodeNames := range s.searcherAssignment[req.IndexName] {
			if len(nodeNames) == 0 {
				err := fmt.Errorf("no nodes assigned")
				s.logger.Warn(err.Error(), zap.String("index_name", req.IndexName), zap.String("shard_name", shardName))
				continue
			}
			shuffleNodes(nodeNames)

			if _, ok := assignedNodes[nodeNames[0]]; !ok {
				assignedNodes[nodeNames[0]] = []string{}
			}
			assignedNodes[nodeNames[0]] = append(assignedNodes[nodeNames[0]], shardName)
		}
	} else {
		assignedNodes[s.cluster.LocalNodeName()] = req.ShardNames
	}

	type searchResponse struct {
		nodeName   string
		indexName  string
		shardNames []string
		resp       *proto.SearchResponse
		err        error
	}
	responsesChan := make(chan searchResponse, len(assignedNodes))

	baseCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	eg, ctx := errgroup.WithContext(baseCtx)

	for nodeName, shardNames := range assignedNodes {
		nodeName := nodeName

		request := &proto.SearchRequest{}
		copier.Copy(request, req)
		request.ShardNames = shardNames
		request.Num = request.Start + request.Num
		request.Start = 0

		s.logger.Debug("searching", zap.String("node_name", nodeName), zap.String("index_name", request.IndexName), zap.Strings("shard_names", request.ShardNames))

		eg.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if nodeName == s.cluster.LocalNodeName() {
					resp := &proto.SearchResponse{
						IndexName: request.IndexName,
						Documents: make([]*proto.Document, 0),
						Hits:      0,
					}

					// local node
					readers := make([]*bluge.Reader, 0)
					for _, shardName := range request.ShardNames {
						reader, err := s.indexReaders.Get(request.IndexName, shardName)
						if err != nil {
							s.logger.Warn(err.Error(), zap.String("index_name", request.IndexName), zap.String("shard_name", shardName))
							continue
						}
						readers = append(readers, reader.BlugeReader())
					}

					if len(readers) == 0 {
						err := fmt.Errorf("no index readers are assigned")
						s.logger.Warn(err.Error(), zap.String("index_name", request.IndexName), zap.Strings("shard_names", request.ShardNames))
						responsesChan <- searchResponse{
							nodeName:   nodeName,
							indexName:  request.IndexName,
							shardNames: request.ShardNames,
							resp:       resp,
							err:        nil,
						}
						return nil
					}

					var queryOpts map[string]interface{}
					if err := json.Unmarshal(request.Query.Options, &queryOpts); err != nil {
						s.logger.Error(err.Error(), zap.Any("query", request.Query))
						responsesChan <- searchResponse{
							nodeName:   nodeName,
							indexName:  request.IndexName,
							shardNames: request.ShardNames,
							resp:       nil,
							err:        err,
						}
						return err
					}
					query, err := phalanxqueries.NewQuery(request.Query.Type, queryOpts)
					if err != nil {
						s.logger.Error(err.Error(), zap.Any("query", query))
						responsesChan <- searchResponse{
							nodeName:   nodeName,
							indexName:  request.IndexName,
							shardNames: request.ShardNames,
							resp:       nil,
							err:        err,
						}
						return err
					}

					blugeRequest := bluge.NewTopNSearch(int(request.Num), query).
						SetFrom(int(request.Start)).
						WithStandardAggregations().
						ExplainScores().
						IncludeLocations()

					if request.SortBy != "" {
						blugeRequest.SortBy([]string{request.SortBy})
					} else {
						blugeRequest.SortBy([]string{"-_score"})
					}

					// Set aggregations
					aggs, err := phalanxaggregations.NewAggregations(request.Aggregations)
					if err != nil {
						s.logger.Error(err.Error(), zap.String("index_name", request.IndexName))
						responsesChan <- searchResponse{
							nodeName:   nodeName,
							indexName:  request.IndexName,
							shardNames: request.ShardNames,
							resp:       nil,
							err:        err,
						}
						return err
					}
					for name, agg := range aggs {
						blugeRequest.AddAggregation(name, agg)
					}

					docMatchIter, err := bluge.MultiSearch(ctx, blugeRequest, readers...)
					if err != nil {
						s.logger.Error(err.Error(), zap.String("index_name", request.IndexName))
						responsesChan <- searchResponse{
							nodeName:   nodeName,
							indexName:  request.IndexName,
							shardNames: request.ShardNames,
							resp:       nil,
							err:        err,
						}
						return err
					}

					// Get hits.
					resp.Hits = docMatchIter.Aggregations().Count()

					docMatch, err := docMatchIter.Next()
					if err != nil {
						s.logger.Error(err.Error())
						responsesChan <- searchResponse{
							nodeName:   nodeName,
							indexName:  request.IndexName,
							shardNames: request.ShardNames,
							resp:       nil,
							err:        err,
						}
						return err
					}

					indexMapping, err := s.metastore.GetMapping(request.IndexName)
					if err != nil {
						s.logger.Error(err.Error(), zap.String("index_name", request.IndexName))
						responsesChan <- searchResponse{
							nodeName:   nodeName,
							indexName:  request.IndexName,
							shardNames: request.ShardNames,
							resp:       nil,
							err:        err,
						}
						return err
					}

					// Make highlights.

					highlightRequests := make(map[string]*phalanxhighlight.HighlightRequest)
					for fieldName, highlight := range request.Highlights {
						fmt.Println(fieldName, highlight)
						opts := make(map[string]interface{})
						if err := json.Unmarshal(highlight.Highlighter.Options, &opts); err != nil {
							s.logger.Error(err.Error(), zap.String("field_name", fieldName), zap.String("highlighter_type", highlight.Highlighter.Type))
							responsesChan <- searchResponse{
								nodeName:   nodeName,
								indexName:  request.IndexName,
								shardNames: request.ShardNames,
								resp:       nil,
								err:        err,
							}
							return err
						}
						highliter, err := phalanxhighlight.NewHighlighter(highlight.Highlighter.Type, opts)
						if err != nil {
							s.logger.Error(err.Error(), zap.String("field_name", fieldName), zap.String("highlighter_type", highlight.Highlighter.Type))
							responsesChan <- searchResponse{
								nodeName:   nodeName,
								indexName:  request.IndexName,
								shardNames: request.ShardNames,
								resp:       nil,
								err:        err,
							}
							return err
						}

						highlightRequests[fieldName] = &phalanxhighlight.HighlightRequest{
							Highlighter: highliter,
							Num:         int(highlight.Num),
						}
					}

					// Make docs
					for err == nil && docMatch != nil {
						// Load stored fields.
						doc := &proto.Document{}
						fields := make(map[string][]interface{})
						highlights := make(map[string][]string)
						err := docMatch.VisitStoredFields(func(field string, value []byte) bool {
							switch field {
							case mapping.IdFieldName:
								doc.Id = string(value)
							case mapping.TimestampFieldName:
								timestamp, err := bluge.DecodeDateTime(value)
								if err != nil {
									s.logger.Error(err.Error(), zap.String("index_name", req.IndexName), zap.Any("field", field))
								}
								doc.Timestamp = timestamp.UTC().UnixNano()
							default:
								exists := false
								for _, reqField := range request.Fields {
									if wildcard.Match(reqField, field) {
										exists = true
										break
									}
								}
								if exists {
									// decode field value
									fieldType, err := indexMapping.GetFieldType(field)
									if err != nil {
										s.logger.Error(err.Error(), zap.String("index_name", req.IndexName), zap.String("field_name", field))
										return true
									}

									if ok := fields[field]; ok == nil {
										fields[field] = make([]interface{}, 0)
									}

									switch fieldType {
									case mapping.TextField:
										fields[field] = append(fields[field], string(value))
										fo, err := indexMapping.GetFieldOptions(field)
										if err != nil {
											s.logger.Error(err.Error(), zap.String("index_name", req.IndexName), zap.String("field_name", field))
											return true
										}
										if fo&bluge.HighlightMatches != 0 {
											if highlightRequest, ok := highlightRequests[field]; ok {
												if _, ok := highlights[field]; !ok {
													highlights[field] = make([]string, 0)
												}
												highlights[field] = append(highlights[field], highlightRequest.Highlighter.BestFragments(docMatch.Locations[field], value, highlightRequest.Num)...)
											}
										}
									case mapping.NumericField:
										f64Value, err := bluge.DecodeNumericFloat64(value)
										if err != nil {
											s.logger.Error(err.Error(), zap.String("index_name", req.IndexName), zap.Any("field", field))
										}
										fields[field] = append(fields[field], f64Value)
									case mapping.DatetimeField:
										timestamp, err := bluge.DecodeDateTime(value)
										if err != nil {
											s.logger.Error(err.Error(), zap.String("index_name", req.IndexName), zap.Any("field", field))
										}
										fields[field] = append(fields[field], timestamp.Format(time.RFC3339))
									case mapping.GeoPointField:
										lat, lon, err := bluge.DecodeGeoLonLat(value)
										if err != nil {
											s.logger.Error(err.Error(), zap.String("index_name", req.IndexName), zap.Any("field", field))
										}
										fields[field] = append(fields[field], geo.Point{Lat: lat, Lon: lon})
									}
								}
							}
							return true
						})
						if err != nil {
							s.logger.Error(err.Error(), zap.String("index_name", req.IndexName))
							responsesChan <- searchResponse{
								nodeName:   nodeName,
								indexName:  request.IndexName,
								shardNames: request.ShardNames,
								resp:       nil,
								err:        err,
							}
							return err
						}

						// Set doc score.
						doc.Score = docMatch.Score

						// Serialize fields.
						fieldsBytes, err := json.Marshal(fields)
						if err != nil {
							s.logger.Error(err.Error(), zap.String("index_name", req.IndexName), zap.String("doc_id", doc.Id), zap.Any("fields", fields))
							responsesChan <- searchResponse{
								nodeName:   nodeName,
								indexName:  request.IndexName,
								shardNames: request.ShardNames,
								resp:       nil,
								err:        err,
							}
							return err
						}
						doc.Fields = fieldsBytes

						// Serialize highlights.
						highlightsBytes, err := json.Marshal(highlights)
						if err != nil {
							s.logger.Error(err.Error(), zap.String("index_name", req.IndexName), zap.String("doc_id", doc.Id), zap.Any("highlights", highlights))
							responsesChan <- searchResponse{
								nodeName:   nodeName,
								indexName:  request.IndexName,
								shardNames: request.ShardNames,
								resp:       nil,
								err:        err,
							}
							return err
						}
						doc.Highlights = highlightsBytes

						resp.Documents = append(resp.Documents, doc)

						docMatch, err = docMatchIter.Next()
						if err != nil {
							s.logger.Error(err.Error(), zap.String("index_name", req.IndexName))
							responsesChan <- searchResponse{
								nodeName:   nodeName,
								indexName:  request.IndexName,
								shardNames: request.ShardNames,
								resp:       nil,
								err:        err,
							}
							return err
						}
					}

					// Make aggregation responses.
					resp.Aggregations = make(map[string]*proto.AggregationResponse)
					for name, agg := range request.Aggregations {
						switch agg.Type {
						case phalanxaggregations.AggregationType_name[phalanxaggregations.AggregationTypeTerms]:
							fallthrough
						case phalanxaggregations.AggregationType_name[phalanxaggregations.AggregationTypeRange]:
							fallthrough
						case phalanxaggregations.AggregationType_name[phalanxaggregations.AggregationTypeDateRange]:
							buckets := docMatchIter.Aggregations().Buckets(name)
							resp.Aggregations[name] = &proto.AggregationResponse{
								Buckets: make(map[string]float64),
							}
							for _, bucket := range buckets {
								resp.Aggregations[name].Buckets[bucket.Name()] = float64(bucket.Count())
							}
						case phalanxaggregations.AggregationType_name[phalanxaggregations.AggregationTypeSum]:
							fallthrough
						case phalanxaggregations.AggregationType_name[phalanxaggregations.AggregationTypeMin]:
							fallthrough
						case phalanxaggregations.AggregationType_name[phalanxaggregations.AggregationTypeMax]:
							fallthrough
						case phalanxaggregations.AggregationType_name[phalanxaggregations.AggregationTypeAvg]:
							metric := docMatchIter.Aggregations().Metric(name)
							resp.Aggregations[name] = &proto.AggregationResponse{
								Buckets: make(map[string]float64),
							}
							resp.Aggregations[name].Buckets["value"] = metric
						}
					}

					// Search successfull.
					responsesChan <- searchResponse{
						nodeName:   nodeName,
						indexName:  request.IndexName,
						shardNames: request.ShardNames,
						resp:       resp,
						err:        nil,
					}

					return nil
				} else {
					metadata, err := s.cluster.NodeMetadata(nodeName)
					if err != nil {
						s.logger.Error(err.Error(), zap.String("index_name", req.IndexName), zap.String("node_name", nodeName))
						responsesChan <- searchResponse{
							nodeName:   nodeName,
							indexName:  request.IndexName,
							shardNames: request.ShardNames,
							resp:       nil,
							err:        err,
						}
						return err
					}

					nodeAddr, err := s.cluster.NodeAddress(nodeName)
					if err != nil {
						s.logger.Error(err.Error(), zap.String("index_name", req.IndexName), zap.String("node_name", nodeName))
						responsesChan <- searchResponse{
							nodeName:   nodeName,
							indexName:  request.IndexName,
							shardNames: request.ShardNames,
							resp:       nil,
							err:        err,
						}
						return err
					}

					grpcAddr := fmt.Sprintf("%s:%d", nodeAddr, metadata.GrpcPort)
					client, ok := s.clients[grpcAddr]
					if !ok {
						err := errors.ErrNodeDoesNotFound
						s.logger.Error(err.Error(), zap.String("node_name", nodeName), zap.String("grpc_address", grpcAddr))
						responsesChan <- searchResponse{
							nodeName:   nodeName,
							indexName:  request.IndexName,
							shardNames: request.ShardNames,
							resp:       nil,
							err:        err,
						}
						return err
					}

					remoteResp, err := client.Search(ctx, request)
					if err != nil {
						s.logger.Error(err.Error(), zap.String("index_name", req.IndexName))
						responsesChan <- searchResponse{
							nodeName:   nodeName,
							indexName:  request.IndexName,
							shardNames: request.ShardNames,
							resp:       nil,
							err:        err,
						}
						return err
					}

					// Search successfull.
					responsesChan <- searchResponse{
						nodeName:   nodeName,
						indexName:  request.IndexName,
						shardNames: request.ShardNames,
						resp:       remoteResp,
						err:        nil,
					}
					return nil
				}
			}
		})
	}

	if err := eg.Wait(); err != nil {
		s.logger.Error(err.Error(), zap.String("index_name", req.IndexName))
		return nil, err
	}
	close(responsesChan)

	// Merge responses.
	resp := &proto.SearchResponse{}
	resp.Documents = make([]*proto.Document, 0)
	resp.IndexName = req.IndexName
	resp.Aggregations = make(map[string]*proto.AggregationResponse)
	for response := range responsesChan {
		if response.err != nil {
			s.logger.Error(response.err.Error(), zap.String("node_name", response.nodeName), zap.String("index_name", response.indexName), zap.Strings("shard_names", response.shardNames))
		}

		// Merge hits.
		resp.Hits = resp.Hits + response.resp.Hits

		// Merge documents.
		resp.Documents = mergeDocs(req.SortBy, resp.Documents, response.resp.Documents)

		// Merge aggregations.
		for aggName, aggResp := range response.resp.Aggregations {

			aggReq, ok := req.Aggregations[aggName]
			if !ok {
				s.logger.Warn("Aggregation not found", zap.String("agg_name", aggName), zap.String("index_name", req.IndexName))
				continue
			}
			switch aggReq.Type {
			case phalanxaggregations.AggregationType_name[phalanxaggregations.AggregationTypeTerms]:
				fallthrough
			case phalanxaggregations.AggregationType_name[phalanxaggregations.AggregationTypeRange]:
				fallthrough
			case phalanxaggregations.AggregationType_name[phalanxaggregations.AggregationTypeDateRange]:
				fallthrough
			case phalanxaggregations.AggregationType_name[phalanxaggregations.AggregationTypeSum]:
				if _, ok := resp.Aggregations[aggName]; ok {
					for bucketName, bucketCount := range aggResp.Buckets {
						if _, ok := resp.Aggregations[aggName].Buckets[bucketName]; !ok {
							resp.Aggregations[aggName].Buckets[bucketName] = 0.0
						}
						resp.Aggregations[aggName].Buckets[bucketName] = resp.Aggregations[aggName].Buckets[bucketName] + bucketCount
					}
				} else {
					resp.Aggregations[aggName] = aggResp
				}
			case phalanxaggregations.AggregationType_name[phalanxaggregations.AggregationTypeMin]:
				if _, ok := resp.Aggregations[aggName]; ok {
					if _, ok := resp.Aggregations[aggName].Buckets["value"]; !ok {
						resp.Aggregations[aggName].Buckets["value"] = aggResp.Buckets["value"]
					} else {
						if resp.Aggregations[aggName].Buckets["value"] > aggResp.Buckets["value"] {
							resp.Aggregations[aggName].Buckets["value"] = aggResp.Buckets["value"]
						}
					}
				} else {
					resp.Aggregations[aggName] = aggResp
				}
			case phalanxaggregations.AggregationType_name[phalanxaggregations.AggregationTypeMax]:
				if _, ok := resp.Aggregations[aggName]; ok {
					if _, ok := resp.Aggregations[aggName].Buckets["value"]; !ok {
						resp.Aggregations[aggName].Buckets["value"] = aggResp.Buckets["value"]
					} else {
						if resp.Aggregations[aggName].Buckets["value"] < aggResp.Buckets["value"] {
							resp.Aggregations[aggName].Buckets["value"] = aggResp.Buckets["value"]
						}
					}
				} else {
					resp.Aggregations[aggName] = aggResp
				}
			case phalanxaggregations.AggregationType_name[phalanxaggregations.AggregationTypeAvg]:
				if _, ok := resp.Aggregations[aggName]; ok {
					if _, ok := resp.Aggregations[aggName].Buckets["value"]; !ok {
						resp.Aggregations[aggName].Buckets["value"] = aggResp.Buckets["value"]
					} else {
						resp.Aggregations[aggName].Buckets["value"] = (resp.Aggregations[aggName].Buckets["value"] + aggResp.Buckets["value"]) / float64(2.0)
					}
				} else {
					resp.Aggregations[aggName] = aggResp
				}
			}

		}
	}

	// Extract the specified range of documents.
	if int(req.Start+req.Num) > len(resp.Documents) {
		resp.Documents = resp.Documents[req.Start:]
	} else {
		resp.Documents = resp.Documents[req.Start : req.Start+req.Num]
	}

	// Extract top n aggregations.
	for aggName, aggResp := range resp.Aggregations {
		aggReq := req.Aggregations[aggName]

		buckets := phalanxaggregations.SortByCount(aggResp.Buckets)

		opts := make(map[string]interface{})
		if err := json.Unmarshal(aggReq.Options, &opts); err != nil {
			s.logger.Error(err.Error(), zap.String("aggregation_name", aggName))
			return nil, err
		}

		if sizeValue, ok := opts["size"]; ok {
			if size, ok := sizeValue.(float64); ok {
				if len(buckets) > int(size) {
					buckets = buckets[:int(size)]
				}
			}
		}

		newBuckets := make(map[string]float64)
		for _, bucket := range buckets {
			newBuckets[bucket.Name] = bucket.Count
		}

		resp.Aggregations[aggName].Buckets = newBuckets
	}

	return resp, nil
}

type sortOrder int

const (
	sortOrderAsc sortOrder = iota
	sortOrderDesc
)

func mergeDocs(sortBy string, docs1 []*proto.Document, docs2 []*proto.Document) []*proto.Document {
	if len(docs1) == 0 {
		return docs2
	}

	if len(docs2) == 0 {
		return docs1
	}

	order := sortOrderAsc
	field := sortBy
	if strings.HasPrefix(sortBy, "-") {
		order = sortOrderDesc
		field = sortBy[1:]
	}

	retDocs := make([]*proto.Document, 0)

	var sortValue1 float64
	var sortValue2 float64

	if field == mapping.ScoreFieldName {
		sortValue1 = docs1[0].Score
		sortValue2 = docs2[0].Score
	} else {
		fields1 := make(map[string]interface{})
		json.Unmarshal(docs1[0].Fields, &fields1)

		fields2 := make(map[string]interface{})
		json.Unmarshal(docs2[0].Fields, &fields2)

		var ok bool
		sortValue1, ok = fields1[field].(float64)
		if !ok {
			sortValue1 = 0.0
		}
		sortValue2, ok = fields2[field].(float64)
		if !ok {
			sortValue2 = 0.0
		}
	}

	for len(docs1) > 0 && len(docs2) > 0 {
		// Add document with high scores to the list.
		var doc *proto.Document
		if order == sortOrderDesc {
			if sortValue1 > sortValue2 {
				doc, docs1 = docs1[0], docs1[1:]
			} else {
				doc, docs2 = docs2[0], docs2[1:]
			}
		} else {
			if sortValue1 < sortValue2 {
				doc, docs1 = docs1[0], docs1[1:]
			} else {
				doc, docs2 = docs2[0], docs2[1:]
			}
		}
		retDocs = append(retDocs, doc)
	}

	// Append the remaining list to the end.
	retDocs = append(retDocs, docs1...)
	retDocs = append(retDocs, docs2...)

	return retDocs
}
