# Analyzer

Analyzer is the process of converting unstructured text, like the body of an email or a product description, into a structured data that’s optimized for search.  
The analyzer consists of three elements: `char_filtes`, `tokenizer`, and `token_filters`. It processes unstructured text in the order of `char_filtes`, `tokenizer`, and `token_filters`.  
The format of the analyzer definition is as follows:
```
{
    "char_filters": <CHAR_FILTERS>,
    "tokenizer": <TOKENIZER>,
    "token_filters": <TOKEN_FILTERS>
}
```
- `<CHAR_FILTERS>`: (Optional) Char filter settings. If you write multiple char_filter in this part, you can process that order. See [Char filters](/analyzer/char_filters.md) section.

- `<TOKENIZER>`: (Required) Tokenizer setting. Specifies a tokenizer to split the text into tokens. See [Tokenizers](/analyzer/tokenizers.md) section.

- `<TOKEN_FILTERS>`: (Optional) Token filter settings. If you write multiple token_filter in this part, you can process that order.  See [Token filters](/analyzer/token_filters.md) section.


## Example

```
{
    "char_filters": [
        {
            "name": "ascii_folding"
        },
        {
            "name": "unicode_normalize",
            "options": {
                "form": "NFKC"
            }
        }
    ],
    "tokenizer": {
        "name": "unicode"
    },
    "token_filters": [
        {
            "name": "lower_case"
        },
        {
            "name": "ngram",
            "options": {
                "min_length": 1,
                "max_length": 2
            }
        }
    ]
}
```
