package aggregations

import (
	"encoding/json"
	"sort"

	"github.com/blugelabs/bluge/search"
	"github.com/mosuka/phalanx/proto"
)

type AggregationType int

const (
	AggregationTypeUnknown AggregationType = iota
	AggregationTypeTerms
	AggregationTypeRange
	AggregationTypeDateRange
	AggregationTypeSum
	AggregationTypeMin
	AggregationTypeMax
	AggregationTypeAvg
)

// Maps for AggregationType.
var (
	AggregationType_name = map[AggregationType]string{
		AggregationTypeUnknown:   "unknown",
		AggregationTypeTerms:     "terms",
		AggregationTypeRange:     "range",
		AggregationTypeDateRange: "date_range",
		AggregationTypeSum:       "sum",
		AggregationTypeMin:       "min",
		AggregationTypeMax:       "max",
		AggregationTypeAvg:       "avg",
	}
	AggregationType_value = map[string]AggregationType{
		"unknown":    AggregationTypeUnknown,
		"terms":      AggregationTypeTerms,
		"range":      AggregationTypeRange,
		"date_range": AggregationTypeDateRange,
		"sum":        AggregationTypeSum,
		"min":        AggregationTypeMin,
		"max":        AggregationTypeMax,
		"avg":        AggregationTypeAvg,
	}
)

func NewAggregations(requests map[string]*proto.AggregationRequest) (map[string]search.Aggregation, error) {
	aggs := make(map[string]search.Aggregation)
	for name, request := range requests {
		switch request.Type {
		case AggregationType_name[AggregationTypeTerms]:
			opts := make(map[string]interface{})
			if err := json.Unmarshal(request.Options, &opts); err != nil {
				return nil, err
			}
			agg, err := NewTermsAggregationWithOptions(opts)
			if err != nil {
				return nil, err
			}

			aggs[name] = agg
		case AggregationType_name[AggregationTypeRange]:
			opts := make(map[string]interface{})
			if err := json.Unmarshal(request.Options, &opts); err != nil {
				return nil, err
			}
			agg, err := NewRangeAggregationWithOptions(opts)
			if err != nil {
				return nil, err
			}

			aggs[name] = agg
		case AggregationType_name[AggregationTypeDateRange]:
			opts := make(map[string]interface{})
			if err := json.Unmarshal(request.Options, &opts); err != nil {
				return nil, err
			}
			agg, err := NewDateRangeAggregationWithOptions(opts)
			if err != nil {
				return nil, err
			}

			aggs[name] = agg
		case AggregationType_name[AggregationTypeSum]:
			opts := make(map[string]interface{})
			if err := json.Unmarshal(request.Options, &opts); err != nil {
				return nil, err
			}
			agg, err := NewSumWithOptions(opts)
			if err != nil {
				return nil, err
			}

			aggs[name] = agg
		case AggregationType_name[AggregationTypeMin]:
			opts := make(map[string]interface{})
			if err := json.Unmarshal(request.Options, &opts); err != nil {
				return nil, err
			}
			agg, err := NewMinWithOptions(opts)
			if err != nil {
				return nil, err
			}

			aggs[name] = agg
		case AggregationType_name[AggregationTypeMax]:
			opts := make(map[string]interface{})
			if err := json.Unmarshal(request.Options, &opts); err != nil {
				return nil, err
			}
			agg, err := NewMaxWithOptions(opts)
			if err != nil {
				return nil, err
			}

			aggs[name] = agg
		case AggregationType_name[AggregationTypeAvg]:
			opts := make(map[string]interface{})
			if err := json.Unmarshal(request.Options, &opts); err != nil {
				return nil, err
			}
			agg, err := NewAvgWithOptions(opts)
			if err != nil {
				return nil, err
			}

			aggs[name] = agg
		}
	}
	return aggs, nil
}

func SortByCount(values map[string]float64) PairList {
	pl := make(PairList, len(values))
	i := 0
	for k, v := range values {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}

type Pair struct {
	Name  string
	Count float64
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Count < p[j].Count }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
