package queries

import (
	"encoding/json"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/numeric/geo"
)

type GeoBoundingPolygonQueryOptions struct {
	Points []geo.Point `json:"points"`
	Field  string      `json:"field"`
	Boost  float64     `json:"boost"`
}

func NewGeoBoundingPolygonQueryOptions() GeoBoundingPolygonQueryOptions {
	return GeoBoundingPolygonQueryOptions{}
}

// Create new GeoBoundingPolygonQuery with given options.
// Options example:
// {
//   "points": [
//     {
//       "lon": 40.73,
//       "lat": -74.1
//     },
//     {
//       "lon": 40.73,
//       "lat": -74.1
//     },
//     {
//       "lon": 40.73,
//       "lat": -74.1
//     },
//     {
//       "lon": 40.73,
//       "lat": -74.1
//     },
//     {
//       "lon": 40.73,
//       "lat": -74.1
//     }
//   ],
//   "field": "location",
//   "boost": 1.0
// }
func NewGeoBoundingPolygonQueryWithMap(opts map[string]interface{}) (*bluge.GeoBoundingPolygonQuery, error) {
	bytes, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	options := NewGeoBoundingPolygonQueryOptions()
	if err := json.Unmarshal(bytes, &options); err != nil {
		return nil, err
	}

	return NewGeoBoundingPolygonQueryWithOptions(options)
}

func NewGeoBoundingPolygonQueryWithOptions(opts GeoBoundingPolygonQueryOptions) (*bluge.GeoBoundingPolygonQuery, error) {
	geoBoundingPolygonQuery := bluge.NewGeoBoundingPolygonQuery(opts.Points)

	// field is optional.
	if opts.Field != "" {
		geoBoundingPolygonQuery.SetField(opts.Field)
	}

	// boost is optional.
	if opts.Boost >= 0.0 {
		geoBoundingPolygonQuery.SetBoost(opts.Boost)
	}

	return geoBoundingPolygonQuery, nil
}
