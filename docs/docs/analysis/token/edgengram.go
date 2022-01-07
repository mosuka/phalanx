package token

import (
	"fmt"

	"github.com/blugelabs/bluge/analysis/token"
)

// Create new EdgeNgramFilter with given options.
// Options example:
// {
//   "back": false,
//   "min_length": 1,
//   "max_length": 2
// }
func NewEdgeNgramFilterWithOptions(opts map[string]interface{}) (*token.EdgeNgramFilter, error) {
	backValue, ok := opts["back"]
	if !ok {
		return nil, fmt.Errorf("back option does not exist")
	}
	back, ok := backValue.(bool)
	if !ok {
		return nil, fmt.Errorf("back option is unexpected")
	}
	var side token.Side
	if back {
		side = token.BACK
	} else {
		side = token.FRONT
	}

	minLengthValue, ok := opts["min_length"]
	if !ok {
		return nil, fmt.Errorf("min_length option does not exist")
	}
	minLengthNum, ok := minLengthValue.(float64)
	if !ok {
		return nil, fmt.Errorf("min_length option is unexpected")
	}
	minLength := int(minLengthNum)

	maxLengthValue, ok := opts["max_length"]
	if !ok {
		return nil, fmt.Errorf("max_length option does not exist")
	}
	maxLengthNum, ok := maxLengthValue.(float64)
	if !ok {
		return nil, fmt.Errorf("max_length option is unexpected")
	}
	maxLength := int(maxLengthNum)

	return token.NewEdgeNgramFilter(side, minLength, maxLength), nil
}
