# Index Mapping

Index mapping is the definition of how documents and the fields they contain are stored and indexed.  

The format of the field definition to be included in the index is as follows:
```
{
    <FIELD_NAME>: {
        "type": <FIELD_TYPE>,
        "options": <FIELD_OPTIONS>,
        [ "analyzer": <ANALYZER> ]
    }
    ...
}
```
- `<FIELD_NAME>`: The name of the field you want to include in the index.  


- `<FIELD_TYPE>`: The type of field you want to include in the index.  
The following types can be defined:
    - `text`: Unstructured natural language text or keywords.
    - `numeric`: Numeric types, such as long and double, used to express amounts.
    - `datetime`: DateTime string, such as date and time, formatted in RFC3339.
    - `geo_point`: Latitude and longitude points.


- `<FIELD_OPTIONS>`:  Specifies the options for how the field values are registered in the index.
The options that can be set differ for each field type.
    - `index`: (Optional, boolean) Whether or not to register the value in the index. Set to true if you want to search by the value of the field. (`text`, `numeric`, `datetime`, `geo_point`)
    - `store`: (Optional, boolean) Whether or not to store the original value. Set to true if you want to return the field values of document retrieved. (`text`, `numeric`, `datetime`, `geo_point`)
    - `term_positions`: (Optional, boolean) Set to true if you want to include information such as the position of the term in the index. (`text`)
    - `highlight`: (Optional, boolean) Set to true if you want to highlight the terms in the matching part of the query. (`text`)
    - `sortable`: (Optional, boolean) Set to true if you want to sort by field value. (`text`, `numeric`, `datetime`, `geo_point`)
    - `aggregatable`: (Optional, boolean) Set to true if you want to aggregate the values of the fields.　(`text`, `numeric`, `datetime`, `geo_point`)


- `<ANALYZER>`: (Optional, JSON) You only need to define an analyzer if you define a `text` field.  
The Analyzer defines how to analyze the value of a text field. See [Analyzer](/analyzer.md) section.


## Example

```
{
    "id": {
        "type": "numeric",
        "options": {
            "index": true,
            "store": true,
            "sortable": true,
            "aggregatable": true
        }
    },
    "text": {
        "type": "text",
        "options": {
            "index": true,
            "store": true,
            "term_positions": true,
            "highlight": true,
            "sortable": true,
            "aggregatable": true
        },
        "analyzer": {
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
                }
            ]
        }
    }
}
```