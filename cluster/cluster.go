package cluster

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
		clusterLogger.Error(err.Error(), zap.Any("config", config))
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

func (c *Cluster) Join(seeds []string) (int, error) {
	return c.memberList.Join(seeds)
}

func (c *Cluster) Leave(timeout time.Duration) error {
	return c.memberList.Leave(timeout)
}

func (c *Cluster) LocalNodeName() string {
	return c.memberList.LocalNode().Name
}

func (c *Cluster) LocalNodeMetadata() (*NodeMetadata, error) {
	return NewNodeMetadataWithBytes(c.memberList.LocalNode().Meta)
}

func (c *Cluster) NodeMetadata(nodeName string) (*NodeMetadata, error) {
	nodes := c.memberList.Members()
	for _, node := range nodes {
		if node.Name == nodeName {
			return NewNodeMetadataWithBytes(node.Meta)
		}
	}

	return nil, errors.ErrNodeDoesNotFound
}

func (c *Cluster) NodeAddress(nodeName string) (string, error) {
	nodes := c.memberList.Members()
	for _, node := range nodes {
		if node.Name == nodeName {
			return node.Addr.String(), nil
		}
	}

	return "", errors.ErrNodeDoesNotFound
}

func (c *Cluster) NodePort(nodeName string) (uint16, error) {
	nodes := c.memberList.Members()
	for _, node := range nodes {
		if node.Name == nodeName {
			return node.Port, nil
		}
	}

	return 0, errors.ErrNodeDoesNotFound
}

func (c *Cluster) NodeState(nodeName string) (NodeState, error) {
	nodes := c.memberList.Members()
	for _, node := range nodes {
		if node.Name == nodeName {
			return makeNodeState(node.State), nil
		}
	}

	return NodeStateUnknown, errors.ErrNodeDoesNotFound
}

func (c *Cluster) IsSeedNode() bool {
	return c.isSeedNode
}

func (c *Cluster) IsIndexer() bool {
	metadata, err := c.LocalNodeMetadata()
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

func (c *Cluster) IsSearcher() bool {
	metadata, err := c.LocalNodeMetadata()
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

func (c *Cluster) Nodes() []string {
	members := make([]string, 0)
	for _, member := range c.memberList.Members() {
		members = append(members, member.Name)
	}
	return members
}

func (c *Cluster) ClusterEvents() <-chan ClusterEvent {
	return c.clusterEvents
}

func (c *Cluster) Start() error {
	go func() {
		for {
			select {
			case cancel := <-c.stopWatching:
				if cancel {
					return
				}
			case nodeEvent := <-c.nodeEventDeliegate.NodeEvents:
				c.logger.Info("received node event", zap.Any("node_event", nodeEvent))

				clusterEvent := ClusterEvent{
					NodeEvent: nodeEvent,
					Members:   c.Nodes(),
				}

				switch nodeEvent.Type {
				case NodeEventTypeJoin:
					if nodeEvent.NodeMetadata.IsIndexer() {
						if !c.indexerHash.Contains(nodeEvent.NodeName) {
							c.indexerHash.AddWithWeight(nodeEvent.NodeName, 1.0)
						}
					}

					if nodeEvent.NodeMetadata.IsSearcher() {
						if !c.searcherHash.Contains(nodeEvent.NodeName) {
							c.searcherHash.AddWithWeight(nodeEvent.NodeName, 1.0)
						}
					}
				case NodeEventTypeUpdate:
					if nodeEvent.NodeMetadata.IsIndexer() {
						if !c.indexerHash.Contains(nodeEvent.NodeName) {
							c.indexerHash.AddWithWeight(nodeEvent.NodeName, 1.0)
						}
					}

					if nodeEvent.NodeMetadata.IsSearcher() {
						if !c.searcherHash.Contains(nodeEvent.NodeName) {
							c.searcherHash.AddWithWeight(nodeEvent.NodeName, 1.0)
						}
					}
				case NodeEventTypeLeave:
					if nodeEvent.NodeMetadata.IsIndexer() {
						if c.indexerHash.Contains(nodeEvent.NodeName) {
							c.indexerHash.Remove(nodeEvent.NodeName)
						}
					}

					if nodeEvent.NodeMetadata.IsSearcher() {
						if c.searcherHash.Contains(nodeEvent.NodeName) {
							c.searcherHash.Remove(nodeEvent.NodeName)
						}
					}
				}

				c.clusterEvents <- clusterEvent
			}
		}
	}()

	return nil
}

func (c *Cluster) Stop() error {
	c.stopWatching <- true

	return nil
}

func (c *Cluster) LookupIndexer(key string) string {
	return c.indexerHash.Lookup(key)
}

func (c *Cluster) LookupSearchers(key string, numNodes int) []string {
	return c.searcherHash.LookupTopN(key, numNodes)
}
