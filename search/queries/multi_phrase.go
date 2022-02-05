package queries

import (
	"encoding/json"

	"github.com/blugelabs/bluge"
)

type MultiPhraseQueryOptions struct {
	Terms [][]string `json:"terms"`
	Field string     `json:"field"`
	Boost float64    `json:"boost"`
	Slop  int        `json:"slop"`
}

func NewMultiPhraseQueryOptions() MultiPhraseQueryOptions {
	return MultiPhraseQueryOptions{}
}

// Create new MultiPhraseQuery with given options.
// Options example:
// {
//   "terms": [
//     ["foo", "bar"],
//     ["baz"]
//   ],
//   "field": "description",
//   "slop": 1,
//   "boost": 1.0
// }
func NewMultiPhraseQueryWithMap(opts map[string]interface{}) (*bluge.MultiPhraseQuery, error) {
	bytes, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	options := NewMultiPhraseQueryOptions()
	if err := json.Unmarshal(bytes, &options); err != nil {
		return nil, err
	}

	return NewMultiPhraseQueryWithOptions(options)
}

func NewMultiPhraseQueryWithOptions(opts MultiPhraseQueryOptions) (*bluge.MultiPhraseQuery, error) {
	multiPhraseQuery := bluge.NewMultiPhraseQuery(opts.Terms)

	// slop is optional.
	if opts.Slop >= 0 {
		multiPhraseQuery.SetSlop(opts.Slop)
	}

	// field is optional.
	if opts.Field != "" {
		multiPhraseQuery.SetField(opts.Field)
	}

	// boost is optional.
	if opts.Boost >= 0.0 {
		multiPhraseQuery.SetBoost(opts.Boost)
	}

	return multiPhraseQuery, nil
}
