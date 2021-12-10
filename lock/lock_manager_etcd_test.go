package lock

import (
	"testing"

	"github.com/mosuka/phalanx/logging"
)

func TestEtcdLockManagerWithUri(t *testing.T) {
	uri := "etcd://phalanx-test/locks/wikipedia_en?endpoints=localhost:2379"
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	_, err := NewEtcdLockManagerWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
}

func TestEtcdLockManagerLock(t *testing.T) {
	uri := "etcd://phalanx-test/locks/wikipedia_en?endpoints=localhost:2379"
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	lockManager, err := NewEtcdLockManagerWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	rev, err := lockManager.Lock()
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	if rev == 0 {
		t.Fatalf("expecting the revision greater than 0.\n")
	}

	lockManager.Unlock()
}

func TestEtcdLockManagerLockDeadlineExceeded(t *testing.T) {
	uri := "etcd://phalanx-test/locks/wikipedia_en?endpoints=localhost:2379"
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	lockManager1, err := NewEtcdLockManagerWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	lockManager2, err := NewEtcdLockManagerWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	rev, err := lockManager1.Lock()
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	if rev == 0 {
		t.Fatalf("expecting the revision greater than 0.\n")
	}

	_, err = lockManager2.Lock()
	if err == nil {
		t.Fatalf("expecting the context deadline exceeded.\n")
	}

	if err := lockManager1.Unlock(); err != nil {
		t.Fatalf("%v\n", err)
	}

	if err := lockManager2.Unlock(); err != nil {
		t.Fatalf("%v\n", err)
	}
}
