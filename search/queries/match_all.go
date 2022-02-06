package queries

import (
	"encoding/json"

	"github.com/blugelabs/bluge"
)

type MatchAllQueryOptions struct {
	Boost float64 `json:"boost"`
}

func NewMatchAllQueryOptions() MatchAllQueryOptions {
	return MatchAllQueryOptions{}
}

// Create new MatchAllQuery with given options.
// Options example:
// {
//   "boost": 1.0,
// }
func NewMatchAllQueryWithMap(opts map[string]interface{}) (*bluge.MatchAllQuery, error) {
	bytes, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	options := NewMatchAllQueryOptions()
	if err := json.Unmarshal(bytes, &options); err != nil {
		return nil, err
	}

	return NewMatchAllQueryWithOptions(options)
}

func NewMatchAllQueryWithOptions(opts MatchAllQueryOptions) (*bluge.MatchAllQuery, error) {
	matchAllQuery := bluge.NewMatchAllQuery()

	// boost is optional.
	if opts.Boost >= 0.0 {
		matchAllQuery.SetBoost(opts.Boost)
	}

	return matchAllQuery, nil
}
