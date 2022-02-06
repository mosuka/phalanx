package queries

import (
	"encoding/json"

	"github.com/blugelabs/bluge"
)

type WildlcardQueryOptions struct {
	Wildcard string  `json:"wildcard"`
	Field    string  `json:"field"`
	Boost    float64 `json:"boost"`
}

func NewWildcardQueryOptions() WildlcardQueryOptions {
	return WildlcardQueryOptions{}
}

// Create new WildcardQuery with given options.
// Options example:
// {
//   "wildcard": "h*",
//   "field": "description",
//   "boost": 1.0
// }
func NewWildcardQueryWithMap(opts map[string]interface{}) (*bluge.WildcardQuery, error) {
	bytes, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	options := NewWildcardQueryOptions()
	err = json.Unmarshal(bytes, &options)
	if err != nil {
		return nil, err
	}

	return NewWildcardQueryWithOptions(options)
}

func NewWildcardQueryWithOptions(opts WildlcardQueryOptions) (*bluge.WildcardQuery, error) {
	// wildcard is required.
	wildcardQuery := bluge.NewWildcardQuery(opts.Wildcard)

	// field is optional.
	if opts.Field != "" {
		wildcardQuery.SetField(opts.Field)
	}

	// boost is optional.
	if opts.Boost >= 0.0 {
		wildcardQuery.SetBoost(opts.Boost)
	}

	return wildcardQuery, nil
}
