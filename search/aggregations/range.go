package aggregations

import (
	"fmt"

	"github.com/blugelabs/bluge/search"
	"github.com/blugelabs/bluge/search/aggregations"
)

// Create new RangeAggregation with given options.
// Options example:
// {
//   "field": "id",
//   "ranges": {
//     "low": {
//       "from": 0,
//       "to": 500
//     },
//     "medium": {
//       "from": 500,
//       "to": 1000
//     },
//     "high": {
//       "from": 1000,
//       "to": 1500
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

		from, ok := rangeMap["from"].(float64)
		if !ok {
			return nil, fmt.Errorf("range %v from option is unexpected: %v", name, rangeMap["from"])
		}

		to, ok := rangeMap["to"].(float64)
		if !ok {
			return nil, fmt.Errorf("range %v to option is unexpected: %v", name, rangeMap["to"])
		}

		rangesAgg.AddRange(aggregations.NamedRange(name, from, to))
	}

	return rangesAgg, nil
}
