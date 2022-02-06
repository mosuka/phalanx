package queries

import (
	"encoding/json"

	"github.com/blugelabs/bluge"
)

type MatchNoneQueryOptions struct {
	Boost float64 `json:"boost"`
}

func NewMatchNoneQueryOptions() MatchNoneQueryOptions {
	return MatchNoneQueryOptions{}
}

// Create new MatchNoneQuery with given options.
// Options example:
// {
//   "boost": 1.0
// }
func NewMatchNoneQueryWithMap(opts map[string]interface{}) (*bluge.MatchNoneQuery, error) {
	bytes, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	options := NewMatchNoneQueryOptions()
	if err := json.Unmarshal(bytes, &options); err != nil {
		return nil, err
	}

	return NewMatchNoneQueryWithOptions(options)
}

func NewMatchNoneQueryWithOptions(opts MatchNoneQueryOptions) (*bluge.MatchNoneQuery, error) {
	matchNoneQuery := bluge.NewMatchNoneQuery()

	// boost is optional.
	if opts.Boost >= 0.0 {
		matchNoneQuery.SetBoost(opts.Boost)
	}

	return matchNoneQuery, nil
}
