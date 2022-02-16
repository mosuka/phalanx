package highlight

import (
	"encoding/json"

	"github.com/blugelabs/bluge/search/highlight"
)

const (
	DefaultColor = "BgYellow"
)

type AnsiHighlighterOptions struct {
	FragmentSize int    `json:"fragment_size"`
	Color        string `json:"color"`
	Separator    string `json:"separator"`
}

func NewAnsiHighlighterOptions() AnsiHighlighterOptions {
	return AnsiHighlighterOptions{
		FragmentSize: DefaultFragmentSize,
		Color:        DefaultColor,
		Separator:    DefaultSeparator,
	}
}

// Create new TermsAggregation with given options.
// Options example:
// {
//   "fragment_size": 100,
//   "color": "FgCyan",
//   "separator": "…"
// }
func NewAnsiHighlighterWithMap(opts map[string]interface{}) (*highlight.SimpleHighlighter, error) {
	bytes, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	options := NewAnsiHighlighterOptions()
	if err := json.Unmarshal(bytes, &options); err != nil {
		return nil, err
	}

	return NewAnsiHighlighterWithOptions(options)
}

func NewAnsiHighlighterWithOptions(opts AnsiHighlighterOptions) (*highlight.SimpleHighlighter, error) {
	var fragmenter *highlight.SimpleFragmenter
	if opts.FragmentSize > 0 {
		fragmenter = highlight.NewSimpleFragmenterSized(opts.FragmentSize)
	} else {
		fragmenter = highlight.NewSimpleFragmenter()
	}

	var formatter *highlight.ANSIFragmentFormatter
	switch opts.Color {
	case "Reset":
		formatter = highlight.NewANSIFragmentFormatterColor(highlight.Reset)
	case "Bright":
		formatter = highlight.NewANSIFragmentFormatterColor(highlight.Bright)
	case "Dim":
		formatter = highlight.NewANSIFragmentFormatterColor(highlight.Dim)
	case "Underscore":
		formatter = highlight.NewANSIFragmentFormatterColor(highlight.Underscore)
	case "Blink":
		formatter = highlight.NewANSIFragmentFormatterColor(highlight.Blink)
	case "Reverse":
		formatter = highlight.NewANSIFragmentFormatterColor(highlight.Reverse)
	case "Hidden":
		formatter = highlight.NewANSIFragmentFormatterColor(highlight.Hidden)
	case "FgBlack":
		formatter = highlight.NewANSIFragmentFormatterColor(highlight.FgBlack)
	case "FgRed":
		formatter = highlight.NewANSIFragmentFormatterColor(highlight.FgRed)
	case "FgGreen":
		formatter = highlight.NewANSIFragmentFormatterColor(highlight.FgGreen)
	case "FgYellow":
		formatter = highlight.NewANSIFragmentFormatterColor(highlight.FgYellow)
	case "FgBlue":
		formatter = highlight.NewANSIFragmentFormatterColor(highlight.FgBlue)
	case "FgMagenta":
		formatter = highlight.NewANSIFragmentFormatterColor(highlight.FgMagenta)
	case "FgCyan":
		formatter = highlight.NewANSIFragmentFormatterColor(highlight.FgCyan)
	case "FgWhite":
		formatter = highlight.NewANSIFragmentFormatterColor(highlight.FgWhite)
	case "BgBlack":
		formatter = highlight.NewANSIFragmentFormatterColor(highlight.BgBlack)
	case "BgRed":
		formatter = highlight.NewANSIFragmentFormatterColor(highlight.BgRed)
	case "BgGreen":
		formatter = highlight.NewANSIFragmentFormatterColor(highlight.BgGreen)
	case "BgYellow":
		formatter = highlight.NewANSIFragmentFormatterColor(highlight.BgYellow)
	case "BgBlue":
		formatter = highlight.NewANSIFragmentFormatterColor(highlight.BgBlue)
	case "BgMagenta":
		formatter = highlight.NewANSIFragmentFormatterColor(highlight.BgMagenta)
	case "BgCyan":
		formatter = highlight.NewANSIFragmentFormatterColor(highlight.BgCyan)
	case "BgWhite":
		formatter = highlight.NewANSIFragmentFormatterColor(highlight.BgWhite)
	case "":
		formatter = highlight.NewANSIFragmentFormatter()
	default:
		return highlight.NewANSIHighlighterColor(opts.Color), nil
	}

	var separator string
	if opts.Separator != "" {
		separator = opts.Separator
	} else {
		separator = DefaultSeparator
	}

	return highlight.NewSimpleHighlighter(fragmenter, formatter, separator), nil
}
