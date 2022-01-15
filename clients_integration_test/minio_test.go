//go:build integration

package clients_integration_test

import (
	"testing"

	"github.com/mosuka/phalanx/clients"
)

func TestNewMinioClientWithUri(t *testing.T) {
	uri := "minio://phalanx-test/indexes/test?endpoint=localhost:9000"

	if _, err := clients.NewMinioClientWithUri(uri); err != nil {
		t.Fatalf("error %v\n", err)
	}
}
