package mapping

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/analysis"
	"github.com/blugelabs/bluge/analysis/analyzer"
	"github.com/blugelabs/bluge/analysis/char"
	"github.com/blugelabs/bluge/analysis/token"
	"github.com/blugelabs/bluge/analysis/tokenizer"
	"github.com/blugelabs/bluge/numeric/geo"
	"github.com/ikawaha/blugeplugin/analysis/lang/ja"
	"github.com/ikawaha/kagome-dict/dict"
	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome-dict/uni"
	"github.com/mosuka/phalanx/errors"
	"golang.org/x/text/unicode/norm"
)

const IdFieldName = "_id"
const TimestampFieldName = "_timestamp"
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
	NumericField            = "numeric"
	DatetimeField           = "datetime"
	GeoPointField           = "geo_point"
)

type CharFilter string

const (
	AsciiFoldingCharFilter       CharFilter = "ascii_folding"
	HtmlCharFilter                          = "html"
	RegexpCharFilter                        = "regexp"
	UnicodeNormalizeCharFilter              = "unicode_normalize"
	ZeroWidthNonJoinerCharFilter            = "zero_width_non_joiner"
)

type Tokenizer string

const (
	CharacterTokenizer   Tokenizer = "character"
	ExceptionTokenizer             = "exception"
	JapaneseTokenizer              = "japanese"
	LetterTokenizer                = "letter"
	RegexpTokenizer                = "regexp"
	SingleTokenTokenizer           = "single_token"
	UnicodeTokenizer               = "unicode"
	WebTokenizer                   = "web"
	WhitespaceTokenizer            = "whitespace"
)

type TokenFilter string

const (
	ApostropheTokenFilter         TokenFilter = "apostrophe"
	CamelCaseTokenFilter                      = "camel_case"
	DictionaryCompoundTokenFilter             = "dictionary_compound"
	EdgeNgramTokenFilter                      = "edge_ngram"
	ElisionTokenFilter                        = "elision"
	KeywordMarkerTokenFilter                  = "keyword_marker"
	LengthTokenFilter                         = "length"
	LowerCaseTokenFilter                      = "lower_case"
	NgramTokenFilter                          = "ngram"
	PorterStemmerTokenFilter                  = "porter_stemmer"
	ReverseTokenFilter                        = "reverse"
	ShingleTokenFilter                        = "shingle"
	StopTokensTokenFilter                     = "stop_tokens"
	TruncateTokenFilter                       = "truncate"
	UnicodeNormalizeTokenFilter               = "unicode_normalize"
	UniqueTermTokenFilter                     = "unique_term"
)

type CharFilterSetting struct {
	Name    CharFilter             `json:"name"`
	Options map[string]interface{} `json:"options"`
}

type TokenizerSetting struct {
	Name    Tokenizer              `json:"name"`
	Options map[string]interface{} `json:"options"`
}

type TokenFilterSetting struct {
	Name    TokenFilter            `json:"name"`
	Options map[string]interface{} `json:"options"`
}

type AnalyzerSetting struct {
	CharFilterSettings  []CharFilterSetting  `json:"char_filters"`
	TokenizerSetting    TokenizerSetting     `json:"tokenizer"`
	TokenFilterSettings []TokenFilterSetting `json:"token_filters"`
}

type FieldOptions struct {
	Index         bool `json:"index"`
	Store         bool `json:"store"`
	TermPositions bool `json:"term_positions"`
	Highlight     bool `json:"highlight"`
	Sortable      bool `json:"sortable"`
	Aggregatable  bool `json:"aggregatable"`
}

