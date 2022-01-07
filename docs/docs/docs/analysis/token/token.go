package token

type TokenFilter string

const (
	ApostropheTokenFilter         TokenFilter = "apostrophe"
	CamelCaseTokenFilter          TokenFilter = "camel_case"
	DictionaryCompoundTokenFilter TokenFilter = "dictionary_compound"
	EdgeNgramTokenFilter          TokenFilter = "edge_ngram"
	ElisionTokenFilter            TokenFilter = "elision"
	KeywordMarkerTokenFilter      TokenFilter = "keyword_marker"
	LengthTokenFilter             TokenFilter = "length"
	LowerCaseTokenFilter          TokenFilter = "lower_case"
	NgramTokenFilter              TokenFilter = "ngram"
	PorterStemmerTokenFilter      TokenFilter = "porter_stemmer"
	ReverseTokenFilter            TokenFilter = "reverse"
	ShingleTokenFilter            TokenFilter = "shingle"
	StopTokensTokenFilter         TokenFilter = "stop_tokens"
	TruncateTokenFilter           TokenFilter = "truncate"
	UnicodeNormalizeTokenFilter   TokenFilter = "unicode_normalize"
	UniqueTermTokenFilter         TokenFilter = "unique_term"
)

type TokenFilterSetting struct {
	Name    TokenFilter            `json:"name"`
	Options map[string]interface{} `json:"options"`
}
