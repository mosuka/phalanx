package token

import (
	"fmt"

	"github.com/blugelabs/bluge/analysis"
	"github.com/blugelabs/bluge/analysis/token"
)

// Create new ElisionFilter with given options.
// Options example:
// {
//   "articles": [
//     "ar"
//   ]
// }
func NewElisionFilterWithOptions(opts map[string]interface{}) (*token.ElisionFilter, error) {
	articlesValue, ok := opts["articles"]
	if !ok {
		return nil, fmt.Errorf("articles option does not exist")
	}
	articles, ok := articlesValue.([]interface{})
	if !ok {
		return nil, fmt.Errorf("articles option is unexpected")
	}
	articleMap := analysis.NewTokenMap()
	for _, article := range articles {
		str, ok := article.(string)
		if !ok {
			return nil, fmt.Errorf("articles is unexpected")
		}
		articleMap.AddToken(str)
	}

	return token.NewElisionFilter(articleMap), nil
}
