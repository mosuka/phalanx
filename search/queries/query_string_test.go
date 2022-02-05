package queries

import (
	"encoding/json"
	"io/ioutil"
	"testing"
)

func TestNewQueryStringQueryWithMap(t *testing.T) {
	queryFile := "../../testdata/test_query_string_query.json"

	bytes, err := ioutil.ReadFile(queryFile)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	var opts map[string]interface{}
	if err := json.Unmarshal(bytes, &opts); err != nil {
		t.Fatalf("%v\n", err)
	}

	_, err = NewQueryStringQueryWithMap(opts)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
}
