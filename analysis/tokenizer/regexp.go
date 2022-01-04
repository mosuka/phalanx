package tokenizer

import (
	"fmt"
	"regexp"

	"github.com/blugelabs/bluge/analysis/tokenizer"
)

// Create new RegexTokenizer with given options.
// Options example:
// {
//   "pattern": "[0-9a-zA-Z_]*"
// }
func NewRegexpTokenizerWithOptions(opts map[string]interface{}) (*tokenizer.RegexpTokenizer, error) {
	patternValue, ok := opts["pattern"]
	if !ok {
		return nil, fmt.Errorf("pattern option does not exist")
	}
	pattern, ok := patternValue.(string)
	if !ok {
		return nil, fmt.Errorf("form option is unexpected")
	}

	return tokenizer.NewRegexpTokenizer(regexp.MustCompile(pattern)), nil
}
