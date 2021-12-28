# Phalanx

Phalanx is a cloud-native distributed search engine written in [Go](https://golang.org) built on top of [Bluge](https://github.com/blugelabs/bluge) that provides endpoints through [gRPC](https://grpc.io/) and traditional RESTful API.  
Phalanx implements a cluster formation by [hashicorp/memberlist](https://github.com/hashicorp/memberlist) and managing index metadata on [etcd](https://etcd.io/), so it is easy to bring up a fault-tolerant cluster.  
Metrics for system operation can also be output in Prometheus exposition format, so that monitoring can be done immediately using [Prometheus](https://prometheus.io/).  
Phalanx is using object storage for the storage layer, it is only responsible for the computation layer, such as indexing and retrieval processes. Therefore, scaling is easy, and you can simply add new nodes to the cluster.  
Currently, it is an alpha version and only supports [MinIO](https://min.io/) as the storage layer, but in the future it will support [Amazon S3](https://aws.amazon.com/s3/), [Google Cloud Storage](https://cloud.google.com/storage), and [Azure Blob Storage](https://azure.microsoft.com/en-us/services/storage/blobs/).  


## Build

Building Phalanx as following:

```bash
% git clone https://github.com/mosuka/phalanx.git
% cd phalanx
% make build
```


## Binary

You can see the binary file when build successful like so:

```bash
% ls ./bin
phalanx
```


## Start Phalanx on a local machine using a local file system

Phalanx can be started on a local machine using a local file system as a metastore. The following command starts with a configuration file:

```
% ./bin/phalanx --index-metastore-uri=file:///tmp/phalanx/metastore
```

A metastore is a place where various information about an index is stored.  

### Create index on local file system

If you have started Phalanx to use the local file system, you can use this command to create an index.

```
% curl -XPUT -H 'Content-type: application/json' http://localhost:8000/v1/indexes/wikipedia_en --data-binary @./examples/create_index_wikipedia_en_local.json
```

In `create_index_example_en_local.json` used in the above command, the URI of the local filesystem is specified in `index_uri` and `lock_uri`.
`index_mapping` defines what kind of fields the index has. `num_shards` specifies how many shards the index will be divided into.  
Both of the above commands will create an index named `example_en`.


## Start Phalanx on local machine with MinIO and etcd

To experience Phalanx functionality, let's start Phalanx with MinIO and etcd. 
This repository has a docker-compose.yml file. With it, you can easily launch Phalanx, MinIO and etcd on Docker.

```
% docker-compose up
```

Once the container has been started, you can check the MinIO and etcd data in your browser at the following URL.

- MinIO  
http://localhost:9001/dashboard

- ETCD Keeper  
http://localhost:8080/etcdkeeper/

### Create index with MinIO and etcd

If you have started Phalanx to use MinIO and etcd, use this command to create the index.

```
% curl -XPUT -H 'Content-type: application/json' http://localhost:8000/v1/indexes/example_en --data-binary @./examples/create_index_example_en.json
```

In the `create_index_example_en.json` used in the above command, `index_uri` is a MinIO URI and `lock_uri` is an etcd URI. This means that indexes will be created in MinIO, and locks for those indexes will be created in etcd. Phalanx uses etcd as a distributed lock manager.


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
    "example_en": {
      "shard-Dyb1CXqJ": "node-YA0Zso3w",
      "shard-OSFMC5gL": "node-YA0Zso3w",
      "shard-TQu8fyHA": "node-YA0Zso3w",
      "shard-UfilJ5I4": "node-YA0Zso3w",
      "shard-WLJEezNT": "node-YA0Zso3w",
      "shard-eH6LOGpc": "node-YA0Zso3w",
      "shard-jWU7v3MR": "node-YA0Zso3w",
      "shard-sng0xmKr": "node-YA0Zso3w",
      "shard-tKKy1LdN": "node-YA0Zso3w",
      "shard-vpI7ExL5": "node-YA0Zso3w"
    }
  },
  "indexes": {
    "example_en": {
      "index_lock_uri": "etcd://phalanx/locks/example_en",
      "index_uri": "minio://phalanx/indexes/example_en",
      "shards": {
        "shard-Dyb1CXqJ": {
          "shard_lock_uri": "etcd://phalanx/locks/example_en/shard-Dyb1CXqJ",
          "shard_uri": "minio://phalanx/indexes/example_en/shard-Dyb1CXqJ"
        },
        "shard-OSFMC5gL": {
          "shard_lock_uri": "etcd://phalanx/locks/example_en/shard-OSFMC5gL",
          "shard_uri": "minio://phalanx/indexes/example_en/shard-OSFMC5gL"
        },
        "shard-TQu8fyHA": {
          "shard_lock_uri": "etcd://phalanx/locks/example_en/shard-TQu8fyHA",
          "shard_uri": "minio://phalanx/indexes/example_en/shard-TQu8fyHA"
        },
        "shard-UfilJ5I4": {
          "shard_lock_uri": "etcd://phalanx/locks/example_en/shard-UfilJ5I4",
          "shard_uri": "minio://phalanx/indexes/example_en/shard-UfilJ5I4"
        },
        "shard-WLJEezNT": {
          "shard_lock_uri": "etcd://phalanx/locks/example_en/shard-WLJEezNT",
          "shard_uri": "minio://phalanx/indexes/example_en/shard-WLJEezNT"
        },
        "shard-eH6LOGpc": {
          "shard_lock_uri": "etcd://phalanx/locks/example_en/shard-eH6LOGpc",
          "shard_uri": "minio://phalanx/indexes/example_en/shard-eH6LOGpc"
        },
        "shard-jWU7v3MR": {
          "shard_lock_uri": "etcd://phalanx/locks/example_en/shard-jWU7v3MR",
          "shard_uri": "minio://phalanx/indexes/example_en/shard-jWU7v3MR"
        },
        "shard-sng0xmKr": {
          "shard_lock_uri": "etcd://phalanx/locks/example_en/shard-sng0xmKr",
          "shard_uri": "minio://phalanx/indexes/example_en/shard-sng0xmKr"
        },
        "shard-tKKy1LdN": {
          "shard_lock_uri": "etcd://phalanx/locks/example_en/shard-tKKy1LdN",
          "shard_uri": "minio://phalanx/indexes/example_en/shard-tKKy1LdN"
        },
        "shard-vpI7ExL5": {
          "shard_lock_uri": "etcd://phalanx/locks/example_en/shard-vpI7ExL5",
          "shard_uri": "minio://phalanx/indexes/example_en/shard-vpI7ExL5"
        }
      }
    }
  },
  "nodes": {
    "node-YA0Zso3w": {
      "addr": "172.19.0.4",
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
    "example_en": {
      "shard-Dyb1CXqJ": [
        "node-YA0Zso3w"
      ],
      "shard-OSFMC5gL": [
        "node-YA0Zso3w"
      ],
      "shard-TQu8fyHA": [
        "node-YA0Zso3w"
      ],
      "shard-UfilJ5I4": [
        "node-YA0Zso3w"
      ],
      "shard-WLJEezNT": [
        "node-YA0Zso3w"
      ],
      "shard-eH6LOGpc": [
        "node-YA0Zso3w"
      ],
      "shard-jWU7v3MR": [
        "node-YA0Zso3w"
      ],
      "shard-sng0xmKr": [
        "node-YA0Zso3w"
      ],
      "shard-tKKy1LdN": [
        "node-YA0Zso3w"
      ],
      "shard-vpI7ExL5": [
        "node-YA0Zso3w"
      ]
    }
  }
}
```


## Add / Update documents

```
% ./bin/phalanx_docs.sh -i id ./testdata/enwiki-20211201-pages-articles-multistream-1000.jsonl | curl -XPUT -H 'Content-type: application/x-ndjson' http://localhost:8000/v1/indexes/wikipedia_en/documents --data-binary @-
```


## Delete documents

```
% jq -r '.id' ./testdata/enwiki-20211201-pages-articles-multistream-1000.jsonl | curl -XDELETE -H 'Content-type: text/plain' http://localhost:8000/v1/indexes/wikipedia_en/documents --data-binary @-
```


## Search

```
% curl -XPOST -H 'Content-type: text/plain' http://localhost:8000/v1/indexes/wikipedia_en/_search --data-binary @./examples/search.json | jq .
```


## Delete index

The following command will delete the index `example_en` with the specified name. This command will delete the index file on the object storage and the index metadata on the metastore.

```
% curl -XDELETE http://localhost:8000/v1/indexes/wikipedia_en
```


## Docker container

### Build Docker container image

You can build the Docker container image like so:

```
% make docker-build
```

### Pull Docker container image from docker.io

You can also use the Docker container image already registered in docker.io like so:

```
% docker pull mosuka/phalanx:latest
```

See https://hub.docker.com/r/mosuka/phalanx/tags/

### Start on Docker

Running a Blast data node on Docker. Start Blast node like so:

```bash
% docker run --rm --name phalanx-node1 \
    -p 2000:2000 \
    -p 5000:5000 \
    -p 8000:8000 \
    mosuka/phalanx:latest start \
      --host=0.0.0.0 \
      --bind-port=2000 \
      --grpc-port=5000 \
      --http-port=8000 \
      --roles=indexer,searcher \
      --index-metastore-uri=file:///tmp/phalanx/metadata
```
