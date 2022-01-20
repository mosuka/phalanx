# Run with AWS

Phalanx supports [Amazon S3](https://aws.amazon.com/s3/) and [Amazon DynamoDB](https://aws.amazon.com/dynamodb/).

If you want to try it on your local machine, you can start [LocalStack](https://localstack.cloud/) with the following command to prepare a pseudo AWS environment.

```
% docker-compose up localstack
```


## Start Phalanx cluster with DynamoDB metastore

Start the first node:

```
% ./bin/phalanx --index-metastore-uri=dynamodb://phalanx-metadata
```

Start the second node:

```
% ./bin/phalanx --index-metastore-uri=dynamodb://phalanx-metadata --bind-port=2001 --grpc-port=5001 --http-port=8001 --seed-addresses=localhost:2000
```

Start the third node:

```
% ./bin/phalanx --index-metastore-uri=dynamodb://phalanx-metadata --bind-port=2002 --grpc-port=5002 --http-port=8002 --seed-addresses=localhost:2000
```

Use the following command to create an index of 10 shards. Start the required nodes the same way.


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
	"num_shards": 10,
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
