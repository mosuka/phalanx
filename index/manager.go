package index

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/numeric/geo"
	querystr "github.com/blugelabs/query_string"
	"github.com/hashicorp/memberlist"
	"github.com/jinzhu/copier"
	"github.com/mosuka/phalanx/clients"
	"github.com/mosuka/phalanx/directory"
	"github.com/mosuka/phalanx/errors"
	"github.com/mosuka/phalanx/mapping"
	"github.com/mosuka/phalanx/membership"
	"github.com/mosuka/phalanx/metastore"
	"github.com/mosuka/phalanx/proto"
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

type Manager struct {
	node               *membership.Node
	ms                 *metastore.Metastore
	certificateFile    string
	commonName         string
	logger             *zap.Logger
	indexWriters       *IndexWriters
	indexReaders       *IndexReaders
	stopWatching       chan bool
	indexerAssignment  map[string]map[string]string
	searcherAssignment map[string]map[string][]string
	clients            map[string]*clients.IndexClient
	mutex              sync.RWMutex
}

func NewManager(node *membership.Node, ms *metastore.Metastore, certificateFile string, commonName string, logger *zap.Logger) (*Manager, error) {
	managerLogger := logger.Named("manager")

	return &Manager{
		node:               node,
		ms:                 ms,
		certificateFile:    certificateFile,
		commonName:         commonName,
		logger:             logger,
		indexWriters:       NewIndexWriters(managerLogger),
		indexReaders:       NewIndexReaders(managerLogger),
		stopWatching:       make(chan bool),
		indexerAssignment:  map[string]map[string]string{},
		searcherAssignment: map[string]map[string][]string{},
		clients:            map[string]*clients.IndexClient{},
		mutex:              sync.RWMutex{},
	}, nil
}

func (m *Manager) Start() error {
	// Watch metastore events and cluster events.
	go func() {
		for {
			select {
			case cancel := <-m.stopWatching:
				// check
				if cancel {
					return
				}
			case event := <-m.ms.Events():
				switch {
				case event.Type == metastore.MetastoreEventTypePutIndex:
					m.logger.Info("received metastore event", zap.String("type", metastore.MetastoreEventType_name[event.Type]), zap.String("index", event.Index))
				case event.Type == metastore.MetastoreEventTypeDeleteIndex:
					m.logger.Info("received metastore event", zap.String("type", metastore.MetastoreEventType_name[event.Type]), zap.String("index", event.Index))
				case event.Type == metastore.MetastoreEventTypePutShard:
					m.logger.Info("received metastore event", zap.String("type", metastore.MetastoreEventType_name[event.Type]), zap.String("index", event.Index), zap.String("shard", event.Shard))
					m.assignShardsToNode()
				case event.Type == metastore.MetastoreEventTypeDeleteShard:
					m.logger.Info("received metastore event", zap.String("type", metastore.MetastoreEventType_name[event.Type]), zap.String("index", event.Index), zap.String("shard", event.Shard))
					m.assignShardsToNode()
				}
			case event := <-m.node.Events():
				m.logger.Info("received membership event", zap.String("name", event.Node), zap.String("type", membership.EventType_name[event.Type]))

				switch event.Type {
				case membership.EventTypeJoin:
					if m.node.IsSeedNode() || event.Node != m.node.Name() {
						m.assignShardsToNode()
					}
				case membership.EventTypeUpdate:
					m.assignShardsToNode()
				case membership.EventTypeLeave:
					m.assignShardsToNode()
				}
			}
		}
	}()

	return nil
}

func (m *Manager) Stop() error {
	m.stopWatching <- true

	// Close all index writers.
	if err := m.indexWriters.CloseAll(); err != nil {
		m.logger.Error("failed to close index writers", zap.Error(err))
		return err
	}

	// Close all index readers.
	if err := m.indexReaders.CloseAll(); err != nil {
		m.logger.Error("failed to close index writers", zap.Error(err))
		return err
	}

	// Close all index clients.
	for address, indexClient := range m.clients {
		if err := indexClient.Close(); err != nil {
			m.logger.Error("failed to close index client", zap.Error(err), zap.String("address", address))
			return err
		}
		// m.logger.Debug("closing index client", zap.String("address", address))
	}

	return nil
}

