# Run with MinIO and etcd

To experience Phalanx functionality, let's start Phalanx with [MinIO](https://min.io/) and [etcd](https://etcd.io/). 
This repository has a docker-compose.yml file. With it, you can easily launch MinIO and etcd on Docker.

```
% docker-compose up --force-recreate etcd etcdkeeper minio
```

Once the container has been started, you can check the MinIO and etcd data in your browser at the following URL.

- MinIO  
http://localhost:9001/dashboard

- ETCD Keeper  
http://localhost:8080/etcdkeeper/


## Start Phalanx with etcd metastore

```
% ./bin/phalanx --index-metastore-uri=etcd://phalanx-metadata
```


### Create index with MinIO and etcd

Use MinIO as index storage, and create a lock on etcd to avoid write conflicts.

```
% curl -XPUT -H 'Content-type: application/json' http://localhost:8000/v1/indexes/example --data-binary '
{
	"index_uri": "minio://phalanx-indexes/example",
	"lock_uri": "etcd://phalanx-locks/example",
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