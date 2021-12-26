package metastore_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

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

	if err := fsMetastore.Put("/hello.txt", []byte("world")); err != nil {
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

	fsMetastore.Put("/hello.txt", []byte("hello"))

	content, err := fsMetastore.Get("/hello.txt")
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

	fsMetastore.Put("/hello.txt", []byte("hello"))

	err = fsMetastore.Delete("/hello.txt")
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

	if exists, err := fsMetastore.Exists("/hello.txt"); err != nil {
		t.Fatalf("%v\n", err)
	} else {
		if exists != false {
			t.Fatalf("expect false, but %v\n", exists)
		}
	}

	fsMetastore.Put("/hello.txt", []byte("hello"))

	if exists, err := fsMetastore.Exists("/hello.txt"); err != nil {
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

	fsMetastore.Put("/hello.txt", []byte("hello"))
	fsMetastore.Put("/world.txt", []byte("world"))

	paths, err := fsMetastore.List("/")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	sort.Slice(paths, func(i, j int) bool { return paths[i] < paths[j] })

	if !reflect.DeepEqual(paths, []string{"/hello.txt", "/world.txt"}) {
		t.Fatalf("unexpected %v\v", paths)
	}
}
