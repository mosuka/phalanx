package metastore

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/mosuka/phalanx/logging"
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

	_, err = NewFileSystemStorageWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

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

	_, err = NewFileSystemStorageWithPath(path, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

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

	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	path := filepath.ToSlash(tmpDir)

	metastore, err := NewFileSystemStorageWithPath(path, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	if err := metastore.Put("/hello.txt", []byte("hello")); err != nil {
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

	metastore, err := NewFileSystemStorageWithPath(path, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	metastore.Put("/hello.txt", []byte("hello"))

	content, err := metastore.Get("/hello.txt")
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

	metastore, err := NewFileSystemStorageWithPath(path, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	metastore.Put("/hello.txt", []byte("hello"))

	err = metastore.Delete("/hello.txt")
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

	metastore, err := NewFileSystemStorageWithPath(path, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	if exists, err := metastore.Exists("/hello.txt"); err != nil {
		t.Fatalf("%v\n", err)
	} else {
		if exists != false {
			t.Fatalf("expect false, but %v\n", exists)
		}
	}

	metastore.Put("/hello.txt", []byte("hello"))

	if exists, err := metastore.Exists("/hello.txt"); err != nil {
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

	metastore, err := NewFileSystemStorageWithPath(path, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	metastore.Put("/hello.txt", []byte("hello"))
	metastore.Put("/world.txt", []byte("world"))

	paths, err := metastore.List("/")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	sort.Slice(paths, func(i, j int) bool { return paths[i] < paths[j] })

	if !reflect.DeepEqual(paths, []string{"/hello.txt", "/world.txt"}) {
		t.Fatalf("unexpected %v\v", paths)
	}
}
