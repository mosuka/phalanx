////go:build integration

package metastore_integration_test

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

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

	dynamodbStorage, err := metastore.NewDynamodbStorage(uri, logger)
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

	dynamodbStorage, err := metastore.NewDynamodbStorage(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer dynamodbStorage.Close()

	_, err = dynamodbStorage.Get("/wikipedia_en.json")
	if err != metastore.ErrRecordNotFound {
		t.Fatalf("unexpected value. %v\n", err)
	}

	err = dynamodbStorage.Put("/wikipedia_en.json", []byte("{}"))
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	content, err := dynamodbStorage.Get("/wikipedia_en.json")
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

	dynamodbStorage, err := metastore.NewDynamodbStorage(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer dynamodbStorage.Close()

	err = dynamodbStorage.Put("/wikipedia_en.json", []byte("{}"))
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	err = dynamodbStorage.Put("/wikipedia_en.json", []byte("{}"))
	if err != metastore.ErrDuplicateRecord {
		t.Fatalf("unexpected value. %v\n", err)
	}

	content, err := dynamodbStorage.Get("/wikipedia_en.json")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	if string(content) != "{}" {
		t.Fatalf("unexpected value. %v\n", string(content))
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

	dynamodbStorage, err := metastore.NewDynamodbStorage(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer dynamodbStorage.Close()

	dynamodbStorage.Put("/wikipedia_en.json", []byte("{}"))
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	dynamodbStorage.Delete("/wikipedia_en.json")
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

	dynamodbStorage, err := metastore.NewDynamodbStorage(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer dynamodbStorage.Close()

	exists, err := dynamodbStorage.Exists("/wikipedia_en.json")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	if exists != false {
		t.Fatalf("unexpected value. %v\n", exists)
	}

	dynamodbStorage.Put("/wikipedia_en.json", []byte("{}"))
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	exists, err = dynamodbStorage.Exists("/wikipedia_en.json")
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

	dynamodbStorage, err := metastore.NewDynamodbStorage(uri, logger)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer dynamodbStorage.Close()

	dynamodbStorage.Put("/hello.txt", []byte("hello"))
	dynamodbStorage.Put("/world.txt", []byte("world"))

	paths, err := dynamodbStorage.List("/")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	sort.Slice(paths, func(i, j int) bool { return paths[i] < paths[j] })

	if !reflect.DeepEqual(paths, []string{"/hello.txt", "/world.txt"}) {
		t.Fatalf("unexpected %v\v", paths)
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
