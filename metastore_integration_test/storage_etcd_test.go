//go:build integration

package metastore_integration_test

import (
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/mosuka/phalanx/logging"
	"github.com/mosuka/phalanx/metastore"
	"github.com/thanhpk/randstr"
)

func TestEtcdStorageWithUri(t *testing.T) {
	err := godotenv.Load(filepath.FromSlash("../.env"))
	if err != nil {
		t.Errorf("Failed to load .env file")
	}

	tmpDir := randstr.String(8)
	uri := fmt.Sprintf("etcd://phalanx-test/metastore/newtest/%s", tmpDir)
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	etcdStorage, err := metastore.NewEtcdStorageWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer etcdStorage.Close()
}

func TestEtcdStoragePut(t *testing.T) {
	err := godotenv.Load(filepath.FromSlash("../.env"))
	if err != nil {
		t.Errorf("Failed to load .env file")
	}

	tmpDir := randstr.String(8)
	uri := fmt.Sprintf("etcd://phalanx-test/metastore/puttest/%s", tmpDir)
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
	err := godotenv.Load(filepath.FromSlash("../.env"))
	if err != nil {
		t.Errorf("Failed to load .env file")
	}

	tmpDir := randstr.String(8)
	uri := fmt.Sprintf("etcd://phalanx-test/metastore/gettest/%s", tmpDir)
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
	err := godotenv.Load(filepath.FromSlash("../.env"))
	if err != nil {
		t.Errorf("Failed to load .env file")
	}

	tmpDir := randstr.String(8)
	uri := fmt.Sprintf("etcd://phalanx-test/metastore/deletetest/%s", tmpDir)
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
	err := godotenv.Load(filepath.FromSlash("../.env"))
	if err != nil {
		t.Errorf("Failed to load .env file")
	}

	tmpDir := randstr.String(8)
	uri := fmt.Sprintf("etcd://phalanx-test/metastore/existstest/%s", tmpDir)
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
	err := godotenv.Load(filepath.FromSlash("../.env"))
	if err != nil {
		t.Errorf("Failed to load .env file")
	}

	tmpDir := randstr.String(8)
	uri := fmt.Sprintf("etcd://phalanx-test/metastore/listtest/%s", tmpDir)
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

func TestEtcdStorageEvents(t *testing.T) {
	err := godotenv.Load(filepath.FromSlash("../.env"))
	if err != nil {
		t.Errorf("Failed to load .env file")
	}

	tmpDir := randstr.String(8)
	uri := fmt.Sprintf("etcd://phalanx-test/metastore/eventstest/%s", tmpDir)
	logger := logging.NewLogger("INFO", "", 500, 3, 30, false)

	etcdStorage, err := metastore.NewEtcdStorageWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer etcdStorage.Close()

	eventList := make([]metastore.StorageEvent, 0)
	done := make(chan bool)

	events := etcdStorage.Events()

	go func() {
		for {
			select {
			case cancel := <-done:
				// check
				if cancel {
					return
				}
			case event := <-events:
				eventList = append(eventList, event)
			}
		}
	}()

	etcdStorage.Put("/hello.txt", []byte("hello"))
	etcdStorage.Put("/world.txt", []byte("world"))

	// wait for events to be processed
	time.Sleep(1 * time.Second)

	done <- true

	actual := len(eventList)
	expected := 2
	if actual != expected {
		t.Fatalf("expected %v, but %v\n", expected, actual)
	}
}
