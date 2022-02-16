# Tokenizers

A tokenizer receives a stream of characters, breaks it up into individual tokens (usually individual words), and outputs a stream of tokens.  

The format of the char filter definition is as follows:
```
{
    "name": <TOKENIZER_NAME>,
    "options": <TOKENIZER_OPTIONS>
}
```
- `<TOKENIZER_NAME>`: 
- `<TOKENIZER_OPTIONS>`: 

The following char filters are available:
- Character
- Exception
- Kagome
- Letter
- Regular Expression
- Single Token
- Unicode
- Web
- Whitespace


## Character

Outputs tokens with the specified rune. The following parameters can be set for `rune`.  
- `graphic`: Such characters include letters, marks, numbers, punctuation, symbols, and spaces.
- `print`: Such characters include letters, marks, numbers, punctuation, symbols, and the ASCII space character.
- `control`: Control characters.
- `letter`: Letter characters.
- `mark`: Mark characters.
- `number`: Number characters.
- `punct`: Unicode punctuation characters.
- `space`: space character as defined by Unicode's White Space property; in the Latin-1 space this is '\t', '\n', '\v', '\f', '\r', ' ', U+0085 (NEL), U+00A0 (NBSP).
- `symbol`: Symbolic characters

Example:
```json
{
    "name": "unicode_normalize",
    "options": {
        "rune": "letter"
    }    
}
```

## Exception

Split strings that match multiple regular expression patterns into tokens with UnicodeTokenizer.  

Example:
```json
{
    "name": "exception",
    "options": {
        "patterns": [
            "[hH][tT][tT][pP][sS]?://(\S)*",
            "[fF][iI][lL][eE]://(\S)*",
            "[fF][tT][pP]://(\S)*",
            "\S+@\S+"
        ]
    }    
}
```

## Kagome

Use [Kagome](https://github.com/ikawaha/kagome), a morphological analyzer for Japanese, to split Japanese text into tokens.

- `dictionary`: You can set `IPADIC` or `UniDIC`.
- `stop_tags`: You can specify the Japanese part of speech to be removed. The specified part of speech will not be output as a token.
- `base_form`: Converts the token of the specified Japanese part of speech to its base form. Example, convert `美しく` to `美しい`.

Example:
```json
{
    "name": "kagome",
    "options": {
        "dictionary": "IPADIC",
        "stop_tags": [
            "接続詞",
            "助詞",
            "助詞-格助詞",
            "助詞-格助詞-一般",
            "助詞-格助詞-引用",
            "助詞-格助詞-連語",
            "助詞-接続助詞",
            "助詞-係助詞",
            "助詞-副助詞",
            "助詞-間投助詞",
            "助詞-並立助詞",
            "助詞-終助詞",
            "助詞-副助詞／並立助詞／終助詞",
            "助詞-連体化",
            "助詞-副詞化",
            "助詞-特殊",
            "助動詞",
            "記号",
            "記号-一般",
            "記号-読点",
            "記号-句点",
            "記号-空白",
            "記号-括弧開",
            "記号-括弧閉",
            "その他-間投",
            "フィラー",
            "非言語音"
        ],
        "base_forms": [
            "動詞",
            "形容詞",
            "形容動詞"
        ]
    }    
}
```


## Letter

This is the same as specifying `letter` for `rune` option in the Character tokenizer.  

Example:
```json
{
    "name": "letter"
}
```


## Regular Expression

Outputs strings that matches the specified regular expression as a token.  

Example:
```json
{
    "name": "regexp",
    "pattern": "[0-9a-zA-Z_]*"
}
```


## Single Token

Output text as a single token.  

Example:
```json
{
    "name": "single_token",
}
```


## Unicode

Output tokens based on Unicode character categories.  

Example:
```json
{
    "name": "unicode",
}
```


## Web

Extracts E-mail, URL, Twitter handle, and Twitter hashtag from web content based on Exception tokenizer and outputs the token.  

Example:
```json
{
    "name": "web",
}
```


## Whitespace

Outputs tokens by splitting the text in whitespace.  

Example:
```json
{
    "name": "whitespace",
}
```
