package highlight

import (
	"encoding/json"
	"fmt"

	"github.com/blugelabs/bluge/search/highlight"
)

const (
	DefaultPreTag  = "<mark>"
	DefaultPostTag = "</mark>"
)

type HtmlHighlighterOptions struct {
	FragmentSize int    `json:"fragment_size"`
	PreTag       string `json:"pre_tag"`
	PostTag      string `json:"post_tag"`
	Separator    string `json:"separator"`
}

func NewHtmlHighlighterOptions() HtmlHighlighterOptions {
	return HtmlHighlighterOptions{
		FragmentSize: DefaultFragmentSize,
		PreTag:       DefaultPreTag,
		PostTag:      DefaultPostTag,
		Separator:    DefaultSeparator,
	}
}

// Create new TermsAggregation with given options.
// Options example:
// {
//   "fragment_size": 100,
//   "pre_tag": "<em>",
//   "post_tag": "</em>",
//   "separator": "…"
// }
func NewHtmlHighlighterWithMap(opts map[string]interface{}) (*highlight.SimpleHighlighter, error) {
	bytes, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	options := NewHtmlHighlighterOptions()
	if err := json.Unmarshal(bytes, &options); err != nil {
		return nil, err
	}

	return NewHtmlHighlighterWithOptions(options)
}

func NewHtmlHighlighterWithOptions(opts HtmlHighlighterOptions) (*highlight.SimpleHighlighter, error) {
	fmt.Println("opts", opts)

	var fragmenter *highlight.SimpleFragmenter
	if opts.FragmentSize > 0 {
		fragmenter = highlight.NewSimpleFragmenterSized(opts.FragmentSize)
	} else {
		fragmenter = highlight.NewSimpleFragmenter()
	}

	var formatter *highlight.HTMLFragmentFormatter
	if opts.PreTag == "" && opts.PostTag == "" {
		formatter = highlight.NewHTMLFragmentFormatter()
	} else if opts.PreTag != "" && opts.PostTag != "" {
		formatter = highlight.NewHTMLFragmentFormatterTags(opts.PreTag, opts.PostTag)
	} else {
		return nil, fmt.Errorf("pre_tag or post_tag option is not set")
	}

	var separator string
	if opts.Separator != "" {
		separator = opts.Separator
	} else {
		separator = DefaultSeparator
	}

	return highlight.NewSimpleHighlighter(fragmenter, formatter, separator), nil
}
