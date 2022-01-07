package token

import (
	"fmt"

	"github.com/blugelabs/bluge/analysis"
	"github.com/blugelabs/bluge/analysis/token"
)

// Create new StopTokensFilter with given options.
// Options example:
// {
//   "stop_tokens": [
//     "a",
//     "an",
//     "and",
//     "are",
//     "as",
//     "at",
//     "be",
//     "but",
//     "by",
//     "for",
//     "if",
//     "in",
//     "into",
//     "is",
//     "it",
//     "no",
//     "not",
//     "of",
//     "on",
//     "or",
//     "such",
//     "that",
//     "the",
//     "their",
//     "then",
//     "there",
//     "these",
//     "they",
//     "this",
//     "to",
//     "was",
//     "will",
//     "with"
//   ]
// }
func NewStopTokensFilterWithOptions(opts map[string]interface{}) (*token.StopTokensFilter, error) {
	stopTokensValue, ok := opts["stop_tokens"]
	if !ok {
		return nil, fmt.Errorf("stop_tokens option does not exist")
	}
	stopTokens, ok := stopTokensValue.([]interface{})
	if !ok {
		return nil, fmt.Errorf("stop_tokens option is unexpected")
	}
	stopTokenMap := analysis.NewTokenMap()
	for _, stopToken := range stopTokens {
		token, ok := stopToken.(string)
		if !ok {
			return nil, fmt.Errorf("base_form is unexpected: %v", stopToken)
		}
		stopTokenMap.AddToken(token)
	}

	return token.NewStopTokensFilter(stopTokenMap), nil
}
