package aggregations

import (
	"fmt"

	"github.com/blugelabs/bluge/search"
	"github.com/blugelabs/bluge/search/aggregations"
)

// Create new TermsAggregation with given options.
// Options example:
// {
//   "field": "tags",
//   "min_length": 2,
//   "max_length": 10,
//   "size": 10
// }
func NewTermsAggregationWithOptions(opts map[string]interface{}) (*aggregations.TermsAggregation, error) {
	fieldValue, ok := opts["field"]
	if !ok {
		return nil, fmt.Errorf("field option does not exist")
	}
	field, ok := fieldValue.(string)
	if !ok {
		return nil, fmt.Errorf("field option is unexpected: %v", fieldValue)
	}
	if len(field) == 0 {
		return nil, fmt.Errorf("field option is empty")
	}

	minLength := -1
	if minLengthValue, ok := opts["min_length"]; ok {
		if minLengthNum, ok := minLengthValue.(float64); !ok {
			return nil, fmt.Errorf("min_length option is unexpected: %v", minLengthValue)
		} else {
			minLength = int(minLengthNum)
		}
	}

	maxLength := -1
	if maxLengthValue, ok := opts["max_length"]; ok {
		if maxLengthNum, ok := maxLengthValue.(float64); !ok {
			return nil, fmt.Errorf("max_length option is unexpected: %v", maxLengthValue)
		} else {
			maxLength = int(maxLengthNum)
		}
	}

	size := 10
	if sizeValue, ok := opts["size"]; ok {
		if sizeNum, ok := sizeValue.(float64); !ok {
			return nil, fmt.Errorf("size option is unexpected: %v", sizeValue)
		} else {
			size = int(sizeNum)
		}
	}

	return aggregations.NewTermsAggregation(aggregations.FilterText(search.Field(field), func(bytes []byte) bool {
		switch {
		case len(bytes) < minLength && minLength > 0:
			return false
		case len(bytes) > maxLength && maxLength > 0:
			return false
		}
		return true
	}), size), nil
}
