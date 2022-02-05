package queries

import (
	"encoding/json"

	"github.com/blugelabs/bluge"
)

type RegexpQueryOptions struct {
	Regexp string  `json:"regexp"`
	Field  string  `json:"field"`
	Boost  float64 `json:"boost"`
}

func NewRegexpQueryOptions() RegexpQueryOptions {
	return RegexpQueryOptions{}
}

// Create new RegexpQuery with given options.
// Options example:
// {
//   "regexp": "hel.*",
//   "field": "description",
//   "boost": 1.0
// }
func NewRegexpQueryWithMap(opts map[string]interface{}) (*bluge.RegexpQuery, error) {
	bytes, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	options := NewRegexpQueryOptions()
	err = json.Unmarshal(bytes, &options)
	if err != nil {
		return nil, err
	}

	return NewRegexpQueryWithOptions(options)
}

func NewRegexpQueryWithOptions(opts RegexpQueryOptions) (*bluge.RegexpQuery, error) {
	// regexp is required.
	regexpQuery := bluge.NewRegexpQuery(opts.Regexp)

	// field is optional.
	if opts.Field != "" {
		regexpQuery.SetField(opts.Field)
	}

	// boost is optional.
	if opts.Boost >= 0.0 {
		regexpQuery.SetBoost(opts.Boost)
	}

	return regexpQuery, nil
}
