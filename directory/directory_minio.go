package directory

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/index"
	segment "github.com/blugelabs/bluge_segment_api"
	minio "github.com/minio/minio-go/v7"
	"github.com/mosuka/phalanx/clients"
	"github.com/mosuka/phalanx/errors"
	"github.com/mosuka/phalanx/lock"
	"go.uber.org/zap"
)

type uint64Slice []uint64

func (e uint64Slice) Len() int           { return len(e) }
func (e uint64Slice) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
func (e uint64Slice) Less(i, j int) bool { return e[i] < e[j] }

func MinioIndexConfig(uri string, lockManager lock.LockManager, logger *zap.Logger) bluge.Config {
	return bluge.DefaultConfigWithDirectory(func() index.Directory {
		return NewMinioDirectoryWithUri(uri, lockManager, logger)
	})
}

type MinioDirectory struct {
	bucket         string
	path           string
	client         *minio.Client
	ctx            context.Context
	requestTimeout time.Duration
	lockManager    lock.LockManager
	logger         *zap.Logger
}

func NewMinioDirectoryWithUri(uri string, lockManager lock.LockManager, logger *zap.Logger) *MinioDirectory {
	directoryLogger := logger.Named("minio")

	client, err := clients.NewMinioClientWithUri(uri)
	if err != nil {
		logger.Error("failed to create minio client", zap.Error(err))
		return nil
	}

	// Parse URI.
	u, err := url.Parse(uri)
	if err != nil {
		logger.Error("failed to parse URI", zap.Error(err))
		return nil
	}
	if u.Scheme != SchemeType_name[SchemeTypeMinio] {
		err := errors.ErrInvalidUri
		logger.Error("scheme is not etcd", zap.Error(err))
		return nil
	}

	return &MinioDirectory{
		client:         client,
		bucket:         u.Host,
		path:           u.Path,
		ctx:            context.Background(),
		requestTimeout: 3 * time.Second,
		lockManager:    lockManager,
		logger:         directoryLogger,
	}
}

func (d *MinioDirectory) exists() (bool, error) {
	ctx, cancel := context.WithTimeout(d.ctx, d.requestTimeout)
	defer cancel()

	// Check to see if we already own this bucket (which happens if you run this twice)
	exists, err := d.client.BucketExists(ctx, d.bucket)
	if err != nil {
		d.logger.Error("failed to check the bucket existence", zap.Error(err))
		return false, err
	}

	return exists, nil
}

func (d *MinioDirectory) fileName(kind string, id uint64) string {
	return fmt.Sprintf("%012x", id) + kind
}

func (d *MinioDirectory) Setup(readOnly bool) error {
	exists, err := d.exists()
	if err != nil {
		d.logger.Error("failed to check the bucket existence", zap.Error(err))
		return err
	}

	if !exists {
		ctx, cancel := context.WithTimeout(d.ctx, d.requestTimeout)
		defer cancel()

		region, err := d.client.GetBucketLocation(ctx, d.bucket)
		if err != nil {
			d.logger.Error("failed to get bucket location (region)", zap.Error(err), zap.String("bucket", d.bucket))
			return err
		}

		opts := minio.MakeBucketOptions{
			Region: region,
		}

		err = d.client.MakeBucket(ctx, d.bucket, opts)
		if err != nil {
			d.logger.Error("failed to make the bucket", zap.Error(err), zap.String("region", region), zap.String("bucket", d.bucket))
			return err
		}
	}

	return nil
}

func (d *MinioDirectory) List(kind string) ([]uint64, error) {
	opts := minio.ListObjectsOptions{
		Prefix:    d.path,
		Recursive: true,
	}

	ctx, cancel := context.WithTimeout(d.ctx, d.requestTimeout)
	defer cancel()

	var rv uint64Slice
	for object := range d.client.ListObjects(ctx, d.bucket, opts) {
		if object.Err != nil {
			d.logger.Error("failed to list objects", zap.Error(object.Err))
			return nil, object.Err
		}
		if filepath.Ext(object.Key) != kind {
			continue
		}

		// E.g. indexes/wikipedia_en/000000000004.seg -> 000000000004
		base := filepath.Base(object.Key)
		base = base[:len(base)-len(kind)]

		var epoch uint64
		epoch, err := strconv.ParseUint(base, 16, 64)
		if err != nil {
			d.logger.Error("failed to parse identifier", zap.Error(object.Err), zap.String("base", base))
			return nil, err
		}
		rv = append(rv, epoch)
	}

	sort.Sort(sort.Reverse(rv))

	return rv, nil
}

