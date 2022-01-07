package token

import (
	"fmt"

	"github.com/blugelabs/bluge/analysis/token"
)

// Create new NgramFilter with given options.
// Options example:
// {
//   "min_length": 1,
//   "max_length": 2
// }
func NewNgramFilterWithOptions(opts map[string]interface{}) (*token.NgramFilter, error) {
	minLengthValue, ok := opts["min_length"]
	if !ok {
		return nil, fmt.Errorf("min_length option does not exist")
	}
	minLengthNum, ok := minLengthValue.(float64)
	if !ok {
		return nil, fmt.Errorf("min_length option is unexpected: %v", minLengthValue)
	}
	minLength := int(minLengthNum)

	maxLengthValue, ok := opts["max_length"]
	if !ok {
		return nil, fmt.Errorf("max_length option does not exist")
	}
	maxLengthNum, ok := maxLengthValue.(float64)
	if !ok {
		return nil, fmt.Errorf("max_length option is unexpected: %v", maxLengthValue)
	}
	maxLength := int(maxLengthNum)

	return token.NewNgramFilter(minLength, maxLength), nil
}
