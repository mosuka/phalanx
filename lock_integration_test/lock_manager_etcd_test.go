//go:build integration

package lock_integration_test

import (
	"testing"

	"github.com/mosuka/phalanx/lock"
	"github.com/mosuka/phalanx/logging"
)

func TestEtcdLockManagerWithUri(t *testing.T) {
	uri := "etcd://phalanx-test/locks/example_en/shard-05he7Bph?endpoints=localhost:2379"
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	etcdLock, err := lock.NewEtcdLockManagerWithUri(uri, logger)
	if err != nil {
		t.Fatalf("error %v\n", err)
	}
	defer etcdLock.Close()
}

func TestEtcdLockManagerLock(t *testing.T) {
	uri := "etcd://phalanx-test/locks/example_en/shard-05he7Bph?endpoints=localhost:2379"
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	etcdLock, err := lock.NewEtcdLockManagerWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer etcdLock.Close()

	rev, err := etcdLock.Lock()
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	if rev == 0 {
		t.Fatalf("expecting the revision greater than 0.\n")
	}

	defer etcdLock.Unlock()
}
