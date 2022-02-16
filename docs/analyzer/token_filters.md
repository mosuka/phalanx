# Token filters

Token filters accept a stream of tokens from a [tokenizer](./analyzer/tokenizers.md) and can modify tokens (e.g., lowercasing), delete tokens (e.g., remove stopwords).

The format of the char filter definition is as follows:
```
{
    "name": <TOKEN_FILTER_NAME>,
    "options": <TOKEN_FILTER_OPTIONS>
}
```
- `<TOKEN_FILTER_NAME>`: 
- `<TOKEN_FILTER_OPTIONS>`: 

The following char filters are available:
- Apostrophe
- Camel Case
- Dictionary Compound
- Edge Ngram
- Elision
- Keyword Marker
- Length
- Lower Case
- Ngram
- Porter Stemmer
- Reverse
- Shingle
- Stop Tokens
- Truncate
- Unicode Normalize
- Unique Term



## Apostrophe

Remove strings even after the apostrophe of the token.  

Example:  
```json
{
    "name": "apostrophe"
}
```


## Camel Case

Split the CamelCase token further. For example, `GoLang` will be split into `Go` and `Lang`.

Example:  
```json
{
    "name": "camel_case"
}
```


## Dictionary Compound

The token is further divided based on the dictionary. In the following example, `softball` will be split into two tokens, `soft` and `ball`.  

Example:  
```json
{
    "name": "dictionary_compound",
    "options": {
        "words": [
            "soft",
            "softest",
            "ball"
        ],
        "min_word_size": 5,
        "min_sub_word_size": 2,
        "max_sub_word_size": 15,
        "only_longest_match": false
    }
}
```


## Edge Ngram

Generates edge n-gram tokens of sizes within the given range.  

Example:  
```json
{
    "name": "edge_ngram",
    "options": {
        "back": false,
        "min_length": 1,
        "max_length": 2
    }
}
```


## Elision

Output tokens without the prefix specified by `articles`. In the example below, the token `ar'word` will be output as a token `word`.  

Example:  
```json
{
    "name": "elision",
    "options": {
        "articles": [
            "ar"
        ]
    }
}
```


## Keyword Marker

Set the `KeyWord` member variable to `true` for tokens that match the string specified by the `keywords` option. You can mark special tokens.

Example:  
```json
{
    "name": "keyword_marker",
    "options": {
        "keywords": [
            "walk",
            "park"
        ]
    }
}
```


## Length

Removes tokens shorter or longer than specified character lengths.  

Example:  
```json
{
    "name": "length",
    "options": {
        "min_length": 3,
        "max_length": 4
    }
}
```


## Lower Case

Changes token text to lowercase.  

Example:  
```json
{
    "name": "lower_case"
}
```


## Ngram

Forms n-grams of specified lengths from a token.  

Example:  
```json
{
    "name": "ngram",
    "options": {
        "min_length": 1,
        "max_length": 2
    }
}
```


## Porter Stemmer

Provides algorithmic stemming, based on the Porter stemming algorithm.  

Example:  
```json
{
    "name": "porter_stemmer"
}
```


## Reverse

Reverses each token in a stream.  

Example:  
```json
{
    "name": "reverse"
}
```


## Shingle

Add shingles, or word n-grams, to a token stream by concatenating adjacent tokens.  

Example:  
```json
{
    "name": "shingle",
    "options":{
        "min_length": 2,
        "max_length": 2,
        "output_original": true,
        "token_separator": " ",
        "fill": "_"
    }
}
```


## Stop Tokens

Removes stop words from a token stream.

Example:  
```json
{
    "name": "stop_tokens",
    "options":{
        "stop_tokens": [
            "a",
            "an",
            "and",
            "are",
            "as",
            "at",
            "be",
            "but",
            "by",
            "for",
            "if",
            "in",
            "into",
            "is",
            "it",
            "no",
            "not",
            "of",
            "on",
            "or",
            "such",
            "that",
            "the",
            "their",
            "then",
            "there",
            "these",
            "they",
            "this",
            "to",
            "was",
            "will",
            "with"
        ]
    }
}
```


## Truncate

Truncates tokens that exceed a specified character limit.

Example:  
```json
{
    "name": "stop_tokens",
    "options":{
        "length": 5
    }
}
```


## Unicode Normalize

Performs unicode normalization. The following parameters can be set for `form`.  
- `NFD`
- `NFC`
- `NFKD`
- `NFKC`

Example:
```json
{
    "name": "unicode_normalize",
    "options": {
        "form": "NFKC"
    }    
}
```


## Unique Term

Removes duplicate tokens from a stream.

Example:
```json
{
    "name": "unique_term"
}
```
