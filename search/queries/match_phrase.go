package queries

import (
	"encoding/json"

	"github.com/blugelabs/bluge"
	"github.com/mosuka/phalanx/analysis/analyzer"
)

type MatchPhraseQueryOptions struct {
	Phrase   string                   `json:"phrase"`
	Field    string                   `json:"field"`
	Boost    float64                  `json:"boost"`
	Slop     int                      `json:"slop"`
	Analyzer analyzer.AnalyzerSetting `json:"analyzer"`
}

func NewMatchPhraseQueryOptions() MatchPhraseQueryOptions {
	return MatchPhraseQueryOptions{}
}

// Create new MatchPhraseQuery with given options.
// Options example:
// {
//   "phrase": "hello world",
//   "field": "description",
//   "slop": 1,
//   "boost": 1.0,
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
func NewMatchPhraseQueryWithMap(opts map[string]interface{}) (*bluge.MatchPhraseQuery, error) {
	bytes, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	options := NewMatchPhraseQueryOptions()
	if err := json.Unmarshal(bytes, &options); err != nil {
		return nil, err
	}

	return NewMatchPhraseQueryWithOptions(options)
}

func NewMatchPhraseQueryWithOptions(opts MatchPhraseQueryOptions) (*bluge.MatchPhraseQuery, error) {
	matchPhraseQuery := bluge.NewMatchPhraseQuery(opts.Phrase)

	// field is optional.
	if opts.Field != "" {
		matchPhraseQuery.SetField(opts.Field)
	}

	// boost is optional.
	if opts.Boost != 0.0 {
		matchPhraseQuery.SetBoost(opts.Boost)
	}

	// slop is optional.
	if opts.Slop != 0 {
		matchPhraseQuery.SetSlop(opts.Slop)
	}

	// analyzer is optional.
	if analyzer, err := analyzer.NewAnalyzer(opts.Analyzer); err == nil {
		matchPhraseQuery.SetAnalyzer(analyzer)
	}

	return matchPhraseQuery, nil
}
