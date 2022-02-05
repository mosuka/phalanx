package queries

import (
	"encoding/json"

	"github.com/blugelabs/bluge"
)

type FuzzyQueryOptions struct {
	Term      string  `json:"term"`
	Field     string  `json:"field"`
	Boost     float64 `json:"boost"`
	Prefix    int     `json:"prefix"`
	Fuzziness int     `json:"fuzziness"`
}

func NewFuzzyQueryOptions() FuzzyQueryOptions {
	return FuzzyQueryOptions{}
}

// Create new FuzzyQuery with given options.
// Options example:
// {
//   "term": "hello",
//   "prefix": 1,
//   "fuzziness": 1,
//   "field": "description",
//   "boost": 1.0
// }
func NewFuzzyQueryWithMap(opts map[string]interface{}) (*bluge.FuzzyQuery, error) {
	bytes, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	options := NewFuzzyQueryOptions()
	if err := json.Unmarshal(bytes, &options); err != nil {
		return nil, err
	}

	return NewFuzzyQueryWithOptions(options)
}

func NewFuzzyQueryWithOptions(opts FuzzyQueryOptions) (*bluge.FuzzyQuery, error) {
	fuzzyQuery := bluge.NewFuzzyQuery(opts.Term)

	// field is optional.
	if opts.Field != "" {
		fuzzyQuery.SetField(opts.Field)
	}

	// boost is optional.
	if opts.Boost >= 0.0 {
		fuzzyQuery.SetBoost(opts.Boost)
	}

	// prefix is optional.
	if opts.Prefix > 0 {
		fuzzyQuery.SetPrefix(opts.Prefix)
	}

	// fuzziness is optional.
	if opts.Fuzziness > 0 {
		fuzzyQuery.SetFuzziness(opts.Fuzziness)
	}

	return fuzzyQuery, nil
}
