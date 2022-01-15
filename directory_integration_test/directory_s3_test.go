package directory_integration_test_test

import (
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
	"github.com/mosuka/phalanx/directory"
	"github.com/mosuka/phalanx/logging"
)

func TestNewS3DirectoryWithUri(t *testing.T) {
	err := godotenv.Load(filepath.FromSlash("../.env"))
	if err != nil {
		t.Errorf("Failed to load .env file")
	}

	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	uri := "s3://phalanx-test/indexes/test"
	lockUri := "etcd://phalanx-test/locks/test"

	directory := directory.NewS3DirectoryWithUri(uri, lockUri, logger)
	if directory == nil {
		t.Fatalf("failed to create S3 directory\n")
	}
}

func TestS3DirectorySetup(t *testing.T) {
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	uri := "s3://phalanx-test/indexes/test"
	lockUri := "etcd://phalanx-test/locks/test"

	directory := directory.NewS3DirectoryWithUri(uri, lockUri, logger)
	if directory == nil {
		t.Fatalf("failed to create S3 directory\n")
	}

	if err := directory.Setup(false); err != nil {
		t.Fatalf("failed to setup S3 directory\n")
	}
}
