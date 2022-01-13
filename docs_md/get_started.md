# Getting started

## Start Phalanx on a local machine using a local file system

Phalanx can be started on a local machine using a local file system as a metastore. The following command starts with a configuration file:

```
% ./bin/phalanx --index-metastore-uri=file:///tmp/phalanx/metastore
```

A metastore is a place where various information about an index is stored.  

### Create index on local file system

If you have started Phalanx to use the local file system, you can use this command to create an index.

```
% curl -XPUT -H 'Content-type: application/json' http://localhost:8000/v1/indexes/example --data-binary '
{
	"index_uri": "file:///tmp/phalanx/indexes/example",
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

## Health check

These endpoints should be used for Phalanx health checks.

### Liveness check

If Phalanx is running properly, it will return HTTP status 200.

```
% curl -XGET http://localhost:8000/livez | jq .
```

```json
{
  "state":"alive"
}
```

### Readiness check

If Phalanx is ready to accept the traffic, it will return HTTP Status 200.

```
% curl -XGET http://localhost:8000/readyz | jq .
```

```json
{
  "state":"ready"
}
```

But this endpoint is not yet fully implemented.


## Metrics exposition

This endpoint returns Phalanx metrics in Prometheus exposition format.

```
% curl -XGET http://localhost:8000/metrics
```

```text
# HELP phalanx_grpc_server_handled_total Total number of RPCs completed on the server, regardless of success or failure.
# TYPE phalanx_grpc_server_handled_total counter
phalanx_grpc_server_handled_total{grpc_code="Aborted",grpc_method="AddDocuments",grpc_service="index.Index",grpc_type="unary"} 0
phalanx_grpc_server_handled_total{grpc_code="Aborted",grpc_method="Cluster",grpc_service="index.Index",grpc_type="unary"} 0
...
phalanx_grpc_server_started_total{grpc_method="Metrics",grpc_service="index.Index",grpc_type="unary"} 1
phalanx_grpc_server_started_total{grpc_method="ReadinessCheck",grpc_service="index.Index",grpc_type="unary"} 0
phalanx_grpc_server_started_total{grpc_method="Search",grpc_service="index.Index",grpc_type="unary"} 0
```


## Cluster status

This endpoint returns the latest cluster status.
- `nodes`: Lists the nodes that are joining in the cluster.
- `indexes`: Lists the indexes served by the cluster.
- `indexer_assignment`: Lists which node is responsible for the shard in the index.
- `searcher_assignment`: Lists which nodes are responsible for the shard in the index.

```
% curl -XGET http://localhost:8000/cluster | jq .
```

```json
{
  "indexer_assignment": {
    "wikipedia_en": {
      "shard-73iAEf8K": "node-duIMwfjn",
      "shard-CRzZVi2b": "node-duIMwfjn",
      "shard-Wh7VO5Lp": "node-duIMwfjn",
      "shard-YazeIhze": "node-duIMwfjn",
      "shard-cXyt4esz": "node-duIMwfjn",
      "shard-hUM3HWQW": "node-duIMwfjn",
      "shard-jH3sTtc7": "node-duIMwfjn",
      "shard-viI2Dm3V": "node-duIMwfjn",
      "shard-y1tMwCEP": "node-duIMwfjn",
      "shard-y7VRCIlU": "node-duIMwfjn"
    }
  },
  "indexes": {
    "wikipedia_en": {
      "index_lock_uri": "",
      "index_uri": "file:///tmp/phalanx/indexes/wikipedia_en",
      "shards": {
        "shard-73iAEf8K": {
          "shard_lock_uri": "",
          "shard_uri": "file:///tmp/phalanx/indexes/wikipedia_en/shard-73iAEf8K"
        },
        "shard-CRzZVi2b": {
          "shard_lock_uri": "",
          "shard_uri": "file:///tmp/phalanx/indexes/wikipedia_en/shard-CRzZVi2b"
        },
        "shard-Wh7VO5Lp": {
          "shard_lock_uri": "",
          "shard_uri": "file:///tmp/phalanx/indexes/wikipedia_en/shard-Wh7VO5Lp"
        },
        "shard-YazeIhze": {
          "shard_lock_uri": "",
          "shard_uri": "file:///tmp/phalanx/indexes/wikipedia_en/shard-YazeIhze"
        },
        "shard-cXyt4esz": {
          "shard_lock_uri": "",
          "shard_uri": "file:///tmp/phalanx/indexes/wikipedia_en/shard-cXyt4esz"
        },
        "shard-hUM3HWQW": {
          "shard_lock_uri": "",
          "shard_uri": "file:///tmp/phalanx/indexes/wikipedia_en/shard-hUM3HWQW"
        },
        "shard-jH3sTtc7": {
          "shard_lock_uri": "",
          "shard_uri": "file:///tmp/phalanx/indexes/wikipedia_en/shard-jH3sTtc7"
        },
        "shard-viI2Dm3V": {
          "shard_lock_uri": "",
          "shard_uri": "file:///tmp/phalanx/indexes/wikipedia_en/shard-viI2Dm3V"
        },
        "shard-y1tMwCEP": {
          "shard_lock_uri": "",
          "shard_uri": "file:///tmp/phalanx/indexes/wikipedia_en/shard-y1tMwCEP"
        },
        "shard-y7VRCIlU": {
          "shard_lock_uri": "",
          "shard_uri": "file:///tmp/phalanx/indexes/wikipedia_en/shard-y7VRCIlU"
        }
      }
    }
  },
  "nodes": {
    "node-duIMwfjn": {
      "addr": "0.0.0.0",
      "meta": {
        "grpc_port": 5000,
        "http_port": 8000,
        "roles": [
          "indexer",
          "searcher"
        ]
      },
      "port": 3000,
      "state": "alive"
    }
  },
  "searcher_assignment": {
    "wikipedia_en": {
      "shard-73iAEf8K": [
        "node-duIMwfjn"
      ],
      "shard-CRzZVi2b": [
        "node-duIMwfjn"
      ],
      "shard-Wh7VO5Lp": [
        "node-duIMwfjn"
      ],
      "shard-YazeIhze": [
        "node-duIMwfjn"
      ],
      "shard-cXyt4esz": [
        "node-duIMwfjn"
      ],
      "shard-hUM3HWQW": [
        "node-duIMwfjn"
      ],
      "shard-jH3sTtc7": [
        "node-duIMwfjn"
      ],
      "shard-viI2Dm3V": [
        "node-duIMwfjn"
      ],
      "shard-y1tMwCEP": [
        "node-duIMwfjn"
      ],
      "shard-y7VRCIlU": [
        "node-duIMwfjn"
      ]
    }
  }
}
```


## Add / Update documents

```
% curl -XPUT -H 'Content-type: application/x-ndjson' http://localhost:8000/v1/indexes/example/documents --data-binary '
{"_id":"1", "id":1, "text":"This is an example document 1."}
{"_id":"2", "id":2, "text":"This is an example document 2."}
{"_id":"3", "id":3, "text":"This is an example document 3."}
'
```


## Delete documents

```
% curl -XDELETE -H 'Content-type: text/plain' http://localhost:8000/v1/indexes/example/documents --data-binary '
1
2
3
'
```


## Search

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
'
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


## Delete index

The following command will delete the index `wikipedia_en` with the specified name. This command will delete the index file on the object storage and the index metadata on the metastore.

```
% curl -XDELETE http://localhost:8000/v1/indexes/example
```
