package mapping

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/blugelabs/bluge/analysis"
)

func tokenStream(termStrs ...string) analysis.TokenStream {
	tokenStream := make([]*analysis.Token, len(termStrs))
	index := 0
	for i, termStr := range termStrs {
		tokenStream[i] = &analysis.Token{
			Term:         []byte(termStr),
			PositionIncr: 1,
			Start:        index,
			End:          index + len(termStr),
		}
		index += len(termStr)
	}
	return tokenStream
}

func TestNewMapping(t *testing.T) {
	indexMappingFile := "../testdata/test_mapping.json"

	bytes, err := ioutil.ReadFile(indexMappingFile)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	_, err = NewMapping(bytes)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
}

func TestGetFieldType(t *testing.T) {
	indexMappingFile := "../testdata/test_mapping.json"

	bytes, _ := ioutil.ReadFile(indexMappingFile)

	mapping, _ := NewMapping(bytes)

	fieldType, err := mapping.GetFieldType("numeric_field")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	if fieldType != NumericField {
		t.Fatalf("%v is not %v\n", fieldType, DatetimeField)
	}

	fieldType, err = mapping.GetFieldType("geo_point_field")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	if fieldType != GeoPointField {
		t.Fatalf("%v is not %v\n", fieldType, GeoPointField)
	}

	fieldType, err = mapping.GetFieldType("datetime_field")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	if fieldType != DatetimeField {
		t.Fatalf("%v is not %v\n", fieldType, DatetimeField)
	}

	fieldType, err = mapping.GetFieldType("text_field")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	if fieldType != TextField {
		t.Fatalf("%v is not %v\n", fieldType, TextField)
	}
}

func TestGetFieldOptions(t *testing.T) {
	indexMappingFile := "../testdata/test_mapping.json"

	bytes, _ := ioutil.ReadFile(indexMappingFile)

	mapping, _ := NewMapping(bytes)

	fieldOptions, err := mapping.GetFieldOptions("field_optrions_test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	if fieldOptions != 63 {
		t.Fatalf("%v is not %v\n", fieldOptions, 63)
	}
}

