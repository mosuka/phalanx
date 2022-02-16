# Highlights

Highlights enable you to get highlighted snippets from one or more fields in your search results so you can show users where the query matches are.

The format of the highlight definition is as follows:
```
{
    <FIELD_NAME>: {
        "highlighter": {
            "type": <HIGHLIGHTER_TYPE>,
            "options": <HIGHLIGHTER_OPTIONS>
        },
        "num": <NUMBER_OF_FRAGMENTS>
    }
}
```
- `<HIGHLIGHTER_TYPE>`: Highlighter name.
- `<HIGHLIGHTER_OPTIONS>`: Highlighter options.
- `<NUMBER_OF_FRAGMENTS>`: Number of fragments.


## ANSI Highlighter

Example:
```json
{
    "field_a": {
        "highlighter": {
            "type": "ansi",
            "options": {
                "fragment_size": 100,
                "color": "FgCyan",
                "separator": "…"
            }
        },
        "num": 3
    }
}
```


## HTML Highlighter

Example:
```json
{
    "field_a": {
        "highlighter": {
            "type": "html",
            "options": {
                "fragment_size": 100,
                "pre_tag": "<mark>",
                "post_tag": "</mark>",
                "separator": "…"
            }
        },
        "num": 3
    }
}
```
