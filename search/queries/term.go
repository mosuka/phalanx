package queries

import (
	"encoding/json"

	"github.com/blugelabs/bluge"
)

type TermQueryOptions struct {
	Term  string  `json:"term"`
	Field string  `json:"field"`
	Boost float64 `json:"boost"`
}

func NewTermQueryOptions() TermQueryOptions {
	return TermQueryOptions{}
}

// Create new TermQuery with given options.
// Options example:
// {
//   "term": "hello",
//   "field": "description",
//   "boost": 1.0
// }
func NewTermQueryWithMap(opts map[string]interface{}) (*bluge.TermQuery, error) {
	bytes, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	options := NewTermQueryOptions()
	err = json.Unmarshal(bytes, &options)
	if err != nil {
		return nil, err
	}

	return NewTermQueryWithOptions(options)
}

func NewTermQueryWithOptions(opts TermQueryOptions) (*bluge.TermQuery, error) {
	// term is required.
	termQuery := bluge.NewTermQuery(opts.Term)

	// field is optional.
	if opts.Field != "" {
		termQuery.SetField(opts.Field)
	}

	// boost is optional.
	if opts.Boost >= 0.0 {
		termQuery.SetBoost(opts.Boost)
	}

	return termQuery, nil
}
