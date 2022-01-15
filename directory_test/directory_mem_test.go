package directory_integration_test_test

import (
	"github.com/thanhpk/randstr"
	"testing"

	"github.com/mosuka/phalanx/directory"
	"github.com/mosuka/phalanx/logging"
)

func TestNewInMemoryDirectoryWithUri(t *testing.T) {
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	path := randstr.String(8)
	uri := "mem://" + path

	directory := directory.NewInMemoryDirectoryWithUri(uri, logger)
	if directory == nil {
		t.Fatalf("failed to create In-Memory directory\n")
	}
}
