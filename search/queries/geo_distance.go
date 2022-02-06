package queries

import (
	"encoding/json"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/numeric/geo"
)

type GeoDistanceQueryOptions struct {
	Point    geo.Point `json:"point"`
	Distance string    `json:"distance"`
	Field    string    `json:"field"`
	Boost    float64   `json:"boost"`
}

func NewGeoDistanceQueryOptions() GeoDistanceQueryOptions {
	return GeoDistanceQueryOptions{}
}

// Create new GeoDistanceQuery with given options.
// Options example:
// {
//   "point": {
//     "lon": 40.73,
//     "lat": -74.1
//   },
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
	geoDistanceQuery := bluge.NewGeoDistanceQuery(opts.Point.Lon, opts.Point.Lat, opts.Distance)

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
