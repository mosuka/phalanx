package index

import (
	"sync"

	"github.com/blugelabs/bluge"
	"github.com/mosuka/phalanx/directory"
	"github.com/mosuka/phalanx/errors"
	"github.com/mosuka/phalanx/lock"
	"github.com/mosuka/phalanx/metastore"
	"go.uber.org/zap"
)

type IndexWriters struct {
	writerMap map[string]map[string]*bluge.Writer
	mutex     sync.RWMutex
	logger    *zap.Logger
}

func NewIndexWriters(logger *zap.Logger) *IndexWriters {
	writerLogger := logger.Named("writer")

	return &IndexWriters{
		writerMap: make(map[string]map[string]*bluge.Writer),
		logger:    writerLogger,
	}
}

func (i *IndexWriters) indexes() []string {
	indexes := make([]string, 0, len(i.writerMap))
	for index := range i.writerMap {
		indexes = append(indexes, index)
	}

	return indexes
}

func (i *IndexWriters) Indexes() []string {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	return i.indexes()
}

func (i *IndexWriters) shards(indexName string) []string {
	_, ok := i.writerMap[indexName]
	if !ok {
		return nil
	}

	shards := make([]string, 0, len(i.writerMap[indexName]))
	for shard := range i.writerMap[indexName] {
		shards = append(shards, shard)
	}

	return shards
}

func (i *IndexWriters) Shards(indexName string) []string {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	return i.shards(indexName)
}

func (i *IndexWriters) contains(indexName string, shardName string) bool {
	_, ok := i.writerMap[indexName]
	if !ok {
		return false
	}

	_, ok = i.writerMap[indexName][shardName]

	return ok
}

func (i *IndexWriters) Contains(indexName string, shardName string) bool {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	return i.contains(indexName, shardName)
}

func (i *IndexWriters) open(indexName string, shardName string, indexMetadata *metastore.IndexMetadata, shardMetadata *metastore.ShardMetadata) error {
	// Create lock manager
	lockManager, err := lock.NewLockManagerWithUri(shardMetadata.ShardLockUri, i.logger)
	if err != nil {
		return err
	}

	// Make directory config
	config, err := directory.NewIndexConfigWithUri(shardMetadata.ShardUri, lockManager, i.logger)
	if err != nil {
		return err
	}
	if indexMetadata.DefaultSearchField != "" {
		config.DefaultSearchField = indexMetadata.DefaultSearchField
	}
	// config.DefaultSearchAnalyzer = req.DefaultSearchAnalyzer
	// config.DefaultSimilarity = req.DefaultSearchSimilarity

	// Open index writer.
	writer, err := bluge.OpenWriter(config)
	if err != nil {
		return err
	}

	_, ok := i.writerMap[indexName]
	if !ok {
		i.writerMap[indexName] = make(map[string]*bluge.Writer)
	}

	i.writerMap[indexName][shardName] = writer

	return nil
}

func (i *IndexWriters) Open(indexName string, shardName string, indexMetadata *metastore.IndexMetadata, shardMetadata *metastore.ShardMetadata) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	// Open index writer.
	return i.open(indexName, shardName, indexMetadata, shardMetadata)
}

func (i *IndexWriters) get(indexName string, shardName string) (*bluge.Writer, error) {
	_, ok := i.writerMap[indexName]
	if !ok {
		return nil, errors.ErrIndexDoesNotExist
	}

	writer, ok := i.writerMap[indexName][shardName]
	if !ok {
		return nil, errors.ErrShardDoesNotExist
	}

	return writer, nil
}

func (i *IndexWriters) Get(indexName string, shardName string) (*bluge.Writer, error) {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	return i.get(indexName, shardName)
}

func (i *IndexWriters) close(indexName string, shardName string) error {
	_, ok := i.writerMap[indexName]
	if !ok {
		return errors.ErrIndexDoesNotExist
	}

	writer, ok := i.writerMap[indexName][shardName]
	if !ok {
		return errors.ErrShardDoesNotExist
	}

	if err := writer.Close(); err != nil {
		return err
	}

	delete(i.writerMap[indexName], shardName)

	if len(i.writerMap[indexName]) == 0 {
		delete(i.writerMap, indexName)
	}

	return nil
}

func (i *IndexWriters) Close(indexName string, shardName string) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	return i.close(indexName, shardName)
}

func (i *IndexWriters) Reopen(indexName string, shardName string, indexMetadata *metastore.IndexMetadata, shardMetadata *metastore.ShardMetadata) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	// Close index writer.
	if err := i.close(indexName, shardName); err != nil {
		return err
	}

	// Open index writer.
	if err := i.open(indexName, shardName, indexMetadata, shardMetadata); err != nil {
		return err
	}

	return nil
}

func (i *IndexWriters) CloseAll() error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	for _, indexName := range i.indexes() {
		for _, shardName := range i.shards(indexName) {
			if err := i.close(indexName, shardName); err != nil {
				return err
			}
		}
	}

	return nil
}
