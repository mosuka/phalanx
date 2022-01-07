package hy

import (
	"github.com/blugelabs/bluge/analysis"
	"github.com/blugelabs/bluge/analysis/lang/hy"
	"github.com/blugelabs/bluge/analysis/token"
	"github.com/blugelabs/bluge/analysis/tokenizer"
	"golang.org/x/text/unicode/norm"
)

func Analyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Tokenizer: tokenizer.NewUnicodeTokenizer(),
		TokenFilters: []analysis.TokenFilter{
			token.NewLowerCaseFilter(),
			token.NewUnicodeNormalizeFilter(norm.NFKC),
			hy.StopWordsFilter(),
		},
	}
}
