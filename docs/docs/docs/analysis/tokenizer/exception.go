package tokenizer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/blugelabs/bluge/analysis/tokenizer"
)

// Create new ExceptionsTokenizer with given options.
// Options example:
// {
//   "patterns": [
//     "[hH][tT][tT][pP][sS]?://(\S)*",
//     "[fF][iI][lL][eE]://(\S)*",
//     "[fF][tT][pP]://(\S)*",
//     "\S+@\S+"
//   ]
// }
func NewExceptionsTokenizerWithOptions(opts map[string]interface{}) (*tokenizer.ExceptionsTokenizer, error) {
	patternsValue, ok := opts["patterns"]
	if !ok {
		return nil, fmt.Errorf("patterns option does not exist")
	}
	patterns, ok := patternsValue.([]interface{})
	if !ok {
		return nil, fmt.Errorf("patterns option is unexpected")
	}
	patternStrs := make([]string, 0)
	for _, pattern := range patterns {
		str, ok := pattern.(string)
		if !ok {
			return nil, fmt.Errorf("patterns option is unexpected")
		}
		patternStrs = append(patternStrs, str)
	}

	pattern := strings.Join(patternStrs, "|")
	r, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("patterns option is unexpected")
	}

	return tokenizer.NewExceptionsTokenizer(r, tokenizer.NewUnicodeTokenizer()), nil
}
