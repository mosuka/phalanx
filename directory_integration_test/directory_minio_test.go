package directory_integration_test_test

import (
	"testing"

	"github.com/mosuka/phalanx/directory"
	"github.com/mosuka/phalanx/logging"
)

func TestNewMinioDirectoryWithUri(t *testing.T) {
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	uri := "minio://phalanx-test/indexes/test?endpoint=localhost:9000&access_key=minio&secret_key=miniosecret&secure=false&region=us-east-1"
	lockUri := "etcd://phalanx-test/locks/test?endpoints=localhost:2379"

	directory := directory.NewMinioDirectoryWithUri(uri, lockUri, logger)
	if directory == nil {
		t.Fatalf("failed to create MinIO directory\n")
	}
}

func TestMinIODirectorySetup(t *testing.T) {
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	uri := "minio://phalanx-test/indexes/test?endpoint=localhost:9000&access_key=minio&secret_key=miniosecret&secure=false&region=us-east-1"
	lockUri := "etcd://phalanx-test/locks/test?endpoints=localhost:2379"

	directory := directory.NewMinioDirectoryWithUri(uri, lockUri, logger)
	if directory == nil {
		t.Fatalf("failed to create S3 directory\n")
	}

	if err := directory.Setup(false); err != nil {
		t.Fatalf("failed to setup S3 directory\n")
	}
}
