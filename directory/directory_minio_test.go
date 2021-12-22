package directory

// import (
// 	"testing"

// 	"github.com/blugelabs/bluge/index"
// 	"github.com/mosuka/phalanx/lock"
// 	"github.com/mosuka/phalanx/logging"
// )

// func TestNewMinioDirectoryWithUri(t *testing.T) {
// 	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

// 	lock_uri := "etcd://phalanx-test/locks/wikipedia_en?endpoints=localhost:2379"
// 	lockManager, err := lock.NewLockManagerWithUri(lock_uri, logger)
// 	if err != nil {
// 		t.Fatalf("%v\n", err)
// 	}

// 	uri := "minio://phalanx-test/indexes/wikipedia_en?endpoint=localhost:9000&access_key=minio&secret_key=miniosecret&secure=false&region=us-east-1"
// 	directory := NewMinioDirectoryWithUri(uri, lockManager, logger)
// 	if directory == nil {
// 		t.Fatalf("failed to create MinIO directory\n")
// 	}
// }

// func TestMinioDirectorySetup(t *testing.T) {
// 	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

// 	lock_uri := "etcd://phalanx-test/locks/wikipedia_en?endpoints=localhost:2379"
// 	lockManager, err := lock.NewLockManagerWithUri(lock_uri, logger)
// 	if err != nil {
// 		t.Fatalf("%v\n", err)
// 	}

// 	uri := "minio://phalanx-test/indexes/wikipedia_en?endpoint=localhost:9000&access_key=minio&secret_key=miniosecret&secure=false&region=us-east-1"
// 	directory := NewMinioDirectoryWithUri(uri, lockManager, logger)
// 	if directory == nil {
// 		t.Fatalf("failed to create MinIO directory\n")
// 	}

// 	if err := directory.Setup(false); err != nil {
// 		t.Fatalf("%v\n", err)
// 	}
// }

// func TestMinioDirectoryList(t *testing.T) {
// 	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

// 	lock_uri := "etcd://phalanx-test/locks/wikipedia_en?endpoints=localhost:2379"
// 	lockManager, err := lock.NewLockManagerWithUri(lock_uri, logger)
// 	if err != nil {
// 		t.Fatalf("%v\n", err)
// 	}

// 	uri := "minio://phalanx-test/indexes/wikipedia_en?endpoint=localhost:9000&access_key=minio&secret_key=miniosecret&secure=false&region=us-east-1"
// 	directory := NewMinioDirectoryWithUri(uri, lockManager, logger)
// 	if directory == nil {
// 		t.Fatalf("failed to create MinIO directory\n")
// 	}

// 	if err := directory.Setup(false); err != nil {
// 		t.Fatalf("%v\n", err)
// 	}

// 	kind := index.ItemKindSegment

// 	_, err = directory.List(kind)
// 	if err != nil {
// 		t.Fatalf("%v\n", err)
// 	}
// }

// func TestMinioDirectoryLoad(t *testing.T) {
// 	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

// 	lock_uri := "etcd://phalanx-test/locks/wikipedia_en?endpoints=localhost:2379"
// 	lockManager, err := lock.NewLockManagerWithUri(lock_uri, logger)
// 	if err != nil {
// 		t.Fatalf("%v\n", err)
// 	}

// 	uri := "minio://phalanx-test/indexes/wikipedia_en?endpoint=localhost:9000&access_key=minio&secret_key=miniosecret&secure=false&region=us-east-1"
// 	directory := NewMinioDirectoryWithUri(uri, lockManager, logger)
// 	if directory == nil {
// 		t.Fatalf("failed to create MinIO directory\n")
// 	}

// 	if err := directory.Setup(false); err != nil {
// 		t.Fatalf("%v\n", err)
// 	}

// 	kind := index.ItemKindSegment

// 	ids, err := directory.List(kind)
// 	if err != nil {
// 		t.Fatalf("%v\n", err)
// 	}

// 	id := ids[0]

// 	_, _, err = directory.Load(kind, id)
// 	if err != nil {
// 		t.Fatalf("%v\n", err)
// 	}
// }

// func TestMinioDirectoryLock(t *testing.T) {
// 	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

// 	lock_uri := "etcd://phalanx-test/locks/wikipedia_en?endpoints=localhost:2379"
// 	lockManager, err := lock.NewLockManagerWithUri(lock_uri, logger)
// 	if err != nil {
// 		t.Fatalf("%v\n", err)
// 	}

// 	uri := "minio://phalanx-test/indexes/wikipedia_en?endpoint=localhost:9000&access_key=minio&secret_key=miniosecret&secure=false&region=us-east-1"
// 	directory := NewMinioDirectoryWithUri(uri, lockManager, logger)
// 	if directory == nil {
// 		t.Fatalf("failed to create MinIO directory\n")
// 	}

// 	if err := directory.Setup(false); err != nil {
// 		t.Fatalf("%v\n", err)
// 	}

// 	if err := directory.Lock(); err != nil {
// 		t.Fatalf("%v\n", err)
// 	}
// 	defer directory.Unlock()
// }
