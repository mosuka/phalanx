package token

import (
	"fmt"

	"github.com/blugelabs/bluge/analysis/token"
)

// Create new TruncateTokenFilter with given options.
// Options example:
// {
//   "length": 5
// }
func NewTruncateTokenFilterWithOptions(opts map[string]interface{}) (*token.TruncateTokenFilter, error) {
	lengthValue, ok := opts["length"]
	if !ok {
		return nil, fmt.Errorf("length option does not exist")
	}
	lengthNum, ok := lengthValue.(float64)
	if !ok {
		return nil, fmt.Errorf("length option is unexpected: %v", lengthValue)
	}
	length := int(lengthNum)

	return token.NewTruncateTokenFilter(length), nil
}
