package membership

import (
	"fmt"
	"time"

	"github.com/hashicorp/memberlist"
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

type Node struct {
	memberList           *memberlist.Memberlist
	nodeEventDeliegate   *NodeEventDelegate
	nodeMetadataDelegate *NodeMetadataDelegate
	logger               *zap.Logger
	nodeEvents           chan NodeEvent
	stopWatching         chan bool
	isSeedNode           bool
	indexerHash          *rendezvous.Ring
	searcherHash         *rendezvous.Ring
}

func NewNode(host string, bindPort int, nodeMetadata NodeMetadata, isSeedNode bool, logger *zap.Logger) (*Node, error) {
	nodeLogger := logger.Named("node")

	nodeEventDeliegate := NewNodeEventDelegate(nodeLogger)

	nodeMetadataDelegate := NewNodeMetadataDelegate(nodeMetadata, nodeLogger)

	config := memberlist.DefaultLocalConfig()
	config.Name = generateNodeName()
	config.BindAddr = host
	config.BindPort = bindPort
	config.AdvertiseAddr = host
	config.AdvertisePort = bindPort
	config.Events = nodeEventDeliegate
	config.Delegate = nodeMetadataDelegate

	members, err := memberlist.Create(config)
	if err != nil {
		nodeLogger.Error("Failed to create member list", zap.Error(err), zap.Any("config", config))
		return nil, err
	}
	// members.LocalNode().Meta, err = nodeMetadata.Bytes()
	// if err != nil {
	// 	nodeLogger.Error("Failed to set node metadata", zap.Error(err))
	// }
	// members.UpdateNode(10 * time.Second)

	return &Node{
		memberList:           members,
		nodeEventDeliegate:   nodeEventDeliegate,
		nodeMetadataDelegate: nodeMetadataDelegate,
		logger:               nodeLogger,
		nodeEvents:           make(chan NodeEvent, 10),
		stopWatching:         make(chan bool),
		isSeedNode:           isSeedNode,
		indexerHash:          rendezvous.New(),
		searcherHash:         rendezvous.New(),
	}, nil
}

func (n *Node) Join(seeds []string) (int, error) {
	return n.memberList.Join(seeds)
}

func (n *Node) Leave(timeout time.Duration) error {
	return n.memberList.Leave(timeout)
}

func (n *Node) Name() string {
	return n.memberList.LocalNode().Name
}

func (n *Node) Metadata() (*NodeMetadata, error) {
	return NewNodeMetadataWithBytes(n.memberList.LocalNode().Meta)
}

func (n *Node) IsSeedNode() bool {
	return n.isSeedNode
}

func (n *Node) IsIndexer() bool {
	metadata, err := n.Metadata()
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

func (n *Node) IsSearcher() bool {
	metadata, err := n.Metadata()
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

func (n *Node) Member(name string) *memberlist.Node {
	for _, member := range n.Members() {
		if member.Name == name {
			return member
		}
	}

	return nil
}

func (n *Node) Members() []*memberlist.Node {
	return n.memberList.Members()
}

func (n *Node) Events() <-chan NodeEvent {
	return n.nodeEvents
}

func (n *Node) Start() error {
	go func() {
		for {
			select {
			case cancel := <-n.stopWatching:
				if cancel {
					return
				}
			case nodeEvent := <-n.nodeEventDeliegate.NodeEvents:
				n.logger.Info("Received node event", zap.Any("nodeEvent", nodeEvent))

				// make member list
				members := make([]string, 0)
				for _, member := range n.Members() {
					members = append(members, member.Name)
				}
				nodeEvent.Members = members

				switch nodeEvent.Type {
				case EventTypeJoin:
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
				case EventTypeUpdate:
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
				case EventTypeLeave:
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

				n.nodeEvents <- nodeEvent
			}
		}
	}()

	return nil
}

func (n *Node) Stop() error {
	n.stopWatching <- true

	return nil
}

func (n *Node) LookupIndexer(key string) string {
	return n.indexerHash.Lookup(key)
}

func (n *Node) LookupSearchers(key string, numNodes int) []string {
	return n.searcherHash.LookupTopN(key, numNodes)
}
