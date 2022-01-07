package tokenizer

import (
	"fmt"

	"github.com/blugelabs/bluge/analysis"
	"github.com/ikawaha/blugeplugin/analysis/lang/ja"
	"github.com/ikawaha/kagome-dict/dict"
	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome-dict/uni"
)

// Create new KagomeTokenizer with given options.
// Options example:
// {
//   "dictionary": "IPADIC",
//   "stop_tags": [
//     "接続詞",
//     "助詞",
//     "助詞-格助詞",
//     "助詞-格助詞-一般",
//     "助詞-格助詞-引用",
//     "助詞-格助詞-連語",
//     "助詞-接続助詞",
//     "助詞-係助詞",
//     "助詞-副助詞",
//     "助詞-間投助詞",
//     "助詞-並立助詞",
//     "助詞-終助詞",
//     "助詞-副助詞／並立助詞／終助詞",
//     "助詞-連体化",
//     "助詞-副詞化",
//     "助詞-特殊",
//     "助動詞",
//     "記号",
//     "記号-一般",
//     "記号-読点",
//     "記号-句点",
//     "記号-空白",
//     "記号-括弧開",
//     "記号-括弧閉",
//     "その他-間投",
//     "フィラー",
//     "非言語音"
//   ],
//   "base_forms": [
//     "動詞",
//     "形容詞",
//     "形容動詞"
//   ]
// }
func NewKagomeTokenizerWithOptions(opts map[string]interface{}) (analysis.Tokenizer, error) {
	dictionaryValue, ok := opts["dictionary"]
	if !ok {
		return nil, fmt.Errorf("dict option does not exist")
	}
	dictionaryStr, ok := dictionaryValue.(string)
	if !ok {
		return nil, fmt.Errorf("dict option is unexpected")
	}
	var dictionary *dict.Dict
	switch dictionaryStr {
	case "IPADIC":
		dictionary = ipa.Dict()
	case "UniDIC":
		dictionary = uni.Dict()
	}

	stopTagsValue, ok := opts["stop_tags"]
	if !ok {
		return nil, fmt.Errorf("stop_tags option does not exist")
	}
	stopTags, ok := stopTagsValue.([]interface{})
	if !ok {
		return nil, fmt.Errorf("stop_tags option is unexpected")
	}
	stopTagsTokenMap := analysis.NewTokenMap()
	for _, stopTag := range stopTags {
		token, ok := stopTag.(string)
		if !ok {
			return nil, fmt.Errorf("stop_tag is unexpected: %v", stopTag)
		}
		stopTagsTokenMap.AddToken(token)
	}

	baseFormsValue, ok := opts["base_forms"]
	if !ok {
		return nil, fmt.Errorf("stop_tags option does not exist")
	}
	baseForms, ok := baseFormsValue.([]interface{})
	if !ok {
		return nil, fmt.Errorf("stop_tags option is unexpected")
	}
	baseFormsTokenMap := analysis.NewTokenMap()
	for _, baseForm := range baseForms {
		token, ok := baseForm.(string)
		if !ok {
			return nil, fmt.Errorf("stop_tag is unexpected: %v", baseForm)
		}
		baseFormsTokenMap.AddToken(token)
	}

	return ja.NewJapaneseTokenizer(dictionary, ja.StopTagsFilter(stopTagsTokenMap), ja.BaseFormFilter(baseFormsTokenMap)), nil
}