func TestAsciiFoldingCharFilter(t *testing.T) {
	indexMappingFile := "../testdata/test_mapping.json"

	b, _ := ioutil.ReadFile(indexMappingFile)

	mapping, _ := NewMapping(b)

	a, err := mapping.GetAnalyzer("ascii_folding_char_filter_test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	actual := a.CharFilters[0].Filter([]byte(`Ápple Àpple Äpple Âpple Ãpple Åpple`))
	expected := []byte(`Apple Apple Apple Apple Apple Apple`)
	if !bytes.Equal(actual, expected) {
		t.Fatalf("`%s` is not `%s`\n", string(actual), string(expected))
	}
}

func TestHtmlCharFilter(t *testing.T) {
	indexMappingFile := "../testdata/test_mapping.json"

	b, _ := ioutil.ReadFile(indexMappingFile)

	mapping, _ := NewMapping(b)

	a, err := mapping.GetAnalyzer("html_char_filter_test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	actual := a.CharFilters[0].Filter([]byte(`<html><head><title>title</title></head><body>Body</body></html>`))
	expected := []byte(`   title   Body  `)
	if !bytes.Equal(actual, expected) {
		t.Fatalf("`%s` is not `%s`\n", string(actual), string(expected))
	}
}

func TestRegexpCharFilter(t *testing.T) {
	indexMappingFile := "../testdata/test_mapping.json"

	b, _ := ioutil.ReadFile(indexMappingFile)

	mapping, _ := NewMapping(b)

	a, err := mapping.GetAnalyzer("regexp_char_filter_test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	actual := a.CharFilters[0].Filter([]byte(`I use Bleve.`))
	expected := []byte(`I use Bluge.`)
	if !bytes.Equal(actual, expected) {
		t.Fatalf("`%s` is not `%s`\n", string(actual), string(expected))
	}
}

func TestUnicodeNormalizeCharFilter(t *testing.T) {
	indexMappingFile := "../testdata/test_mapping.json"

	b, _ := ioutil.ReadFile(indexMappingFile)

	mapping, _ := NewMapping(b)

	a, err := mapping.GetAnalyzer("unicode_normalize_char_filter_test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	actual := a.CharFilters[0].Filter([]byte(`ﾊﾟﾅｿﾆｯｸ`))
	expected := []byte(`パナソニック`)
	if !bytes.Equal(actual, expected) {
		t.Fatalf("`%s` is not `%s`\n", string(actual), string(expected))
	}
}

// func TestZeroWidthNonJoinerCharFilter(t *testing.T) {
// 	indexMappingFile := "../testdata/test_mapping.json"

// 	b, _ := ioutil.ReadFile(indexMappingFile)

// 	mapping, _ := NewMapping(b)

// 	a, err := mapping.GetAnalyzer("zero_width_non_joiner_char_filter_test")
// 	if err != nil {
// 		t.Fatalf("%v\n", err)
// 	}
// 	s := "HERE>\u200C<HERE"
// 	actual := a.CharFilters[0].Filter([]byte(s))
// 	expected := []byte(`HERE><HERE`)
// 	if bytes.Compare(actual, expected) != 0 {
// 		t.Fatalf("`%s` is not `%s`\n", string(actual), string(expected))
// 	}
// }

func TestCharacterTokenizer(t *testing.T) {
	indexMappingFile := "../testdata/test_mapping.json"

	b, _ := ioutil.ReadFile(indexMappingFile)

	mapping, _ := NewMapping(b)

	a, err := mapping.GetAnalyzer("character_tokenizer_test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	actual := a.Tokenizer.Tokenize([]byte(`dominique@mcdiabetes.com`))
	expected := analysis.TokenStream{
		{
			Start:        0,
			End:          9,
			Term:         []byte("dominique"),
			PositionIncr: 1,
			Type:         analysis.AlphaNumeric,
		},
		{
			Start:        10,
			End:          20,
			Term:         []byte("mcdiabetes"),
			PositionIncr: 1,
			Type:         analysis.AlphaNumeric,
		},
		{
			Start:        21,
			End:          24,
			Term:         []byte("com"),
			PositionIncr: 1,
			Type:         analysis.AlphaNumeric,
		},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("`%v` is not `%v`\n", actual, expected)
	}
}

func TestExceptionTokenizer(t *testing.T) {
	indexMappingFile := "../testdata/test_mapping.json"

	b, _ := ioutil.ReadFile(indexMappingFile)

	mapping, _ := NewMapping(b)

	a, err := mapping.GetAnalyzer("exception_tokenizer_test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	actual := a.Tokenizer.Tokenize([]byte(`please email minoru.osuka@gmail.com the URL https://github.com/mosuka`))
	expected := analysis.TokenStream{
		{
			Term:         []byte("please"),
			PositionIncr: 1,
			Start:        0,
			End:          6,
		},
		{
			Term:         []byte("email"),
			PositionIncr: 1,
			Start:        7,
			End:          12,
		},
		{
			Term:         []byte("minoru.osuka@gmail.com"),
			PositionIncr: 1,
			Start:        13,
			End:          35,
		},
		{
			Term:         []byte("the"),
			PositionIncr: 1,
			Start:        36,
			End:          39,
		},
		{
			Term:         []byte("URL"),
			PositionIncr: 1,
			Start:        40,
			End:          43,
		},
		{
			Term:         []byte("https://github.com/mosuka"),
			PositionIncr: 1,
			Start:        44,
			End:          69,
		},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("`%v` is not `%v`\n", actual, expected)
	}
}

func TestJapaneseTokenizer(t *testing.T) {
	indexMappingFile := "../testdata/test_mapping.json"

	b, _ := ioutil.ReadFile(indexMappingFile)

	mapping, _ := NewMapping(b)

	a, err := mapping.GetAnalyzer("kagome_tokenizer_test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	actual := a.Tokenizer.Tokenize([]byte(`私は鰻。は。ねこはいます。`))
	expected := analysis.TokenStream{
		{
			Start:        0,
			End:          3,
			Term:         []byte("私"),
			PositionIncr: 1,
			Type:         analysis.Ideographic,
		},
		{
			Start:        6,
			End:          9,
			Term:         []byte("鰻"),
			PositionIncr: 2,
			Type:         analysis.Ideographic,
		},
		{
			Start:        18,
			End:          24,
			Term:         []byte("ねこ"),
			PositionIncr: 4,
			Type:         analysis.Ideographic,
		},
		{
			Start:        27,
			End:          30,
			Term:         []byte("いる"),
			PositionIncr: 2,
			Type:         analysis.Ideographic,
		},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("`%v` is not `%v`\n", actual, expected)
	}
}

func TestLetterTokenizer(t *testing.T) {
	indexMappingFile := "../testdata/test_mapping.json"

	b, _ := ioutil.ReadFile(indexMappingFile)

	mapping, _ := NewMapping(b)

	a, err := mapping.GetAnalyzer("letter_tokenizer_test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	actual := a.Tokenizer.Tokenize([]byte(`http://localhost`))
	expected := analysis.TokenStream{
		{
			Start:        0,
			End:          4,
			Term:         []byte("http"),
			PositionIncr: 1,
		},
		{
			Start:        7,
			End:          16,
			Term:         []byte("localhost"),
			PositionIncr: 1,
		},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("`%v` is not `%v`\n", actual, expected)
	}
}

func TestRegexpTokenizer(t *testing.T) {
	indexMappingFile := "../testdata/test_mapping.json"

	b, _ := ioutil.ReadFile(indexMappingFile)

	mapping, _ := NewMapping(b)

	a, err := mapping.GetAnalyzer("regexp_tokenizer_test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	actual := a.Tokenizer.Tokenize([]byte(`Chatha Edwards Sr.`))
	expected := analysis.TokenStream{
		{
			Start:        0,
			End:          6,
			Term:         []byte("Chatha"),
			PositionIncr: 1,
			Type:         analysis.AlphaNumeric,
		},
		{
			Start:        7,
			End:          14,
			Term:         []byte("Edwards"),
			PositionIncr: 1,
			Type:         analysis.AlphaNumeric,
		},
		{
			Start:        15,
			End:          17,
			Term:         []byte("Sr"),
			PositionIncr: 1,
			Type:         analysis.AlphaNumeric,
		},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("`%v` is not `%v`\n", actual, expected)
	}
}

func TestSingleTokenTokenizer(t *testing.T) {
	indexMappingFile := "../testdata/test_mapping.json"

	b, _ := ioutil.ReadFile(indexMappingFile)

	mapping, _ := NewMapping(b)

	a, err := mapping.GetAnalyzer("single_token_tokenizer_test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	actual := a.Tokenizer.Tokenize([]byte(`Hello World`))
	expected := analysis.TokenStream{
		{
			Start:        0,
			End:          11,
			Term:         []byte("Hello World"),
			PositionIncr: 1,
			Type:         analysis.AlphaNumeric,
		},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("`%v` is not `%v`\n", actual, expected)
	}
}

func TestUnicodeTokenizer(t *testing.T) {
	indexMappingFile := "../testdata/test_mapping.json"

	b, _ := ioutil.ReadFile(indexMappingFile)

	mapping, _ := NewMapping(b)

	a, err := mapping.GetAnalyzer("unicode_tokenizer_test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	actual := a.Tokenizer.Tokenize([]byte(`Hello World`))
	expected := analysis.TokenStream{
		{
			Start:        0,
			End:          5,
			Term:         []byte("Hello"),
			PositionIncr: 1,
			Type:         analysis.AlphaNumeric,
		},
		{
			Start:        6,
			End:          11,
			Term:         []byte("World"),
			PositionIncr: 1,
			Type:         analysis.AlphaNumeric,
		},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("`%v` is not `%v`\n", actual, expected)
	}
}

func TestWebTokenizer(t *testing.T) {
	indexMappingFile := "../testdata/test_mapping.json"

	b, _ := ioutil.ReadFile(indexMappingFile)

	mapping, _ := NewMapping(b)

	a, err := mapping.GetAnalyzer("web_tokenizer_test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	actual := a.Tokenizer.Tokenize([]byte(`Hello minoru.osuka@gmail.com`))
	expected := analysis.TokenStream{
		{
			Start:        0,
			End:          5,
			Term:         []byte("Hello"),
			PositionIncr: 1,
			Type:         analysis.AlphaNumeric,
		},
		{
			Start:        6,
			End:          28,
			Term:         []byte("minoru.osuka@gmail.com"),
			PositionIncr: 1,
			Type:         analysis.AlphaNumeric,
		},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("`%v` is not `%v`\n", actual, expected)
	}
}

func TestWhitespaceTokenizer(t *testing.T) {
	indexMappingFile := "../testdata/test_mapping.json"

	b, _ := ioutil.ReadFile(indexMappingFile)

	mapping, _ := NewMapping(b)

	a, err := mapping.GetAnalyzer("whitespace_tokenizer_test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	actual := a.Tokenizer.Tokenize([]byte(`Hello World.`))
	expected := analysis.TokenStream{
		{
			Start:        0,
			End:          5,
			Term:         []byte("Hello"),
			PositionIncr: 1,
			Type:         analysis.AlphaNumeric,
		},
		{
			Start:        6,
			End:          12,
			Term:         []byte("World."),
			PositionIncr: 1,
			Type:         analysis.AlphaNumeric,
		},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("`%v` is not `%v`\n", actual, expected)
	}
}

func TestApostropheTokenFilter(t *testing.T) {
	indexMappingFile := "../testdata/test_mapping.json"

	b, _ := ioutil.ReadFile(indexMappingFile)

	mapping, _ := NewMapping(b)

	a, err := mapping.GetAnalyzer("apostrophe_token_filter_test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	input := analysis.TokenStream{
		&analysis.Token{
			Term:         []byte("Türkiye'de"),
			PositionIncr: 1,
		},
	}

	actual := a.TokenFilters[0].Filter(input)

	expected := analysis.TokenStream{
		&analysis.Token{
			Term:         []byte("Türkiye"),
			PositionIncr: 1,
		},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("`%v` is not `%v`\n", actual, expected)
	}
}

func TestCamelCaseTokenFilter(t *testing.T) {
	indexMappingFile := "../testdata/test_mapping.json"

	b, _ := ioutil.ReadFile(indexMappingFile)

	mapping, _ := NewMapping(b)

	a, err := mapping.GetAnalyzer("camel_case_token_filter_test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	input := tokenStream("...aMACMac123macILoveGolang")

	actual := a.TokenFilters[0].Filter(input)

	expected := tokenStream("...", "a", "MAC", "Mac", "123", "mac", "I", "Love", "Golang")

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("`%v` is not `%v`\n", actual, expected)
	}
}

func TestDictionaryCompoundTokenFilter(t *testing.T) {
	indexMappingFile := "../testdata/test_mapping.json"

	b, _ := ioutil.ReadFile(indexMappingFile)

	mapping, _ := NewMapping(b)

	a, err := mapping.GetAnalyzer("dictionary_compound_token_filter_test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	input := analysis.TokenStream{
		&analysis.Token{
			Term:         []byte("i"),
			Start:        0,
			End:          1,
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("like"),
			Start:        2,
			End:          6,
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("to"),
			Start:        7,
			End:          9,
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("play"),
			Start:        10,
			End:          14,
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("softball"),
			Start:        15,
			End:          23,
			PositionIncr: 1,
		},
	}

	actual := a.TokenFilters[0].Filter(input)

	expected := analysis.TokenStream{
		&analysis.Token{
			Term:         []byte("i"),
			Start:        0,
			End:          1,
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("like"),
			Start:        2,
			End:          6,
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("to"),
			Start:        7,
			End:          9,
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("play"),
			Start:        10,
			End:          14,
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("softball"),
			Start:        15,
			End:          23,
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("soft"),
			Start:        15,
			End:          19,
			PositionIncr: 0,
		},
		&analysis.Token{
			Term:         []byte("ball"),
			Start:        19,
			End:          23,
			PositionIncr: 0,
		},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("`%v` is not `%v`\n", actual, expected)
	}
}

func TestEdgeNgramTokenFilter(t *testing.T) {
	indexMappingFile := "../testdata/test_mapping.json"

	b, _ := ioutil.ReadFile(indexMappingFile)

	mapping, _ := NewMapping(b)

	a, err := mapping.GetAnalyzer("edge_ngram_token_filter_test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	input := analysis.TokenStream{
		&analysis.Token{
			Term:         []byte("abcde"),
			PositionIncr: 1,
		},
	}

	actual := a.TokenFilters[0].Filter(input)

	expected := analysis.TokenStream{
		&analysis.Token{
			Term:         []byte("a"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("ab"),
			PositionIncr: 0,
		},
		&analysis.Token{
			Term:         []byte("abc"),
			PositionIncr: 0,
		},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("`%v` is not `%v`\n", actual, expected)
	}
}

func TestElisionTokenFilter(t *testing.T) {
	indexMappingFile := "../testdata/test_mapping.json"

	b, _ := ioutil.ReadFile(indexMappingFile)

	mapping, _ := NewMapping(b)

	a, err := mapping.GetAnalyzer("elision_token_filter_test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	input := analysis.TokenStream{
		&analysis.Token{
			Term:         []byte("ar" + string('\'') + "word"),
			PositionIncr: 1,
		},
	}

	actual := a.TokenFilters[0].Filter(input)

	expected := analysis.TokenStream{
		&analysis.Token{
			Term:         []byte("word"),
			PositionIncr: 1,
		},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("`%v` is not `%v`\n", actual, expected)
	}
}

func TestKeywordMarkerTokenFilter(t *testing.T) {
	indexMappingFile := "../testdata/test_mapping.json"

	b, _ := ioutil.ReadFile(indexMappingFile)

	mapping, _ := NewMapping(b)

	a, err := mapping.GetAnalyzer("keyword_marker_token_filter_test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	input := analysis.TokenStream{
		&analysis.Token{
			Term:         []byte("a"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("walk"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("in"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("the"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("park"),
			PositionIncr: 1,
		},
	}

	actual := a.TokenFilters[0].Filter(input)

	expected := analysis.TokenStream{
		&analysis.Token{
			Term:         []byte("a"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("walk"),
			KeyWord:      true,
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("in"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("the"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("park"),
			KeyWord:      true,
			PositionIncr: 1,
		},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("`%v` is not `%v`\n", actual, expected)
	}
}

func TestLengthTokenFilter(t *testing.T) {
	indexMappingFile := "../testdata/test_mapping.json"

	b, _ := ioutil.ReadFile(indexMappingFile)

	mapping, _ := NewMapping(b)

	a, err := mapping.GetAnalyzer("length_token_filter_test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	input := analysis.TokenStream{
		&analysis.Token{
			Term:         []byte("1"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("two"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("three"),
			PositionIncr: 1,
		},
	}

	actual := a.TokenFilters[0].Filter(input)

	expected := analysis.TokenStream{
		&analysis.Token{
			Term:         []byte("two"),
			PositionIncr: 2,
		},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("`%v` is not `%v`\n", actual, expected)
	}
}

func TestLowerCaseTokenFilter(t *testing.T) {
	indexMappingFile := "../testdata/test_mapping.json"

	b, _ := ioutil.ReadFile(indexMappingFile)

	mapping, _ := NewMapping(b)

	a, err := mapping.GetAnalyzer("lower_case_token_filter_test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	input := analysis.TokenStream{
		&analysis.Token{
			Term:         []byte("ONE"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("two"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("ThReE"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("steven's"),
			PositionIncr: 1,
		},
		// these characters are chosen in particular
		// because the utf-8 encoding of the lower-case
		// version has a different length
		// Rune İ(304) width 2 - Lower i(105) width 1
		// Rune Ⱥ(570) width 2 - Lower ⱥ(11365) width 3
		// Rune Ⱦ(574) width 2 - Lower ⱦ(11366) width 3
		&analysis.Token{
			Term:         []byte("İȺȾCAT"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("ȺȾCAT"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("ὈΔΥΣΣ"),
			PositionIncr: 1,
		},
	}

	actual := a.TokenFilters[0].Filter(input)

	expected := analysis.TokenStream{
		&analysis.Token{
			Term:         []byte("one"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("two"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("three"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("steven's"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("iⱥⱦcat"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("ⱥⱦcat"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("ὀδυσς"),
			PositionIncr: 1,
		},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("`%v` is not `%v`\n", actual, expected)
	}
}

func TestNgramTokenFilter(t *testing.T) {
	indexMappingFile := "../testdata/test_mapping.json"

	b, _ := ioutil.ReadFile(indexMappingFile)

	mapping, _ := NewMapping(b)

	a, err := mapping.GetAnalyzer("ngram_token_filter_test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	input := analysis.TokenStream{
		&analysis.Token{
			Term:         []byte("abcde"),
			PositionIncr: 1,
		},
	}

	actual := a.TokenFilters[0].Filter(input)

	expected := analysis.TokenStream{
		&analysis.Token{
			Term:         []byte("a"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("ab"),
			PositionIncr: 0,
		},
		&analysis.Token{
			Term:         []byte("abc"),
			PositionIncr: 0,
		},
		&analysis.Token{
			Term:         []byte("b"),
			PositionIncr: 0,
		},
		&analysis.Token{
			Term:         []byte("bc"),
			PositionIncr: 0,
		},
		&analysis.Token{
			Term:         []byte("bcd"),
			PositionIncr: 0,
		},
		&analysis.Token{
			Term:         []byte("c"),
			PositionIncr: 0,
		},
		&analysis.Token{
			Term:         []byte("cd"),
			PositionIncr: 0,
		},
		&analysis.Token{
			Term:         []byte("cde"),
			PositionIncr: 0,
		},
		&analysis.Token{
			Term:         []byte("d"),
			PositionIncr: 0,
		},
		&analysis.Token{
			Term:         []byte("de"),
			PositionIncr: 0,
		},
		&analysis.Token{
			Term:         []byte("e"),
			PositionIncr: 0,
		},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("`%v` is not `%v`\n", actual, expected)
	}
}

func TestPorterStemmerTokenFilter(t *testing.T) {
	indexMappingFile := "../testdata/test_mapping.json"

	b, _ := ioutil.ReadFile(indexMappingFile)

	mapping, _ := NewMapping(b)

	a, err := mapping.GetAnalyzer("porter_stemmer_token_filter_test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	input := analysis.TokenStream{
		&analysis.Token{
			Term:         []byte("walking"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("talked"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("business"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("protected"),
			KeyWord:      true,
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("cat"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("done"),
			PositionIncr: 1,
		},
		// a term which does stem, but does not change length
		&analysis.Token{
			Term:         []byte("marty"),
			PositionIncr: 1,
		},
	}

	actual := a.TokenFilters[0].Filter(input)

	expected := analysis.TokenStream{
		&analysis.Token{
			Term:         []byte("walk"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("talk"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("busi"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("protected"),
			KeyWord:      true,
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("cat"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("done"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("marti"),
			PositionIncr: 1,
		},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("`%v` is not `%v`\n", actual, expected)
	}
}

func TestReverseTokenFilter(t *testing.T) {
	indexMappingFile := "../testdata/test_mapping.json"

	b, _ := ioutil.ReadFile(indexMappingFile)

	mapping, _ := NewMapping(b)

	a, err := mapping.GetAnalyzer("reverse_token_filter_test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	input := analysis.TokenStream{
		&analysis.Token{},
		&analysis.Token{
			Term:         []byte("one"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("TWo"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("thRee"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("four's"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("what's this in reverse"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("œ∑´®†"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("İȺȾCAT÷≥≤µ123"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("!@#$%^&*()"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("cafés"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("¿Dónde estás?"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("Me gustaría una cerveza."),
			PositionIncr: 1,
		},
	}

	actual := a.TokenFilters[0].Filter(input)

	expected := analysis.TokenStream{
		&analysis.Token{},
		&analysis.Token{
			Term:         []byte("eno"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("oWT"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("eeRht"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("s'ruof"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("esrever ni siht s'tahw"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("†®´∑œ"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("321µ≤≥÷TACȾȺİ"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte(")(*&^%$#@!"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("séfac"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("?sátse ednóD¿"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte(".azevrec anu aíratsug eM"),
			PositionIncr: 1,
		},
	}

	for i := 0; i < len(expected); i++ {
		if !bytes.Equal(actual[i].Term, expected[i].Term) {
			t.Errorf("[%d] expected %s got %s",
				i+1, expected[i].Term, actual[i].Term)
		}
	}
}

func TestShingleTokenFilter(t *testing.T) {
	indexMappingFile := "../testdata/test_mapping.json"

	b, _ := ioutil.ReadFile(indexMappingFile)

	mapping, _ := NewMapping(b)

	a, err := mapping.GetAnalyzer("shingle_token_filter_test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	input := analysis.TokenStream{
		&analysis.Token{
			Term:         []byte("the"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("quick"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("brown"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("fox"),
			PositionIncr: 1,
		},
	}

	actual := a.TokenFilters[0].Filter(input)

	expected := analysis.TokenStream{
		&analysis.Token{
			Term:         []byte("the"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("quick"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("the quick"),
			Type:         analysis.Shingle,
			PositionIncr: 0,
		},
		&analysis.Token{
			Term:         []byte("brown"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("quick brown"),
			Type:         analysis.Shingle,
			PositionIncr: 0,
		},
		&analysis.Token{
			Term:         []byte("fox"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("brown fox"),
			Type:         analysis.Shingle,
			PositionIncr: 0,
		},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("`%v` is not `%v`\n", actual, expected)
	}
}

func TestStopTokensTokenFilter(t *testing.T) {
	indexMappingFile := "../testdata/test_mapping.json"

	b, _ := ioutil.ReadFile(indexMappingFile)

	mapping, _ := NewMapping(b)

	a, err := mapping.GetAnalyzer("stop_tokens_token_filter_test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	input := analysis.TokenStream{
		&analysis.Token{
			Term:         []byte("a"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("walk"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("in"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("the"),
			PositionIncr: 1,
		},
		&analysis.Token{
			Term:         []byte("park"),
			PositionIncr: 1,
		},
	}

	actual := a.TokenFilters[0].Filter(input)

	expected := analysis.TokenStream{
		&analysis.Token{
			Term:         []byte("walk"),
			PositionIncr: 2,
		},
		&analysis.Token{
			Term:         []byte("park"),
			PositionIncr: 3,
		},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("`%v` is not `%v`\n", actual, expected)
	}
}

func TestTruncateTokenFilter(t *testing.T) {
	indexMappingFile := "../testdata/test_mapping.json"

	b, _ := ioutil.ReadFile(indexMappingFile)

	mapping, _ := NewMapping(b)

	a, err := mapping.GetAnalyzer("truncate_token_filter_test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	input := analysis.TokenStream{
		&analysis.Token{
			Term:         []byte("abcdefgh"),
			PositionIncr: 1,
		},
	}

	actual := a.TokenFilters[0].Filter(input)

	expected := analysis.TokenStream{
		&analysis.Token{
			Term:         []byte("abcde"),
			PositionIncr: 1,
		},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("`%v` is not `%v`\n", actual, expected)
	}
}

func TestUnicodeNormalizeTokenFilter(t *testing.T) {
	indexMappingFile := "../testdata/test_mapping.json"

	b, _ := ioutil.ReadFile(indexMappingFile)

	mapping, _ := NewMapping(b)

	a, err := mapping.GetAnalyzer("unicode_normalize_token_filter_test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	input := analysis.TokenStream{
		&analysis.Token{
			Term:         []byte("ｳﾞｨｯﾂ"),
			PositionIncr: 1,
		},
	}

	actual := a.TokenFilters[0].Filter(input)

	expected := analysis.TokenStream{
		&analysis.Token{
			Term:         []byte("ヴィッツ"),
			PositionIncr: 1,
		},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("`%v` is not `%v`\n", actual, expected)
	}
}

func TestUniqueTermTokenFilter(t *testing.T) {
	indexMappingFile := "../testdata/test_mapping.json"

	b, _ := ioutil.ReadFile(indexMappingFile)

	mapping, _ := NewMapping(b)

	a, err := mapping.GetAnalyzer("unique_term_token_filter_test")
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	input := tokenStream("a", "a", "A", "a", "a", "A")

	actual := a.TokenFilters[0].Filter(input)

	expected := analysis.TokenStream{
		&analysis.Token{
			Term:         []byte("a"),
			PositionIncr: 1,
			Start:        0,
			End:          1,
		},
		&analysis.Token{
			Term:         []byte("A"),
			PositionIncr: 2,
			Start:        2,
			End:          3,
		},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("`%v` is not `%v`\n", actual, expected)
	}
}
