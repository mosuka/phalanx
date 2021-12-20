package membership

import (
	"fmt"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/mosuka/phalanx/errors"
	"github.com/mosuka/rendezvous"
	"github.com/thanhpk/randstr"
	"go.uber.org/zap"
)

const (
	nodeNamePrefix = "node-"
)

func generateNodeName() string {
	return fmt.Sprintf("%s%s", nodeNamePrefix, randstr.String(8))
}

type Cluster struct {
	memberList           *memberlist.Memberlist
	nodeEventDeliegate   *NodeEventDelegate
	nodeMetadataDelegate *NodeMetadataDelegate
	logger               *zap.Logger
	clusterEvents        chan ClusterEvent
	stopWatching         chan bool
	isSeedNode           bool
	indexerHash          *rendezvous.Ring
	searcherHash         *rendezvous.Ring
}

func NewCluster(host string, bindPort int, nodeMetadata NodeMetadata, isSeedNode bool, logger *zap.Logger) (*Cluster, error) {
	clusterLogger := logger.Named("cluster")

	nodeEventDeliegate := NewNodeEventDelegate(clusterLogger)
	nodeMetadataDelegate := NewNodeMetadataDelegate(nodeMetadata, clusterLogger)

	config := memberlist.DefaultLocalConfig()
	config.Name = generateNodeName()
	config.BindAddr = host
	config.BindPort = bindPort
	config.AdvertiseAddr = host
	config.AdvertisePort = bindPort
	config.Events = nodeEventDeliegate
	config.Delegate = nodeMetadataDelegate

	memberList, err := memberlist.Create(config)
	if err != nil {
		clusterLogger.Error("Failed to create member list", zap.Error(err), zap.Any("config", config))
		return nil, err
	}
	// members.LocalNode().Meta, err = nodeMetadata.Bytes()
	// if err != nil {
	// 	nodeLogger.Error("Failed to set node metadata", zap.Error(err))
	// }
	// members.UpdateNode(10 * time.Second)

	return &Cluster{
		memberList:           memberList,
		nodeEventDeliegate:   nodeEventDeliegate,
		nodeMetadataDelegate: nodeMetadataDelegate,
		logger:               clusterLogger,
		clusterEvents:        make(chan ClusterEvent, 10),
		stopWatching:         make(chan bool),
		isSeedNode:           isSeedNode,
		indexerHash:          rendezvous.New(),
		searcherHash:         rendezvous.New(),
	}, nil
}

func (n *Cluster) Join(seeds []string) (int, error) {
	return n.memberList.Join(seeds)
}

func (n *Cluster) Leave(timeout time.Duration) error {
	return n.memberList.Leave(timeout)
}

func (n *Cluster) LocalNodeName() string {
	return n.memberList.LocalNode().Name
}

func (n *Cluster) LocalNodeMetadata() (*NodeMetadata, error) {
	return NewNodeMetadataWithBytes(n.memberList.LocalNode().Meta)
}

func (n *Cluster) NodeMetadata(nodeName string) (*NodeMetadata, error) {
	nodes := n.memberList.Members()
	for _, node := range nodes {
		if node.Name == nodeName {
			return NewNodeMetadataWithBytes(node.Meta)
		}
	}

	return nil, errors.ErrNodeDoesNotFound
}

func (n *Cluster) NodeAddress(nodeName string) (string, error) {
	nodes := n.memberList.Members()
	for _, node := range nodes {
		if node.Name == nodeName {
			return node.Addr.String(), nil
		}
	}

	return "", errors.ErrNodeDoesNotFound
}

func (n *Cluster) NodePort(nodeName string) (uint16, error) {
	nodes := n.memberList.Members()
	for _, node := range nodes {
		if node.Name == nodeName {
			return node.Port, nil
		}
	}

	return 0, errors.ErrNodeDoesNotFound
}

func (n *Cluster) NodeState(nodeName string) (NodeState, error) {
	nodes := n.memberList.Members()
	for _, node := range nodes {
		if node.Name == nodeName {
			return makeNodeState(node.State), nil
		}
	}

	return NodeStateUnknown, errors.ErrNodeDoesNotFound
}

func (n *Cluster) IsSeedNode() bool {
	return n.isSeedNode
}

func (n *Cluster) IsIndexer() bool {
	metadata, err := n.LocalNodeMetadata()
	if err != nil {
		return false
	}
	for _, role := range metadata.Roles {
		if role == NodeRoleIndexer {
			return true
		}
	}
	return false
}

func (n *Cluster) IsSearcher() bool {
	metadata, err := n.LocalNodeMetadata()
	if err != nil {
		return false
	}
	for _, role := range metadata.Roles {
		if role == NodeRoleSearcher {
			return true
		}
	}
	return false
}

func (n *Cluster) Nodes() []string {
	members := make([]string, 0)
	for _, member := range n.memberList.Members() {
		members = append(members, member.Name)
	}
	return members
}

func (n *Cluster) ClusterEvents() <-chan ClusterEvent {
	return n.clusterEvents
}

func (n *Cluster) Start() error {
	go func() {
		for {
			select {
			case cancel := <-n.stopWatching:
				if cancel {
					return
				}
			case nodeEvent := <-n.nodeEventDeliegate.NodeEvents:
				n.logger.Info("Received node event", zap.Any("nodeEvent", nodeEvent))

				clusterEvent := ClusterEvent{
					NodeEvent: nodeEvent,
					Members:   n.Nodes(),
				}

				switch nodeEvent.Type {
				case NodeEventTypeJoin:
					if nodeEvent.Meta.IsIndexer() {
						if !n.indexerHash.Contains(nodeEvent.Node) {
							n.indexerHash.AddWithWeight(nodeEvent.Node, 1.0)
						}
					}

					if nodeEvent.Meta.IsSearcher() {
						if !n.searcherHash.Contains(nodeEvent.Node) {
							n.searcherHash.AddWithWeight(nodeEvent.Node, 1.0)
						}
					}
				case NodeEventTypeUpdate:
					if nodeEvent.Meta.IsIndexer() {
						if !n.indexerHash.Contains(nodeEvent.Node) {
							n.indexerHash.AddWithWeight(nodeEvent.Node, 1.0)
						}
					}

					if nodeEvent.Meta.IsSearcher() {
						if !n.searcherHash.Contains(nodeEvent.Node) {
							n.searcherHash.AddWithWeight(nodeEvent.Node, 1.0)
						}
					}
				case NodeEventTypeLeave:
					if nodeEvent.Meta.IsIndexer() {
						if n.indexerHash.Contains(nodeEvent.Node) {
							n.indexerHash.Remove(nodeEvent.Node)
						}
					}

					if nodeEvent.Meta.IsSearcher() {
						if n.searcherHash.Contains(nodeEvent.Node) {
							n.searcherHash.Remove(nodeEvent.Node)
						}
					}
				}

				n.clusterEvents <- clusterEvent
			}
		}
	}()

	return nil
}

func (n *Cluster) Stop() error {
	n.stopWatching <- true

	return nil
}

func (n *Cluster) LookupIndexer(key string) string {
	return n.indexerHash.Lookup(key)
}

func (n *Cluster) LookupSearchers(key string, numNodes int) []string {
	return n.searcherHash.LookupTopN(key, numNodes)
}
