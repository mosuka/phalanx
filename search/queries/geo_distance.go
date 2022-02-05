package queries

import (
	"encoding/json"

	"github.com/blugelabs/bluge"
)

type GeoDistanceQueryOptions struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	Distance  string  `json:"distance"`
	Field     string  `json:"field"`
	Boost     float64 `json:"boost"`
}

func NewGeoDistanceQueryOptions() GeoDistanceQueryOptions {
	return GeoDistanceQueryOptions{}
}

// Create new GeoDistanceQuery with given options.
// Options example:
// {
//   "longitude": 40.73,
//   "latitude": -74.1,
//   "distance": "1km",
//   "field": "location",
//   "boost": 1.0
// }
func NewGeoDistanceQueryWithMap(opts map[string]interface{}) (*bluge.GeoDistanceQuery, error) {
	bytes, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	options := NewGeoDistanceQueryOptions()
	if err := json.Unmarshal(bytes, &options); err != nil {
		return nil, err
	}

	return NewGeoDistanceQueryWithOptions(options)

}

func NewGeoDistanceQueryWithOptions(opts GeoDistanceQueryOptions) (*bluge.GeoDistanceQuery, error) {
	geoDistanceQuery := bluge.NewGeoDistanceQuery(opts.Longitude, opts.Latitude, opts.Distance)

	// field is optional.
	if opts.Field != "" {
		geoDistanceQuery.SetField(opts.Field)
	}

	// boost is optional.
	if opts.Boost >= 0.0 {
		geoDistanceQuery.SetBoost(opts.Boost)
	}

	return geoDistanceQuery, nil
}
