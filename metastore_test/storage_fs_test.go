package metastore_test

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/mosuka/phalanx/logging"
	"github.com/mosuka/phalanx/metastore"
	"github.com/mosuka/phalanx/util"
)

func TestNewFileSystemStorageWithUri(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "phalanx-test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer os.RemoveAll(tmpDir)

	path := filepath.ToSlash(tmpDir)
	uri := "file://" + path

	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	fsMetastore, err := metastore.NewFileSystemStorageWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer fsMetastore.Close()

	if !util.FileExists(path) {
		t.Fatalf("directory does not exist.\n")
	}
}

func TestNewFileSystemStorageWithPath(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "phalanx-test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer os.RemoveAll(tmpDir)

	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	path := filepath.ToSlash(tmpDir)

	fsMetastore, err := metastore.NewFileSystemStorageWithPath(path, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer fsMetastore.Close()

	if !util.FileExists(path) {
		t.Fatalf("directory does not exist.\n")
	}
}

func TestFileSystemStoragePut(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "phalanx-test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer os.RemoveAll(tmpDir)

	path := filepath.ToSlash(tmpDir)
	uri := "file://" + path

	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	fsMetastore, err := metastore.NewFileSystemStorageWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer fsMetastore.Close()

	if err := fsMetastore.Put(context.Background(), "/hello.txt", []byte("world")); err != nil {
		t.Fatalf("%v\n", err)
	}
}

func TestFileSystemStorageGet(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "phalanx-test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer os.RemoveAll(tmpDir)

	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	path := filepath.ToSlash(tmpDir)

	fsMetastore, err := metastore.NewFileSystemStorageWithPath(path, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer fsMetastore.Close()

	ctx := context.Background()

	fsMetastore.Put(ctx, "/hello.txt", []byte("hello"))

	content, err := fsMetastore.Get(ctx, "/hello.txt")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	if string(content) != "hello" {
		t.Fatalf("the data has not been written correctly. %v\n", string(content))
	}
}

func TestFileSystemStorageDelete(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "phalanx-test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer os.RemoveAll(tmpDir)

	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	path := filepath.ToSlash(tmpDir)

	fsMetastore, err := metastore.NewFileSystemStorageWithPath(path, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer fsMetastore.Close()

	ctx := context.Background()

	fsMetastore.Put(ctx, "/hello.txt", []byte("hello"))

	err = fsMetastore.Delete(ctx, "/hello.txt")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
}

func TestFileSystemStorageExists(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "phalanx-test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer os.RemoveAll(tmpDir)

	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	path := filepath.ToSlash(tmpDir)

	fsMetastore, err := metastore.NewFileSystemStorageWithPath(path, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer fsMetastore.Close()

	ctx := context.Background()

	if exists, err := fsMetastore.Exists(ctx, "/hello.txt"); err != nil {
		t.Fatalf("%v\n", err)
	} else {
		if exists != false {
			t.Fatalf("expect false, but %v\n", exists)
		}
	}

	fsMetastore.Put(ctx, "/hello.txt", []byte("hello"))

	if exists, err := fsMetastore.Exists(ctx, "/hello.txt"); err != nil {
		t.Fatalf("%v\n", err)
	} else {
		if !exists {
			t.Fatalf("expect true, but %v\n", exists)
		}
	}
}

func TestFileSystemStorageList(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "phalanx-test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer os.RemoveAll(tmpDir)

	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	path := filepath.ToSlash(tmpDir)

	fsMetastore, err := metastore.NewFileSystemStorageWithPath(path, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer fsMetastore.Close()

	ctx := context.Background()

	fsMetastore.Put(ctx, "/hello.txt", []byte("hello"))
	fsMetastore.Put(ctx, "/world.txt", []byte("world"))

	paths, err := fsMetastore.List(ctx, "/")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	sort.Slice(paths, func(i, j int) bool { return paths[i] < paths[j] })

	if !reflect.DeepEqual(paths, []string{filepath.FromSlash("/hello.txt"), filepath.FromSlash("/world.txt")}) {
		t.Fatalf("unexpected %v\v", paths)
	}
}

func TestFileSystemStorageEvents(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "phalanx-test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer os.RemoveAll(tmpDir)

	logger := logging.NewLogger("INFO", "", 500, 3, 30, false)

	path := filepath.ToSlash(tmpDir)

	fsMetastore, err := metastore.NewFileSystemStorageWithPath(path, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer fsMetastore.Close()

	eventList := make([]metastore.StorageEvent, 0)
	done := make(chan bool)

	events := fsMetastore.Events()

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

	ctx := context.Background()

	fsMetastore.Put(ctx, "/hello.txt", []byte("hello"))
	fsMetastore.Put(ctx, "/hello2.txt", []byte("hello2"))

	// wait for events to be processed
	time.Sleep(5 * time.Second)

	done <- true

	actual := len(eventList)
	expected := 2
	if actual != expected {
		t.Fatalf("expected %v, but %v\n", expected, actual)
	}
}
