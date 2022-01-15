//go:build integration

package lock_integration_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
	"github.com/mosuka/phalanx/lock"
	"github.com/mosuka/phalanx/logging"
	"github.com/thanhpk/randstr"
)

func TestNewDynamoDBLockManagerWithUri(t *testing.T) {
	err := godotenv.Load(filepath.FromSlash("../.env"))
	if err != nil {
		t.Errorf("Failed to load .env file")
	}

	tmpDir := randstr.String(8)
	uri := fmt.Sprintf("dynamodb://phalanx-locks-test/%s", tmpDir)
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	dynamoLock, err := lock.NewDynamoDBLockManagerWithUri(uri, logger)
	if err != nil {
		t.Fatalf("error %v\n", err)
	}
	defer dynamoLock.Close()
}

func TestDynamoDBLockManagerLock(t *testing.T) {
	err := godotenv.Load(filepath.FromSlash("../.env"))
	if err != nil {
		t.Errorf("Failed to load .env file")
	}

	tmpDir := randstr.String(8)
	uri := fmt.Sprintf("dynamodb://phalanx-locks-test/%s", tmpDir)
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	dynamoLock, err := lock.NewDynamoDBLockManagerWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer dynamoLock.Close()

	if _, err := dynamoLock.Lock(); err != nil {
		t.Fatalf("%v\n", err)
	}

	defer dynamoLock.Unlock()
}

func TestDynamoDBLockManagerLockTimeout(t *testing.T) {
	err := godotenv.Load(filepath.FromSlash("../.env"))
	if err != nil {
		t.Errorf("Failed to load .env file")
	}

	tmpDir := randstr.String(8)
	uri := fmt.Sprintf("dynamodb://phalanx-locks-test/%s", tmpDir)
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	dynamoLock1, err := lock.NewDynamoDBLockManagerWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer dynamoLock1.Close()

	dynamoLock2, err := lock.NewDynamoDBLockManagerWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer dynamoLock2.Close()

	if _, err := dynamoLock1.Lock(); err != nil {
		t.Fatalf("%v\n", err)
	}

	_, err = dynamoLock2.Lock()
	if err == nil {
		t.Fatalf("expect error: contextdeadline exceeded\n")
	}

	defer dynamoLock1.Unlock()
	defer dynamoLock2.Unlock()
}
