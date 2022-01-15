package directory_integration_test_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/mosuka/phalanx/directory"
	"github.com/mosuka/phalanx/logging"
)

func TestNewFileSystemDirectoryWithUri(t *testing.T) {
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	tmpDir, err := ioutil.TempDir("", "phalanx-test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer os.RemoveAll(tmpDir)

	path := filepath.ToSlash(tmpDir)
	uri := "file://" + path

	directory := directory.NewFileSystemDirectoryWithUri(uri, logger)
	if directory == nil {
		t.Fatalf("failed to create S3 directory\n")
	}
}
