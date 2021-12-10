package index

import (
	"sync"

	"github.com/mosuka/rendezvous"
)

type ShardHash struct {
	shardHashMap map[string]*rendezvous.Ring
	mutex        sync.RWMutex
}

func NewShardHashMap() *ShardHash {
	return &ShardHash{
		shardHashMap: make(map[string]*rendezvous.Ring),
		mutex:        sync.RWMutex{},
	}
}

func (h *ShardHash) Indexes() []string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	indexes := make([]string, 0)
	for indexName := range h.shardHashMap {
		indexes = append(indexes, indexName)
	}

	return indexes
}

func (h *ShardHash) Contains(indexName string, shardName string) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	hash, ok := h.shardHashMap[indexName]
	if !ok {
		return false
	}
	return hash.Contains(shardName)
}

func (h *ShardHash) Get(indexName string, key string) string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	hash, ok := h.shardHashMap[indexName]
	if !ok {
		return ""
	}
	return hash.Lookup(key)
}

func (h *ShardHash) Set(indexName string, shardName string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	hash, ok := h.shardHashMap[indexName]
	if !ok {
		hash = rendezvous.New()
		h.shardHashMap[indexName] = hash
	}

	hash.Add(shardName)
}

func (h *ShardHash) Delete(indexName string, shardName string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	hash, ok := h.shardHashMap[indexName]
	if !ok {
		return
	}
	hash.Remove(shardName)

	if len(hash.List()) == 0 {
		delete(h.shardHashMap, indexName)
	}
}

func (h *ShardHash) Len(indexName string) int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	hash, ok := h.shardHashMap[indexName]
	if !ok {
		return 0
	}

	return hash.Len()
}

func (h *ShardHash) List(indexName string) []string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	hash, ok := h.shardHashMap[indexName]
	if !ok {
		return []string{}
	}

	return hash.List()
}
