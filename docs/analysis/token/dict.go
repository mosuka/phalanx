package token

import (
	"fmt"

	"github.com/blugelabs/bluge/analysis"
	"github.com/blugelabs/bluge/analysis/token"
)

// Create new DictionaryCompoundFilter with given options.
// Options example:
// {
//   "words": [
//     "soft",
//     "softest",
//     "ball"
//   ],
//   "min_word_size": 5,
//   "min_sub_word_size": 2,
//   "max_sub_word_size": 15,
//   "only_longest_match": false
// }
func NewDictionaryCompoundFilterWithOptions(opts map[string]interface{}) (*token.DictionaryCompoundFilter, error) {
	wordsValue, ok := opts["words"]
	if !ok {
		return nil, fmt.Errorf("words option does not exist")
	}
	words, ok := wordsValue.([]interface{})
	if !ok {
		return nil, fmt.Errorf("words option is unexpected")
	}
	wordMap := analysis.NewTokenMap()
	for _, word := range words {
		str, ok := word.(string)
		if !ok {
			return nil, fmt.Errorf("word is unexpected")
		}
		wordMap.AddToken(str)
	}

	minWordSizeValue, ok := opts["min_word_size"]
	if !ok {
		return nil, fmt.Errorf("min_word_size option does not exist")
	}
	minWordSizeNum, ok := minWordSizeValue.(float64)
	if !ok {
		return nil, fmt.Errorf("min_word_size option is unexpected")
	}
	minWordSize := int(minWordSizeNum)

	minSubWordSizeValue, ok := opts["min_sub_word_size"]
	if !ok {
		return nil, fmt.Errorf("min_sub_word_size option does not exist")
	}
	minSubWordSizeNum, ok := minSubWordSizeValue.(float64)
	if !ok {
		return nil, fmt.Errorf("min_sub_word_size option is unexpected")
	}
	minSubWordSize := int(minSubWordSizeNum)

	maxSubWordSizeValue, ok := opts["max_sub_word_size"]
	if !ok {
		return nil, fmt.Errorf("max_sub_word_size option does not exist")
	}
	maxSubWordSizeNum, ok := maxSubWordSizeValue.(float64)
	if !ok {
		return nil, fmt.Errorf("max_sub_word_size option is unexpected")
	}
	maxSubWordSize := int(maxSubWordSizeNum)

	onlyLongestMatchValue, ok := opts["only_longest_match"]
	if !ok {
		return nil, fmt.Errorf("only_longest_match option does not exist")
	}
	onlyLongestMatch, ok := onlyLongestMatchValue.(bool)
	if !ok {
		return nil, fmt.Errorf("only_longest_match option is unexpected")
	}

	return token.NewDictionaryCompoundFilter(wordMap, minWordSize, minSubWordSize, maxSubWordSize, onlyLongestMatch), nil
}
