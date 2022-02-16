# Run with AWS

Phalanx supports [Amazon S3](https://aws.amazon.com/s3/) and [etcd](https://etcd.io/).

If you want to try it on your local machine, you can start [LocalStack](https://localstack.cloud/) and [etcd](https://etcd.io/) with the following command to prepare a pseudo AWS environment.

```
% docker-compose up --force-recreate etcd etcdkeeper localstack
```


## Start Phalanx cluster with DynamoDB metastore

```
% ./bin/phalanx --index-metastore-uri=etcd://phalanx-metadata
```


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
