package queries

import (
	"encoding/json"

	"github.com/blugelabs/bluge"
)

type QuerySetting struct {
	Type    string                 `json:"type"`
	Options map[string]interface{} `json:"options"`
}

type BooleanQueryOptions struct {
	Must      []QuerySetting `json:"must"`
	MustNot   []QuerySetting `json:"must_not"`
	Should    []QuerySetting `json:"should"`
	MinShould int            `json:"min_should"`
	Boost     float64        `json:"boost"`
}

func NewBooleanQueryOptions() BooleanQueryOptions {
	return BooleanQueryOptions{}
}

// Create new BooleanQuery with given options.
// Options example:
// {
//   "must": [
//     {
//       "type": "term",
//       "options": {
//         "term": "hello",
//         "field": "description",
//         "boost": 1.0
//       }
//     },
//     {
//       "type": "term",
//       "options": {
//         "term": "world",
//         "field": "description",
//         "boost": 1.0
//       }
//     }
//   ],
//   "must_not": [
//     {
//       "type": "term",
//       "options": {
//         "term": "bye",
//         "field": "description",
//         "boost": 1.0
//       }
//     },
//     {
//       "type": "term",
//       "options": {
//         "term": "さようなら",
//         "field": "description",
//         "boost": 1.0
//       }
//     }
//   ],
//   "should": [
//     {
//       "type": "term",
//       "options": {
//         "term": "こんにちは",
//         "field": "description",
//         "boost": 1.0
//       }
//     },
//     {
//       "type": "term",
//       "options": {
//         "term": "世界",
//         "field": "description",
//         "boost": 1.0
//       }
//     }
//   ],
//   "min_should": 1,
//   "boost": 1.0
// }
func NewBooleanQueryWithMap(opts map[string]interface{}) (*bluge.BooleanQuery, error) {
	bytes, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	options := NewBooleanQueryOptions()
	if err := json.Unmarshal(bytes, &options); err != nil {
		return nil, err
	}

	return NewBooleanQueryWithOptions(options)
}

func NewBooleanQueryWithOptions(opts BooleanQueryOptions) (*bluge.BooleanQuery, error) {
	booleanQuery := bluge.NewBooleanQuery()

	for _, mustQuery := range opts.Must {
		if query, err := NewQuery(mustQuery.Type, mustQuery.Options); err == nil {
			booleanQuery.AddMust(query)
		}
	}

	for _, mustNotQuery := range opts.MustNot {
		if query, err := NewQuery(mustNotQuery.Type, mustNotQuery.Options); err == nil {
			booleanQuery.AddMustNot(query)
		}
	}

	for _, shouldQuery := range opts.Should {
		if query, err := NewQuery(shouldQuery.Type, shouldQuery.Options); err == nil {
			booleanQuery.AddShould(query)
		}
	}

	// min_should is optional.
	if opts.MinShould > 0 {
		booleanQuery.SetMinShould(opts.MinShould)
	}

	// boost is optional.
	if opts.Boost >= 0.0 {
		booleanQuery.SetBoost(opts.Boost)
	}

	return booleanQuery, nil
}
