package cluster

import (
	"github.com/hashicorp/memberlist"
	"go.uber.org/zap"
)

type NodeEventType int

const (
	NodeEventTypeUnknown NodeEventType = iota
	NodeEventTypeJoin
	NodeEventTypeLeave
	NodeEventTypeUpdate
)

// Enum value maps for NodeEventType.
var (
	NodeEventType_name = map[NodeEventType]string{
		NodeEventTypeUnknown: "unknown",
		NodeEventTypeJoin:    "join",
		NodeEventTypeLeave:   "leave",
		NodeEventTypeUpdate:  "update",
	}
	NodeEventType_value = map[string]NodeEventType{
		"unknown": NodeEventTypeUnknown,
		"join":    NodeEventTypeJoin,
		"leave":   NodeEventTypeLeave,
		"update":  NodeEventTypeUpdate,
	}
)

type NodeState int

const (
	NodeStateUnknown NodeState = iota
	NodeStateAlive
	NodeStateSuspect
	NodeStateDead
	NodeStateLeft
)

// Enum value maps for NodeState.
var (
	NodeState_name = map[NodeState]string{
		NodeStateUnknown: "unknown",
		NodeStateAlive:   "alive",
		NodeStateSuspect: "suspect",
		NodeStateDead:    "dead",
		NodeStateLeft:    "left",
	}
	NodeState_value = map[string]NodeState{
		"unknown": NodeStateUnknown,
		"alive":   NodeStateAlive,
		"suspect": NodeStateSuspect,
		"dead":    NodeStateDead,
		"left":    NodeStateLeft,
	}
)

type NodeEvent struct {
	Type         NodeEventType
	NodeName     string
	NodeMetadata *NodeMetadata
	NodeState    NodeState
}

type ClusterEvent struct {
	NodeEvent NodeEvent
	Members   []string
}

type NodeEventDelegate struct {
	NodeEvents chan NodeEvent
	logger     *zap.Logger
}

func NewNodeEventDelegate(logger *zap.Logger) *NodeEventDelegate {
	delegateLogger := logger.Named("event_delegate")

	return &NodeEventDelegate{
		NodeEvents: make(chan NodeEvent, 10),
		logger:     delegateLogger,
	}
}

func (d *NodeEventDelegate) NotifyJoin(node *memberlist.Node) {
	d.NodeEvents <- makeNodeEvent(NodeEventTypeJoin, node)
}
func (d *NodeEventDelegate) NotifyLeave(node *memberlist.Node) {
	d.NodeEvents <- makeNodeEvent(NodeEventTypeLeave, node)
}
func (d *NodeEventDelegate) NotifyUpdate(node *memberlist.Node) {
	d.NodeEvents <- makeNodeEvent(NodeEventTypeUpdate, node)
}

func makeNodeState(state memberlist.NodeStateType) NodeState {
	switch state {
	case memberlist.StateAlive:
		return NodeStateAlive
	case memberlist.StateSuspect:
		return NodeStateSuspect
	case memberlist.StateDead:
		return NodeStateDead
	case memberlist.StateLeft:
		return NodeStateLeft
	default:
		return NodeStateUnknown
	}
}

func makeNodeEvent(eventType NodeEventType, node *memberlist.Node) NodeEvent {
	nodeMeta, err := NewNodeMetadataWithBytes(node.Meta)
	if err != nil {
		nodeMeta = NewNodeMetadata()
	}

	state := makeNodeState(node.State)

	return NodeEvent{
		Type:         eventType,
		NodeName:     node.Name,
		NodeState:    state,
		NodeMetadata: nodeMeta,
	}
}

type NodeMetadataDelegate struct {
	metadata NodeMetadata
	logger   *zap.Logger
}

func NewNodeMetadataDelegate(metadata NodeMetadata, logger *zap.Logger) *NodeMetadataDelegate {
	delegateLogger := logger.Named("metadata_delegate")

	return &NodeMetadataDelegate{
		metadata: metadata,
		logger:   delegateLogger,
	}
}

func (d *NodeMetadataDelegate) NodeMeta(limit int) []byte {
	data, err := d.metadata.Marshal()
	if err != nil {
		return []byte{}
	}

	return data
}

func (d *NodeMetadataDelegate) LocalState(join bool) []byte {
	return []byte{}
}
func (d *NodeMetadataDelegate) NotifyMsg(msg []byte) {
	// d.logger.Debug("notify msg", zap.ByteString("msg", msg))
}

func (d *NodeMetadataDelegate) GetBroadcasts(overhead, limit int) [][]byte {
	return [][]byte{}
}

func (d *NodeMetadataDelegate) MergeRemoteState(buf []byte, join bool) {
	// d.logger.Debug("merge remote state", zap.ByteString("buf", buf), zap.Bool("join", join))
}
