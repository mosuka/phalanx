package char

type CharFilter string

const (
	AsciiFoldingCharFilter       CharFilter = "ascii_folding"
	HtmlCharFilter               CharFilter = "html"
	RegexpCharFilter             CharFilter = "regexp"
	UnicodeNormalizeCharFilter   CharFilter = "unicode_normalize"
	ZeroWidthNonJoinerCharFilter CharFilter = "zero_width_non_joiner"
)

type CharFilterSetting struct {
	Name    CharFilter             `json:"name"`
	Options map[string]interface{} `json:"options"`
}