func (m *Manager) assignShardsToNode() error {
	searchReplicationFactor := 3
	indexerAssignment := make(map[string]map[string]string)    // index/shard/node
	searcherAssignment := make(map[string]map[string][]string) // index/shard/nodes

	// Assign shards to indexers and searchers.
	for _, indexName := range m.ms.GetIndexNames() {
		for _, shardName := range m.ms.GetShardNames(indexName) {
			// Assign indexer.
			if _, ok := indexerAssignment[indexName]; !ok {
				indexerAssignment[indexName] = make(map[string]string)
			}
			indexerAssignment[indexName][shardName] = m.node.LookupIndexer(shardName)

			// Assign searchers.
			if _, ok := searcherAssignment[indexName]; !ok {
				searcherAssignment[indexName] = make(map[string][]string)
			}
			searcherAssignment[indexName][shardName] = m.node.LookupSearchers(shardName, searchReplicationFactor)
		}
	}

	m.indexerAssignment = indexerAssignment
	m.searcherAssignment = searcherAssignment
	// fmt.Println("indexerAssignment:")
	// for indexName, shardIndexerAssignment := range indexerAssignment {
	// 	fmt.Println("  index name:", indexName)
	// 	for shardName, nodeName := range shardIndexerAssignment {
	// 		fmt.Println("    shard name:", shardName)
	// 		fmt.Println("      node name:", nodeName)
	// 	}
	// }
	// fmt.Println("searcherAssignment:")
	// for indexName, shardIndexerAssignment := range searcherAssignment {
	// 	fmt.Println("  index name:", indexName)
	// 	for shardName, nodeNames := range shardIndexerAssignment {
	// 		fmt.Println("    shard name:", shardName)
	// 		for _, nodeName := range nodeNames {
	// 			fmt.Println("      node name:", nodeName)
	// 		}
	// 	}
	// }

	// Open the index writers for assigned shards.
	for assignedIndexName, shardAssignment := range m.indexerAssignment {
		for assignedShardName, assignedNodeName := range shardAssignment {
			isAssigned := assignedNodeName == m.node.Name()
			if isAssigned {
				if !m.indexWriters.Contains(assignedIndexName, assignedShardName) {
					indexMetadata, err := m.ms.GetIndexMetadata(assignedIndexName)
					if err != nil {
						m.logger.Warn("failed to get index metadata", zap.String("index_name", assignedIndexName))
						continue
					}

					shardMetadata, err := m.ms.GetShardMetadata(assignedIndexName, assignedShardName)
					if err != nil {
						m.logger.Warn("failed to get shard metadata", zap.String("index_name", assignedIndexName), zap.String("shard_name", assignedShardName))
						continue
					}

					// m.logger.Info("open index writer", zap.String("index_name", assignedIndexName), zap.String("shard_name", assignedShardName))
					if err := m.indexWriters.Open(assignedIndexName, assignedShardName, indexMetadata, shardMetadata); err != nil {
						m.logger.Warn("failed to open index writer", zap.Error(err), zap.String("index_name", assignedIndexName), zap.String("shard_name", assignedShardName))
						continue
					}
				}
			} else {
				if m.indexWriters.Contains(assignedIndexName, assignedShardName) {
					// Close the index writer for the shard assigned to the other node.
					// m.logger.Info("close index writer", zap.String("index_name", assignedIndexName), zap.String("shard_name", assignedShardName))
					if err := m.indexWriters.Close(assignedIndexName, assignedShardName); err != nil {
						m.logger.Warn("failed to close index writer", zap.Error(err), zap.String("index_name", assignedIndexName), zap.String("shard_name", assignedShardName))
						continue
					}
				}
			}
		}
	}

	// Close the index writers for unassigned shards.
	for _, openedIndexName := range m.indexWriters.Indexes() {
		for _, openedShardName := range m.indexWriters.Shards(openedIndexName) {
			assignedNodeName, ok := m.indexerAssignment[openedIndexName][openedShardName]
			if !ok {
				// Close the index writer for shard that doesn't already exist.
				// m.logger.Info("close index writer", zap.String("index_name", openedIndexName), zap.String("shard_name", openedShardName))
				if err := m.indexWriters.Close(openedIndexName, openedShardName); err != nil {
					m.logger.Warn("failed to close index writer", zap.Error(err), zap.String("index_name", openedIndexName), zap.String("shard_name", openedShardName))
					continue
				}
			}

			isAssigned := assignedNodeName == m.node.Name()
			if !isAssigned {
				// Close the index writer for the shard assigned to the other node.
				// m.logger.Info("close index writer", zap.String("index_name", openedIndexName), zap.String("shard_name", openedShardName))
				if err := m.indexWriters.Close(openedIndexName, openedShardName); err != nil {
					m.logger.Warn("failed to close index writer", zap.Error(err), zap.String("index_name", openedIndexName), zap.String("shard_name", openedShardName))
					continue
				}
			}
		}
	}

	// open searchers
	for assignedIndexName, shardAssignment := range m.searcherAssignment {
		for assignedShardName, assignedNodeNames := range shardAssignment {
			isAssigned := false
			for _, assignedNodeName := range assignedNodeNames {
				if assignedNodeName == m.node.Name() {
					isAssigned = true
					break
				}
			}
			if isAssigned {
				indexMetadata, err := m.ms.GetIndexMetadata(assignedIndexName)
				if err != nil {
					m.logger.Warn("failed to get index metadata", zap.String("index_name", assignedIndexName))
					continue
				}

				shardMetadata, err := m.ms.GetShardMetadata(assignedIndexName, assignedShardName)
				if err != nil {
					m.logger.Warn("failed to get shard metadata", zap.Error(err), zap.String("index_name", assignedIndexName), zap.String("shard_name", assignedShardName))
					continue
				}
				if shardMetadata == nil {
					m.logger.Warn("shard metadata already gone", zap.String("index_name", assignedIndexName), zap.String("shard_name", assignedShardName))
					continue
				}

				if m.indexReaders.Contains(assignedIndexName, assignedShardName) {
					if m.indexReaders.Version(assignedIndexName, assignedShardName) != shardMetadata.ShardVersion {
						// Reopen the index reader for the assigned shard.
						// m.logger.Info("reopen index reader", zap.String("index_name", assignedIndexName), zap.String("shard_name", assignedShardName))
						if err := m.indexReaders.Reopen(assignedIndexName, assignedShardName, indexMetadata, shardMetadata); err != nil {
							m.logger.Warn("failed to reopen index reader", zap.Error(err), zap.String("index_name", assignedIndexName), zap.String("shard_name", assignedShardName))
							continue
						}
					}
				} else {
					// Open the index reader for the assigned shard.
					// m.logger.Info("open index reader", zap.String("index_name", assignedIndexName), zap.String("shard_name", assignedShardName))
					if err := m.indexReaders.Open(assignedIndexName, assignedShardName, indexMetadata, shardMetadata); err != nil {
						m.logger.Warn("failed to open index reader", zap.Error(err), zap.String("index_name", assignedIndexName), zap.String("shard_name", assignedShardName))
						continue
					}
				}
			} else {
				// Close the index reader for the shard assigned to the other node.
				if m.indexReaders.Contains(assignedIndexName, assignedShardName) {
					// m.logger.Info("close index reader", zap.String("index_name", assignedIndexName), zap.String("shard_name", assignedShardName))
					if err := m.indexReaders.Close(assignedIndexName, assignedShardName); err != nil {
						m.logger.Warn("failed to close index reader", zap.Error(err), zap.String("index_name", assignedIndexName), zap.String("shard_name", assignedShardName))
						continue
					}
				}
			}
		}
	}

	// Close the index readers for unassigned shards.
	for _, indexName := range m.indexReaders.Indexes() {
		for _, shardName := range m.indexReaders.Shards(indexName) {
			assignedNodenames, ok := m.searcherAssignment[indexName][shardName]
			if !ok {
				// Close the index reader for shard that doesn't already exist.
				// m.logger.Info("close index reader", zap.String("index_name", indexName), zap.String("shard_name", shardName))
				if err := m.indexReaders.Close(indexName, shardName); err != nil {
					m.logger.Warn("failed to close index reader", zap.Error(err), zap.String("index_name", indexName), zap.String("shard_name", shardName))
				}
				continue
			}

			isAssigned := false
			for _, assignedNodeName := range assignedNodenames {
				if assignedNodeName == m.node.Name() {
					isAssigned = true
					break
				}
			}
			if !isAssigned {
				// Close the index reader for the shard assigned to the other node.
				// m.logger.Info("close index reader", zap.String("index_name", indexName), zap.String("shard_name", shardName))
				if err := m.indexReaders.Close(indexName, shardName); err != nil {
					m.logger.Warn("failed to close index reader", zap.Error(err), zap.String("index_name", indexName), zap.String("shard_name", shardName))
				}
				continue
			}
		}
	}

	// open clients
	for _, member := range m.node.Members() {
		if member.Name != m.node.Name() {
			metadata, err := membership.NewNodeMetadataWithBytes(member.Meta)
			if err != nil {
				m.logger.Warn("failed to create node metadata", zap.Error(err), zap.String("node_name", member.Name))
				continue
			}
			grpcAddress := fmt.Sprintf("%s:%d", member.Addr.String(), metadata.GrpcPort)

			if _, ok := m.clients[grpcAddress]; !ok {
				// m.logger.Info("open index client", zap.String("node_name", member.Name), zap.String("address", grpcAddress))
				client, err := clients.NewIndexClientWithTLS(grpcAddress, m.certificateFile, m.commonName)
				if err != nil {
					m.logger.Warn("failed to open index client", zap.Error(err), zap.String("node_name", member.Name))
					continue
				}
				m.mutex.Lock()
				m.clients[grpcAddress] = client
				m.mutex.Unlock()
			}
		}
	}

	// close clients
	for address, client := range m.clients {
		notFound := true
		for _, member := range m.node.Members() {
			metadata, err := membership.NewNodeMetadataWithBytes(member.Meta)
			if err != nil {
				m.logger.Warn("failed to create node metadata", zap.Error(err), zap.String("node_name", member.Name))
				continue
			}
			grpcAddress := fmt.Sprintf("%s:%d", member.Addr.String(), metadata.GrpcPort)
			if address == grpcAddress {
				notFound = false
				break
			}
		}
		if notFound {
			// m.logger.Debug("close index client", zap.String("address", address))
			if err := client.Close(); err != nil {
				m.logger.Warn("failed to close index client", zap.Error(err), zap.String("address", client.Address()))
				continue
			}
			m.mutex.Lock()
			delete(m.clients, address)
			m.mutex.Unlock()
		}
	}

	return nil
}

