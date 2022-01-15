package directory_integration_test_test

import (
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
	"github.com/mosuka/phalanx/directory"
	"github.com/mosuka/phalanx/logging"
)

func TestNewMinioDirectoryWithUri(t *testing.T) {
	err := godotenv.Load(filepath.FromSlash("../.env"))
	if err != nil {
		t.Errorf("Failed to load .env file")
	}

	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	uri := "minio://phalanx-test/indexes/test"
	lockUri := "etcd://phalanx-test/locks/test"

	directory := directory.NewMinioDirectoryWithUri(uri, lockUri, logger)
	if directory == nil {
		t.Fatalf("failed to create MinIO directory\n")
	}
}

func TestMinIODirectorySetup(t *testing.T) {
	err := godotenv.Load(filepath.FromSlash("../.env"))
	if err != nil {
		t.Errorf("Failed to load .env file")
	}

	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	uri := "minio://phalanx-test/indexes/test"
	lockUri := "etcd://phalanx-test/locks/test"

	directory := directory.NewMinioDirectoryWithUri(uri, lockUri, logger)
	if directory == nil {
		t.Fatalf("failed to create S3 directory\n")
	}

	if err := directory.Setup(false); err != nil {
		t.Fatalf("failed to setup S3 directory\n")
	}
}
