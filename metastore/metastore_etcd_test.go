package metastore

import (
	"reflect"
	"sort"
	"testing"

	"github.com/mosuka/phalanx/logging"
)

func TestEtcdMetastoreWithUri(t *testing.T) {
	uri := "etcd://phalanx-test/metastore?endpoints=localhost:2379"
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	_, err := NewEtcdMetastoreWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
}

func TestEtcdMetastorePut(t *testing.T) {
	uri := "etcd://phalanx-test/metastore?endpoints=localhost:2379"
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	metastore, err := NewEtcdMetastoreWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	if err := metastore.Put("/wikipedia_en.json", []byte("{}")); err != nil {
		t.Fatalf("%v\n", err)
	}
}

func TestEtcdMetastoreGet(t *testing.T) {
	uri := "etcd://phalanx-test/metastore?endpoints=localhost:2379"
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	metastore, err := NewEtcdMetastoreWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	metastore.Put("/wikipedia_en.json", []byte("{}"))

	content, err := metastore.Get("/wikipedia_en.json")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	if string(content) != "{}" {
		t.Fatalf("the data has not been written correctly. %v\n", string(content))
	}
}

func TestEtcdMetastoreDelete(t *testing.T) {
	uri := "etcd://phalanx-test/metastore?endpoints=localhost:2379"
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	metastore, err := NewEtcdMetastoreWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	metastore.Put("/wikipedia_en.json", []byte("{}"))

	if err := metastore.Delete("/wikipedia_en.json"); err != nil {
		t.Fatalf("%v\n", err)
	}
}

func TestEtcdMetastoreExists(t *testing.T) {
	uri := "etcd://phalanx-test/metastore?endpoints=localhost:2379"
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	metastore, err := NewEtcdMetastoreWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	if exists, err := metastore.Exists("/wikipedia_en.json"); err != nil {
		t.Fatalf("%v\n", err)
	} else {
		if exists != false {
			t.Fatalf("expect false, but %v\n", exists)
		}
	}

	metastore.Put("/wikipedia_en.json", []byte("{}"))

	if exists, err := metastore.Exists("/wikipedia_en.json"); err != nil {
		t.Fatalf("%v\n", err)
	} else {
		if !exists {
			t.Fatalf("expect true, but %v\n", exists)
		}
	}
}

func TestEtcdMetastoreList(t *testing.T) {
	uri := "etcd://phalanx-test/metastore?endpoints=localhost:2379"
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	metastore, err := NewEtcdMetastoreWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	metastore.Put("/wikipedia_en.json", []byte("{}"))
	metastore.Put("/wikipedia_ja.json", []byte("{}"))

	paths, err := metastore.List("/")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	sort.Slice(paths, func(i, j int) bool { return paths[i] < paths[j] })

	if !reflect.DeepEqual(paths, []string{"/wikipedia_en.json", "/wikipedia_ja.json"}) {
		t.Fatalf("unexpected %v\v", paths)
	}
}
