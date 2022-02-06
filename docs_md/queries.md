# Queries

Phanalx provides a full Query DSL (Domain Specific Language) based on JSON to define queries.  


## Boolean query

A query that matches documents matching boolean combinations of other queries.  
Result documents satisfy all of the `must` queries and satisfy none of the `must_not` queries.
Also satisfy any of the `should` queries will score higher.

- `must`: A list of queries. The clause (query) must appear in matching documents and will contribute to the score.
- `must_not`: A list of queries. The clause (query) must not appear in the matching documents.
- `should`: A list of queries. The clause (query) should appear in the matching document.
- `min_should`: To specify the number of should clauses returned documents must match.
- `boost`: To boost a query. By default, the boost factor is 1.0. Although the boost factor must be positive, it can be less than 1 (for example, it could be 0.2).

```json
{
  "type": "boolean",
  "options": {
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
}
```


## Date range query

This query is for a range of date values.
The datetime string must be RFC3339, and both endpoints cannot be empty string.
`inclusive_start` and `inclusive_end` control the inclusion of the endpoints.

- `start`: Start datetime string must be RFC3339.
- `end` : End datetime string must be RFC3339.
- `inclusive_start`: Specifies whether or not to include the start datetime.
- `inclusive_end`: Specifies whether or not to include the end datetime.
- `field`: Specify the target field name.
- `boost`: To boost a query. By default, the boost factor is 1.0. Although the boost factor must be positive, it can be less than 1 (for example, it could be 0.2).

```json
{
  "type": "date_range",
  "options": {
    "start": "2022-01-01T00:00:00Z",
    "end": "2023-01-01T00:00:00Z",
    "inclusive_start": true,
    "inclusive_end": false,
    "field": "description",
    "boost": 1.0
  }
}
```


## Fuzzy query

Fuzzy query finds documents containing terms within a specific fuzziness of the specified term.
The default fuzziness is 1. 

The current implementation uses Levenshtein edit distance as the fuzziness metric.

- `term`: Specify the term to search for.
- `prefix`: Number of beginning characters left unchanged when creating expansions.
- `fuzziness`: Maximum edit distance allowed for matching.
- `field`: Specify the target field name.
- `boost`: To boost a query. By default, the boost factor is 1.0. Although the boost factor must be positive, it can be less than 1 (for example, it could be 0.2).

```json
{
  "type": "fuzzy",
  "options": {
    "term": "hello",
    "prefix": 1,
    "fuzziness": 1,
    "field": "description",
    "boost": 1.0
  }
}
```


## Geo bounding box query

This query is for performing geo bounding box searches. The arguments describe the position of the box and documents which have an indexed geo point inside the box will be returned.

- `top_left_point.lon`: Longitude of the corner where the bounding box starts.
- `top_left_point.lat`: Latitude of the corner where the bounding box starts.
- `bottom_right_point.lon`: Longitude of the corner where the bounding box ends.
- `bottom_right_point.lat`: Latitude of the corner where the bounding box ends.
- `field`: Specify the target field name.
- `boost`: To boost a query. By default, the boost factor is 1.0. Although the boost factor must be positive, it can be less than 1 (for example, it could be 0.2).

```json
{
  "type": "geo_bounding_box",
  "options": {
    "top_left_point": {
      "lon": 40.73,
      "lat": -74.1
    },
    "bottom_right_point": {
      "lon": 40.01,
      "lat": -71.12
    },
    "field": "location",
    "boost": 1.0
  }
}
```


## Geo bounding polygon query

This query is for performing geo bounding polygon searches. The arguments describe the position of the polygon and documents which have an indexed geo point inside the box will be returned.

- `points`: A list of longitude and latitude that expresses the location of the polygon's corners.
- `field`: Specify the target field name.
- `boost`: To boost a query. By default, the boost factor is 1.0. Although the boost factor must be positive, it can be less than 1 (for example, it could be 0.2).

```json
{
  "type": "get_bounding_polygon",
  "options": {
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
}
```


## Geo Distance query

This is for performing geo distance searches. The arguments describe a position and a distance.
Documents which have an indexed geo point which is less than or equal to the provided distance from the given position will be returned.

