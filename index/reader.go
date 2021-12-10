package index

import (
	"sync"

	"github.com/blugelabs/bluge"
	"github.com/mosuka/phalanx/directory"
	"github.com/mosuka/phalanx/errors"
	"github.com/mosuka/phalanx/lock"
	"go.uber.org/zap"
)

type IndexReader struct {
	*bluge.Reader
	version int64
}

func NewIndexReader(reader *bluge.Reader, version int64) *IndexReader {
	return &IndexReader{
		Reader:  reader,
		version: version,
	}
}

func (r *IndexReader) BlugeReader() *bluge.Reader {
	return r.Reader
}

func (r *IndexReader) Version() int64 {
	return r.version
}

type IndexReaders struct {
	readerMap map[string]map[string]*IndexReader
	mutex     sync.RWMutex
	logger    *zap.Logger
}

func NewIndexReaders(logger *zap.Logger) *IndexReaders {
	readerLogger := logger.Named("reader")

	return &IndexReaders{
		readerMap: make(map[string]map[string]*IndexReader),
		logger:    readerLogger,
	}
}

func (i *IndexReaders) indexes() []string {
	indexes := make([]string, 0, len(i.readerMap))
	for index := range i.readerMap {
		indexes = append(indexes, index)
	}

	return indexes
}

func (i *IndexReaders) Indexes() []string {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	return i.indexes()
}

func (i *IndexReaders) shards(indexName string) []string {
	_, ok := i.readerMap[indexName]
	if !ok {
		return nil
	}

	shards := make([]string, 0, len(i.readerMap[indexName]))
	for shard := range i.readerMap[indexName] {
		shards = append(shards, shard)
	}

	return shards
}

func (i *IndexReaders) Shards(indexName string) []string {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	return i.shards(indexName)
}

func (i *IndexReaders) contains(indexName string, shardName string) bool {
	_, ok := i.readerMap[indexName]
	if !ok {
		return false
	}

	_, ok = i.readerMap[indexName][shardName]

	return ok
}

func (i *IndexReaders) Contains(indexName string, shardName string) bool {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	return i.contains(indexName, shardName)
}

func (i *IndexReaders) open(indexName string, shardName string, indexMetadata *IndexMetadata, shardMetadata *ShardMetadata) error {
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
	reader, err := bluge.OpenReader(config)
	if err != nil {
		return err
	}

	_, ok := i.readerMap[indexName]
	if !ok {
		i.readerMap[indexName] = make(map[string]*IndexReader)
	}

	i.readerMap[indexName][shardName] = NewIndexReader(reader, shardMetadata.ShardVersion)

	return nil
}

func (i *IndexReaders) Open(indexName string, shardName string, indexMetadata *IndexMetadata, shardMetadata *ShardMetadata) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	// Open index reader.
	return i.open(indexName, shardName, indexMetadata, shardMetadata)
}

func (i *IndexReaders) get(indexName string, shardName string) (*IndexReader, error) {
	_, ok := i.readerMap[indexName]
	if !ok {
		return nil, errors.ErrIndexDoesNotExist
	}

	reader, ok := i.readerMap[indexName][shardName]
	if !ok {
		return nil, errors.ErrShardDoesNotExist
	}

	return reader, nil
}

func (i *IndexReaders) Get(indexName string, shardName string) (*IndexReader, error) {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	return i.get(indexName, shardName)
}

func (i *IndexReaders) Version(indexName string, shardName string) int64 {
	reader, err := i.get(indexName, shardName)
	if err != nil {
		return 0
	}

	return reader.Version()
}

func (i *IndexReaders) close(indexName string, shardName string) error {
	_, ok := i.readerMap[indexName]
	if !ok {
		return errors.ErrIndexDoesNotExist
	}

	reader, ok := i.readerMap[indexName][shardName]
	if !ok {
		return errors.ErrShardDoesNotExist
	}

	if err := reader.Close(); err != nil {
		return err
	}

	delete(i.readerMap[indexName], shardName)

	if len(i.readerMap[indexName]) == 0 {
		delete(i.readerMap, indexName)
	}

	return nil
}

func (i *IndexReaders) Close(indexName string, shardName string) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	return i.close(indexName, shardName)
}

func (i *IndexReaders) Reopen(indexName string, shardName string, indexMetadata *IndexMetadata, shardMetadata *ShardMetadata) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	// Close index reader.
	if err := i.close(indexName, shardName); err != nil {
		return err
	}

	// Open index reader.
	if err := i.open(indexName, shardName, indexMetadata, shardMetadata); err != nil {
		return err
	}

	return nil
}

func (i *IndexReaders) CloseAll() error {
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
