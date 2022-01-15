//go:build integration

package clients_integration_test

import (
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
	"github.com/mosuka/phalanx/clients"
)

func TestNewDynamoDBClientWithUri(t *testing.T) {
	err := godotenv.Load(filepath.FromSlash("../.env"))
	if err != nil {
		t.Errorf("Failed to load .env file")
	}

	uri := "dynamodb://phalanx-locks-test/index/shard"

	if _, err := clients.NewDynamoDBClientWithUri(uri); err != nil {
		t.Fatalf("error %v\n", err)
	}
}
