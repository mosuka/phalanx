package highlight

import (
	"github.com/blugelabs/bluge/search/highlight"
	"github.com/mosuka/phalanx/errors"
)

const (
	DefaultFragmentSize = 200
	DefaultSeparator    = highlight.DefaultSeparator
)

type HighlighterType int

const (
	HighlighterTypeUnknown HighlighterType = iota
	HighlighterTypeAnsi
	HighlighterTypeHtml
)

// Maps for HighliterType.
var (
	HighlighterType_name = map[HighlighterType]string{
		HighlighterTypeUnknown: "unknown",
		HighlighterTypeAnsi:    "ansi",
		HighlighterTypeHtml:    "html",
	}
	HighlighterType_value = map[string]HighlighterType{
		"unknown": HighlighterTypeUnknown,
		"ansi":    HighlighterTypeAnsi,
		"html":    HighlighterTypeHtml,
	}
)

func NewHighlighter(highlighterType string, highlighterOpts map[string]interface{}) (*highlight.SimpleHighlighter, error) {
	switch HighlighterType_value[highlighterType] {
	case HighlighterTypeAnsi:
		return NewAnsiHighlighterWithMap(highlighterOpts)
	case HighlighterTypeHtml:
		return NewHtmlHighlighterWithMap(highlighterOpts)
	default:
		return nil, errors.ErrUnknownHighlighterType
	}
}

type HighlightRequest struct {
	Highlighter highlight.Highlighter
	Num         int
}
