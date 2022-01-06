package aggregations

import (
	"fmt"

	"github.com/blugelabs/bluge/search"
	"github.com/blugelabs/bluge/search/aggregations"
)

// Create new RangeAggregation with given options.
// Each bucket represents the number of documents
// that the condition between `low`` or more and less than `high`.
// low <= number < high
// Options example:
// {
//   "field": "id",
//   "ranges": {
//     "low": {
//       "low": 0,
//       "high": 500
//     },
//     "medium": {
//       "low": 500,
//       "high": 1000
//     },
//     "high": {
//       "low": 1000,
//       "high": 1500
//     }
//   }
// }
func NewRangeAggregationWithOptions(opts map[string]interface{}) (*aggregations.RangeAggregation, error) {
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

	rangesAgg := aggregations.Ranges(search.Field(field))

	ranges, ok := opts["ranges"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("ranges option does not exist")
	}
	for name, rangeValue := range ranges {
		rangeMap, ok := rangeValue.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("range %v option is unexpected: %v", name, rangeValue)
		}

		low, ok := rangeMap["low"].(float64)
		if !ok {
			return nil, fmt.Errorf("range %v low option is unexpected: %v", name, rangeMap["low"])
		}

		high, ok := rangeMap["high"].(float64)
		if !ok {
			return nil, fmt.Errorf("range %v high option is unexpected: %v", name, rangeMap["high"])
		}

		rangesAgg.AddRange(aggregations.NamedRange(name, low, high))
	}

	return rangesAgg, nil
}
