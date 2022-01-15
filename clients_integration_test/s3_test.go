//go:build integration

package clients_integration_test

import (
	"testing"

	"github.com/mosuka/phalanx/clients"
)

func TestNewS3ClientWithUri(t *testing.T) {
	uri := "s3://phalanx-test/indexes/test?endpoint=http://localhost:4572"

	if _, err := clients.NewS3ClientWithUri(uri); err != nil {
		t.Fatalf("error %v\n", err)
	}
}
