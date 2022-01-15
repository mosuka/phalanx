//go:build integration

package clients_integration_test

import (
	"testing"

	"github.com/mosuka/phalanx/clients"
)

func TestNewEtcdClientWithUri(t *testing.T) {
	uri := "etcd://phalanx-test/metastore/test?endpoints=localhost:2379"

	if _, err := clients.NewEtcdClientWithUri(uri); err != nil {
		t.Fatalf("error %v\n", err)
	}
}
