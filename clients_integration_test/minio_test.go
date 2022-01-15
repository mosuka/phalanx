//go:build integration

package clients_integration_test

import (
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
	"github.com/mosuka/phalanx/clients"
)

func TestNewMinioClientWithUri(t *testing.T) {
	err := godotenv.Load(filepath.FromSlash("../.env"))
	if err != nil {
		t.Errorf("Failed to load .env file")
	}

	uri := "minio://phalanx-test/indexes/test"

	if _, err := clients.NewMinioClientWithUri(uri); err != nil {
		t.Fatalf("error %v\n", err)
	}
}
