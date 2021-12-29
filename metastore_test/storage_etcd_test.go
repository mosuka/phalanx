//go:build integration

package metastore_test

import (
	"reflect"
	"sort"
	"testing"

	"github.com/mosuka/phalanx/logging"
	"github.com/mosuka/phalanx/metastore"
)

func TestEtcdStorageWithUri(t *testing.T) {
	uri := "etcd://phalanx-test/metastore?endpoints=localhost:2379"
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	etcdStorage, err := metastore.NewEtcdStorageWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer etcdStorage.Close()
}

func TestEtcdStoragePut(t *testing.T) {
	uri := "etcd://phalanx-test/metastore/puttest?endpoints=localhost:2379"
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
	uri := "etcd://phalanx-test/metastore/gettest?endpoints=localhost:2379"
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
	uri := "etcd://phalanx-test/metastore/deletetest?endpoints=localhost:2379"
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
	uri := "etcd://phalanx-test/metastore/existstest?endpoints=localhost:2379"
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
	uri := "etcd://phalanx-test/metastore/listtest?endpoints=localhost:2379"
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
