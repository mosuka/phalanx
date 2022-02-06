# Queries

Phanalx provides a full Query DSL (Domain Specific Language) based on JSON to define queries.  


## Boolean query

A query that matches documents matching boolean combinations of other queries.  
Result documents satisfy all of the `must` queries and satisfy none of the `must_not` queries.
Also satisfy any of the `should` queries will score higher.

```json
{
  "must": [
    {
      "type": "term",
      "options": {
        "term": "hello",
        "field": "description",
        "boost": 1.0
      }
    },
    {
      "type": "term",
      "options": {
        "term": "world",
        "field": "description",
        "boost": 1.0
      }
    }
  ],
  "must_not": [
    {
      "type": "term",
      "options": {
        "term": "bye",
        "field": "description",
        "boost": 1.0
      }
    },
    {
      "type": "term",
      "options": {
        "term": "さようなら",
        "field": "description",
        "boost": 1.0
      }
    }
  ],
  "should": [
    {
      "type": "term",
      "options": {
        "term": "こんにちは",
        "field": "description",
        "boost": 1.0
      }
    },
    {
      "type": "term",
      "options": {
        "term": "世界",
        "field": "description",
        "boost": 1.0
      }
    }
  ],
  "min_should": 1,
  "boost": 1.0
}
```


## Date range query

This query is for a range of date values.
The date string must be RFC3339, and both endpoints cannot be `nil`.
`inclusive_start` and `inclusive_end` control the inclusion of the endpoints.

```json
{
  "start": "2022-01-01T00:00:00Z",
  "end": "2023-01-01T00:00:00Z",
  "inclusive_start": true,
  "inclusive_end": false,
  "field": "description",
  "boost": 1.0
}
```


## Fazzy query

Fuzzy query finds documents containing terms within a specific fuzziness of the specified term.
The default fuzziness is 1. 

The current implementation uses Levenshtein edit distance as the fuzziness metric.

```json
{
  "term": "hello",
  "prefix": 1,
  "fuzziness": 1,
  "field": "description",
  "boost": 1.0
}
```


## Geo bounding box query

This query is for performing geo bounding box searches. The arguments describe the position of the box and documents which have an indexed geo point inside the box will be returned.

```json
{
  "top_left_point": {
    "lon": 40.73,
    "lat": -74.1
  },
  "bottom_right_point": {
    "lon": 40.01,{
  "wildcard": "h*",
  "field": "description",
  "boost": 1.0
}
    "lat": -71.12
  },
  "field": "location",
  "boost": 1.0
}
```


## Geo bounding polygon query

This query is for performing geo bounding polygon searches. The arguments describe the position of the polygon and documents which have an indexed geo point inside the box will be returned.

```json
{
  "points": [
    {
      "lon": 40.73,
      "lat": -74.1
    },
    {
      "lon": 31.21,
      "lat": -80.11
    },
    {
      "lon": 24.87,
      "lat": -94.31
    }
  ],
  "field": "location",
  "boost": 1.0
}
```


## Geo Distance query

This is for performing geo distance searches. The arguments describe a position and a distance.
Documents which have an indexed geo point which is less than or equal to the provided distance from the given position will be returned.

```json
{
  "point": {
    "lon": 40.73,
    "lat": -74.1
  },
  "distance": "1km",
  "field": "location",
  "boost": 1.0
}
```


## Match query

This query is for matching text. An Analyzer is chosen based on the field. Input text is analyzed using this analyzer. Token terms resulting from this analysis are used to perform term searches. Result documents must satisfy at least one of these term searches.

```json
{
  "match": "hello",
  "field": "description",
  "prefix": 1,
  "fuzziness": 1,
  "boost": 1.0,
  "operator": "AND",
  "analyzer": {
    "char_filters": [
      {
        "name": "unicode_normalize",
        "options": {
          "form": "NFKC"
        }
      }
    ],
    "tokenizer": {
      "name": "whitespace"
    },
    "token_filters": [
      {
        "name": "lower_case"
      }
    ]
  }
}
```


## Match all query

This query will match all documents in the index.

```json
{
  "boost": 1.0
}
```


## Match none query

This query will not match any documents in the index.

```json
{
  "boost": 1.0
}
```


## Match phrase query

This query is for matching phrases in the index. An Analyzer is chosen based on the field.
Input text is analyzed using this analyzer. Token terms resulting from this analysis are used to build a search phrase.  
Result documents must match this phrase. Queried field must have been indexed with `IncludeTermVectors` set to true.

```json
{
  "phrase": "hello world",
  "field": "description",
  "slop": 1,
  "boost": 1.0,
  "analyzer": {
    "char_filters": [
      {
        "name": "unicode_normalize",
        "options": {
          "form": "NFKC"
        }
      }
    ],
    "tokenizer": {
      "name": "whitespace"
    },
    "token_filters": [
      {
        "name": "lower_case"
      }
    ]
  }
}
```


## Multi phrase query

This query is for finding term phrases in the index.
It is like match phrase query, but each position in the phrase may be satisfied by a list of terms as opposed to just one.
At least one of the terms must exist in the correct order, at the correct index offsets, in the specified field. Queried field must have been indexed with `IncludeTermVectors` set to true.

```json
{
  "terms": [
    [
      "foo",
      "bar"
    ],
    [
      "baz"
    ]
  ],
  "field": "description",
  "slop": 1,
  "boost": 1.0
}
```


## Numeric range query

This query is for ranges of numeric values. Either, but not both endpoints can be `nil`. 

```json
{
  "min": 0.0,
  "max": 1.0,
  "inclusive_min": true,
  "inclusive_max": false,
  "field": "description",
  "boost": 1.0
}
```


## Prefix query

This query which finds documents containing terms that start with the specified prefix.

```json
{
  "prefix": "hel",
  "field": "description",
  "boost": 1.0
}
```


## Query string query

This query uses a syntax to parse and split the provided query string based on operators, such as AND or NOT. The query then analyzes each split text independently before returning matching documents.

```json
{
  "query": "hello AND world",
  "date_format": "RFC3339",
  "analyzers": {
    "title": {
      "char_filters": [
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
    },
    "description": {
      "char_filters": [
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


## Regexp query

This query which finds documents containing terms that match the specified regular expression.

```json
{
  "regexp": "hel.*",
  "field": "description",
  "boost": 1.0
}
```


## Term query

This query searches an exact term match in the index.

```json
{
  "term": "hello",
  "field": "description",
  "boost": 1.0
}
```


## Term range query

This query searches ranges of text terms.
Either, but not both endpoints can be "".

```json
{
  "min": "a",
  "max": "z",
  "inclusive_min": true,
  "inclusive_max": false,
  "field": "description",
  "boost": 1.0
}
```


## Wildcard query

This query finds documents containing terms that match the specified wildcard.  In the wildcard pattern '*' will match any sequence of 0 or more characters, and '?' will match any single character.

```json
{
  "wildcard": "h*",
  "field": "description",
  "boost": 1.0
}
```
