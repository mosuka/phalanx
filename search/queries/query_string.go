package queries

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/blugelabs/bluge"
	querystr "github.com/blugelabs/query_string"
	"github.com/mosuka/phalanx/analysis/analyzer"
)

type QueryStringQueryOptions struct {
	Query      string                              `json:"query"`
	DateFormat string                              `json:"date_format"`
	Analyzers  map[string]analyzer.AnalyzerSetting `json:"analyzers"`
}

func NewQueryStringQueryOptions() QueryStringQueryOptions {
	return QueryStringQueryOptions{}
}

// Create new QueryStringQuery with given options.
// Options example:
// {
//   "query": "hello +world",
//   "date_format": "RFC3339",
//   "analyzers": {
//     "title": {
//       "char_filters": [
//         {
//           "name": "unicode_normalize",
//           "options": {
//             "form": "NFKC"
//           }
//         }
//       ],
//       "tokenizer": {
//         "name": "unicode"
//       },
//       "token_filters": [
//         {
//           "name": "lower_case"
//         }
//       ]
//     },
//     "description": {
//       "char_filters": [
//         {
//           "name": "unicode_normalize",
//           "options": {
//             "form": "NFKC"
//           }
//         }
//       ],
//       "tokenizer": {
//         "name": "unicode"
//       },
//       "token_filters": [
//         {
//           "name": "lower_case"
//         }
//       ]
//     }
//   },
// }
func NewQueryStringQueryWithMap(opts map[string]interface{}) (bluge.Query, error) {
	bytes, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	options := NewQueryStringQueryOptions()
	err = json.Unmarshal(bytes, &options)
	if err != nil {
		return nil, err
	}

	return NewQueryStringQueryWithOptions(options)
}

func NewQueryStringQueryWithOptions(opts QueryStringQueryOptions) (bluge.Query, error) {
	queryOpts := querystr.DefaultOptions()

	// date_format is optional.
	switch opts.DateFormat {
	case "Layout":
		queryOpts = queryOpts.WithDateFormat(time.Layout)
	case "ASCII":
		queryOpts = queryOpts.WithDateFormat(time.ANSIC)
	case "UnixDate":
		queryOpts = queryOpts.WithDateFormat(time.UnixDate)
	case "RubyDate":
		queryOpts = queryOpts.WithDateFormat(time.RubyDate)
	case "RFC822":
		queryOpts = queryOpts.WithDateFormat(time.RFC822)
	case "RFC822Z":
		queryOpts = queryOpts.WithDateFormat(time.RFC822Z)
	case "RFC850":
		queryOpts = queryOpts.WithDateFormat(time.RFC850)
	case "RFC1123":
		queryOpts = queryOpts.WithDateFormat(time.RFC1123)
	case "RFC1123Z":
		queryOpts = queryOpts.WithDateFormat(time.RFC1123Z)
	case "RFC3339":
		queryOpts = queryOpts.WithDateFormat(time.RFC3339)
	case "RFC3339Nano":
		queryOpts = queryOpts.WithDateFormat(time.RFC3339Nano)
	case "Kitchen":
		queryOpts = queryOpts.WithDateFormat(time.Kitchen)
	case "Stamp":
		queryOpts = queryOpts.WithDateFormat(time.Stamp)
	case "StampMilli":
		queryOpts = queryOpts.WithDateFormat(time.StampMilli)
	case "StampMicro":
		queryOpts = queryOpts.WithDateFormat(time.StampMicro)
	case "StampNano":
		queryOpts = queryOpts.WithDateFormat(time.StampNano)
	default:
		queryOpts = queryOpts.WithDateFormat(time.RFC3339)
	}

	for fieldName, analyzerSetting := range opts.Analyzers {
		analyzer, err := analyzer.NewAnalyzer(analyzerSetting)
		if err != nil {
			return nil, fmt.Errorf("failed to create analyzer for %s: %v", fieldName, err)
		}
		queryOpts = queryOpts.WithAnalyzerForField(fieldName, analyzer)
	}

	query, err := querystr.ParseQueryString(opts.Query, queryOpts)
	if err != nil {
		return nil, err
	}

	return query, nil
}