func (m *Manager) Cluster(req *proto.ClusterRequest) (*proto.ClusterResponse, error) {
	resp := &proto.ClusterResponse{}

	resp.Nodes = make(map[string]*proto.Node)
	for _, member := range m.node.Members() {
		// Deserialize the node metadata
		nodeMetadata, err := membership.NewNodeMetadataWithBytes(member.Meta)
		if err != nil {
			nodeMetadata = &membership.NodeMetadata{}
		}

		nodeRoles := make([]proto.NodeRole, 0)
		for _, role := range nodeMetadata.Roles {
			switch role {
			case membership.NodeRoleIndexer:
				nodeRoles = append(nodeRoles, proto.NodeRole_NODE_ROLE_INDEXER)
			case membership.NodeRoleSearcher:
				nodeRoles = append(nodeRoles, proto.NodeRole_NODE_ROLE_SEARCHER)
			default:
				nodeRoles = append(nodeRoles, proto.NodeRole_NODE_ROLE_UNKNOWN)
			}
		}

		state := proto.NodeState_NODE_STATE_UNKNOWN
		switch member.State {
		case memberlist.StateAlive:
			state = proto.NodeState_NODE_STATE_ALIVE
		case memberlist.StateSuspect:
			state = proto.NodeState_NODE_STATE_SUSPECT
		case memberlist.StateDead:
			state = proto.NodeState_NODE_STATE_DEAD
		case memberlist.StateLeft:
			state = proto.NodeState_NODE_STATE_LEFT
		}

		node := &proto.Node{
			Addr: member.Addr.String(),
			Port: uint32(member.Port),
			Meta: &proto.NodeMeta{
				GrpcPort: uint32(nodeMetadata.GrpcPort),
				HttpPort: uint32(nodeMetadata.HttpPort),
				Roles:    nodeRoles,
			},
			State: state,
		}

		resp.Nodes[member.Name] = node
	}

	resp.Indexes = make(map[string]*proto.IndexMetadata)
	for _, indexName := range m.ms.GetIndexNames() {
		indexMetadata, err := m.ms.GetIndexMetadata(indexName)
		if err != nil {
			m.logger.Warn("failed to get index metadata", zap.Error(err), zap.String("index_name", indexName))
			continue
		}
		indexMeta := &proto.IndexMetadata{
			IndexUri:     indexMetadata.IndexUri,
			IndexLockUri: indexMetadata.IndexLockUri,
			Shards:       make(map[string]*proto.ShardMetadata),
		}

		for _, shardName := range m.ms.GetShardNames(indexName) {
			shardMetadata, err := m.ms.GetShardMetadata(indexName, shardName)
			if err != nil {
				m.logger.Warn("failed to get shard metadata", zap.Error(err), zap.String("index_name", indexName), zap.String("shard_name", shardName))
				continue
			}

			indexMeta.Shards[shardName] = &proto.ShardMetadata{
				ShardUri:     shardMetadata.ShardUri,
				ShardLockUri: shardMetadata.ShardLockUri,
			}
		}

		resp.Indexes[indexName] = indexMeta
	}

	// indexer assignment
	indexerAssingmentBytes, err := json.Marshal(m.indexerAssignment)
	if err != nil {
		return nil, err
	}
	resp.IndexerAssignment = indexerAssingmentBytes

	// searcher assignment
	searcherAssingmentBytes, err := json.Marshal(m.searcherAssignment)
	if err != nil {
		return nil, err
	}
	resp.SearcherAssignment = searcherAssingmentBytes

	return resp, nil
}

