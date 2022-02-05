package queries

import (
	"encoding/json"

	"github.com/blugelabs/bluge"
)

type NumericRangeQueryOptions struct {
	Min          float64 `json:"min"`
	Max          float64 `json:"max"`
	InclusiveMin bool    `json:"inclusive_min"`
	InclusiveMax bool    `json:"inclusive_max"`
	Field        string  `json:"field"`
	Boost        float64 `json:"boost"`
}

func NewNumericRangeQueryOption() NumericRangeQueryOptions {
	return NumericRangeQueryOptions{
		Min:          bluge.MinNumeric,
		Max:          bluge.MaxNumeric,
		InclusiveMin: true,
		InclusiveMax: false,
	}
}

// Create new NumericRangeQuery with given options.
// Options example:
// {
//   "min": 0.0,
//   "max": 1.0,
//   "inclusive_min": true,
//   "inclusive_max": false,
//   "field": "description",
//   "boost": 1.0
// }
func NewNumericRangeQueryWithMap(opts map[string]interface{}) (*bluge.NumericRangeQuery, error) {
	bytes, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	options := NewNumericRangeQueryOption()
	err = json.Unmarshal(bytes, &options)
	if err != nil {
		return nil, err
	}

	return NewNumericRangeQueryWithOptions(options)
}

func NewNumericRangeQueryWithOptions(opts NumericRangeQueryOptions) (*bluge.NumericRangeQuery, error) {
	numericRangeQuery := bluge.NewNumericRangeInclusiveQuery(opts.Min, opts.Max, opts.InclusiveMin, opts.InclusiveMax)

	// field is optional.
	if opts.Field != "" {
		numericRangeQuery.SetField(opts.Field)
	}

	// boost is optional.
	if opts.Boost >= 0.0 {
		numericRangeQuery.SetBoost(opts.Boost)
	}

	return numericRangeQuery, nil
}
