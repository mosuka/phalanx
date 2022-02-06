package queries

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/blugelabs/bluge"
)

var (
	MinTime = time.Unix(0, 0).UTC()
	MaxTime = time.Unix(1<<63-1, 999999999).UTC()
)

type DateRangeQueryOptions struct {
	Start          string  `json:"start"`
	End            string  `json:"end"`
	InclusiveStart bool    `json:"inclusive_start"`
	InclusiveEnd   bool    `json:"inclusive_end"`
	Field          string  `json:"field"`
	Boost          float64 `json:"boost"`
}

func NewDateRangeQueryOptions() DateRangeQueryOptions {
	return DateRangeQueryOptions{
		Start:          MinTime.Format(time.RFC3339),
		End:            MaxTime.Format(time.RFC3339),
		InclusiveStart: true,
		InclusiveEnd:   false,
	}
}

// Create new DateRangeQuery with given options.
// Options example:
// {
//   "start": "2022-01-01T00:00:00Z",
//   "end": "2023-01-01T00:00:00Z",
//   "inclusive_start": true,
//   "inclusive_end": false,
//   "field": "description",
//   "boost": 1.0
// }
func NewDateRangeQueryWithMap(opts map[string]interface{}) (*bluge.DateRangeQuery, error) {
	bytes, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	options := NewDateRangeQueryOptions()
	if err := json.Unmarshal(bytes, &options); err != nil {
		return nil, err
	}

	return NewDateRangeQueryWithOptions(options)

}

func NewDateRangeQueryWithOptions(opts DateRangeQueryOptions) (*bluge.DateRangeQuery, error) {
	var start time.Time
	if opts.Start != "" {
		var err error
		start, err = time.Parse(time.RFC3339, opts.Start)
		if err != nil {
			return nil, fmt.Errorf("start option is unexpected: %v", opts.Start)
		}
	}

	var end time.Time
	if opts.End != "" {
		var err error
		end, err = time.Parse(time.RFC3339, opts.End)
		if err != nil {
			return nil, fmt.Errorf("end option is unexpected: %v", opts.End)
		}
	}

	dateRangeQuery := bluge.NewDateRangeInclusiveQuery(start, end, opts.InclusiveStart, opts.InclusiveEnd)

	// field is optional.
	if opts.Field != "" {
		dateRangeQuery.SetField(opts.Field)
	}

	// boost is optional.
	if opts.Boost >= 0.0 {
		dateRangeQuery.SetBoost(opts.Boost)
	}

	return dateRangeQuery, nil
}
