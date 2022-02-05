package queries

import (
	"encoding/json"
	"io/ioutil"
	"testing"
)

func TestNewTermRangeQueryWithMap(t *testing.T) {
	queryFile := "../../testdata/test_term_range_query.json"

	bytes, err := ioutil.ReadFile(queryFile)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	var opts map[string]interface{}
	if err := json.Unmarshal(bytes, &opts); err != nil {
		t.Fatalf("%v\n", err)
	}

	_, err = NewTermRangeQueryWithMap(opts)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
}
