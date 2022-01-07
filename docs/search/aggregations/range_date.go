package aggregations

import (
	"fmt"

	"github.com/blugelabs/bluge/search"
	"github.com/blugelabs/bluge/search/aggregations"
	"github.com/mosuka/phalanx/mapping"
)

// Create new DateRangeAggregation with given options.
// Each bucket represents the number of documents
// that the condition between `start` or more and less than `end`.
// start <= datetime < end
// Options example:
// {
//   "field": "timestamp",
//   "ranges": {
//     "year_before_last": {
//       "start": "2020-01-01T00:00:00Z",
//       "end": "2021-01-01T00:00:00Z"
//     },
//     "last_year": {
//       "start": "2021-01-01T00:00:00Z",
//       "end": "2022-01-01T00:00:00Z"
//     },
//     "this_year": {
//       "start": "2022-01-01T00:00:00Z",
//       "end": "2023-01-01T00:00:00Z"
//     }
//   }
// }
func NewDateRangeAggregationWithOptions(opts map[string]interface{}) (*aggregations.DateRangeAggregation, error) {
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

	dateRangesAgg := aggregations.DateRanges(search.Field(field))

	ranges, ok := opts["ranges"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("ranges option does not exist")
	}
	for name, rangeValue := range ranges {
		rangeMap, ok := rangeValue.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("range %v option is unexpected: %v", name, rangeValue)
		}

		startStr, ok := rangeMap["start"].(string)
		if !ok {
			return nil, fmt.Errorf("range %v start option is unexpected: %v", name, rangeMap["start"])
		}
		start, err := mapping.MakeDateTimeWithRfc3339(startStr)
		if err != nil {
			return nil, fmt.Errorf("range %v start option is unexpected: %v", name, startStr)
		}

		endStr, ok := rangeMap["end"].(string)
		if !ok {
			return nil, fmt.Errorf("range %v end option is unexpected: %v", name, rangeMap["high"])
		}
		end, err := mapping.MakeDateTimeWithRfc3339(endStr)
		if err != nil {
			return nil, fmt.Errorf("range %v start option is unexpected: %v", name, endStr)
		}

		dateRangesAgg.AddRange(aggregations.NewNamedDateRange(name, start, end))
	}

	return dateRangesAgg, nil
}
