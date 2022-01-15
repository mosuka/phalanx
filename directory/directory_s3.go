package directory

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	phalanxerrors "github.com/mosuka/phalanx/errors"
	"io"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/index"
	segment "github.com/blugelabs/bluge_segment_api"
	"github.com/mosuka/phalanx/clients"
	"github.com/mosuka/phalanx/lock"
	"go.uber.org/zap"
)

func S3IndexConfig(uri string, lockUri string, logger *zap.Logger) bluge.Config {
	return bluge.DefaultConfigWithDirectory(func() index.Directory {
		return NewS3DirectoryWithUri(uri, lockUri, logger)
	})
}

type S3Directory struct {
	bucket         string
	path           string
	client         *s3.Client
	ctx            context.Context
	requestTimeout time.Duration
	lockUri        string
	lockManager    lock.LockManager
	logger         *zap.Logger
}

func NewS3DirectoryWithUri(uri string, lockUri string, logger *zap.Logger) *S3Directory {
	directoryLogger := logger.Named("s3")

	client, err := clients.NewS3ClientWithUri(uri)
	if err != nil {
		logger.Error(err.Error(), zap.String("uri", uri))
		return nil
	}

	// Parse URI.
	u, err := url.Parse(uri)
	if err != nil {
		logger.Error(err.Error(), zap.String("uri", uri))
		return nil
	}
	if u.Scheme != SchemeType_name[SchemeTypeS3] {
		err := phalanxerrors.ErrInvalidUri
		logger.Error(err.Error(), zap.String("uri", uri))
		return nil
	}

	return &S3Directory{
		client:         client,
		bucket:         u.Host,
		path:           u.Path,
		ctx:            context.Background(),
		requestTimeout: 3 * time.Second,
		lockUri:        lockUri,
		lockManager:    nil,
		logger:         directoryLogger,
	}
}

func (d *S3Directory) fileName(kind string, id uint64) string {
	return fmt.Sprintf("%012x", id) + kind
}

func (d *S3Directory) Setup(readOnly bool) error {
	d.logger.Info("setup", zap.String("bucket", d.bucket), zap.String("path", d.path))

	//ctx, cancel := context.WithTimeout(d.ctx, d.requestTimeout)
	//defer cancel()

	input := &s3.CreateBucketInput{
		Bucket: aws.String(d.bucket),
	}

	_, err := d.client.CreateBucket(d.ctx, input)
	if err != nil {
		var bne *types.BucketAlreadyExists
		if errors.As(err, &bne) {
			d.logger.Info(err.Error(), zap.String("bucket", d.bucket))
		} else {
			d.logger.Error(err.Error(), zap.String("bucket", d.bucket))
			return err
		}
	}

	return nil
}

func (d *S3Directory) List(kind string) ([]uint64, error) {
	ctx, cancel := context.WithTimeout(d.ctx, d.requestTimeout)
	defer cancel()

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(d.bucket),
	}

	list, err := d.client.ListObjectsV2(ctx, input)
	if err != nil {
		d.logger.Error(err.Error(), zap.String("bucket", d.bucket))
		return nil, err
	}

	var rv uint64Slice
	for _, object := range list.Contents {
		if filepath.Ext(*object.Key) != kind {
			continue
		}

		// E.g. indexes/wikipedia_en/000000000004.seg -> 000000000004
		base := filepath.Base(*object.Key)
		base = base[:len(base)-len(kind)]

		var epoch uint64
		epoch, err := strconv.ParseUint(base, 16, 64)
		if err != nil {
			d.logger.Error(err.Error(), zap.String("base", base))
			return nil, err
		}
		rv = append(rv, epoch)
	}

	sort.Sort(sort.Reverse(rv))

	return rv, nil
}

func (d *S3Directory) Load(kind string, id uint64) (*segment.Data, io.Closer, error) {
	path := filepath.Join(d.path, d.fileName(kind, id))

	ctx, cancel := context.WithTimeout(d.ctx, d.requestTimeout)
	defer cancel()

	input := &s3.GetObjectInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(path),
	}

	object, err := d.client.GetObject(ctx, input)
	if err != nil {
		d.logger.Error(err.Error(), zap.String("bucket", d.bucket), zap.String("path", path))
		return nil, nil, err
	}

	data, err := ioutil.ReadAll(object.Body)
	if err != nil {
		d.logger.Error(err.Error())
		return nil, nil, err
	}

	return segment.NewDataBytes(data), nil, nil
}

func (d *S3Directory) Persist(kind string, id uint64, w index.WriterTo, closeCh chan struct{}) error {
	var buf bytes.Buffer
	size, err := w.WriteTo(&buf, closeCh)
	if err != nil {
		d.logger.Error(err.Error())
		return err
	}

	reader := bytes.NewReader(buf.Bytes())

	path := filepath.Join(d.path, d.fileName(kind, id))

	ctx, cancel := context.WithTimeout(d.ctx, d.requestTimeout)
	defer cancel()

	input := &s3.PutObjectInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(path),
		Body:   reader,
	}

	if _, err := d.client.PutObject(ctx, input); err != nil {
		d.logger.Error(err.Error(), zap.String("bucket", d.bucket), zap.String("path", path), zap.Int64("size", size))

		// TODO: Remove the failed file.
		return err
	}

	return nil
}

func (d *S3Directory) Remove(kind string, id uint64) error {
	path := filepath.Join(d.path, d.fileName(kind, id))

	ctx, cancel := context.WithTimeout(d.ctx, d.requestTimeout)
	defer cancel()

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(path),
	}

	if _, err := d.client.DeleteObject(ctx, input); err != nil {
		d.logger.Error(err.Error(), zap.String("bucket", d.bucket), zap.String("path", path))
		return err
	}

	return nil
}

func (d *S3Directory) Stats() (uint64, uint64) {
	numFilesOnDisk := uint64(0)
	numBytesUsedDisk := uint64(0)

	ctx, cancel := context.WithTimeout(d.ctx, d.requestTimeout)
	defer cancel()

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(d.bucket),
		Prefix: aws.String(d.path),
	}

	list, err := d.client.ListObjectsV2(ctx, input)
	if err != nil {
		d.logger.Error(err.Error(), zap.String("bucket", d.bucket))
		return 0, 0
	}

	for _, object := range list.Contents {
		numFilesOnDisk++
		numBytesUsedDisk += uint64(object.Size)
	}

	return numFilesOnDisk, numBytesUsedDisk
}

func (d *S3Directory) Sync() error {
	return nil
}

func (d *S3Directory) Lock() error {
	// Create lock manager
	lockManager, err := lock.NewLockManagerWithUri(d.lockUri, d.logger)
	if err != nil {
		d.logger.Error(err.Error(), zap.String("lock_uri", d.lockUri))
		return err
	}
	d.lockManager = lockManager

	if _, err := d.lockManager.Lock(); err != nil {
		d.logger.Error(err.Error())
		return err
	}

	return nil
}

func (d *S3Directory) Unlock() error {
	if err := d.lockManager.Unlock(); err != nil {
		d.logger.Error(err.Error())
		return err
	}

	d.lockManager.Close()

	return nil
}
