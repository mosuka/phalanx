package char

import (
	"fmt"
	"regexp"

	"github.com/blugelabs/bluge/analysis/char"
)

// Create new RegexpCharFilter with given options.
// Options example:
// {
//   "pattern": "foo",
//   "replacement": "var"
// }
func NewRegexpCharFilterWithOptions(opts map[string]interface{}) (*char.RegexpCharFilter, error) {
	patternValue, ok := opts["pattern"]
	if !ok {
		return nil, fmt.Errorf("pattern option does not exist")
	}
	pattern, ok := patternValue.(string)
	if !ok {
		return nil, fmt.Errorf("form option is unexpected")
	}

	replacementValue, ok := opts["replacement"]
	if !ok {
		return nil, fmt.Errorf("pattern option does not exist")
	}
	replacement, ok := replacementValue.(string)
	if !ok {
		return nil, fmt.Errorf("form option is unexpected")
	}

	charFilter := char.NewRegexpCharFilter(regexp.MustCompile(pattern), []byte(replacement))

	return charFilter, nil
}