- `point`: Specify the longitude and latitude of the center point.
- `distance`: Specify the distance from the center point. The units that can be specified are `in`/`inches`, `yd`/`yards`, `ft`/`feet`, `km`/`kilometers`, `nm`/`nauticalmiles`, `mm`/`millimeters`, `cm`/`centimeters`, `mi`/`mile`, `m`/`meters`.
- `field`: Specify the target field name.
- `boost`: To boost a query. By default, the boost factor is 1.0. Although the boost factor must be positive, it can be less than 1 (for example, it could be 0.2).

```json
{
  "type": "geo_distance",
  "options": {
    "point": {
      "lon": 40.73,
      "lat": -74.1
    },
    "distance": "1km",
    "field": "location",
    "boost": 1.0
  }
}
```


## Match query

This query is for matching text. An Analyzer is chosen based on the field. Input text is analyzed using this analyzer. Token terms resulting from this analysis are used to perform term searches. Result documents must satisfy at least one of these term searches.

- `match`: Specify the text to search for.
- `prefix`: Number of beginning characters left unchanged when creating expansions.
- `fuzziness`: Maximum edit distance allowed for matching.
- `operator`: Specifies the operator to be applied when searching for terms analyzed by the analyzer. Can be specified are `AND` or `OR`.
- `analyzer`: Specifies the analyzer to analyze the specified text. If omitted, the default analyzer will be applied. See [Analyzer](/analyzer.md) section for details on how to specify the analyzer.
- `field`: Specify the target field name.
- `boost`: To boost a query. By default, the boost factor is 1.0. Although the boost factor must be positive, it can be less than 1 (for example, it could be 0.2).

```json
{
  "type": "match",
  "options": {
    "match": "hello",
    "prefix": 1,
    "fuzziness": 1,
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
    },
    "field": "description",
    "boost": 1.0
  }
}
```


## Match all query

This query will match all documents in the index.

- `boost`: To boost a query. By default, the boost factor is 1.0. Although the boost factor must be positive, it can be less than 1 (for example, it could be 0.2).

```json
{
  "type": "match_all",
  "options": {
    "boost": 1.0
  }
}
```


## Match none query

This query will not match any documents in the index.

- `boost`: To boost a query. By default, the boost factor is 1.0. Although the boost factor must be positive, it can be less than 1 (for example, it could be 0.2).

```json
{
  "type": "match_none",
  "options": {
    "boost": 1.0
  }
}
```


## Match phrase query

This query is for matching phrases in the index. An Analyzer is chosen based on the field.
Input text is analyzed using this analyzer. Token terms resulting from this analysis are used to build a search phrase.  
Result documents must match this phrase. Queried field must have been indexed with `IncludeTermVectors` set to true.

- `phrase`: Specify the text to search for.
- `slop`: A phrase query matches terms up to a configurable slop (which defaults to 0) in any order.
- `analyzer`: Specifies the analyzer to analyze the specified text. If omitted, the default analyzer will be applied. See [Analyzer](/analyzer.md) section for details on how to specify the analyzer.
- `field`: Specify the target field name.
- `boost`: To boost a query. By default, the boost factor is 1.0. Although the boost factor must be positive, it can be less than 1 (for example, it could be 0.2).

```json
{
  "type": "match_phrase",
  "options": {
    "phrase": "hello world",
    "slop": 1,
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
    },
    "field": "description",
    "boost": 1.0
  }
}
```


## Multi phrase query

This query is for finding term phrases in the index.
It is like match phrase query, but each position in the phrase may be satisfied by a list of terms as opposed to just one.
At least one of the terms must exist in the correct order, at the correct index offsets, in the specified field. Queried field must have been indexed with `IncludeTermVectors` set to true.

- `terms`: Specify a list of multiple terms to search for phrases.
- `slop`: A phrase query matches terms up to a configurable slop (which defaults to 0) in any order.
- `field`: Specify the target field name.
- `boost`: To boost a query. By default, the boost factor is 1.0. Although the boost factor must be positive, it can be less than 1 (for example, it could be 0.2).

