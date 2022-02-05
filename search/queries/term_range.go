package queries

import (
	"encoding/json"

	"github.com/blugelabs/bluge"
)

type TermRangeQueryOptions struct {
	Min          string  `json:"min"`
	Max          string  `json:"max"`
	InclusiveMin bool    `json:"inclusive_min"`
	InclusiveMax bool    `json:"inclusive_max"`
	Field        string  `json:"field"`
	Boost        float64 `json:"boost"`
}

func NewTermRangeQueryOptions() TermRangeQueryOptions {
	return TermRangeQueryOptions{
		Min:          "",
		Max:          "",
		InclusiveMin: true,
		InclusiveMax: false,
	}
}

// Create new TermRangeQuery with given options.
// Options example:
// {
//   "min": "a",
//   "max": "z",
//   "inclusive_min": true,
//   "inclusive_max": false,
//   "field": "description",
//   "boost": 1.0
// }
func NewTermRangeQueryWithMap(opts map[string]interface{}) (*bluge.TermRangeQuery, error) {
	bytes, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	options := NewTermRangeQueryOptions()
	err = json.Unmarshal(bytes, &options)
	if err != nil {
		return nil, err
	}

	return NewTermRangeQueryWithOptions(options)
}

func NewTermRangeQueryWithOptions(opts TermRangeQueryOptions) (*bluge.TermRangeQuery, error) {
	termRangeQuery := bluge.NewTermRangeInclusiveQuery(opts.Min, opts.Max, opts.InclusiveMin, opts.InclusiveMax)

	// field is optional.
	if opts.Field != "" {
		termRangeQuery.SetField(opts.Field)
	}

	// boost is optional.
	if opts.Boost >= 0.0 {
		termRangeQuery.SetBoost(opts.Boost)
	}

	return termRangeQuery, nil
}
