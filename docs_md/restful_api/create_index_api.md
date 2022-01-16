# Create Index API

This API creates a new index.

## Request

```
PUT /v1/indexes/<INDEX_NAME>
```


## Path parameters

- `<INDEX_NAME>`: (Required, string) Name of the index you want to create.


## Request body

```
{
    "index_uri": <INDEX_URI>,
    "lock_uri": <LOCK_URI>,
    "index_mapping": {
        <INDEX_MAPPING>
    },
    "num_shards": <NUM_SHARDS>,
	"default_search_field": <DEFAULT_SEARCH_FIELD>,
	"default_analyzer": {
        <DEFAULT_ANALYZER>
	}
}
```

- `<INDEX_URI>`: (Required, string) Path of the index.  
Specifies the path to create the index. It currently supports the following:
  - Local file system:  
  e.g., `file:///var/lib/phalanx-indexes/wikipedia_en`  
  - [MinIO](https://min.io/):  
  e.g., `minio://phalanx-indexes/wikipedia_en`  


- `<LOCK_URI>`: (Optional, string) Path of the lock objects.  
Specifies the path to create the locks. It currently supports the following:
  - Local file system:  
  e.g., `file:///var/lib/phalanx-locks/wikipedia_en`  
  - [etcd](https://etcd.io/):  
  e.g., `etcd://phalanx-locks/wikipedia_en`  


- `<INDEX_MAPPING>`: (Required, JSON) Mapping for fields in the index.  
See [Index Mapping](/index_mapping.md) section.


- `<NUM_SHARDS>`: (Optional, integer) Number of shards in the index.  
Defaults to 1.


- `<DEFAULT_SEARCH_FIELD>`: (Optional, string) Default search field in the index.  
Defaults to `_all`.


- `<DEFAULT_ANALYZER>`: (Optional, JSON) Default analyuzer to use in the index.  
See [Analyzer](/analyzer.md) section. Defaults to use [StandardAnalyzer](https://github.com/blugelabs/bluge/blob/master/analysis/analyzer/standard.go) that will be the same as the next setting.  
```json
{
    "tokenizer": {
        "name": "unicode"
    },
    "token_filters": [
        {
            "name": "lower_case"
        }
    ]
}
```


## Examples

```
% curl -XPUT -H 'Content-type: application/json' http://localhost:8000/v1/indexes/example --data-binary '
{
	"index_uri": "file:///tmp/phalanx-indexes/example",
	"index_mapping": {
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
	},
	"num_shards": 1,
	"default_search_field": "_all",
	"default_analyzer": {
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
'
```
