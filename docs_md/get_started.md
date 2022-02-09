# Getting started

## Start Phalanx on a local machine using a local file system

Phalanx can be started on a local machine using a local file system as a metastore. The following command starts with a configuration file:

```
% ./bin/phalanx --index-metastore-uri=file:///tmp/phalanx-metadata
```

Multiple Phalanx nodes can be started on the local machine to run a pseudo-cluster.
You can easily add nodes to the cluster with the previously started node as the seed node, using the following command.

Start the second node:

```
% ./bin/phalanx --index-metastore-uri=file:///tmp/phalanx-metadata --bind-port=2001 --grpc-port=5001 --http-port=8001 --seed-addresses=localhost:2000
```

Start the third node:

```
% ./bin/phalanx --index-metastore-uri=file:///tmp/phalanx-metadata --bind-port=2002 --grpc-port=5002 --http-port=8002 --seed-addresses=localhost:2000
```


A metastore is a place where various information about an index is stored.
See the [metadata store section](/metadata_store.md) for details.

### Create index on local file system

If you have started Phalanx to use the local file system, you can use this command to create an index.
If you have started Phalanx to use the local file system, you can use this command to create an index. If you want to use an object store for index storage, you need to specify `lock_uri` as well.
See the [index store section](/index_store.md) and the [lock store section](/lock_store.md)for details.

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
	"num_shards": 6,
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
    "example": {
      "shard-60zYKCLJ": "node-1P4Hkvhy",
      "shard-EIPjSbHG": "node-1P4Hkvhy",
      "shard-HYuLwqp8": "node-1P4Hkvhy",
      "shard-rIWGHdkS": "node-1P4Hkvhy",
      "shard-sNHs6Kyh": "node-1P4Hkvhy",
      "shard-wrAZqmV6": "node-1P4Hkvhy"
    }
  },
  "indexes": {
    "example": {
      "index_lock_uri": "",
      "index_mapping": {
        "id": {
          "type": "numeric",
          "options": {
            "index": true,
            "store": true,
            "term_positions": false,
            "highlight": false,
            "sortable": true,
            "aggregatable": true
          },
          "analyzer": {
            "char_filters": null,
            "tokenizer": {
              "name": "",
              "options": null
            },
            "token_filters": null
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
                "name": "ascii_folding",
                "options": null
              },
              {
                "name": "unicode_normalize",
                "options": {
                  "form": "NFKC"
                }
              }
            ],
            "tokenizer": {
              "name": "unicode",
              "options": null
            },
            "token_filters": [
              {
                "name": "lower_case",
                "options": null
              }
            ]
          }
        }
      },
      "index_uri": "file:///tmp/phalanx-indexes/example",
      "shards": {
        "shard-60zYKCLJ": {
          "shard_lock_uri": "",
          "shard_uri": "file:///tmp/phalanx-indexes/example/shard-60zYKCLJ"
        },
        "shard-EIPjSbHG": {
          "shard_lock_uri": "",
          "shard_uri": "file:///tmp/phalanx-indexes/example/shard-EIPjSbHG"
        },
        "shard-HYuLwqp8": {
          "shard_lock_uri": "",
          "shard_uri": "file:///tmp/phalanx-indexes/example/shard-HYuLwqp8"
        },
        "shard-rIWGHdkS": {
          "shard_lock_uri": "",
          "shard_uri": "file:///tmp/phalanx-indexes/example/shard-rIWGHdkS"
        },
        "shard-sNHs6Kyh": {
          "shard_lock_uri": "",
          "shard_uri": "file:///tmp/phalanx-indexes/example/shard-sNHs6Kyh"
        },
        "shard-wrAZqmV6": {
          "shard_lock_uri": "",
          "shard_uri": "file:///tmp/phalanx-indexes/example/shard-wrAZqmV6"
        }
      }
    }
  },
  "nodes": {
    "node-1P4Hkvhy": {
      "addr": "0.0.0.0",
      "meta": {
        "grpc_port": 5002,
        "http_port": 8002,
        "roles": [
          "indexer",
          "searcher"
        ]
      },
      "port": 2002,
      "state": "alive"
    },
    "node-snjdTQVL": {
      "addr": "0.0.0.0",
      "meta": {
        "grpc_port": 5001,
        "http_port": 8001,
        "roles": [
          "indexer",
          "searcher"
        ]
      },
      "port": 2001,
      "state": "alive"
    },
    "node-z8PozpGp": {
      "addr": "0.0.0.0",
      "meta": {
        "grpc_port": 5000,
        "http_port": 8000,
        "roles": [
          "indexer",
          "searcher"
        ]
      },
      "port": 2000,
      "state": "alive"
    }
  },
  "searcher_assignment": {
    "example": {
      "shard-60zYKCLJ": [
        "node-1P4Hkvhy",
        "node-snjdTQVL",
        "node-z8PozpGp"
      ],
      "shard-EIPjSbHG": [
        "node-1P4Hkvhy",
        "node-z8PozpGp",
        "node-snjdTQVL"
      ],
      "shard-HYuLwqp8": [
        "node-1P4Hkvhy",
        "node-z8PozpGp",
        "node-snjdTQVL"
      ],
      "shard-rIWGHdkS": [
        "node-1P4Hkvhy",
        "node-z8PozpGp",
        "node-snjdTQVL"
      ],
      "shard-sNHs6Kyh": [
        "node-1P4Hkvhy",
        "node-z8PozpGp",
        "node-snjdTQVL"
      ],
      "shard-wrAZqmV6": [
        "node-1P4Hkvhy",
        "node-snjdTQVL",
        "node-z8PozpGp"
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
{"_id":"4", "id":4, "text":"This is an example document 4."}
{"_id":"5", "id":5, "text":"This is an example document 5."}
{"_id":"6", "id":6, "text":"This is an example document 6."}
{"_id":"7", "id":7, "text":"This is an example document 7."}
{"_id":"8", "id":8, "text":"This is an example document 8."}
{"_id":"9", "id":9, "text":"This is an example document 9."}
{"_id":"10", "id":10, "text":"This is an example document 10."}
{"_id":"11", "id":11, "text":"This is an example document 11."}
{"_id":"12", "id":12, "text":"This is an example document 12."}
'
```


## Delete documents

```
% curl -XDELETE -H 'Content-type: text/plain' http://localhost:8000/v1/indexes/example/documents --data-binary '
1
2
3
4
5
6
7
8
9
10
11
12
'
```


## Search

```
% curl -XPOST -H 'Content-type: application/json' http://localhost:8000/v1/indexes/example/_search --data-binary '
{
    "query": {
        "type": "boolean",
        "options": {
            "must": [
                {
                    "type": "query_string",
                    "options": {
                        "query": "*"
                    }
                }
            ],
            "min_should": 1,
            "boost": 1.0
        }
    },
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
      "this_year": 12,
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
      "score": 0.08287343490634301,
      "timestamp": 1643177200809294300
    },
    {
      "fields": {
        "id": 2,
        "text": "This is an example document 2."
      },
      "id": "2",
      "score": 0.08287343490634301,
      "timestamp": 1643177200809316000
    },
    {
      "fields": {
        "id": 4,
        "text": "This is an example document 4."
      },
      "id": "4",
      "score": 0.047891143480830164,
      "timestamp": 1643177200809422600
    },
    {
      "fields": {
        "id": 7,
        "text": "This is an example document 7."
      },
      "id": "7",
      "score": 0.047891143480830164,
      "timestamp": 1643177200809437000
    },
    {
      "fields": {
        "id": 10,
        "text": "This is an example document 10."
      },
      "id": "10",
      "score": 0.047891143480830164,
      "timestamp": 1643177200809448400
    },
    {
      "fields": {
        "id": 11,
        "text": "This is an example document 11."
      },
      "id": "11",
      "score": 0.047891143480830164,
      "timestamp": 1643177200809457700
    },
    {
      "fields": {
        "id": 5,
        "text": "This is an example document 5."
      },
      "id": "5",
      "score": 0.06069608755660118,
      "timestamp": 1643177200809490000
    },
    {
      "fields": {
        "id": 6,
        "text": "This is an example document 6."
      },
      "id": "6",
      "score": 0.06069608755660118,
      "timestamp": 1643177200809517000
    },
    {
      "fields": {
        "id": 12,
        "text": "This is an example document 12."
      },
      "id": "12",
      "score": 0.06069608755660118,
      "timestamp": 1643177200809526300
    },
    {
      "fields": {
        "id": 3,
        "text": "This is an example document 3."
      },
      "id": "3",
      "score": 0.06069608755660118,
      "timestamp": 1643177200809375200
    }
  ],
  "hits": 12,
  "index_name": "example"
}
```


## Delete index

The following command will delete the index `wikipedia_en` with the specified name. This command will delete the index file on the object storage and the index metadata on the metastore.

```
% curl -XDELETE http://localhost:8000/v1/indexes/example
```
