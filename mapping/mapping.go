package mapping

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/analysis"
	"github.com/blugelabs/bluge/analysis/analyzer"
	"github.com/blugelabs/bluge/numeric/geo"
	phalanxanalyzer "github.com/mosuka/phalanx/analysis/analyzer"
	"github.com/mosuka/phalanx/errors"
)

const IdFieldName = "_id"
const TimestampFieldName = "_timestamp"
const ScoreFieldName = "_score"
const AllFieldName = "_all"

const DefaultTextFieldOptions = bluge.Index | bluge.Store | bluge.SearchTermPositions | bluge.HighlightMatches
const DefaultNumericFieldOptions = bluge.Index | bluge.Store | bluge.Sortable | bluge.Aggregatable
const DefaultDateTimeFieldOptions = bluge.Index | bluge.Store | bluge.Sortable | bluge.Aggregatable
const DefaultGeoPointFieldOptions = bluge.Index | bluge.Store | bluge.Sortable | bluge.Aggregatable

func IsDateTime(value interface{}) bool {
	strValue, ok := value.(string)
	if !ok {
		return false
	}

	if _, err := time.Parse(time.RFC3339, strValue); err != nil {
		return false
	}

	return true
}

func MakeDateTime(value interface{}) (time.Time, error) {
	strValue, ok := value.(string)
	if !ok {
		return time.Time{}, fmt.Errorf("value is not string")
	}

	return MakeDateTimeWithRfc3339(strValue)
}

func MakeDateTimeWithRfc3339(value string) (time.Time, error) {
	datetimeValue, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, err
	}

	return datetimeValue, nil
}

func IsGeoPoint(value interface{}) bool {
	valueMap, ok := value.(map[string]interface{})
	if !ok {
		return false
	}

	_, hasLat := valueMap["lat"]
	_, hasLon := valueMap["lon"]

	if len(valueMap) != 2 || !hasLat || !hasLon {
		return false
	}

	return true
}

func MakeGeoPoint(value interface{}) (geo.Point, error) {
	fieldValueMap, ok := value.(map[string]interface{})
	if !ok {
		return geo.Point{}, fmt.Errorf("value is not map[string]interface{}")
	}

	return MakeGeoPointWithMap(fieldValueMap)
}

func MakeGeoPointWithMap(value map[string]interface{}) (geo.Point, error) {
	_, hasLat := value["lat"]
	_, hasLon := value["lon"]

	if len(value) != 2 || !hasLat || !hasLon {
		return geo.Point{}, fmt.Errorf("unexpected geo point value")
	}

	valueBytes, err := json.Marshal(value)
	if err != nil {
		return geo.Point{}, err
	}

	var geoPoint geo.Point
	json.Unmarshal(valueBytes, &geoPoint)
	if err != nil {
		return geo.Point{}, err
	}

	return geoPoint, nil
}

func MakeTextField(fieldName string, fieldValue string, fieldOptions bluge.FieldOptions, analyzer bluge.Analyzer) *bluge.TermField {
	field := bluge.NewTextField(fieldName, fieldValue)
	field.FieldOptions = fieldOptions
	field.WithAnalyzer(analyzer)
	return field
}

func MakeNumericField(fieldName string, fieldValue float64, fieldOptions bluge.FieldOptions) *bluge.TermField {
	field := bluge.NewNumericField(fieldName, fieldValue)
	field.FieldOptions = fieldOptions
	return field
}

func MakeDateTimeField(fieldName string, fieldValue time.Time, fieldOptions bluge.FieldOptions) *bluge.TermField {
	field := bluge.NewDateTimeField(fieldName, fieldValue)
	field.FieldOptions = fieldOptions
	return field
}

func MakeGeoPointField(fieldName string, fieldValue geo.Point, fieldOptions bluge.FieldOptions) *bluge.TermField {
	field := bluge.NewGeoPointField(fieldName, fieldValue.Lat, fieldValue.Lon)
	field.FieldOptions = fieldOptions
	return field
}

type FieldType string

const (
	TextField     FieldType = "text"
	NumericField  FieldType = "numeric"
	DatetimeField FieldType = "datetime"
	GeoPointField FieldType = "geo_point"
)

type FieldOptions struct {
	Index         bool `json:"index"`
	Store         bool `json:"store"`
	TermPositions bool `json:"term_positions"`
	Highlight     bool `json:"highlight"`
	Sortable      bool `json:"sortable"`
	Aggregatable  bool `json:"aggregatable"`
}

type FieldSetting struct {
	FieldType       FieldType                       `json:"type"`
	FieldOptions    FieldOptions                    `json:"options"`
	AnalyzerSetting phalanxanalyzer.AnalyzerSetting `json:"analyzer"`
}

type IndexMapping map[string]FieldSetting

func NewMapping(source []byte) (IndexMapping, error) {
	indexMapping := make(IndexMapping)

	if err := json.Unmarshal(source, &indexMapping); err != nil {
		return nil, err
	}

	return indexMapping, nil
}

func (m IndexMapping) getFieldSetting(fieldName string) (*FieldSetting, error) {
	fieldSetting, ok := m[fieldName]
	if !ok {
		return nil, errors.ErrFieldSettingDoesNotExist
	}

	return &fieldSetting, nil
}

func (m IndexMapping) Exists(fieldName string) bool {
	_, ok := m[fieldName]

	return ok
}