func (m *Manager) CreateIndex(req *proto.CreateIndexRequest) (*proto.CreateIndexResponse, error) {
	// Check if the index has already been opened.
	if m.ms.IndexMetadataExists(req.IndexName) {
		err := errors.ErrIndexMetadataAlreadyExists
		m.logger.Error("failed to create index", zap.Error(err), zap.String("index_name", req.IndexName))
		return nil, err
	}

	// Load the index mapping.
	var indexMapping mapping.IndexMapping
	if len(req.IndexMapping) > 0 {
		// Unmarshal from the index mapping in a byte array to a map object.
		if err := json.Unmarshal(req.IndexMapping, &indexMapping); err != nil {
			m.logger.Error("failed to create index", zap.Error(err), zap.String("index_name", req.IndexName))
			return nil, err
		}
	}

	// Make the index metadata.
	indexMetadata := &metastore.IndexMetadata{
		IndexName:           req.IndexName,
		IndexUri:            req.IndexUri,
		IndexLockUri:        req.LockUri,
		IndexMapping:        indexMapping,
		IndexMappingVersion: time.Now().UTC().UnixNano(),
		ShardMetadataMap:    make(map[string]*metastore.ShardMetadata),
	}

	// Make shards
	numShards := req.NumShards
	if numShards == 0 {
		numShards = 1
	}
	for i := uint32(0); i < numShards; i++ {
		// Make the shard metadata.
		shardName := generateShardName()
		shardMetadata := &metastore.ShardMetadata{
			ShardName:    shardName,
			ShardUri:     fmt.Sprintf("%s/%s", req.IndexUri, shardName),
			ShardLockUri: fmt.Sprintf("%s/%s", req.LockUri, shardName),
		}

		// Set the shard metadata to the index metadata.
		indexMetadata.SetShardMetadata(shardName, shardMetadata)
	}

	if err := m.ms.SetIndexMetadata(req.IndexName, indexMetadata); err != nil {
		m.logger.Error("failed to set index metadata", zap.Error(err), zap.String("index_name", req.IndexName))
		return nil, err
	}

	return &proto.CreateIndexResponse{}, nil
}

func (m *Manager) DeleteIndex(req *proto.DeleteIndexRequest) (*proto.DeleteIndexResponse, error) {
	if !m.ms.IndexMetadataExists(req.IndexName) {
		err := errors.ErrIndexMetadataDoesNotExist
		m.logger.Error("failed to delete index", zap.Error(err), zap.String("index_name", req.IndexName))
		return nil, err
	}

	indexMetadata, err := m.ms.GetIndexMetadata(req.IndexName)
	if err != nil {
		m.logger.Error("failed to delete index", zap.Error(err), zap.String("index_name", req.IndexName))
		return nil, err
	}
	for shardName, shardMetadata := range indexMetadata.AllShardMetadata() {
		// Check shard directory existence.
		if exists, err := directory.DirectoryExists(shardMetadata.ShardUri); err != nil {
			m.logger.Warn("failed to check index directory existence", zap.Error(err), zap.String("index_name", req.IndexName), zap.String("shard_name", shardName))
		} else {
			if exists {
				// Delete shard directory.
				if err := directory.DeleteDirectory(shardMetadata.ShardUri); err != nil {
					m.logger.Warn("failed to delete shard directory", zap.Error(err), zap.String("index_name", req.IndexName), zap.String("shard_name", shardName))
				}
			} else {
				err := errors.ErrIndexDirectoryDoesNotExist
				m.logger.Warn("shard directory does not exist", zap.Error(err), zap.String("index_name", req.IndexName), zap.String("shard_name", shardName))
			}
		}

		// Delete shard metadata.
		if err := m.ms.DeleteShardMetadata(req.IndexName, shardName); err != nil {
			m.logger.Error("failed to delete shard metadata", zap.Error(err), zap.String("index_name", req.IndexName), zap.String("shard_name", shardName))
			return nil, err
		}
	}
	// Delete index metadata.
	if err := m.ms.DeleteIndexMetadata(req.IndexName); err != nil {
		m.logger.Error("failed to delete index metadata", zap.Error(err), zap.String("index_name", req.IndexName))
		return nil, err
	}

	return &proto.DeleteIndexResponse{}, nil
}

