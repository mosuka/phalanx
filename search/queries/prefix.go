package queries

import (
	"encoding/json"

	"github.com/blugelabs/bluge"
)

type PrefixQueryOptions struct {
	Prefix string  `json:"prefix"`
	Field  string  `json:"field"`
	Boost  float64 `json:"boost"`
}

func NewPrefixQueryOptions() PrefixQueryOptions {
	return PrefixQueryOptions{}
}

// Create new PrefixQuery with given options.
// Options example:
// {
//   "prefix": "hel",
//   "field": "description",
//   "boost": 1.0
// }
func NewPrefixQueryWithMap(opts map[string]interface{}) (*bluge.PrefixQuery, error) {
	bytes, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	options := NewPrefixQueryOptions()
	err = json.Unmarshal(bytes, &options)
	if err != nil {
		return nil, err
	}

	return NewPrefixQueryWithOptions(options)
}

func NewPrefixQueryWithOptions(opts PrefixQueryOptions) (*bluge.PrefixQuery, error) {
	prefixQuery := bluge.NewPrefixQuery(opts.Prefix)

	// field is optional.
	if opts.Field != "" {
		prefixQuery.SetField(opts.Field)
	}

	// boost is optional.
	if opts.Boost >= 0.0 {
		prefixQuery.SetBoost(opts.Boost)
	}

	return prefixQuery, nil
}
