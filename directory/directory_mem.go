package directory

import (
	"net/url"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/index"
	"github.com/mosuka/phalanx/errors"
	"go.uber.org/zap"
)

func InMemoryIndexConfig(uri string, logger *zap.Logger) bluge.Config {
	return bluge.DefaultConfigWithDirectory(func() index.Directory {
		return NewInMemoryDirectoryWithUri(uri, logger)
	})
}

func NewInMemoryDirectoryWithUri(uri string, logger *zap.Logger) *index.InMemoryDirectory {
	inMemoryLogger := logger.Named("in_memory")

	u, err := url.Parse(uri)
	if err != nil {
		inMemoryLogger.Error("failed to parse URI", zap.Error(err), zap.String("uri", uri))
		return nil
	}

	if u.Scheme != SchemeType_name[SchemeTypeMem] {
		err := errors.ErrInvalidUri
		inMemoryLogger.Error("failed to parse URI", zap.Error(err), zap.String("uri", uri))
		return nil
	}

	return index.NewInMemoryDirectory()
}
