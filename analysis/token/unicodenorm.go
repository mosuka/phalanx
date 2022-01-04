package token

import (
	"fmt"

	"github.com/blugelabs/bluge/analysis/token"
	"golang.org/x/text/unicode/norm"
)

// Create new UnicodeNormalizeFilter with given options.
// Options example:
// {
//   "form": "NFKC"
// }
func NewUnicodeNormalizeFilterWithOptions(opts map[string]interface{}) (*token.UnicodeNormalizeFilter, error) {
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
		return token.NewUnicodeNormalizeFilter(norm.NFC), nil
	case "NFD":
		return token.NewUnicodeNormalizeFilter(norm.NFD), nil
	case "NFKC":
		return token.NewUnicodeNormalizeFilter(norm.NFKC), nil
	case "NFKD":
		return token.NewUnicodeNormalizeFilter(norm.NFKD), nil
	default:
		return nil, fmt.Errorf("unknown form option: %v", form)
	}
}
