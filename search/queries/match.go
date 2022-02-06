package queries

import (
	"encoding/json"

	"github.com/blugelabs/bluge"
	"github.com/mosuka/phalanx/analysis/analyzer"
)

var (
	MatchQueryOperator_name = map[bluge.MatchQueryOperator]string{
		bluge.MatchQueryOperatorAnd: "AND",
		bluge.MatchQueryOperatorOr:  "OR",
	}
	MatchQueryOperator_value = map[string]bluge.MatchQueryOperator{
		"AND": bluge.MatchQueryOperatorAnd,
		"OR":  bluge.MatchQueryOperatorOr,
	}
)

type MatchQueryOptions struct {
	Match     string                   `json:"match"`
	Field     string                   `json:"field"`
	Boost     float64                  `json:"boost"`
	Prefix    int                      `json:"prefix"`
	Fuzziness int                      `json:"fuzziness"`
	Operator  string                   `json:"operator"`
	Analyzer  analyzer.AnalyzerSetting `json:"analyzer"`
}

func NewMatchQueryOptions() MatchQueryOptions {
	return MatchQueryOptions{
		Operator: "OR",
	}
}

// Create new MatchQuery with given options.
// Options example:
// {
//   "match": "hello",
//   "field": "description",
//   "prefix": 1,
//   "fuzziness": 1,
//   "boost": 1.0,
//   "operator": "AND",
//   "analyzer": {
//     "char_filters": [
//       {
//         "name": "unicode_normalize",
//         "options": {
//           "form": "NFKC"
//         }
//       }
//     ],
//     "tokenizer": {
//       "name": "whitespace"
//     },
//     "token_filters": [
//       {
//         "name": "lower_case"
//       }
//     ]
//   }
// }
func NewMatchQueryWithMap(opts map[string]interface{}) (*bluge.MatchQuery, error) {
	bytes, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	options := NewMatchQueryOptions()
	if err := json.Unmarshal(bytes, &options); err != nil {
		return nil, err
	}

	return NewMatchQueryWithOptions(options)
}

func NewMatchQueryWithOptions(opts MatchQueryOptions) (*bluge.MatchQuery, error) {
	matchQuery := bluge.NewMatchQuery(opts.Match)

	// field is optional.
	if opts.Field != "" {
		matchQuery.SetField(opts.Field)
	}

	// boost is optional.
	if opts.Boost >= 0.0 {
		matchQuery.SetBoost(opts.Boost)
	}

	// prefix is optional.
	if opts.Prefix > 0 {
		matchQuery.SetPrefix(opts.Prefix)
	}

	// fuzziness is optional.
	if opts.Fuzziness > 0 {
		matchQuery.SetFuzziness(opts.Fuzziness)
	}

	// operator is optional.
	if opts.Operator != "" {
		matchQuery.SetOperator(MatchQueryOperator_value[opts.Operator])
	}

	// analyzer is optional.
	if analyzer, err := analyzer.NewAnalyzer(opts.Analyzer); err == nil {
		matchQuery.SetAnalyzer(analyzer)
	}

	return matchQuery, nil
}
