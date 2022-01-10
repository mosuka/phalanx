# Architecture

Phalanx is a master node-less distributed search engine that separates the computation layer for searching and indexing from the storage layer for persisting the index.
The storage layer is designed to use object storage on public clouds such as [Amazon S3](https://aws.amazon.com/s3/), [Google Cloud Storage](https://cloud.google.com/storage), and [Azure Blob Storage](https://azure.microsoft.com/en-us/services/storage/blobs/).

Phalanx makes it easy to bring up a distributed search engine cluster. A phalanx cluster simply adds nodes when its resources are run out. Of course, it can also simply shut down nodes that are not needed. Indexes are managed by object storage, so there is no need to worry about index placement. No complex operations are required. Clusters are very flexible and scalable.

Phalanx stores index metadata in etcd. The metadata stores the index and the path of the shards under that index. The nodes process the distributed index based on the metadata stored in etcd.

Phalanx also uses etcd as a distributed lock manager to ensure that updates to a single shard are not made on multiple nodes at the same time.

![phalanx_architecture](./architecture.png)
