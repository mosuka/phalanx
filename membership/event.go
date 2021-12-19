package membership

import (
	"github.com/hashicorp/memberlist"
	"go.uber.org/zap"
)

type EventType int

const (
	EventTypeUnknown EventType = iota
	EventTypeJoin
	EventTypeLeave
	EventTypeUpdate
)

// Enum value maps for EventType.
var (
	EventType_name = map[EventType]string{
		EventTypeUnknown: "unknown",
		EventTypeJoin:    "join",
		EventTypeLeave:   "leave",
		EventTypeUpdate:  "update",
	}
	EventType_value = map[string]EventType{
		"unknown": EventTypeUnknown,
		"join":    EventTypeJoin,
		"leave":   EventTypeLeave,
		"update":  EventTypeUpdate,
	}
)

type StateType int

const (
	StateTypeUnknown StateType = iota
	StateTypeAlive
	StateTypeSuspect
	StateTypeDead
	StateTypeLeft
)

// Enum value maps for StateType.
var (
	StateType_name = map[StateType]string{
		StateTypeUnknown: "unknown",
		StateTypeAlive:   "alive",
		StateTypeSuspect: "suspect",
		StateTypeDead:    "dead",
		StateTypeLeft:    "left",
	}
	StateType_value = map[string]StateType{
		"unknown": StateTypeUnknown,
		"alive":   StateTypeAlive,
		"suspect": StateTypeSuspect,
		"dead":    StateTypeDead,
		"left":    StateTypeLeft,
	}
)

type NodeEvent struct {
	Type    EventType
	Node    string
	Meta    *NodeMetadata
	State   StateType
	Members []string
}

type NodeEventDelegate struct {
	NodeEvents chan NodeEvent
	logger     *zap.Logger
}

func NewNodeEventDelegate(logger *zap.Logger) *NodeEventDelegate {
	delegateLogger := logger.Named("node_event_delegate")

	return &NodeEventDelegate{
		NodeEvents: make(chan NodeEvent, 10),
		logger:     delegateLogger,
	}
}

func (d *NodeEventDelegate) NotifyJoin(node *memberlist.Node) {
	d.NodeEvents <- makeNodeEvent(EventTypeJoin, node)
}
func (d *NodeEventDelegate) NotifyLeave(node *memberlist.Node) {
	d.NodeEvents <- makeNodeEvent(EventTypeLeave, node)
}
func (d *NodeEventDelegate) NotifyUpdate(node *memberlist.Node) {
	d.NodeEvents <- makeNodeEvent(EventTypeUpdate, node)
}

func makeNodeEvent(eventType EventType, node *memberlist.Node) NodeEvent {
	nodeMeta, err := NewNodeMetadataWithBytes(node.Meta)
	if err != nil {
		nodeMeta = NewNodeMetadata()
	}

	state := StateTypeUnknown
	switch node.State {
	case memberlist.StateAlive:
		state = StateTypeAlive
	case memberlist.StateSuspect:
		state = StateTypeSuspect
	case memberlist.StateDead:
		state = StateTypeDead
	case memberlist.StateLeft:
		state = StateTypeLeft
	}

	return NodeEvent{
		Type:    eventType,
		Node:    node.Name,
		State:   state,
		Meta:    nodeMeta,
		Members: []string{},
	}
}

type NodeMetadataDelegate struct {
	metadata NodeMetadata
	logger   *zap.Logger
}

func NewNodeMetadataDelegate(metadata NodeMetadata, logger *zap.Logger) *NodeMetadataDelegate {
	delegateLogger := logger.Named("node_metadata_delegate")

	return &NodeMetadataDelegate{
		metadata: metadata,
		logger:   delegateLogger,
	}
}

func (d *NodeMetadataDelegate) NodeMeta(limit int) []byte {
	data, err := d.metadata.Bytes()
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