```json
{
  "type": "multi_phrase",
  "options": {
    "terms": [
      [
        "foo",
        "bar"
      ],
      [
        "baz"
      ]
    ],
    "slop": 1,
    "field": "description",
    "boost": 1.0
  }
}
```


## Numeric range query

This query is for ranges of numeric values. Either, but not both endpoints can be omitted.

- `min`: The minimum value.
- `max`: The maximum value.
- `inclusive_min`: Specifies whether or not to include the minimum value.
- `inclusive_max`: Specifies whether or not to include the maximum value.
- `field`: Specify the target field name.
- `boost`: To boost a query. By default, the boost factor is 1.0. Although the boost factor must be positive, it can be less than 1 (for example, it could be 0.2).

```json
{
  "type": "numeric_range",
  "options": {
    "min": 0.0,
    "max": 1.0,
    "inclusive_min": true,
    "inclusive_max": false,
    "field": "description",
    "boost": 1.0
  }
}
```


## Prefix query

This query which finds documents containing terms that start with the specified prefix.

- `prefix`: Specify a beginning characters of term.
- `field`: Specify the target field name.
- `boost`: To boost a query. By default, the boost factor is 1.0. Although the boost factor must be positive, it can be less than 1 (for example, it could be 0.2).

```json
{
  "type": "prefix",
  "options": {
    "prefix": "hel",
    "field": "description",
    "boost": 1.0
  }
}
```


## Query string query

This query uses a syntax to parse and split the provided query string based on operators, such as AND or NOT. The query then analyzes each split text independently before returning matching documents.

- `query`: Query string you wish to parse and use for search.
- `date_format`: Specify a datetime format. See [time](https://pkg.go.dev/time#pkg-constants) package for details.
- `analyzer`: Specifies the analyzer to analyze the specified text. If omitted, the default analyzer will be applied. See [Analyzer](/analyzer.md) section for details on how to specify the analyzer.

```json
{
  "type": "query_string",
  "options": {
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
}
```


## Regexp query

This query which finds documents containing terms that match the specified regular expression.

- `regexp`: Regular expression for terms to search for.
- `field`: Specify the target field name.
- `boost`: To boost a query. By default, the boost factor is 1.0. Although the boost factor must be positive, it can be less than 1 (for example, it could be 0.2).

```json
{
  "type": "regex",
  "options": {
    "regexp": "hel.*",
    "field": "description",
    "boost": 1.0
  }
}
```


## Term query

This query searches an exact term match in the index.

- `term`: Specify the term to search for.
- `field`: Specify the target field name.
- `boost`: To boost a query. By default, the boost factor is 1.0. Although the boost factor must be positive, it can be less than 1 (for example, it could be 0.2).

```json
{
  "type": "term",
  "options": {
    "term": "hello",
    "field": "description",
    "boost": 1.0
  }
}
```


## Term range query

This query searches ranges of text terms.
Either, but not both endpoints can be "".

- `min`: The minimum value.
- `max`: The maximum value.
- `inclusive_min`: Specifies whether or not to include the minimum value.
- `inclusive_max`: Specifies whether or not to include the maximum value.
- `field`: Specify the target field name.
- `boost`: To boost a query. By default, the boost factor is 1.0. Although the boost factor must be positive, it can be less than 1 (for example, it could be 0.2).

```json
{
  "type": "term_range",
  "options": {
    "min": "a",
    "max": "z",
    "inclusive_min": true,
    "inclusive_max": false,
    "field": "description",
    "boost": 1.0
  }
}
```


## Wildcard query

This query finds documents containing terms that match the specified wildcard.  In the wildcard pattern '*' will match any sequence of 0 or more characters, and '?' will match any single character.

- `wildcard`: Specify the wildcard pattern to search for.
- `field`: Specify the target field name.
- `boost`: To boost a query. By default, the boost factor is 1.0. Although the boost factor must be positive, it can be less than 1 (for example, it could be 0.2).

```json
{
  "type": "wildcard",
  "options": {
    "wildcard": "h*",
    "field": "description",
    "boost": 1.0
  }
}
```
