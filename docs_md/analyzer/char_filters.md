# Char Filters

Character filters are used to preprocess the stream of characters before it is passed to the [tokenizer](/analyzer/tokenizers.md).  

The format of the char filter definition is as follows:
```
{
    "name": <CHAR_FILTER_NAME>,
    "options": <CHAR_FILTER_OPTIONS>
}
```
- `<CHAR_FILTER_NAME>`: 
- `<CHAR_FILTER_OPTIONS>`: 

The following char filters are available:
- ASCII folding
- HTML
- Regular Expression
- Unicode Normalize
- Zero width non-joiner


## ASCII folding

Converts alphabetic, numeric, and symbolic characters that are not in the Basic Latin Unicode block (first 127 ASCII characters) to their ASCII equivalent, if one exists. For example, the filter changes `à` to `a`.  

Example:  
```json
{
    "name": "ascii_folding"
}
```

## HTML

Replace HTML tags to whitespace(` `).  

Example:
```json
{
    "name": "html"
}
```


## Regular Expression

Replaces characters that match the regular expression with the specified characters.  

Example:
```json
{
    "name": "regex",
    "options": {
        "pattern": "foo",
        "replacement": "var"
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


## Zero width non-joiner

Replaces characters that zero width non-joiner(`U+200C`) with the whitespace (` `).  

Example:
```json
{
    "name": "zero_width_non_joiner"
}
```
