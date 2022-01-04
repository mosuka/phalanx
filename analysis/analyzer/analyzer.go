package analyzer

import (
	"fmt"

	"github.com/blugelabs/bluge/analysis"
	"github.com/blugelabs/bluge/analysis/char"
	"github.com/blugelabs/bluge/analysis/token"
	"github.com/blugelabs/bluge/analysis/tokenizer"
	phalanxchar "github.com/mosuka/phalanx/analysis/char"
	phalanxtoken "github.com/mosuka/phalanx/analysis/token"
	phalanxtokenizer "github.com/mosuka/phalanx/analysis/tokenizer"
)

type AnalyzerSetting struct {
	CharFilterSettings  []phalanxchar.CharFilterSetting   `json:"char_filters"`
	TokenizerSetting    phalanxtokenizer.TokenizerSetting `json:"tokenizer"`
	TokenFilterSettings []phalanxtoken.TokenFilterSetting `json:"token_filters"`
}

func NewAnalyzer(config AnalyzerSetting) (*analysis.Analyzer, error) {
	var err error

	// Char filter.
	charFilters := make([]analysis.CharFilter, 0)
	charFilterSettings := config.CharFilterSettings
	for _, charFilterSetting := range charFilterSettings {
		switch charFilterSetting.Name {
		case phalanxchar.AsciiFoldingCharFilter:
			charFilter := char.NewASCIIFoldingFilter()
			charFilters = append(charFilters, charFilter)
		case phalanxchar.HtmlCharFilter:
			charFilter := char.NewHTMLCharFilter()
			charFilters = append(charFilters, charFilter)
		case phalanxchar.RegexpCharFilter:
			charFilter, err := phalanxchar.NewRegexpCharFilterWithOptions(charFilterSetting.Options)
			if err != nil {
				return nil, err
			}
			charFilters = append(charFilters, charFilter)
		case phalanxchar.UnicodeNormalizeCharFilter:
			charFilter, err := phalanxchar.NewUnicodeNormalizeCharFilterWithOptions(charFilterSetting.Options)
			if err != nil {
				return nil, err
			}
			charFilters = append(charFilters, charFilter)
		case phalanxchar.ZeroWidthNonJoinerCharFilter:
			charFilter := char.NewZeroWidthNonJoinerCharFilter()
			charFilters = append(charFilters, charFilter)
		default:
			return nil, fmt.Errorf("unknown char filter: %s", charFilterSetting.Name)
		}
	}

	// Token filter.
	tokenFilters := make([]analysis.TokenFilter, 0)
	tokenFilterSettings := config.TokenFilterSettings
	for _, tokenFilterSetting := range tokenFilterSettings {
		switch tokenFilterSetting.Name {
		case phalanxtoken.ApostropheTokenFilter:
			tokenFilter := token.NewApostropheFilter()
			tokenFilters = append(tokenFilters, tokenFilter)
		case phalanxtoken.CamelCaseTokenFilter:
			tokenFilter := token.NewCamelCaseFilter()
			tokenFilters = append(tokenFilters, tokenFilter)
		case phalanxtoken.DictionaryCompoundTokenFilter:
			tokenFilter, err := phalanxtoken.NewDictionaryCompoundFilterWithOptions(tokenFilterSetting.Options)
			if err != nil {
				return nil, err
			}
			tokenFilters = append(tokenFilters, tokenFilter)
		case phalanxtoken.EdgeNgramTokenFilter:
			tokenFilter, err := phalanxtoken.NewEdgeNgramFilterWithOptions(tokenFilterSetting.Options)
			if err != nil {
				return nil, err
			}
			tokenFilters = append(tokenFilters, tokenFilter)
		case phalanxtoken.ElisionTokenFilter:
			tokenFilter, err := phalanxtoken.NewElisionFilterWithOptions(tokenFilterSetting.Options)
			if err != nil {
				return nil, err
			}
			tokenFilters = append(tokenFilters, tokenFilter)
		case phalanxtoken.KeywordMarkerTokenFilter:
			tokenFilter, err := phalanxtoken.NewKeyWordMarkerFilterWithOptions(tokenFilterSetting.Options)
			if err != nil {
				return nil, err
			}
			tokenFilters = append(tokenFilters, tokenFilter)
		case phalanxtoken.LengthTokenFilter:
			tokenFilter, err := phalanxtoken.NewLengthFilterWithOptions(tokenFilterSetting.Options)
			if err != nil {
				return nil, err
			}
			tokenFilters = append(tokenFilters, tokenFilter)
		case phalanxtoken.LowerCaseTokenFilter:
			tokenFilter := token.NewLowerCaseFilter()
			tokenFilters = append(tokenFilters, tokenFilter)
		case phalanxtoken.NgramTokenFilter:
			tokenFilter, err := phalanxtoken.NewNgramFilterWithOptions(tokenFilterSetting.Options)
			if err != nil {
				return nil, err
			}
			tokenFilters = append(tokenFilters, tokenFilter)
		case phalanxtoken.PorterStemmerTokenFilter:
			tokenFilter := token.NewPorterStemmer()
			tokenFilters = append(tokenFilters, tokenFilter)
		case phalanxtoken.ReverseTokenFilter:
			tokenFilter := token.NewReverseFilter()
			tokenFilters = append(tokenFilters, tokenFilter)
		case phalanxtoken.ShingleTokenFilter:
			tokenFilter, err := phalanxtoken.NewShingleFilterWithOptions(tokenFilterSetting.Options)
			if err != nil {
				return nil, err
			}
			tokenFilters = append(tokenFilters, tokenFilter)
		case phalanxtoken.StopTokensTokenFilter:
			tokenFilter, err := phalanxtoken.NewStopTokensFilterWithOptions(tokenFilterSetting.Options)
			if err != nil {
				return nil, err
			}
			tokenFilters = append(tokenFilters, tokenFilter)
		case phalanxtoken.TruncateTokenFilter:
			tokenFilter, err := phalanxtoken.NewTruncateTokenFilterWithOptions(tokenFilterSetting.Options)
			if err != nil {
				return nil, err
			}
			tokenFilters = append(tokenFilters, tokenFilter)
		case phalanxtoken.UnicodeNormalizeTokenFilter:
			tokenFilter, err := phalanxtoken.NewUnicodeNormalizeFilterWithOptions(tokenFilterSetting.Options)
			if err != nil {
				return nil, err
			}
			tokenFilters = append(tokenFilters, tokenFilter)
		case phalanxtoken.UniqueTermTokenFilter:
			tokenFilter := token.NewUniqueTermFilter()
			tokenFilters = append(tokenFilters, tokenFilter)
		default:
			err := fmt.Errorf("unknown token filter: %v", tokenFilterSetting.Name)
			return nil, err
		}
	}

	// Tokenizer.
	tokenizerSetting := config.TokenizerSetting
	var fieldTokenizer analysis.Tokenizer
	switch tokenizerSetting.Name {
	case phalanxtokenizer.CharacterTokenizer:
		fieldTokenizer, err = phalanxtokenizer.NewCharacterTokenizerWithOptions(tokenizerSetting.Options)
		if err != nil {
			return nil, err
		}
	case phalanxtokenizer.ExceptionTokenizer:
		var err error
		fieldTokenizer, err = phalanxtokenizer.NewExceptionsTokenizerWithOptions(tokenizerSetting.Options)
		if err != nil {
			return nil, err
		}
	case phalanxtokenizer.KagomeTokenizer:
		var err error
		fieldTokenizer, err = phalanxtokenizer.NewKagomeTokenizerWithOptions(tokenizerSetting.Options)
		if err != nil {
			return nil, err
		}
	case phalanxtokenizer.LetterTokenizer:
		fieldTokenizer = tokenizer.NewLetterTokenizer()
	case phalanxtokenizer.RegexpTokenizer:
		var err error
		fieldTokenizer, err = phalanxtokenizer.NewRegexpTokenizerWithOptions(tokenizerSetting.Options)
		if err != nil {
			return nil, err
		}
	case phalanxtokenizer.SingleTokenTokenizer:
		fieldTokenizer = tokenizer.NewSingleTokenTokenizer()
	case phalanxtokenizer.UnicodeTokenizer:
		fieldTokenizer = tokenizer.NewUnicodeTokenizer()
	case phalanxtokenizer.WebTokenizer:
		fieldTokenizer = tokenizer.NewWebTokenizer()
	case phalanxtokenizer.WhitespaceTokenizer:
		fieldTokenizer = tokenizer.NewWhitespaceTokenizer()
	default:
		return nil, fmt.Errorf("unknown tokenizer: %s", tokenizerSetting.Name)
	}

	return &analysis.Analyzer{
		CharFilters:  charFilters,
		Tokenizer:    fieldTokenizer,
		TokenFilters: tokenFilters,
	}, nil
}