func (m IndexMapping) GetFieldType(fieldName string) (FieldType, error) {
	if m.Exists(fieldName) {
		fieldSetting, err := m.getFieldSetting(fieldName)
		if err != nil {
			return "", err
		}
		return fieldSetting.FieldType, nil
	} else {
		fieldNameSlice := strings.Split(fieldName, "_")
		fieldType := FieldType(fieldNameSlice[len(fieldNameSlice)-1])
		switch fieldType {
		case TextField:
			return TextField, nil
		case NumericField:
			return NumericField, nil
		case DatetimeField:
			return DatetimeField, nil
		case GeoPointField:
			return GeoPointField, nil
		default:
			return "", errors.ErrUnknownFieldType
		}
	}
}

func (m IndexMapping) GetFieldOptions(fieldName string) (bluge.FieldOptions, error) {
	fieldSetting, err := m.getFieldSetting(fieldName)
	if err != nil {
		return 0, err
	}

	var fieldOptions bluge.FieldOptions
	if fieldSetting.FieldOptions.Index {
		fieldOptions = fieldOptions | bluge.Index
	}
	if fieldSetting.FieldOptions.Store {
		fieldOptions = fieldOptions | bluge.Store
	}
	if fieldSetting.FieldOptions.TermPositions {
		fieldOptions = fieldOptions | bluge.SearchTermPositions
	}
	if fieldSetting.FieldOptions.Highlight {
		fieldOptions = fieldOptions | bluge.HighlightMatches
	}
	if fieldSetting.FieldOptions.Sortable {
		fieldOptions = fieldOptions | bluge.Sortable
	}
	if fieldSetting.FieldOptions.Aggregatable {
		fieldOptions = fieldOptions | bluge.Aggregatable
	}

	return fieldOptions, nil
}

func (m IndexMapping) GetAnalyzer(fieldName string) (*analysis.Analyzer, error) {
	fieldSetting, err := m.getFieldSetting(fieldName)
	if err != nil {
		return nil, err
	}

	return phalanxanalyzer.NewAnalyzer(fieldSetting.AnalyzerSetting)
}

func (m IndexMapping) MakeDocument(fieldMap map[string]interface{}) (*bluge.Document, error) {
	id, ok := fieldMap[IdFieldName].(string)
	if !ok {
		return nil, errors.ErrDocumentIdDoesNotExist
	}

	// Create document.
	doc := bluge.NewDocument(id)

	// Add timestamp field.
	timestampField := bluge.NewDateTimeField(TimestampFieldName, time.Now().UTC())
	timestampField.FieldOptions = bluge.Index | bluge.Store | bluge.Sortable | bluge.Aggregatable
	doc.AddField(timestampField)

	for fieldName, fieldValueIntr := range fieldMap {
		fieldValues := make([]interface{}, 0)
		switch value := fieldValueIntr.(type) {
		case []interface{}:
			fieldValues = value
		default:
			fieldValues = append(fieldValues, value)
		}

		for _, fieldValue := range fieldValues {
			// Skip system reserved field name.
			switch fieldName {
			case IdFieldName:
				continue
			case TimestampFieldName:
				continue
			case AllFieldName:
				continue
			}

			var field *bluge.TermField
			fieldType, err := m.GetFieldType(fieldName)
			if err != nil {
				return nil, err
			}
			switch fieldType {
			case TextField:
				strValue, ok := fieldValue.(string)
				if !ok {
					return nil, fmt.Errorf("unexpected string value")
				}
				fieldOptions, err := m.GetFieldOptions(fieldName)
				if err != nil {
					fieldOptions = DefaultTextFieldOptions
				}
				fieldAnalyzer, err := m.GetAnalyzer(fieldName)
				if err != nil {
					fieldAnalyzer = analyzer.NewStandardAnalyzer()
				}
				field = MakeTextField(fieldName, strValue, fieldOptions, fieldAnalyzer)
			case NumericField:
				f64Value, ok := fieldValue.(float64)
				if !ok {
					return nil, fmt.Errorf("unexpected numeric value")
				}
				fieldOptions, err := m.GetFieldOptions(fieldName)
				if err != nil {
					fieldOptions = DefaultNumericFieldOptions
				}
				field = MakeNumericField(fieldName, f64Value, fieldOptions)
			case DatetimeField:
				datetimeValue, err := MakeDateTime(fieldValue)
				if err != nil {
					return nil, fmt.Errorf("unexpected datetime value")
				}
				fieldOptions, err := m.GetFieldOptions(fieldName)
				if err != nil {
					fieldOptions = DefaultDateTimeFieldOptions
				}
				field = MakeDateTimeField(fieldName, datetimeValue, fieldOptions)
			case GeoPointField:
				geoPointValue, err := MakeGeoPoint(fieldValue)
				if err != nil {
					return nil, fmt.Errorf("unexpected geo point value")
				}
				fieldOptions, err := m.GetFieldOptions(fieldName)
				if err != nil {
					fieldOptions = DefaultGeoPointFieldOptions
				}
				field = MakeGeoPointField(fieldName, geoPointValue, fieldOptions)
			}
			doc.AddField(field)
		}
	}

	// add _all field
	doc.AddField(bluge.NewCompositeFieldExcluding(AllFieldName, []string{IdFieldName, TimestampFieldName}))

	return doc, nil
}