type FieldSetting struct {
	FieldType       FieldType       `json:"type"`
	FieldOptions    FieldOptions    `json:"options"`
	AnalyzerSetting AnalyzerSetting `json:"analyzer"`
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

	// Char filter.
	charFilters := make([]analysis.CharFilter, 0)
	charFilterSettings := fieldSetting.AnalyzerSetting.CharFilterSettings
	for _, charFilterSetting := range charFilterSettings {
		switch charFilterSetting.Name {
		case AsciiFoldingCharFilter:
			// {
			// 	"name": "ascii_folding"
			// }
			charFilter := char.NewASCIIFoldingFilter()
			charFilters = append(charFilters, charFilter)
		case HtmlCharFilter:
			// {
			// 	"name": "html"
			// }
			charFilter := char.NewHTMLCharFilter()
			charFilters = append(charFilters, charFilter)
		case RegexpCharFilter:
			// {
			// 	"name": "regexp",
			// 	"options": {
			// 		"pattern": "foo",
			//      "replacement": "var"
			// 	}
			// }
			patternValue, ok := charFilterSetting.Options["pattern"]
			if !ok {
				return nil, fmt.Errorf("pattern option does not exist")
			}
			pattern, ok := patternValue.(string)
			if !ok {
				return nil, fmt.Errorf("form option is unexpected")
			}

			replacementValue, ok := charFilterSetting.Options["replacement"]
			if !ok {
				return nil, fmt.Errorf("pattern option does not exist")
			}
			replacement, ok := replacementValue.(string)
			if !ok {
				return nil, fmt.Errorf("form option is unexpected")
			}

			r := regexp.MustCompile(pattern)

			charFilter := char.NewRegexpCharFilter(r, []byte(replacement))
			charFilters = append(charFilters, charFilter)
		case UnicodeNormalizeCharFilter:
			// {
			// 	"name": "unicode_normalize",
			// 	"options": {
			// 		"form": "NFKC"
			// 	}
			// }
			formValue, ok := charFilterSetting.Options["form"]
			if !ok {
				return nil, fmt.Errorf("form option does not exist")
			}
			form, ok := formValue.(string)
			if !ok {
				return nil, fmt.Errorf("form option is unexpected")
			}
			var charFilter analysis.CharFilter
			switch form {
			case "NFC":
				charFilter = ja.NewUnicodeNormalizeCharFilter(norm.NFC)
			case "NFD":
				charFilter = ja.NewUnicodeNormalizeCharFilter(norm.NFD)
			case "NFKC":
				charFilter = ja.NewUnicodeNormalizeCharFilter(norm.NFKC)
			case "NFKD":
				charFilter = ja.NewUnicodeNormalizeCharFilter(norm.NFKD)
			default:
				err := fmt.Errorf("unknown form option")
				return nil, err
			}
			charFilters = append(charFilters, charFilter)
		case ZeroWidthNonJoinerCharFilter:
			// {
			// 	"name": "zero_width_non_joiner"
			// }
			charFilter := char.NewZeroWidthNonJoinerCharFilter()
			charFilters = append(charFilters, charFilter)
		default:
			err := fmt.Errorf("unknown char filter: %s\n", charFilterSetting.Name)
			return nil, err
		}
	}

	// Tokenizer.
	tokenizerSetting := fieldSetting.AnalyzerSetting.TokenizerSetting
	var fieldTokenizer analysis.Tokenizer
	switch tokenizerSetting.Name {
	case CharacterTokenizer:
		// {
		// 	"name": "character"
		// 	"options": {
		// 		"rune": "graphic"
		// 	}
		// }
		runeValue, ok := tokenizerSetting.Options["rune"]
		if !ok {
			return nil, fmt.Errorf("ruen option does not exist")
		}
		runeStr, ok := runeValue.(string)
		if !ok {
			return nil, fmt.Errorf("rune option is unexpected")
		}

		var rune func(r rune) bool
		switch runeStr {
		case "graphic":
			rune = unicode.IsGraphic
		case "print":
			rune = unicode.IsPrint
		case "control":
			rune = unicode.IsControl
		case "letter":
			rune = unicode.IsLetter
		case "mark":
			rune = unicode.IsMark
		case "number":
			rune = unicode.IsNumber
		case "punct":
			rune = unicode.IsPunct
		case "space":
			rune = unicode.IsSpace
		case "symbol":
			rune = unicode.IsSymbol
		default:
			err := fmt.Errorf("unknown rune option: %s\n", runeStr)
			return nil, err
		}
		fieldTokenizer = tokenizer.NewCharacterTokenizer(rune)
	case ExceptionTokenizer:
		// {
		// 	"name": "exception"
		// 	"options": {
		// 		"patterns": [
		//			"[hH][tT][tT][pP][sS]?://(\S)*",
		//			"[fF][iI][lL][eE]://(\S)*",
		//			"[fF][tT][pP]://(\S)*",
		//			"\S+@\S+"
		// 		]
		// 	}
		// }
		patternsValue, ok := tokenizerSetting.Options["patterns"]
		if !ok {
			return nil, fmt.Errorf("patterns option does not exist")
		}
		patterns, ok := patternsValue.([]interface{})
		if !ok {
			return nil, fmt.Errorf("patterns option is unexpected")
		}
		patternStrs := make([]string, 0)
		for _, pattern := range patterns {
			str, ok := pattern.(string)
			if !ok {
				return nil, fmt.Errorf("patterns option is unexpected")
			}
			patternStrs = append(patternStrs, str)
		}

		pattern := strings.Join(patternStrs, "|")
		r, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("patterns option is unexpected")
		}
		fieldTokenizer = tokenizer.NewExceptionsTokenizer(r, tokenizer.NewUnicodeTokenizer())
	case JapaneseTokenizer:
		// {
		// 	"name": "japanese",
		// 	"options": {
		// 		"dictionary": "IPADIC"
		// 		"stop_tags": [
		// 			"接続詞",
		// 			"助詞",
		// 			"助詞-格助詞",
		// 			"助詞-格助詞-一般",
		// 			"助詞-格助詞-引用",
		// 			"助詞-格助詞-連語",
		// 			"助詞-接続助詞",
		// 			"助詞-係助詞",
		// 			"助詞-副助詞",
		// 			"助詞-間投助詞",
		// 			"助詞-並立助詞",
		// 			"助詞-終助詞",
		// 			"助詞-副助詞／並立助詞／終助詞",
		// 			"助詞-連体化",
		// 			"助詞-副詞化",
		// 			"助詞-特殊",
		// 			"助動詞",
		// 			"記号",
		// 			"記号-一般",
		// 			"記号-読点",
		// 			"記号-句点",
		// 			"記号-空白",
		// 			"記号-括弧開",
		// 			"記号-括弧閉",
		// 			"その他-間投",
		// 			"フィラー",
		// 			"非言語音"
		// 		],
		// 		"base_forms": [
		// 			"動詞",
		// 			"形容詞",
		// 			"形容動詞"
		// 		]
		// 	}
		// }
		dictionaryValue, ok := tokenizerSetting.Options["dictionary"]
		if !ok {
			return nil, fmt.Errorf("dict option does not exist")
		}
		dictionaryStr, ok := dictionaryValue.(string)
		if !ok {
			return nil, fmt.Errorf("dict option is unexpected")
		}
		var dictionary *dict.Dict
		switch dictionaryStr {
		case "IPADIC":
			dictionary = ipa.Dict()
		case "UniDIC":
			dictionary = uni.Dict()
		}

		stopTagsValue, ok := tokenizerSetting.Options["stop_tags"]
		if !ok {
			return nil, fmt.Errorf("stop_tags option does not exist")
		}
		stopTags, ok := stopTagsValue.([]interface{})
		if !ok {
			return nil, fmt.Errorf("stop_tags option is unexpected")
		}
		stopTagsTokenMap := analysis.NewTokenMap()
		for _, stopTag := range stopTags {
			token, ok := stopTag.(string)
			if !ok {
				return nil, fmt.Errorf("stop_tag is unexpected: %v\n", stopTag)
			}
			stopTagsTokenMap.AddToken(token)
		}

		baseFormsValue, ok := tokenizerSetting.Options["base_forms"]
		if !ok {
			return nil, fmt.Errorf("stop_tags option does not exist")
		}
		baseForms, ok := baseFormsValue.([]interface{})
		if !ok {
			return nil, fmt.Errorf("stop_tags option is unexpected")
		}
		baseFormsTokenMap := analysis.NewTokenMap()
		for _, baseForm := range baseForms {
			token, ok := baseForm.(string)
			if !ok {
				return nil, fmt.Errorf("stop_tag is unexpected: %v\n", baseForm)
			}
			baseFormsTokenMap.AddToken(token)
		}

		fieldTokenizer = ja.NewJapaneseTokenizer(dictionary, ja.StopTagsFilter(stopTagsTokenMap), ja.BaseFormFilter(baseFormsTokenMap))
	case LetterTokenizer:
		// {
		// 	"name": "letter"
		// }
		fieldTokenizer = tokenizer.NewLetterTokenizer()
	case RegexpTokenizer:
		// {
		// 	"name": "regexp",
		// 	"options": {
		// 		"pattern": "[0-9a-zA-Z_]*"
		// 	}
		// }
		patternValue, ok := tokenizerSetting.Options["pattern"]
		if !ok {
			return nil, fmt.Errorf("pattern option does not exist")
		}
		pattern, ok := patternValue.(string)
		if !ok {
			return nil, fmt.Errorf("form option is unexpected")
		}

		r := regexp.MustCompile(pattern)

		fieldTokenizer = tokenizer.NewRegexpTokenizer(r)
	case SingleTokenTokenizer:
		// {
		// 	"name": "single_token"
		// }
		fieldTokenizer = tokenizer.NewSingleTokenTokenizer()
	case UnicodeTokenizer:
		// {
		// 	"name": "unicode"
		// }
		fieldTokenizer = tokenizer.NewUnicodeTokenizer()
	case WebTokenizer:
		// {
		// 	"name": "web"
		// }
		fieldTokenizer = tokenizer.NewWebTokenizer()
	case WhitespaceTokenizer:
		// {
		// 	"name": "whitespace"
		// }
		fieldTokenizer = tokenizer.NewWhitespaceTokenizer()
	default:
		err := fmt.Errorf("unknown tokenizer: %s\n", tokenizerSetting.Name)
		return nil, err
	}

	// Token filter.
	tokenFilters := make([]analysis.TokenFilter, 0)
	tokenFilterSettings := fieldSetting.AnalyzerSetting.TokenFilterSettings
	for _, tokenFilterSetting := range tokenFilterSettings {
		switch tokenFilterSetting.Name {
		case ApostropheTokenFilter:
			// {
			// 	"name": "apostrophe"
			// }
			tokenFilter := token.NewApostropheFilter()
			tokenFilters = append(tokenFilters, tokenFilter)
		case CamelCaseTokenFilter:
			// {
			// 	"name": "camel_case"
			// }
			tokenFilter := token.NewCamelCaseFilter()
			tokenFilters = append(tokenFilters, tokenFilter)
		case DictionaryCompoundTokenFilter:
			// {
			// 	"name": "dictionary_compound",
			// 	"options": {
			// 		"words": [
			// 			"soft",
			// 			"softest",
			// 			"ball"
			// 		],
			// 		"min_word_size": 5,
			// 		"min_sub_word_size": 2,
			// 		"max_sub_word_size": 15,
			// 		"only_longest_match": false
			// }
			wordsValue, ok := tokenFilterSetting.Options["words"]
			if !ok {
				return nil, fmt.Errorf("words option does not exist")
			}
			words, ok := wordsValue.([]interface{})
			if !ok {
				return nil, fmt.Errorf("words option is unexpected")
			}
			wordMap := analysis.NewTokenMap()
			for _, word := range words {
				str, ok := word.(string)
				if !ok {
					return nil, fmt.Errorf("word is unexpected")
				}
				wordMap.AddToken(str)
			}

			minWordSizeValue, ok := tokenFilterSetting.Options["min_word_size"]
			if !ok {
				return nil, fmt.Errorf("min_word_size option does not exist")
			}
			minWordSizeNum, ok := minWordSizeValue.(float64)
			if !ok {
				return nil, fmt.Errorf("min_word_size option is unexpected")
			}
			minWordSize := int(minWordSizeNum)

			minSubWordSizeValue, ok := tokenFilterSetting.Options["min_sub_word_size"]
			if !ok {
				return nil, fmt.Errorf("min_sub_word_size option does not exist")
			}
			minSubWordSizeNum, ok := minSubWordSizeValue.(float64)
			if !ok {
				return nil, fmt.Errorf("min_sub_word_size option is unexpected")
			}
			minSubWordSize := int(minSubWordSizeNum)

			maxSubWordSizeValue, ok := tokenFilterSetting.Options["max_sub_word_size"]
			if !ok {
				return nil, fmt.Errorf("max_sub_word_size option does not exist")
			}
			maxSubWordSizeNum, ok := maxSubWordSizeValue.(float64)
			if !ok {
				return nil, fmt.Errorf("max_sub_word_size option is unexpected")
			}
			maxSubWordSize := int(maxSubWordSizeNum)

			onlyLongestMatchValue, ok := tokenFilterSetting.Options["only_longest_match"]
			if !ok {
				return nil, fmt.Errorf("only_longest_match option does not exist")
			}
			onlyLongestMatch, ok := onlyLongestMatchValue.(bool)
			if !ok {
				return nil, fmt.Errorf("only_longest_match option is unexpected")
			}

			tokenFilter := token.NewDictionaryCompoundFilter(wordMap, minWordSize, minSubWordSize, maxSubWordSize, onlyLongestMatch)
			tokenFilters = append(tokenFilters, tokenFilter)
		case EdgeNgramTokenFilter:
			// {
			// 	"name": "edge_ngram",
			// 	"options": {
			// 		"back": false,
			// 		"min_length": 1,
			// 		"max_length": 2
			// }
			backValue, ok := tokenFilterSetting.Options["back"]
			if !ok {
				return nil, fmt.Errorf("back option does not exist")
			}
			back, ok := backValue.(bool)
			if !ok {
				return nil, fmt.Errorf("back option is unexpected")
			}
			var side token.Side
			if back {
				side = token.BACK
			} else {
				side = token.FRONT
			}

			minLengthValue, ok := tokenFilterSetting.Options["min_length"]
			if !ok {
				return nil, fmt.Errorf("min_length option does not exist")
			}
			minLengthNum, ok := minLengthValue.(float64)
			if !ok {
				return nil, fmt.Errorf("min_length option is unexpected")
			}
			minLength := int(minLengthNum)

			maxLengthValue, ok := tokenFilterSetting.Options["max_length"]
			if !ok {
				return nil, fmt.Errorf("max_length option does not exist")
			}
			maxLengthNum, ok := maxLengthValue.(float64)
			if !ok {
				return nil, fmt.Errorf("max_length option is unexpected")
			}
			maxLength := int(maxLengthNum)

			tokenFilter := token.NewEdgeNgramFilter(side, minLength, maxLength)
			tokenFilters = append(tokenFilters, tokenFilter)
		case ElisionTokenFilter:
			// {
			// 	"name": "elision",
			// 	"options": {
			// 		"articles": [
			// 			"ar"
			// 		]
			// 	}
			// }
			articlesValue, ok := tokenFilterSetting.Options["articles"]
			if !ok {
				return nil, fmt.Errorf("articles option does not exist")
			}
			articles, ok := articlesValue.([]interface{})
			if !ok {
				return nil, fmt.Errorf("articles option is unexpected")
			}
			articleMap := analysis.NewTokenMap()
			for _, article := range articles {
				str, ok := article.(string)
				if !ok {
					return nil, fmt.Errorf("articles is unexpected")
				}
				articleMap.AddToken(str)
			}

			tokenFilter := token.NewElisionFilter(articleMap)
			tokenFilters = append(tokenFilters, tokenFilter)
		case KeywordMarkerTokenFilter:
			// {
			// 	"name": "keyword_marker"
			// 	"options": {
			// 		"keywords": [
			// 			"walk",
			// 			"park"
			// 		]
			// 	}
			// }
			keywordsValue, ok := tokenFilterSetting.Options["keywords"]
			if !ok {
				return nil, fmt.Errorf("keywords option does not exist")
			}
			keywords, ok := keywordsValue.([]interface{})
			if !ok {
				return nil, fmt.Errorf("keywords option is unexpected")
			}
			keywordMap := analysis.NewTokenMap()
			for _, keyword := range keywords {
				str, ok := keyword.(string)
				if !ok {
					return nil, fmt.Errorf("keyword is unexpected")
				}
				keywordMap.AddToken(str)
			}

			tokenFilter := token.NewKeyWordMarkerFilter(keywordMap)
			tokenFilters = append(tokenFilters, tokenFilter)
		case LengthTokenFilter:
			// {
			// 	"name": "length",
			// 	"options": {
			// 		"min_length": 3,
			// 		"max_length": 4
			// 	}
			// }
			minLengthValue, ok := tokenFilterSetting.Options["min_length"]
			if !ok {
				return nil, fmt.Errorf("min_length option does not exist")
			}
			minLengthNum, ok := minLengthValue.(float64)
			if !ok {
				return nil, fmt.Errorf("min_length option is unexpected")
			}
			minLength := int(minLengthNum)

			maxLengthValue, ok := tokenFilterSetting.Options["max_length"]
			if !ok {
				return nil, fmt.Errorf("max_length option does not exist")
			}
			maxLengthNum, ok := maxLengthValue.(float64)
			if !ok {
				return nil, fmt.Errorf("max_length option is unexpected")
			}
			maxLength := int(maxLengthNum)

			tokenFilter := token.NewLengthFilter(minLength, maxLength)
			tokenFilters = append(tokenFilters, tokenFilter)
		case LowerCaseTokenFilter:
			// {
			// 	"name": "lower_case"
			// }
			tokenFilter := token.NewLowerCaseFilter()
			tokenFilters = append(tokenFilters, tokenFilter)
		case NgramTokenFilter:
			// {
			// 	"name": "ngram",
			// 	"options": {
			// 		"min_length": 1,
			// 		"max_length": 3
			// 	}
			// }
			minLengthValue, ok := tokenFilterSetting.Options["min_length"]
			if !ok {
				return nil, fmt.Errorf("min_length option does not exist")
			}
			minLengthNum, ok := minLengthValue.(float64)
			if !ok {
				return nil, fmt.Errorf("min_length option is unexpected")
			}
			minLength := int(minLengthNum)

			maxLengthValue, ok := tokenFilterSetting.Options["max_length"]
			if !ok {
				return nil, fmt.Errorf("max_length option does not exist")
			}
			maxLengthNum, ok := maxLengthValue.(float64)
			if !ok {
				return nil, fmt.Errorf("max_length option is unexpected")
			}
			maxLength := int(maxLengthNum)

			tokenFilter := token.NewNgramFilter(minLength, maxLength)
			tokenFilters = append(tokenFilters, tokenFilter)
		case PorterStemmerTokenFilter:
			// {
			// 	"name": "porter_stemmer"
			// }
			tokenFilter := token.NewPorterStemmer()
			tokenFilters = append(tokenFilters, tokenFilter)
		case ReverseTokenFilter:
			// {
			// 	"name": "reverse"
			// }
			tokenFilter := token.NewReverseFilter()
			tokenFilters = append(tokenFilters, tokenFilter)
		case ShingleTokenFilter:
			// {
			// 	"name": "shingle",
			// 	"options": {
			// 		"min_length": 2,
			// 		"max_length": 2,
			// 		"output_original": true,
			// 		"token_separator": " ",
			// 		"fill": "_"
			// 	}
			// }
			minLengthValue, ok := tokenFilterSetting.Options["min_length"]
			if !ok {
				return nil, fmt.Errorf("min_length option does not exist")
			}
			minLengthNum, ok := minLengthValue.(float64)
			if !ok {
				return nil, fmt.Errorf("min_length option is unexpected")
			}
			minLength := int(minLengthNum)

			maxLengthValue, ok := tokenFilterSetting.Options["max_length"]
			if !ok {
				return nil, fmt.Errorf("max_length option does not exist")
			}
			maxLengthNum, ok := maxLengthValue.(float64)
			if !ok {
				return nil, fmt.Errorf("max_length option is unexpected")
			}
			maxLength := int(maxLengthNum)

			outputOriginalValue, ok := tokenFilterSetting.Options["output_original"]
			if !ok {
				return nil, fmt.Errorf("output_original option does not exist")
			}
			outputOriginal, ok := outputOriginalValue.(bool)
			if !ok {
				return nil, fmt.Errorf("output_original option is unexpected")
			}

			tokenSeparatorValue, ok := tokenFilterSetting.Options["token_separator"]
			if !ok {
				return nil, fmt.Errorf("token_separator option does not exist")
			}
			tokenSeparator, ok := tokenSeparatorValue.(string)
			if !ok {
				return nil, fmt.Errorf("token_separator option is unexpected")
			}

			fillValue, ok := tokenFilterSetting.Options["fill"]
			if !ok {
				return nil, fmt.Errorf("fill option does not exist")
			}
			fill, ok := fillValue.(string)
			if !ok {
				return nil, fmt.Errorf("fill option is unexpected")
			}

			tokenFilter := token.NewShingleFilter(minLength, maxLength, outputOriginal, tokenSeparator, fill)
			tokenFilters = append(tokenFilters, tokenFilter)
		case StopTokensTokenFilter:
			// {
			// 	"name": "stop_tokens",
			// 	"options": {
			// 		"stop_tokens": [
			// 			"a",
			// 			"an",
			// 			"and",
			// 			"are",
			// 			"as",
			// 			"at",
			// 			"be",
			// 			"but",
			// 			"by",
			// 			"for",
			// 			"if",
			// 			"in",
			// 			"into",
			// 			"is",
			// 			"it",
			// 			"no",
			// 			"not",
			// 			"of",
			// 			"on",
			// 			"or",
			// 			"such",
			// 			"that",
			// 			"the",
			// 			"their",
			// 			"then",
			// 			"there",
			// 			"these",
			// 			"they",
			// 			"this",
			// 			"to",
			// 			"was",
			// 			"will",
			// 			"with"
			// 		]
			// 	}
			// }
			stopTokensValue, ok := tokenFilterSetting.Options["stop_tokens"]
			if !ok {
				return nil, fmt.Errorf("stop_tokens option does not exist")
			}
			stopTokens, ok := stopTokensValue.([]interface{})
			if !ok {
				return nil, fmt.Errorf("stop_tokens option is unexpected")
			}
			stopTokenMap := analysis.NewTokenMap()
			for _, stopToken := range stopTokens {
				token, ok := stopToken.(string)
				if !ok {
					return nil, fmt.Errorf("base_form is unexpected: %v\n", stopToken)
				}
				stopTokenMap.AddToken(token)
			}

			tokenFilter := token.NewStopTokensFilter(stopTokenMap)
			tokenFilters = append(tokenFilters, tokenFilter)
		case TruncateTokenFilter:
			// {
			// 	"name": "truncate",
			// 	"options": {
			// 		"length": 5
			// 	}
			// }
			lengthValue, ok := tokenFilterSetting.Options["length"]
			if !ok {
				return nil, fmt.Errorf("length option does not exist")
			}
			lengthNum, ok := lengthValue.(float64)
			if !ok {
				return nil, fmt.Errorf("length option is unexpected")
			}
			length := int(lengthNum)

			tokenFilter := token.NewTruncateTokenFilter(length)
			tokenFilters = append(tokenFilters, tokenFilter)
		case UnicodeNormalizeTokenFilter:
			// {
			// 	"name": "unicode_normalize",
			// 	"options": {
			// 		"form": "NFKC"
			// 	}
			// }
			formValue, ok := tokenFilterSetting.Options["form"]
			if !ok {
				return nil, fmt.Errorf("form option does not exist")
			}
			form, ok := formValue.(string)
			if !ok {
				return nil, fmt.Errorf("form option is unexpected")
			}
			var tokenFilter analysis.TokenFilter
			switch form {
			case "NFC":
				tokenFilter = token.NewUnicodeNormalizeFilter(norm.NFC)
			case "NFD":
				tokenFilter = token.NewUnicodeNormalizeFilter(norm.NFD)
			case "NFKC":
				tokenFilter = token.NewUnicodeNormalizeFilter(norm.NFKC)
			case "NFKD":
				tokenFilter = token.NewUnicodeNormalizeFilter(norm.NFKD)
			default:
				err := fmt.Errorf("unknown form option")
				return nil, err
			}
			tokenFilters = append(tokenFilters, tokenFilter)
		case UniqueTermTokenFilter:
			// {
			// 	"name": "unique_term"
			// }
			tokenFilter := token.NewUniqueTermFilter()
			tokenFilters = append(tokenFilters, tokenFilter)
		default:
			err := fmt.Errorf("unknown token filter: %s\n", tokenFilterSetting.Name)
			return nil, err
		}
	}

	return &analysis.Analyzer{
		CharFilters:  charFilters,
		Tokenizer:    fieldTokenizer,
		TokenFilters: tokenFilters,
	}, nil
}

func (m IndexMapping) MakeDocument(id string, fieldMap map[string]interface{}) (*bluge.Document, error) {
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
