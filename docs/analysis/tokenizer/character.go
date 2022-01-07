package tokenizer

import (
	"fmt"
	"unicode"

	"github.com/blugelabs/bluge/analysis/tokenizer"
)

// Create new CharacterTokenizer with given options.
// Options example:
// {
//   "rune": "graphic"
// }
func NewCharacterTokenizerWithOptions(opts map[string]interface{}) (*tokenizer.CharacterTokenizer, error) {
	runeValue, ok := opts["rune"]
	if !ok {
		return nil, fmt.Errorf("ruen option does not exist")
	}
	runeStr, ok := runeValue.(string)
	if !ok {
		return nil, fmt.Errorf("rune option is unexpected")
	}

	switch runeStr {
	case "graphic":
		return tokenizer.NewCharacterTokenizer(unicode.IsGraphic), nil
	case "print":
		return tokenizer.NewCharacterTokenizer(unicode.IsPrint), nil
	case "control":
		return tokenizer.NewCharacterTokenizer(unicode.IsControl), nil
	case "letter":
		return tokenizer.NewCharacterTokenizer(unicode.IsLetter), nil
	case "mark":
		return tokenizer.NewCharacterTokenizer(unicode.IsMark), nil
	case "number":
		return tokenizer.NewCharacterTokenizer(unicode.IsNumber), nil
	case "punct":
		return tokenizer.NewCharacterTokenizer(unicode.IsPunct), nil
	case "space":
		return tokenizer.NewCharacterTokenizer(unicode.IsSpace), nil
	case "symbol":
		return tokenizer.NewCharacterTokenizer(unicode.IsSymbol), nil
	default:
		return nil, fmt.Errorf("unknown rune option: %s", runeStr)
	}
}
