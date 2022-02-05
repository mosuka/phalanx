# Run with AWS

Phalanx supports [Amazon S3](https://aws.amazon.com/s3/) and [etcd](https://etcd.io/).

If you want to try it on your local machine, you can start [LocalStack](https://localstack.cloud/) and [etcd](https://etcd.io/) with the following command to prepare a pseudo AWS environment.

```
% docker-compose up --force-recreate etcd etcdkeeper localstack
```


## Start Phalanx cluster with DynamoDB metastore

Start the first node:

```
% ./bin/phalanx --index-metastore-uri=etcd://phalanx-metadata
```

Start the second node:

```
% ./bin/phalanx --index-metastore-uri=etcd://phalanx-metadata --bind-port=2001 --grpc-port=5001 --http-port=8001 --seed-addresses=localhost:2000
```

Start the third node:

```
% ./bin/phalanx --index-metastore-uri=etcd://phalanx-metadata --bind-port=2002 --grpc-port=5002 --http-port=8002 --seed-addresses=localhost:2000
```

Use the following command to create an index of 6 shards. If you need more nodes, start them in the same way as in the above command.


## Create index with S3 and DyunamoDB

Use S3 as index storage, and create a lock on DynamoDB to avoid write conflicts.

```
% curl -XPUT -H 'Content-type: application/json' http://localhost:8000/v1/indexes/example --data-binary '
{
	"index_uri": "s3://phalanx-indexes/example",
	"lock_uri": "dynamodb://phalanx-locks/example",
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


### Cluster status

```
% curl -XGET http://localhost:8000/cluster | jq .
```

```json
{
  "indexer_assignment": {
    "example": {
      "shard-0kZFXhrZ": "node-8giCE4QA",
      "shard-FretswBz": "node-Rh6kmVO8",
      "shard-Lf7rlqwd": "node-Rh6kmVO8",
      "shard-WvUd7lWm": "node-Rh6kmVO8",
      "shard-f4f6jbCi": "node-Rh6kmVO8",
      "shard-kCPFctnO": "node-8giCE4QA"
    }
  },
  "indexes": {
    "example": {
      "index_lock_uri": "etcd://phalanx-locks/example",
      "index_uri": "minio://phalanx-indexes/example",
      "shards": {
        "shard-0kZFXhrZ": {
          "shard_lock_uri": "etcd://phalanx-locks/example/shard-0kZFXhrZ",
          "shard_uri": "minio://phalanx-indexes/example/shard-0kZFXhrZ"
        },
        "shard-FretswBz": {
          "shard_lock_uri": "etcd://phalanx-locks/example/shard-FretswBz",
          "shard_uri": "minio://phalanx-indexes/example/shard-FretswBz"
        },
        "shard-Lf7rlqwd": {
          "shard_lock_uri": "etcd://phalanx-locks/example/shard-Lf7rlqwd",
          "shard_uri": "minio://phalanx-indexes/example/shard-Lf7rlqwd"
        },
        "shard-WvUd7lWm": {
          "shard_lock_uri": "etcd://phalanx-locks/example/shard-WvUd7lWm",
          "shard_uri": "minio://phalanx-indexes/example/shard-WvUd7lWm"
        },
        "shard-f4f6jbCi": {
          "shard_lock_uri": "etcd://phalanx-locks/example/shard-f4f6jbCi",
          "shard_uri": "minio://phalanx-indexes/example/shard-f4f6jbCi"
        },
        "shard-kCPFctnO": {
          "shard_lock_uri": "etcd://phalanx-locks/example/shard-kCPFctnO",
          "shard_uri": "minio://phalanx-indexes/example/shard-kCPFctnO"
        }
      }
    }
  },
  "nodes": {
    "node-09upDjwO": {
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
    "node-8giCE4QA": {
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
    "node-Rh6kmVO8": {
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
      "shard-0kZFXhrZ": [
        "node-8giCE4QA",
        "node-Rh6kmVO8",
        "node-09upDjwO"
      ],
      "shard-FretswBz": [
        "node-Rh6kmVO8",
        "node-09upDjwO",
        "node-8giCE4QA"
      ],
      "shard-Lf7rlqwd": [
        "node-Rh6kmVO8",
        "node-09upDjwO",
        "node-8giCE4QA"
      ],
      "shard-WvUd7lWm": [
        "node-Rh6kmVO8",
        "node-8giCE4QA",
        "node-09upDjwO"
      ],
      "shard-f4f6jbCi": [
        "node-Rh6kmVO8",
        "node-8giCE4QA",
        "node-09upDjwO"
      ],
      "shard-kCPFctnO": [
        "node-8giCE4QA",
        "node-Rh6kmVO8",
        "node-09upDjwO"
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

```
% curl -XDELETE http://localhost:8000/v1/indexes/example
```
