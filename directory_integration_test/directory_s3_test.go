package directory_integration_test_test

import (
	"testing"

	"github.com/mosuka/phalanx/directory"
	"github.com/mosuka/phalanx/logging"
)

func TestNewS3DirectoryWithUri(t *testing.T) {
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	uri := "s3://phalanx-test/indexes/test?endpoint=http://localhost:4566"
	lockUri := "etcd://phalanx-test/locks/test?endpoints=localhost:2379"

	directory := directory.NewS3DirectoryWithUri(uri, lockUri, logger)
	if directory == nil {
		t.Fatalf("failed to create S3 directory\n")
	}
}

func TestS3DirectorySetup(t *testing.T) {
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	uri := "s3://phalanx-test/indexes/test?endpoint=http://localhost:4566"
	lockUri := "etcd://phalanx-test/locks/test?endpoints=localhost:2379"

	directory := directory.NewS3DirectoryWithUri(uri, lockUri, logger)
	if directory == nil {
		t.Fatalf("failed to create S3 directory\n")
	}

	if err := directory.Setup(false); err != nil {
		t.Fatalf("failed to setup S3 directory\n")
	}
}
