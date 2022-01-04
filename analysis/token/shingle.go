package token

import (
	"fmt"

	"github.com/blugelabs/bluge/analysis/token"
)

// Create new ShingleFilter with given options.
// Options example:
// {
//   "min_length": 2,
//   "max_length": 2,
//   "output_original": true,
//   "token_separator": " ",
//   "fill": "_"
// }
func NewShingleFilterWithOptions(opts map[string]interface{}) (*token.ShingleFilter, error) {
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

	outputOriginalValue, ok := opts["output_original"]
	if !ok {
		return nil, fmt.Errorf("output_original option does not exist")
	}
	outputOriginal, ok := outputOriginalValue.(bool)
	if !ok {
		return nil, fmt.Errorf("output_original option is unexpected: %v", outputOriginalValue)
	}

	tokenSeparatorValue, ok := opts["token_separator"]
	if !ok {
		return nil, fmt.Errorf("token_separator option does not exist")
	}
	tokenSeparator, ok := tokenSeparatorValue.(string)
	if !ok {
		return nil, fmt.Errorf("token_separator option is unexpected: %v", tokenSeparatorValue)
	}

	fillValue, ok := opts["fill"]
	if !ok {
		return nil, fmt.Errorf("fill option does not exist")
	}
	fill, ok := fillValue.(string)
	if !ok {
		return nil, fmt.Errorf("fill option is unexpected: %v", fillValue)
	}

	return token.NewShingleFilter(minLength, maxLength, outputOriginal, tokenSeparator, fill), nil
}
