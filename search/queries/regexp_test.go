package queries

import (
	"encoding/json"
	"io/ioutil"
	"testing"
)

func TestNewRegexpQueryWithMap(t *testing.T) {
	queryFile := "../../testdata/test_regexp_query.json"

	bytes, err := ioutil.ReadFile(queryFile)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	var opts map[string]interface{}
	if err := json.Unmarshal(bytes, &opts); err != nil {
		t.Fatalf("%v\n", err)
	}

	_, err = NewRegexpQueryWithMap(opts)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
}
