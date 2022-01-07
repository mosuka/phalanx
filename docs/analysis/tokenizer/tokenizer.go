package tokenizer

type Tokenizer string

const (
	CharacterTokenizer   Tokenizer = "character"
	ExceptionTokenizer   Tokenizer = "exception"
	KagomeTokenizer      Tokenizer = "kagome"
	LetterTokenizer      Tokenizer = "letter"
	RegexpTokenizer      Tokenizer = "regexp"
	SingleTokenTokenizer Tokenizer = "single_token"
	UnicodeTokenizer     Tokenizer = "unicode"
	WebTokenizer         Tokenizer = "web"
	WhitespaceTokenizer  Tokenizer = "whitespace"
)

type TokenizerSetting struct {
	Name    Tokenizer              `json:"name"`
	Options map[string]interface{} `json:"options"`
}
