package aggregations

import (
	"fmt"

	"github.com/blugelabs/bluge/search"
	"github.com/blugelabs/bluge/search/aggregations"
)

// Create new Sum with given options.
// Options example:
// {
//   "field": "price",
// }
func NewSumWithOptions(opts map[string]interface{}) (*aggregations.SingleValueMetric, error) {
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

	return aggregations.Sum(search.Field(field)), nil
}

// Create new Min with given options.
// Options example:
// {
//   "field": "price",
// }
func NewMinWithOptions(opts map[string]interface{}) (*aggregations.SingleValueMetric, error) {
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

	return aggregations.Min(search.Field(field)), nil
}

// Create new Max with given options.
// Options example:
// {
//   "field": "price",
// }
func NewMaxWithOptions(opts map[string]interface{}) (*aggregations.SingleValueMetric, error) {
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

	return aggregations.Max(search.Field(field)), nil
}

// Create new Avg with given options.
// Options example:
// {
//   "field": "price",
// }
func NewAvgWithOptions(opts map[string]interface{}) (*aggregations.WeightedAvgMetric, error) {
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

	return aggregations.Avg(search.Field(field)), nil
}
