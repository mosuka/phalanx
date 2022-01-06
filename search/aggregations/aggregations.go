package aggregations

import "sort"

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
