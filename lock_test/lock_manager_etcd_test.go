package lock_test

import (
	"fmt"
	"net/url"
	"runtime"
	"testing"

	"github.com/mosuka/phalanx/lock"
	"github.com/mosuka/phalanx/logging"
	"go.etcd.io/etcd/pkg/testutil"
	"go.etcd.io/etcd/tests/v3/integration"
)

func TestEtcdLockManagerWithUri(t *testing.T) {
	// Skip this test if windows.
	// See https://github.com/etcd-io/etcd/issues/10854
	if runtime.GOOS == "windows" {
		return
	}

	defer testutil.AfterTest(t)
	integration.BeforeTest(t)
	cluster := integration.NewClusterV3(t, &integration.ClusterConfig{Size: 1, UseTCP: true})
	defer cluster.Terminate(t)

	etcdEndpoints := cluster.RandClient().Endpoints()[0]
	u, err := url.Parse(etcdEndpoints)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	endpoints := fmt.Sprintf("%s:%s", u.Hostname(), u.Port())

	uri := fmt.Sprintf("etcd://phalanx-test/locks/example_en/shard-05he7Bph?endpoints=%s", endpoints)
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	etcdLock, err := lock.NewEtcdLockManagerWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer etcdLock.Close()
}

func TestEtcdLockManagerLock(t *testing.T) {
	// Skip this test if windows.
	// See https://github.com/etcd-io/etcd/issues/10854
	if runtime.GOOS == "windows" {
		return
	}

	defer testutil.AfterTest(t)
	integration.BeforeTest(t)
	cluster := integration.NewClusterV3(t, &integration.ClusterConfig{Size: 1, UseTCP: true})
	defer cluster.Terminate(t)

	etcdEndpoints := cluster.RandClient().Endpoints()[0]
	u, err := url.Parse(etcdEndpoints)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	endpoints := fmt.Sprintf("%s:%s", u.Hostname(), u.Port())

	uri := fmt.Sprintf("etcd://phalanx-test/locks/example_en/shard-05he7Bph?endpoints=%s", endpoints)
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
