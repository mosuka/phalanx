# Search API

This API searches documents from an index.

## Request

```
POST /v1/indexes/<INDEX_NAME>/_search
```


## Path parameters

- `<INDEX_NAME>`: (Required, string) Name of the index you want to search documents.


## Request body

```
{
    "query": <QUERY>,
    "boost": <BOOST>,
    "start": <START>,
    "num": <NUM_DOCS>,
    "sort_by": <SORT_BY>,
    "fields": <FIELDS>,
    "aggregations": <AGGREGATIONS>
}
```

- `<QUERY>`: (Required, string) Query string.  


- `<BOOST>`: (Optional, float) The factor by which scores are multiplied.


- `<START>`: (Optional, integer) Starting document offset.  
Defaults to `0`.


- `<NUM_DOCS>`: (Optional, integer) Number of documents to retrieve from index.  
Defaults to `10`.


- `<SORT_BY>`: (Optional, string) Field name and order to sort.  
If omitted, it will be listed in order of the same score as `-_score`.


- `<FIELDS>`: (Optional, array of strings) Field names to retrieve from document.  


- `<AGGREGATIONS>`: (Optional, JSON) Default analyuzer to use in the index.  
See [Aggregations](/aggregations.md) section.  


## Response body

```
{
  "aggregations": <AGGREGATIONS>,
  "documents": <DOCUMENTS>,
  "hits": <NUM_HITS>,
  "index_name": <INDEX_NAME>
}
```

- `<AGGREGATIONS>`: (Optional, JSON) Aggregation response.  


- `<DOCUMENTS>`: (array of JSON) List of retrieved documents.  
The JSON of the document is in the following format:
```
{
	"fields": {
		<FIELD_NAME>: <FIELD_VALUE>,
		...
	},
	"id": <DOC_ID>,
	"score": <SCORE>,
	"timestamp": <TIMESTAMP>
}
```
	- `<FIELD_NAME>`: 
	- `<FIELD_VALUE>`: 
	- `<DOC_ID>`: 
	- `<SCORE>`: 
	- `<TIMESTAMP>`: 


- `<NUM_HITS>`: (integer) Total number of documents that match the search query.  


- `<INDEX_NAME>`: (Required, string) Name of the index.


## Examples

```
% curl -XPOST -H 'Content-type: application/json' http://localhost:8000/v1/indexes/example/_search --data-binary '
{
    "query": "text:document",
    "boost": 1.0,
    "start": 0,
    "num": 10,
    "sort_by": "-_score",
    "fields": [
        "id",
        "text"
    ],
    "aggregations": {
        "timestamp_date_range": {
            "type": "date_range",
            "options": {
                "field": "_timestamp",
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
}
' | jq .
```

```json
{
  "aggregations": {
    "timestamp_date_range": {
      "last_year": 0,
      "this_year": 3,
      "year_before_last": 0
    }
  },
  "documents": [
    {
      "fields": {
        "id": 1,
        "text": "This is an example document 1."
      },
      "id": "1",
      "score": 0.06069608755660118,
      "timestamp": 1641992086513383200
    },
    {
      "fields": {
        "id": 2,
        "text": "This is an example document 2."
      },
      "id": "2",
      "score": 0.06069608755660118,
      "timestamp": 1641992086513395500
    },
    {
      "fields": {
        "id": 3,
        "text": "This is an example document 3."
      },
      "id": "3",
      "score": 0.06069608755660118,
      "timestamp": 1641992086513399600
    }
  ],
  "hits": 3,
  "index_name": "example"
}
```
