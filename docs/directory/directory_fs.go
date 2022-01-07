package directory

import (
	"net/url"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/index"
	"github.com/mosuka/phalanx/errors"
	"go.uber.org/zap"
)

func FileSystemIndexConfig(uri string, logger *zap.Logger) bluge.Config {
	return bluge.DefaultConfigWithDirectory(func() index.Directory {
		return NewFileSystemDirectoryWithUri(uri, logger)
	})
}

func NewFileSystemDirectoryWithUri(uri string, logger *zap.Logger) *index.FileSystemDirectory {
	fileSystemLogger := logger.Named("file_system")

	// Parse URI.
	u, err := url.Parse(uri)
	if err != nil {
		fileSystemLogger.Error(err.Error(), zap.String("uri", uri))
		return nil
	}

	if u.Scheme != SchemeType_name[SchemeTypeFile] {
		err := errors.ErrInvalidUri
		fileSystemLogger.Error(err.Error(), zap.String("uri", uri))
		return nil
	}

	path := u.Path

	return index.NewFileSystemDirectory(path)
}
