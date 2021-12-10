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

func TestNewFileMetastoreWithUri(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "phalanx-test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer os.RemoveAll(tmpDir)

	path := filepath.ToSlash(tmpDir)
	uri := "file://" + path

	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	_, err = NewFileSystemMetastoreWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	if !util.FileExists(path) {
		t.Fatalf("directory does not exist.\n")
	}
}

func TestNewFileMetastoreWithPath(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "phalanx-test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer os.RemoveAll(tmpDir)

	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	path := filepath.ToSlash(tmpDir)

	_, err = NewFileSystemMetastoreWithPath(path, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	if !util.FileExists(path) {
		t.Fatalf("directory does not exist.\n")
	}
}

func TestFileMetastorePut(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "phalanx-test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer os.RemoveAll(tmpDir)

	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	path := filepath.ToSlash(tmpDir)

	metastore, err := NewFileSystemMetastoreWithPath(path, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	if err := metastore.Put("/hello.txt", []byte("hello")); err != nil {
		t.Fatalf("%v\n", err)
	}
}

func TestFileMetastoreGet(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "phalanx-test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer os.RemoveAll(tmpDir)

	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	path := filepath.ToSlash(tmpDir)

	metastore, err := NewFileSystemMetastoreWithPath(path, logger)
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

func TestFileMetastoreDelete(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "phalanx-test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer os.RemoveAll(tmpDir)

	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	path := filepath.ToSlash(tmpDir)

	metastore, err := NewFileSystemMetastoreWithPath(path, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	metastore.Put("/hello.txt", []byte("hello"))

	err = metastore.Delete("/hello.txt")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
}

func TestFileMetastoreExists(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "phalanx-test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer os.RemoveAll(tmpDir)

	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	path := filepath.ToSlash(tmpDir)

	metastore, err := NewFileSystemMetastoreWithPath(path, logger)
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

func TestFileMetastoreList(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "phalanx-test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer os.RemoveAll(tmpDir)

	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	path := filepath.ToSlash(tmpDir)

	metastore, err := NewFileSystemMetastoreWithPath(path, logger)
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
