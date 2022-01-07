package token

import (
	"fmt"

	"github.com/blugelabs/bluge/analysis"
	"github.com/blugelabs/bluge/analysis/token"
)

// Create new KeyWordMarkerFilter with given options.
// Options example:
// {
//   "keywords": [
//     "walk",
//     "park"
//   ]
// }
func NewKeyWordMarkerFilterWithOptions(opts map[string]interface{}) (*token.KeyWordMarkerFilter, error) {
	keywordsValue, ok := opts["keywords"]
	if !ok {
		return nil, fmt.Errorf("keywords option does not exist")
	}
	keywords, ok := keywordsValue.([]interface{})
	if !ok {
		return nil, fmt.Errorf("keywords option is unexpected")
	}
	keywordMap := analysis.NewTokenMap()
	for _, keyword := range keywords {
		str, ok := keyword.(string)
		if !ok {
			return nil, fmt.Errorf("keyword is unexpected: %v", keyword)
		}
		keywordMap.AddToken(str)
	}

	return token.NewKeyWordMarkerFilter(keywordMap), nil
}
