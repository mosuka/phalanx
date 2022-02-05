package queries

import (
	"encoding/json"

	"github.com/blugelabs/bluge"
)

type GeoBoundingBoxQueryOptions struct {
	TopLeftLongitude     float64 `json:"top_left_longitude"`
	TopLeftLatitude      float64 `json:"top_left_latitude"`
	BottomRightLongitude float64 `json:"bottom_right_longitude"`
	BottomRightLatitude  float64 `json:"bottom_right_latitude"`
	Field                string  `json:"field"`
	Boost                float64 `json:"boost"`
}

func NewGeoBoundingBoxQueryOptions() GeoBoundingBoxQueryOptions {
	return GeoBoundingBoxQueryOptions{}
}

// Create new GeoBoundingBoxQuery with given options.
// Options example:
// {
//   "top_left_longitude": 40.73,
//   "top_left_latitude": -74.1,
//   "bottom_right_longitude": 40.73,
//   "bottom_right_latitude": -74.1,
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
	geoBoundingBoxQuery := bluge.NewGeoBoundingBoxQuery(opts.TopLeftLongitude, opts.TopLeftLatitude, opts.BottomRightLongitude, opts.BottomRightLatitude)

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
