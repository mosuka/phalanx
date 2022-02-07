package wildcard

import (
	"regexp"
	"strings"
)

var wildcardRegexpReplacer = strings.NewReplacer(
	// characters in the wildcard that must
	// be escaped in the regexp
	"+", `\+`,
	"(", `\(`,
	")", `\)`,
	"^", `\^`,
	"$", `\$`,
	".", `\.`,
	"{", `\{`,
	"}", `\}`,
	"[", `\[`,
	"]", `\]`,
	`|`, `\|`,
	`\`, `\\`,
	// wildcard characters
	"*", ".*",
	"?", ".")

func Match(pattern, text string) bool {
	regexpString := wildcardRegexpReplacer.Replace(pattern)

	regexpObj := regexp.MustCompile(regexpString)

	return regexpObj.MatchString(text)
}
