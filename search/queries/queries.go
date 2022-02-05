package queries

import (
	"github.com/blugelabs/bluge"
	"github.com/mosuka/phalanx/errors"
)

type QueryType int

const (
	QueryTypeUnknown QueryType = iota
	QueryTypeBoolean
	QueryTypeDateRange
	QueryTypeFuzzy
	QueryTypeGeoBoundingBox
	QueryTypeGeoBoundingPolygon
	QueryTypeGeoDistance
	QueryTypeMatch
	QueryTypeMatchAll
	QueryTypeMatchNone
	QueryTypeMatchPhrase
	QueryTypeMultiPhrase
	QueryTypeNumericRange
	QueryTypePrefix
	QueryTypeQueryString
	QueryTypeRegexp
	QueryTypeTerm
	QueryTypeTermRange
	QueryTypeWildcard
)

// Maps for QueryType.
var (
	QueryType_name = map[QueryType]string{
		QueryTypeUnknown:            "unknown",
		QueryTypeBoolean:            "boolean",
		QueryTypeDateRange:          "date_range",
		QueryTypeFuzzy:              "fuzzy",
		QueryTypeGeoBoundingBox:     "geo_bounding_box",
		QueryTypeGeoBoundingPolygon: "geo_bounding_polygon",
		QueryTypeGeoDistance:        "geo_distance",
		QueryTypeMatch:              "match",
		QueryTypeMatchAll:           "match_all",
		QueryTypeMatchNone:          "match_none",
		QueryTypeMatchPhrase:        "match_phrase",
		QueryTypeMultiPhrase:        "multi_phrase",
		QueryTypeNumericRange:       "numeric_range",
		QueryTypePrefix:             "prefix",
		QueryTypeQueryString:        "query_string",
		QueryTypeRegexp:             "regexp",
		QueryTypeTerm:               "term",
		QueryTypeTermRange:          "term_range",
		QueryTypeWildcard:           "wildcard",
	}
	QueryType_value = map[string]QueryType{
		"unknown":              QueryTypeUnknown,
		"boolean":              QueryTypeBoolean,
		"date_range":           QueryTypeDateRange,
		"fuzzy":                QueryTypeFuzzy,
		"geo_bounding_box":     QueryTypeGeoBoundingBox,
		"geo_bounding_polygon": QueryTypeGeoBoundingPolygon,
		"geo_distance":         QueryTypeGeoDistance,
		"match":                QueryTypeMatch,
		"match_all":            QueryTypeMatchAll,
		"match_none":           QueryTypeMatchNone,
		"match_phrase":         QueryTypeMatchPhrase,
		"multi_phrase":         QueryTypeMultiPhrase,
		"numeric_range":        QueryTypeNumericRange,
		"prefix":               QueryTypePrefix,
		"query_string":         QueryTypeQueryString,
		"regexp":               QueryTypeRegexp,
		"term":                 QueryTypeTerm,
		"term_range":           QueryTypeTermRange,
		"wildcard":             QueryTypeWildcard,
	}
)

func NewQuery(queryType string, queryOpts map[string]interface{}) (bluge.Query, error) {
	switch QueryType_value[queryType] {
	case QueryTypeBoolean:
		return NewBooleanQueryWithMap(queryOpts)
	case QueryTypeDateRange:
		return NewDateRangeQueryWithMap(queryOpts)
	case QueryTypeFuzzy:
		return NewFuzzyQueryWithMap(queryOpts)
	case QueryTypeGeoBoundingBox:
		return NewGeoBoundingBoxQueryWithMap(queryOpts)
	case QueryTypeGeoBoundingPolygon:
		return NewGeoBoundingPolygonQueryWithMap(queryOpts)
	case QueryTypeGeoDistance:
		return NewGeoDistanceQueryWithMap(queryOpts)
	case QueryTypeMatch:
		return NewMatchQueryWithMap(queryOpts)
	case QueryTypeMatchAll:
		return NewMatchAllQueryWithMap(queryOpts)
	case QueryTypeMatchNone:
		return NewMatchNoneQueryWithMap(queryOpts)
	case QueryTypeMatchPhrase:
		return NewMatchPhraseQueryWithMap(queryOpts)
	case QueryTypeMultiPhrase:
		return NewMultiPhraseQueryWithMap(queryOpts)
	case QueryTypeNumericRange:
		return NewNumericRangeQueryWithMap(queryOpts)
	case QueryTypePrefix:
		return NewPrefixQueryWithMap(queryOpts)
	case QueryTypeQueryString:
		return NewQueryStringQueryWithMap(queryOpts)
	case QueryTypeRegexp:
		return NewRegexpQueryWithMap(queryOpts)
	case QueryTypeTerm:
		return NewTermQueryWithMap(queryOpts)
	case QueryTypeTermRange:
		return NewTermRangeQueryWithMap(queryOpts)
	case QueryTypeWildcard:
		return NewWildcardQueryWithMap(queryOpts)
	default:
		return nil, errors.ErrUnknownQueryType
	}
}
