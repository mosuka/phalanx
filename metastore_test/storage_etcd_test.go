package metastore_test

import (
	"fmt"
	"net/url"
	"reflect"
	"sort"
	"testing"

	"github.com/mosuka/phalanx/logging"
	"github.com/mosuka/phalanx/metastore"
	"go.etcd.io/etcd/pkg/testutil"
	"go.etcd.io/etcd/tests/v3/integration"
)

func TestEtcdStorageWithUri(t *testing.T) {
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

	uri := fmt.Sprintf("etcd://phalanx-test/metastore?endpoints=%s", endpoints)
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	etcdStorage, err := metastore.NewEtcdStorageWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer etcdStorage.Close()
}

func TestEtcdStoragePut(t *testing.T) {
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

	uri := fmt.Sprintf("etcd://phalanx-test/metastore?endpoints=%s", endpoints)
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	etcdStorage, err := metastore.NewEtcdStorageWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer etcdStorage.Close()

	etcdStorage.Put("/wikipedia_en.json", []byte("{}"))
	if err != nil {
		t.Fatalf("%v\n", err)
	}
}

func TestEtcdStorageGet(t *testing.T) {
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

	uri := fmt.Sprintf("etcd://phalanx-test/metastore?endpoints=%s", endpoints)
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	etcdStorage, err := metastore.NewEtcdStorageWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer etcdStorage.Close()

	etcdStorage.Put("/wikipedia_en.json", []byte("{}"))

	content, err := etcdStorage.Get("/wikipedia_en.json")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	if string(content) != "{}" {
		t.Fatalf("unexpected value. %v\n", string(content))
	}
}

func TestEtcdStorageDelete(t *testing.T) {
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

	uri := fmt.Sprintf("etcd://phalanx-test/metastore?endpoints=%s", endpoints)
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	etcdStorage, err := metastore.NewEtcdStorageWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer etcdStorage.Close()

	etcdStorage.Put("/wikipedia_en.json", []byte("{}"))
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	etcdStorage.Delete("/wikipedia_en.json")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
}

func TestEtcdStorageExists(t *testing.T) {
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

	uri := fmt.Sprintf("etcd://phalanx-test/metastore?endpoints=%s", endpoints)
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	etcdStorage, err := metastore.NewEtcdStorageWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer etcdStorage.Close()

	exists, err := etcdStorage.Exists("/wikipedia_en.json")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	if exists != false {
		t.Fatalf("unexpected value. %v\n", exists)
	}

	etcdStorage.Put("/wikipedia_en.json", []byte("{}"))
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	exists, err = etcdStorage.Exists("/wikipedia_en.json")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	if exists != true {
		t.Fatalf("unexpected value. %v\n", exists)
	}
}

func TestEtcdStorageList(t *testing.T) {
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

	uri := fmt.Sprintf("etcd://phalanx-test/metastore?endpoints=%s", endpoints)
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	etcdStorage, err := metastore.NewEtcdStorageWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer etcdStorage.Close()

	etcdStorage.Put("/hello.txt", []byte("hello"))
	etcdStorage.Put("/world.txt", []byte("world"))

	paths, err := etcdStorage.List("/")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	sort.Slice(paths, func(i, j int) bool { return paths[i] < paths[j] })

	if !reflect.DeepEqual(paths, []string{"/hello.txt", "/world.txt"}) {
		t.Fatalf("unexpected %v\v", paths)
	}
}

func TestEtcdStorageWatch(t *testing.T) {
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

	uri := fmt.Sprintf("etcd://phalanx-test/metastore?endpoints=%s", endpoints)
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	etcdStorage, err := metastore.NewEtcdStorageWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer etcdStorage.Close()

	etcdStorage.Put("/test/hello.txt", []byte("hello"))
	etcdStorage.Delete("/test/hello.txt")

	event := <-etcdStorage.Events()
	expected := metastore.StorageEventTypePut
	if event.Type != expected {
		t.Fatalf("%v\n", err)
	}

	event = <-etcdStorage.Events()
	expected = metastore.StorageEventTypeDelete
	if event.Type != expected {
		t.Fatalf("%v\n", err)
	}
}
