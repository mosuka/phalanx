package metastore_test

import (
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mosuka/phalanx/logging"
	"github.com/mosuka/phalanx/metastore"
	mock_metastore "github.com/mosuka/phalanx/mock/metastore"
)

func TestNewMetastore(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockStorage := mock_metastore.NewMockStorage(mockCtrl)

	mockStorage.EXPECT().List(filepath.FromSlash("/")).Return([]string{}, nil)

	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	_, err := metastore.NewMetastore(mockStorage, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
}
