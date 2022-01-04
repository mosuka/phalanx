package char

import (
	"fmt"

	"github.com/blugelabs/bluge/analysis"
	"github.com/ikawaha/blugeplugin/analysis/lang/ja"
	"golang.org/x/text/unicode/norm"
)

// Create new UnicodeNormalizeCharFilter with given options.
// Options example:
// {
//   "form": "NFKC"
// }
func NewUnicodeNormalizeCharFilterWithOptions(opts map[string]interface{}) (analysis.CharFilter, error) {
	formValue, ok := opts["form"]
	if !ok {
		return nil, fmt.Errorf("form option does not exist")
	}
	form, ok := formValue.(string)
	if !ok {
		return nil, fmt.Errorf("form option is unexpected: %v", formValue)
	}
	switch form {
	case "NFC":
		return ja.NewUnicodeNormalizeCharFilter(norm.NFC), nil
	case "NFD":
		return ja.NewUnicodeNormalizeCharFilter(norm.NFD), nil
	case "NFKC":
		return ja.NewUnicodeNormalizeCharFilter(norm.NFKC), nil
	case "NFKD":
		return ja.NewUnicodeNormalizeCharFilter(norm.NFKD), nil
	default:
		return nil, fmt.Errorf("unknown form option: %v", form)
	}
}
