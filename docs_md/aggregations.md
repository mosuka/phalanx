# Aggregations

Aggregation is the compilation of indexed data as metrics, statistics, and other analyses.  

Phalanx provides two types of aggregation.  
- Bucket: Aggregations that group documents into buckets, based on field values, ranges.
- Metric: Aggregations that calculate metrics, such as a sum or average, from field values.

The format of the analyzer definition is as follows:
```
{
    <AGGREGATION_NAME>: {
        "type": <AGGREGATION_TYPE>,
        "options": <AGGREGATION_OPTIONS>
    }
}
```
- `<AGGREGATION_NAME>`: Aggregation name.
- `<AGGREGATION_OPTIONS>`: Aggregation options.


## Bucket

### Terms

The terms aggregation typically operates on field data. Each term seen becomes it’s own bucket, and by default the count metric is applied to each bucket. Finally, at the conclusion of the search, these buckets are sorted by their counts descending, and the top N buckets are returned as part of the result.  

Example:
```json
{
    "tags_terms": {
        "type": "terms",
        "options": {
            "field": "tags",
            "min_length": 2,
            "max_length": 10,
            "size": 10
        }
    }
}
```


### Range

The numeric range aggregation also typically operates on field data. A query time a set of buckets is statically defined, which describe interesting numeric ranges. The aggregation by default includes the count metric, keeping track of how many documents had a numeric field value within the range.  

Example:
```json
{
    "price_range": {
        "type": "range",
        "options": {
            "field": "price",
            "ranges": {
                "low": {
                    "low": 0,
                    "high": 500
                },
                "medium": {
                    "low": 500,
                    "high": 1000
                },
                "high": {
                    "low": 1000,
                    "high": 1500
                }
            }
        }
    }
}
```


### Date Range

The date range aggregation also typically operates on field data. A query time a set of buckets is statically defined, which describe interesting date ranges. The aggregation by default includes the count metric, keeping track of how many documents had a date time field value within the range.  

Example:
```json
{
    "timestamp_range": {
        "type": "date_range",
        "options": {
            "field": "timestamp",
            "ranges": {
                "year_before_last": {
                    "start": "2020-01-01T00:00:00Z",
                    "end": "2021-01-01T00:00:00Z"
                },
                "last_year": {
                    "start": "2021-01-01T00:00:00Z",
                    "end": "2022-01-01T00:00:00Z"
                },
                "this_year": {
                    "start": "2022-01-01T00:00:00Z",
                    "end": "2023-01-01T00:00:00Z"
                }
            }
        }
    }
}
```


### Metric

The following basic single-value metrics are supported:

- Sum
- Min
- Max
- Avg

#### Sum

Calculates the sum of the specified field values.

Example:
```json
{
    "sum_point": {
        "type": "sum",
        "options": {
            "field": "point"
        }
    }
}
```


#### Min

Returns the minimum value of the specified field value.

Example:
```json
{
    "min_price": {
        "type": "min",
        "options": {
            "field": "price"
        }
    }
}
```


#### Max

Returns the maximum value of the specified field value.

Example:
```json
{
    "max_price": {
        "type": "max",
        "options": {
            "field": "price"
        }
    }
}
```


#### Avg

Returns the average value of the specified field values.

Example:
```json
{
    "avg_price": {
        "type": "avg",
        "options": {
            "field": "price"
        }
    }
}
```
