package token

import (
	"fmt"

	"github.com/blugelabs/bluge/analysis/token"
)

// Create new LengthFilter with given options.
// Options example:
// {
//   "min_length": 3,
//   "max_length": 4
// }
func NewLengthFilterWithOptions(opts map[string]interface{}) (*token.LengthFilter, error) {
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
		return nil, fmt.Errorf("max_length option is unexpected: %V", maxLengthValue)
	}
	maxLength := int(maxLengthNum)

	return token.NewLengthFilter(minLength, maxLength), nil
}
