package queries

import (
	"encoding/json"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/numeric/geo"
)

type GeoBoundingBoxQueryOptions struct {
	TopLeftPoint     geo.Point `json:"top_left_point"`
	BottomRightPoint geo.Point `json:"bottom_right_point"`
	Field            string    `json:"field"`
	Boost            float64   `json:"boost"`
}

func NewGeoBoundingBoxQueryOptions() GeoBoundingBoxQueryOptions {
	return GeoBoundingBoxQueryOptions{}
}

// Create new GeoBoundingBoxQuery with given options.
// Options example:
// {
//   "top_left_point": {
//     "lon": 40.73,
//     "lat": -74.1
//   },
//   "bottom_right_point": {
//     "lon": 40.73,
//     "lat": -74.1
//   },
//   "field": "location",
//   "boost": 1.0
// }
func NewGeoBoundingBoxQueryWithMap(opts map[string]interface{}) (*bluge.GeoBoundingBoxQuery, error) {
	bytes, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	options := NewGeoBoundingBoxQueryOptions()
	if err := json.Unmarshal(bytes, &options); err != nil {
		return nil, err
	}

	return NewGeoBoundingBoxQueryWithOptions(options)
}

func NewGeoBoundingBoxQueryWithOptions(opts GeoBoundingBoxQueryOptions) (*bluge.GeoBoundingBoxQuery, error) {
	geoBoundingBoxQuery := bluge.NewGeoBoundingBoxQuery(opts.TopLeftPoint.Lon, opts.TopLeftPoint.Lat, opts.BottomRightPoint.Lon, opts.BottomRightPoint.Lat)

	// field is optional.
	if opts.Field != "" {
		geoBoundingBoxQuery.SetField(opts.Field)
	}

	// boost is optional.
	if opts.Boost >= 0.0 {
		geoBoundingBoxQuery.SetBoost(opts.Boost)
	}

	return geoBoundingBoxQuery, nil
}