func (m *Manager) AddDocuments(req *proto.AddDocumentsRequest) (*proto.AddDocumentsResponse, error) {
	if !m.ms.IndexMetadataExists(req.IndexName) {
		err := errors.ErrIndexMetadataDoesNotExist
		m.logger.Error("failed to add documents", zap.Error(err), zap.String("index_name", req.IndexName))
		return nil, err
	}

	isRootRequest := req.ShardName == ""

	// Assign shards to nodes.
	assignedNodes := make(map[string][]string)
	if isRootRequest {
		for shardName, nodeName := range m.indexerAssignment[req.IndexName] {
			if _, ok := assignedNodes[nodeName]; !ok {
				assignedNodes[nodeName] = []string{}
			}
			assignedNodes[nodeName] = append(assignedNodes[nodeName], shardName)
		}
	} else {
		assignedNodes[m.node.Name()] = []string{req.ShardName}
	}

	// Assign documents.
	addDocumentsRequests := make(map[string]*proto.AddDocumentsRequest)
	if isRootRequest {
		for _, document := range req.Documents {
			shardName := m.ms.GetResponsibleShard(req.IndexName, document.Id)
			if _, ok := addDocumentsRequests[shardName]; !ok {
				addDocumentsRequests[shardName] = &proto.AddDocumentsRequest{
					IndexName: req.IndexName,
					ShardName: shardName,
					Documents: make([]*proto.Document, 0),
				}
			}
			addDocumentsRequests[shardName].Documents = append(addDocumentsRequests[shardName].Documents, document)
		}
	} else {
		addDocumentsRequests[req.ShardName] = req
	}

	type addDocumentsResponse struct {
		indexName string
		shardName string
		err       error
	}

	responsesChan := make(chan addDocumentsResponse, m.ms.NumShards(req.IndexName))

	baseCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	eg, ctx := errgroup.WithContext(baseCtx)

	for nodeName, shardNames := range assignedNodes {
		for _, shardName := range shardNames {
			request, ok := addDocumentsRequests[shardName]
			if !ok {
				m.logger.Warn("failed to get add documents request from add documents requests map", zap.String("index_name", req.IndexName), zap.String("shard_name", shardName))
				continue
			}

			nodeName := nodeName

			eg.Go(func() error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					if nodeName == m.node.Name() {
						m.logger.Info("adding documents", zap.String("node_name", nodeName), zap.String("index_name", request.IndexName), zap.String("shard_name", request.ShardName))

						// Make batch.
						batch := bluge.NewBatch()
						for _, document := range request.Documents {
							// Convert JSON string to map.
							var fieldMap map[string]interface{}
							if err := json.Unmarshal(document.Fields, &fieldMap); err != nil {
								err := errors.ErrInvalidDocument
								m.logger.Error("failed to unmarshal document data", zap.Error(err), zap.String("index_name", request.IndexName), zap.String("shard_name", request.ShardName))
								return err
							}
							// Get mapping.
							indexMapping, err := m.ms.GetMapping(request.IndexName)
							if err != nil {
								m.logger.Error("failed to get index mapping", zap.Error(err), zap.String("index_name", request.IndexName))
								return err
							}
							// Create document.
							doc, err := indexMapping.MakeDocument(document.Id, fieldMap)
							if err != nil {
								m.logger.Error("failed to make document", zap.Error(err), zap.String("index_name", request.IndexName), zap.String("shardTname", request.ShardName), zap.String("doc_id", document.Id))
								return err
							}

							batch.Update(bluge.Identifier(document.Id), doc)
						}

						// Get index writer.
						writer, err := m.indexWriters.Get(request.IndexName, request.ShardName)
						if err != nil {
							m.logger.Error("failed to get index writer", zap.Error(err), zap.String("index_name", request.IndexName), zap.String("shard_name", request.ShardName))
							responsesChan <- addDocumentsResponse{
								indexName: request.IndexName,
								shardName: request.ShardName,
								err:       err,
							}
							return err
						}

						// Execute the batch.
						if err := writer.Batch(batch); err != nil {
							m.logger.Error("failed to add documents", zap.Error(err), zap.String("index_name", request.IndexName), zap.String("shard_name", request.ShardName))
							responsesChan <- addDocumentsResponse{
								indexName: request.IndexName,
								shardName: request.ShardName,
								err:       err,
							}
							return err
						}

						// Update shard version in shard metadata.
						if err := m.ms.TouchShardMetadata(request.IndexName, request.ShardName); err != nil {
							m.logger.Error("failed to update shard version", zap.Error(err), zap.String("index_name", request.IndexName), zap.String("shard_name", request.ShardName))
							responsesChan <- addDocumentsResponse{
								indexName: request.IndexName,
								shardName: request.ShardName,
								err:       err,
							}
							return err
						}
					} else {
						// remote node
						node := m.node.Member(nodeName)
						if node == nil {
							err := errors.ErrNodeNotFound
							m.logger.Error("failed to get node", zap.Error(err), zap.String("node_name", nodeName), zap.String("index_name", request.IndexName), zap.String("shard_name", request.ShardName))
							responsesChan <- addDocumentsResponse{
								indexName: request.IndexName,
								shardName: request.ShardName,
								err:       err,
							}
							return err
						}

						metadata, err := membership.NewNodeMetadataWithBytes(node.Meta)
						if err != nil {
							m.logger.Error("failed to unmarshal node metadata", zap.Error(err))
							responsesChan <- addDocumentsResponse{
								indexName: request.IndexName,
								shardName: request.ShardName,
								err:       err,
							}
							return err
						}

						grpcAddress := fmt.Sprintf("%s:%d", node.Addr.String(), metadata.GrpcPort)
						client, ok := m.clients[grpcAddress]
						if !ok {
							err := errors.ErrNodeNotFound
							m.logger.Error("failed to get client", zap.Error(err), zap.String("node_name", nodeName), zap.String("grpc_address", grpcAddress))
							responsesChan <- addDocumentsResponse{
								indexName: request.IndexName,
								shardName: request.ShardName,
								err:       err,
							}
							return err
						}

						_, err = client.AddDocuments(ctx, request)
						if err != nil {
							m.logger.Error("failed to add documents", zap.Error(err), zap.String("node_name", nodeName), zap.String("address", grpcAddress), zap.String("index_name", request.IndexName), zap.String("shard_name", request.ShardName))
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
		m.logger.Error("failed to add documents", zap.Error(err), zap.String("index_name", req.IndexName))
		return nil, err
	}
	close(responsesChan)

	for response := range responsesChan {
		if response.err != nil {
			m.logger.Error("failed to add documents", zap.Error(response.err), zap.String("index_name", response.indexName), zap.String("shard_name", response.shardName))
			return nil, response.err
		}
	}

	return &proto.AddDocumentsResponse{}, nil
}

func (m *Manager) DeleteDocuments(req *proto.DeleteDocumentsRequest) (*proto.DeleteDocumentsResponse, error) {
	if !m.ms.IndexMetadataExists(req.IndexName) {
		err := errors.ErrIndexMetadataDoesNotExist
		m.logger.Error("failed to delete documents", zap.Error(err), zap.String("index_name", req.IndexName))
		return nil, err
	}

	isRootRequest := req.ShardName == ""

	// Assign shards to nodes.
	assignedNodes := make(map[string][]string)
	if isRootRequest {
		for shardName, nodeName := range m.indexerAssignment[req.IndexName] {
			if _, ok := assignedNodes[nodeName]; !ok {
				assignedNodes[nodeName] = []string{}
			}
			assignedNodes[nodeName] = append(assignedNodes[nodeName], shardName)
		}
	} else {
		assignedNodes[m.node.Name()] = []string{req.ShardName}
	}

	type deleteDocumentsResponse struct {
		indexName string
		shardName string
		err       error
	}
	responsesChan := make(chan deleteDocumentsResponse, m.ms.NumShards(req.IndexName))

	baseCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
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
					if nodeName == m.node.Name() {
						m.logger.Debug("deleting documents", zap.String("node_name", nodeName), zap.String("index_name", request.IndexName), zap.String("shard_name", request.ShardName))

						batch := bluge.NewBatch()
						for _, id := range request.Ids {
							// Add a document ID for deletion to the batch.
							batch.Delete(bluge.Identifier(id))
						}

						// Get index writer.
						writer, err := m.indexWriters.Get(request.IndexName, request.ShardName)
						if err != nil {
							m.logger.Error("failed to get index writer", zap.Error(err), zap.String("index_name", request.IndexName), zap.String("shard_name", request.ShardName))
							responsesChan <- deleteDocumentsResponse{
								indexName: request.IndexName,
								shardName: request.ShardName,
								err:       err,
							}
							return nil
						}

						// Execute the batch.
						if err := writer.Batch(batch); err != nil {
							m.logger.Error("failed to delete documents", zap.Error(err), zap.String("index_name", request.IndexName), zap.String("shard_name", request.ShardName))
							responsesChan <- deleteDocumentsResponse{
								indexName: request.IndexName,
								shardName: request.ShardName,
								err:       err,
							}
							return nil
						}

						// Update shard version in shard metadata.
						if err := m.ms.TouchShardMetadata(request.IndexName, request.ShardName); err != nil {
							m.logger.Error("failed to update shard version", zap.Error(err), zap.String("index_name", request.IndexName), zap.String("shard_name", request.ShardName))
							responsesChan <- deleteDocumentsResponse{
								indexName: request.IndexName,
								shardName: request.ShardName,
								err:       err,
							}
							return nil
						}
					} else {
						// remote node
						node := m.node.Member(nodeName)
						if node == nil {
							err := errors.ErrNodeNotFound
							m.logger.Error("failed to get node", zap.Error(err), zap.String("node_name", nodeName), zap.String("index_name", request.IndexName), zap.String("shard_name", request.ShardName))
							responsesChan <- deleteDocumentsResponse{
								indexName: request.IndexName,
								shardName: request.ShardName,
								err:       err,
							}
							return err
						}

						metadata, err := membership.NewNodeMetadataWithBytes(node.Meta)
						if err != nil {
							m.logger.Error("failed to unmarshal node metadata", zap.Error(err))
							responsesChan <- deleteDocumentsResponse{
								indexName: request.IndexName,
								shardName: request.ShardName,
								err:       err,
							}
							return err
						}

						grpcAddress := fmt.Sprintf("%s:%d", node.Addr.String(), metadata.GrpcPort)
						client, ok := m.clients[grpcAddress]
						if !ok {
							err := errors.ErrNodeNotFound
							m.logger.Error("failed to get client", zap.Error(err), zap.String("node_name", nodeName), zap.String("grpc_address", grpcAddress))
							responsesChan <- deleteDocumentsResponse{
								indexName: request.IndexName,
								shardName: request.ShardName,
								err:       err,
							}
							return err
						}

						_, err = client.DeleteDocuments(ctx, request)
						if err != nil {
							m.logger.Error("failed to delete documents", zap.Error(err), zap.String("node_name", nodeName), zap.String("address", grpcAddress), zap.String("index_name", request.IndexName), zap.String("shard_name", request.ShardName))
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
		m.logger.Error("failed to delete documents", zap.Error(err), zap.String("index_name", req.IndexName))
		return nil, err
	}
	close(responsesChan)

	for response := range responsesChan {
		if response.err != nil {
			m.logger.Error("failed to delete documents", zap.Error(response.err), zap.String("index_name", response.indexName), zap.String("shard_name", response.shardName))
		}
	}

	return &proto.DeleteDocumentsResponse{}, nil
}

func (m *Manager) Search(req *proto.SearchRequest) (*proto.SearchResponse, error) {
	if !m.ms.IndexMetadataExists(req.IndexName) {
		err := errors.ErrIndexMetadataDoesNotExist
		m.logger.Error("failed to search documents", zap.Error(err), zap.String("index_name", req.IndexName))
		return nil, err
	}

	isRootRequest := len(req.ShardNames) == 0

	assignedNodes := make(map[string][]string)
	if isRootRequest {
		for shardName, nodeNames := range m.searcherAssignment[req.IndexName] {
			if len(nodeNames) == 0 {
				m.logger.Warn("no nodes assigned", zap.String("index_name", req.IndexName), zap.String("shard_name", shardName))
				continue
			}
			shuffleNodes(nodeNames)

			if _, ok := assignedNodes[nodeNames[0]]; !ok {
				assignedNodes[nodeNames[0]] = []string{}
			}
			assignedNodes[nodeNames[0]] = append(assignedNodes[nodeNames[0]], shardName)
		}
	} else {
		assignedNodes[m.node.Name()] = req.ShardNames
	}

	type searchResponse struct {
		nodeName   string
		indexName  string
		shardNames []string
		resp       *proto.SearchResponse
		err        error
	}
	responsesChan := make(chan searchResponse, len(assignedNodes))

	baseCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	eg, ctx := errgroup.WithContext(baseCtx)

	for nodeName, shardNames := range assignedNodes {
		nodeName := nodeName

		request := &proto.SearchRequest{}
		copier.Copy(request, req)
		request.ShardNames = shardNames
		request.Num = request.Start + request.Num
		request.Start = 0

		m.logger.Debug("searching", zap.String("node_name", nodeName), zap.String("index_name", request.IndexName), zap.Strings("shard_names", request.ShardNames))

		eg.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if nodeName == m.node.Name() {
					resp := &proto.SearchResponse{
						IndexName: request.IndexName,
						Documents: make([]*proto.Document, 0),
						Hits:      0,
					}

					// local node
					readers := make([]*bluge.Reader, 0)
					for _, shardName := range request.ShardNames {
						reader, err := m.indexReaders.Get(request.IndexName, shardName)
						if err != nil {
							m.logger.Warn("failed to get index reader", zap.Error(err), zap.String("index_name", request.IndexName), zap.String("shard_name", shardName))
							continue
						}
						readers = append(readers, reader.BlugeReader())
					}

					if len(readers) == 0 {
						m.logger.Warn("no index readers are assigned", zap.String("index_name", request.IndexName), zap.Strings("shard_names", request.ShardNames))
						responsesChan <- searchResponse{
							nodeName:   nodeName,
							indexName:  request.IndexName,
							shardNames: request.ShardNames,
							resp:       resp,
							err:        nil,
						}
						return nil
					}

					userQuery, err := querystr.ParseQueryString(request.Query, querystr.DefaultOptions())
					if err != nil {
						m.logger.Error("failed to parse query string", zap.Error(err), zap.String("query", request.Query))
						responsesChan <- searchResponse{
							nodeName:   nodeName,
							indexName:  request.IndexName,
							shardNames: request.ShardNames,
							resp:       nil,
							err:        err,
						}
						return err
					}

					query := bluge.NewBooleanQuery().AddMust(userQuery)
					// TODO: add filter queries
					// .AddMust(filters...)
					if request.Boost > 0.0 {
						query.SetBoost(request.Boost)
					}

					blugeRequest := bluge.NewTopNSearch(int(request.Num), query).
						SetFrom(int(request.Start)).
						WithStandardAggregations().
						ExplainScores()
					// TODO: add aggretations
					// request.AddAggregation(name, aggregation)

					docMatchIter, err := bluge.MultiSearch(ctx, blugeRequest, readers...)
					if err != nil {
						m.logger.Error("failed to execute search", zap.Error(err), zap.String("index_name", request.IndexName))
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
						m.logger.Error("failed to get next document match", zap.Error(err))
						responsesChan <- searchResponse{
							nodeName:   nodeName,
							indexName:  request.IndexName,
							shardNames: request.ShardNames,
							resp:       nil,
							err:        err,
						}
						return err
					}

					indexMapping, err := m.ms.GetMapping(request.IndexName)
					if err != nil {
						m.logger.Error("failed to get index mapping", zap.Error(err), zap.String("index_name", request.IndexName))
						responsesChan <- searchResponse{
							nodeName:   nodeName,
							indexName:  request.IndexName,
							shardNames: request.ShardNames,
							resp:       nil,
							err:        err,
						}
						return err
					}

					// Make docs
					for err == nil && docMatch != nil {
						doc := &proto.Document{}

						// Load stored fields.
						// TODO: Filter only the fields that are needed.
						fields := make(map[string]interface{})
						err := docMatch.VisitStoredFields(func(field string, value []byte) bool {
							switch field {
							case mapping.IdFieldName:
								doc.Id = string(value)
							case mapping.TimestampFieldName:
								timestamp, err := bluge.DecodeDateTime(value)
								if err != nil {
									m.logger.Error("failed to decode field value", zap.Error(err), zap.Any("field", field))
								}
								fields[field] = timestamp.Format(time.RFC3339)
							default:
								// decode field value
								fieldType, err := indexMapping.GetFieldType(field)
								if err != nil {
									m.logger.Error("failed to get field type", zap.Error(err), zap.String("index_name", req.IndexName), zap.String("field_name", field))
									return true
								}
								switch fieldType {
								case mapping.TextField:
									fields[field] = string(value)
								case mapping.NumericField:
									f64Value, err := bluge.DecodeNumericFloat64(value)
									if err != nil {
										m.logger.Error("failed to decode numeric field value", zap.Error(err), zap.Any("field", field))
									}
									fields[field] = f64Value
								case mapping.DatetimeField:
									timestamp, err := bluge.DecodeDateTime(value)
									if err != nil {
										m.logger.Error("failed to decode datetime field value", zap.Error(err), zap.Any("field", field))
									}
									fields[field] = timestamp.Format(time.RFC3339)
								case mapping.GeoPointField:
									lat, lon, err := bluge.DecodeGeoLonLat(value)
									if err != nil {
										m.logger.Error("failed to decode geo point field value", zap.Error(err), zap.Any("field", field))
									}
									fields[field] = geo.Point{Lat: lat, Lon: lon}
								}
							}
							return true
						})
						if err != nil {
							m.logger.Error("failed to load stored fields", zap.Error(err))
							responsesChan <- searchResponse{
								nodeName:   nodeName,
								indexName:  request.IndexName,
								shardNames: request.ShardNames,
								resp:       nil,
								err:        err,
							}
							return err
						}

						// Set doc fields.
						fieldsBytes, err := json.Marshal(fields)
						if err != nil {
							m.logger.Error("failed to marshal document", zap.Error(err), zap.Any("doc", fields))
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

						// Set doc score.
						doc.Score = docMatch.Score

						resp.Documents = append(resp.Documents, doc)

						docMatch, err = docMatchIter.Next()
						if err != nil {
							m.logger.Error("failed to move to the next document match", zap.Error(err))
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
					// remote node
					// request.ShardNames = shardNames

					member := m.node.Member(nodeName)
					if member == nil {
						err := errors.ErrNodeNotFound
						m.logger.Error("failed to get member", zap.Error(err), zap.String("node_name", nodeName))
						responsesChan <- searchResponse{
							nodeName:   nodeName,
							indexName:  request.IndexName,
							shardNames: request.ShardNames,
							resp:       nil,
							err:        err,
						}
						return err
					}

					metadata, err := membership.NewNodeMetadataWithBytes(member.Meta)
					if err != nil {
						m.logger.Error("failed to unmarshal node metadata", zap.Error(err))
						responsesChan <- searchResponse{
							nodeName:   nodeName,
							indexName:  request.IndexName,
							shardNames: request.ShardNames,
							resp:       nil,
							err:        err,
						}
						return err
					}

					grpcAddr := fmt.Sprintf("%s:%d", member.Addr, metadata.GrpcPort)
					client, ok := m.clients[grpcAddr]
					if !ok {
						err := errors.ErrNodeNotFound
						m.logger.Error("failed to get client", zap.Error(err), zap.String("node_name", nodeName), zap.String("grpc_address", grpcAddr))
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
						m.logger.Error("failed to search", zap.Error(err), zap.String("node_name", nodeName))
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
		m.logger.Error("failed to search documents", zap.Error(err), zap.String("index_name", req.IndexName))
		return nil, err
	}
	close(responsesChan)

	// Merge responses.
	resp := &proto.SearchResponse{}
	resp.Documents = make([]*proto.Document, 0)
	resp.IndexName = req.IndexName
	for response := range responsesChan {
		if response.err != nil {
			m.logger.Error("failed to add documents", zap.Error(response.err), zap.String("node_name", response.nodeName), zap.String("index_name", response.indexName), zap.Strings("shard_names", response.shardNames))
		}

		resp.Hits = resp.Hits + response.resp.Hits
		resp.Documents = mergeDocs(resp.Documents, response.resp.Documents)
	}

	if int(req.Start+req.Num) > len(resp.Documents) {
		resp.Documents = resp.Documents[req.Start:]
	} else {
		resp.Documents = resp.Documents[req.Start : req.Start+req.Num]
	}

	return resp, nil
}

func mergeDocs(docs1 []*proto.Document, docs2 []*proto.Document) []*proto.Document {
	if len(docs1) == 0 {
		return docs2
	}

	if len(docs2) == 0 {
		return docs1
	}

	retDocs := make([]*proto.Document, 0)

	for len(docs1) > 0 && len(docs2) > 0 {
		// Add document with high scores to the list.
		var doc *proto.Document
		if docs1[0].Score > docs2[0].Score {
			doc, docs1 = docs1[0], docs1[1:]
		} else {
			doc, docs2 = docs2[0], docs2[1:]
		}
		retDocs = append(retDocs, doc)
	}

	// Append the remaining list to the end.
	retDocs = append(retDocs, docs1...)
	retDocs = append(retDocs, docs2...)

	return retDocs
}
