package errors

import "errors"

var (
	ErrIndexDoesNotExist  = errors.New("index does not exist")
	ErrIndexAlreadyOpened = errors.New("index already opened")
	ErrIndexHasNotOpened  = errors.New("index has not opened")

	ErrShardDoesNotExist = errors.New("shard does not exist")

	ErrIndexDirectoryAlreadyExists = errors.New("index directory already exists")
	ErrIndexDirectoryDoesNotExist  = errors.New("index directory does not exist")
	ErrIndexDirectoryDoesNotMatch  = errors.New("index directory does not match")
	ErrUnsupportedDirectoryType    = errors.New("unsupported directory type")

	ErrDocumentIdDoesNotExist = errors.New("document ID does not exist")
	ErrInvalidDocument        = errors.New("invalid document")

	ErrInvalidUri = errors.New("invalid URI")

	ErrUnsupportedStorageType    = errors.New("unsupported metastore type")
	ErrUnsupportedMetastoreEvent = errors.New("unsupported metastore event")

	ErrIndexMetadataAlreadyExists = errors.New("index metadata already exists")
	ErrShardMetadataAlreadyExists = errors.New("shard metadata already exists")
	ErrIndexMetadataDoesNotExist  = errors.New("index metadata does not exist")
	ErrShardMetadataDoesNotExist  = errors.New("shard metadata does not exist")
	ErrInvalidIndexMetadata       = errors.New("invalid index metadata")
	ErrInvalidShardMetadata       = errors.New("invalid shard metadata")

	ErrShardWritersDoNotExist  = errors.New("shard writers do not exist")
	ErrShardWriterDoesNotExist = errors.New("shard writer does not exist")
	ErrShardReadersDoNotExist  = errors.New("shard readers do not exist")
	ErrShardReaderDoesNotExist = errors.New("shard reader does not exist")

	ErrUnsupportedLockManagerType = errors.New("unsupported lock manager type")
	ErrAlreadyLocked              = errors.New("already locked")
	ErrLockDoesNotExists          = errors.New("lock does not exists")
	ErrLockFailed                 = errors.New("lock failed")

	ErrUnknownFieldType         = errors.New("unknown field type")
	ErrFieldSettingDoesNotExist = errors.New("field setting does not exist")
	ErrUnexpectedFieldSetting   = errors.New("unexpected field setting")
	ErrLockUriIsNotSupported    = errors.New("lock URI is not supported")

	ErrUnknownQueryType = errors.New("unknown query type")

	ErrNodeDoesNotFound = errors.New("node not found")
	ErrInvalidData      = errors.New("invalid data")

	ErrInvalidCredentials = errors.New("invalid credentials")
)
