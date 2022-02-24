//go:build integration

package metastore_integration_test

import (
	"context"
	"fmt"
	"net/url"
	"os"
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

func TestDynamodbStorageWithUri(t *testing.T) {
	err := godotenv.Load(filepath.FromSlash("../.env"))
	if err != nil {
		t.Errorf("Failed to load .env file")
	}

	tmpDir := randstr.String(8)
	uri := fmt.Sprintf("dynamodb://phalanx-test/metastore/newtest/%s?%s", tmpDir, buildQueryFromEnv())
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	dynamodbStorage, err := metastore.NewDynamodbStorageWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer dynamodbStorage.Close()
}

func TestDynamodbStorageGet(t *testing.T) {
	err := godotenv.Load(filepath.FromSlash("../.env"))
	if err != nil {
		t.Errorf("Failed to load .env file")
	}

	tmpDir := randstr.String(8)
	uri := fmt.Sprintf("dynamodb://phalanx-test/metastore/newtest/%s?%s", tmpDir, buildQueryFromEnv())
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	dynamodbStorage, err := metastore.NewDynamodbStorageWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer dynamodbStorage.Close()

	ctx := context.Background()

	_, err = dynamodbStorage.Get(ctx, "/wikipedia_en.json")
	if err != metastore.ErrRecordNotFound {
		t.Fatalf("unexpected value. %v\n", err)
	}

	err = dynamodbStorage.Put(ctx, "/wikipedia_en.json", []byte("{}"))
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	content, err := dynamodbStorage.Get(ctx, "/wikipedia_en.json")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	if string(content) != "{}" {
		t.Fatalf("unexpected value. %v\n", string(content))
	}
}

func TestDynamodbStoragePut(t *testing.T) {
	err := godotenv.Load(filepath.FromSlash("../.env"))
	if err != nil {
		t.Errorf("Failed to load .env file")
	}

	tmpDir := randstr.String(8)
	uri := fmt.Sprintf("dynamodb://phalanx-test/metastore/newtest/%s?%s", tmpDir, buildQueryFromEnv())
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	dynamodbStorage, err := metastore.NewDynamodbStorageWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer dynamodbStorage.Close()

	ctx := context.Background()

	err = dynamodbStorage.Put(ctx, "/wikipedia_en.json", []byte("{}"))
	if err != nil {
		t.Fatalf("%v\n", err)
	}
}

func TestDynamodbStorageDelete(t *testing.T) {
	err := godotenv.Load(filepath.FromSlash("../.env"))
	if err != nil {
		t.Errorf("Failed to load .env file")
	}

	tmpDir := randstr.String(8)
	uri := fmt.Sprintf("dynamodb://phalanx-test/metastore/newtest/%s?%s", tmpDir, buildQueryFromEnv())
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	dynamodbStorage, err := metastore.NewDynamodbStorageWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer dynamodbStorage.Close()

	ctx := context.Background()

	dynamodbStorage.Put(ctx, "/wikipedia_en.json", []byte("{}"))
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	dynamodbStorage.Delete(ctx, "/wikipedia_en.json")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
}

func TestDynamodbStorageExists(t *testing.T) {
	err := godotenv.Load(filepath.FromSlash("../.env"))
	if err != nil {
		t.Errorf("Failed to load .env file")
	}

	tmpDir := randstr.String(8)
	uri := fmt.Sprintf("dynamodb://phalanx-test/metastore/newtest/%s?%s", tmpDir, buildQueryFromEnv())
	logger := logging.NewLogger("INFO", "", 500, 3, 30, false)

	dynamodbStorage, err := metastore.NewDynamodbStorageWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer dynamodbStorage.Close()

	ctx := context.Background()

	exists, err := dynamodbStorage.Exists(ctx, "/wikipedia_en.json")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	if exists != false {
		t.Fatalf("unexpected value. %v\n", exists)
	}

	dynamodbStorage.Put(ctx, "/wikipedia_en.json", []byte("{}"))
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	exists, err = dynamodbStorage.Exists(ctx, "/wikipedia_en.json")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	if exists != true {
		t.Fatalf("unexpected value. %v\n", exists)
	}
}

func TestDynamodbStorageList(t *testing.T) {
	err := godotenv.Load(filepath.FromSlash("../.env"))
	if err != nil {
		t.Errorf("Failed to load .env file")
	}

	tmpDir := randstr.String(8)
	uri := fmt.Sprintf("dynamodb://phalanx-test/metastore/newtest/%s?%s", tmpDir, buildQueryFromEnv())
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	dynamodbStorage, err := metastore.NewDynamodbStorageWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer dynamodbStorage.Close()

	ctx := context.Background()

	dynamodbStorage.Put(ctx, "/hello.txt", []byte("hello"))
	dynamodbStorage.Put(ctx, "/world.txt", []byte("world"))

	paths, err := dynamodbStorage.List(ctx, "/")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	sort.Slice(paths, func(i, j int) bool { return paths[i] < paths[j] })

	if !reflect.DeepEqual(paths, []string{"/hello.txt", "/world.txt"}) {
		t.Fatalf("unexpected %v\v", paths)
	}
}

func TestDynamodbStorageStorageEvents(t *testing.T) {
	err := godotenv.Load(filepath.FromSlash("../.env"))
	if err != nil {
		t.Errorf("Failed to load .env file")
	}

	tmpDir := randstr.String(8)
	uri := fmt.Sprintf("dynamodb://phalanx-test/metastore/eventstest/%s", tmpDir)
	logger := logging.NewLogger("INFO", "", 500, 3, 30, false)

	dynamodbStorage, err := metastore.NewDynamodbStorageWithUri(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer dynamodbStorage.Close()

	eventList := make([]metastore.StorageEvent, 0)
	done2 := make(chan bool)

	events := dynamodbStorage.Events()

	go func() {
		for {
			select {
			case cancel := <-done2:
				// check
				if cancel {
					return
				}
			case event := <-events:
				eventList = append(eventList, event)
			}
		}
	}()

	// Make changes to the database
	// dynamodbStorage.Put("/hello.txt", []byte("hello"))
	// dynamodbStorage.Put("/world.txt", []byte("world"))

	// wait for events to be processed
	time.Sleep(3 * time.Second)
	done2 <- true

	actual := len(eventList)
	expected := 0 // TODO: fix this
	if actual != expected {
		t.Fatalf("expected %v, but %v\n", expected, actual)
	}
}

func buildQueryFromEnv() string {

	vals := url.Values{}

	vals.Add("region", os.Getenv("AWS_DEFAULT_REGION"))
	vals.Add("endpoint_url", os.Getenv("AWS_ENDPOINT_URL"))
	vals.Add("access_key_id", os.Getenv("AWS_ACCESS_KEY_ID"))
	vals.Add("secret_access_key", os.Getenv("AWS_SECRET_ACCESS_KEY"))
	vals.Add("create_table", "true")

	return vals.Encode()
}