func (d *MinioDirectory) Load(kind string, id uint64) (*segment.Data, io.Closer, error) {
	path := filepath.Join(d.path, d.fileName(kind, id))

	opts := minio.GetObjectOptions{}

	ctx, cancel := context.WithTimeout(d.ctx, d.requestTimeout)
	defer cancel()

	object, err := d.client.GetObject(ctx, d.bucket, path, opts)
	if err != nil {
		d.logger.Error("failed to get object", zap.Error(err), zap.String("path", path))
		return nil, nil, err
	}

	data, err := ioutil.ReadAll(object)
	if err != nil {
		d.logger.Error("failed to read object", zap.Error(err), zap.String("path", path))
		return nil, nil, err
	}

	return segment.NewDataBytes(data), nil, nil
}

func (d *MinioDirectory) Persist(kind string, id uint64, w index.WriterTo, closeCh chan struct{}) error {
	var buf bytes.Buffer
	size, err := w.WriteTo(&buf, closeCh)
	if err != nil {
		d.logger.Error("failed to write object", zap.Error(err), zap.String("kind", kind), zap.Uint64("id", id))
		return err
	}

	reader := bufio.NewReader(&buf)

	path := filepath.Join(d.path, d.fileName(kind, id))

	opts := minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	}

	ctx, cancel := context.WithTimeout(d.ctx, d.requestTimeout)
	defer cancel()

	putInfo, err := d.client.PutObject(ctx, d.bucket, path, reader, size, opts)
	if err != nil {
		d.logger.Error("failed to put object", zap.Error(err), zap.String("path", path), zap.Int64("size", size))

		// TODO: Remove the failed file.
		return err
	}
	if size != putInfo.Size {
		d.logger.Error("failed to put object", zap.String("path", path), zap.Int64("expected_size", size), zap.Int64("actual_size", putInfo.Size))

		// TODO: Remove the failed file.
		return err
	}

	return nil
}

func (d *MinioDirectory) Remove(kind string, id uint64) error {
	path := filepath.Join(d.path, d.fileName(kind, id))

	opts := minio.RemoveObjectOptions{
		GovernanceBypass: true,
	}

	ctx, cancel := context.WithTimeout(d.ctx, d.requestTimeout)
	defer cancel()

	err := d.client.RemoveObject(ctx, d.bucket, path, opts)
	if err != nil {
		d.logger.Error("failed to remove object", zap.Error(err), zap.String("path", path))
		return err
	}

	return nil
}

func (d *MinioDirectory) Stats() (uint64, uint64) {
	opts := minio.ListObjectsOptions{
		Prefix:    d.path,
		Recursive: true,
	}

	numFilesOnDisk := uint64(0)
	numBytesUsedDisk := uint64(0)

	ctx, cancel := context.WithTimeout(d.ctx, d.requestTimeout)
	defer cancel()

	for object := range d.client.ListObjects(ctx, d.bucket, opts) {
		if object.Err != nil {
			d.logger.Error("failed to list objects", zap.Error(object.Err))
			return 0, 0
		}

		numFilesOnDisk++
		numBytesUsedDisk += uint64(object.Size)
	}

	return numFilesOnDisk, numBytesUsedDisk
}

func (d *MinioDirectory) Sync() error {
	// d.logger.Debug("sync", zap.String("path", d.path))
	return nil
}

func (d *MinioDirectory) Lock() error {
	if _, err := d.lockManager.Lock(); err != nil {
		d.logger.Error("failed to lock", zap.Error(err), zap.String("path", d.path))
		return err
	}

	return nil
}

func (d *MinioDirectory) Unlock() error {
	if err := d.lockManager.Unlock(); err != nil {
		d.logger.Error("failed to unlock", zap.Error(err), zap.String("path", d.path))
		return err
	}

	return nil
}
