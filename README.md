# Phalanx

Phalanx is a cloud-native distributed search engine written in [Go](https://golang.org) built on top of [Bluge](https://github.com/blugelabs/bluge) that provides endpoints through [gRPC](https://grpc.io/) and traditional RESTful API.  
Phalanx implements a cluster formation by [hashicorp/memberlist](https://github.com/hashicorp/memberlist) and managing index metadata on [etcd](https://etcd.io/), so it is easy to bring up a fault-tolerant cluster.  
Metrics for system operation can also be output in Prometheus exposition format, so that monitoring can be done immediately using [Prometheus](https://prometheus.io/).  
Phalanx is using object storage for the storage layer, it is only responsible for the computation layer, such as indexing and retrieval processes. Therefore, scaling is easy, and you can simply add new nodes to the cluster.  
Currently, it is an alpha version and only supports [MinIO](https://min.io/) as the storage layer, but in the future it will support [Amazon S3](https://aws.amazon.com/s3/), [Google Cloud Storage](https://cloud.google.com/storage), and [Azure Blob Storage](https://azure.microsoft.com/en-us/services/storage/blobs/).  

![phalanx_architecture](./docs_md/architecture.png)


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

You can run a Phalanx node on Docker as follows:

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
