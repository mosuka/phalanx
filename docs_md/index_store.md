# Index Store

The index-store stores the index.  

The index-store supports the following storages:
- Local file system
- [MinIO](https://min.io/)
- [Amazon S3](https://aws.amazon.com/s3/)


## Local file system

It is possible to create indexes on the local file system, but this is useful when running Phalanx on a single node.
You need to take care of the backup or replication of the indexes yourself.

### URI specification

```
file://<PATH_TO_INDEX_DIR>/<INDEX_NAME>
```

- `<PATH_TO_INDEX_DIR>`: (Optional, string) Path to the directory of the indexes. You can specify this directory if necessary.
- `<INDEX_NAME>`: (Required, string) Name of the index. Under this directory, the segment files and other files that make up the index are stored.

### Example

```
file:///var/lib/phalanx-indexes/wikipedia_en
```


## MinIO

You can create indexes on MinIO. This is recommended if you have a multi-node cluster configuration.

### URI specification

```
minio://<BUCKET_NAME>/<PATH_TO_INDEX_DIR>/<INDEX_NAME>
```

- `<BUCKET_NAME>`: (Required, string) Name of the bucket where the index will be stored.
- `<PATH_TO_INDEX_DIR>`: (Optional, string) Path to the directory of the indexes. You can specify this directory if necessary.
- `<INDEX_NAME>`: (Required, string) Name of the index. Under this directory, the segment files and other files that make up the index are stored.

#### URI parameters

- `endpoint`: (Optional, string) MinIO endpoint. e.g. `192.168.1.21:9000`
- `access_key`: (Optional, string) MinIO access key.
- `secret_key`: (Optional, string) MinIO secret key.
- `region`: (Optional, string) MinIO region. e.g. `us-east-1`
- `session_token`: (Optional, string) MinIO session token.

### Environment variables

The following parameters can be specified in the environment variable, but the value specified in the URI parameters will take precedence.

- `MINIO_ENDPOINT`: (Optional, string) MinIO endpoint. e.g. `192.168.1.21:9000`
- `AWS_ACCESS_KEY_ID` / `MINIO_ACCESS_KEY`: (Optional, string) MinIO access key.
- `AWS_SECRET_ACCESS_KEY / MINIO_SECRET_KEY`: (Optional, string) MinIO secret key.
- `AWS_DEFAULT_REGION` / `MINIO_REGION`: (Optional, string) MinIO region. e.g. `us-east-1`
- `AWS_SESSION_TOKEN` / `MINIO_SESSION_TOKEN`: (Optional, string) MinIO session token.
- `MINIO_SECURE`: (Optional, boolean) MinIO secure connection.

### Example

```
minio://phalanx-metastore
```

```
etcd://phalanx-metastore?endpoint=192.168.1.20:9000&region=ap-northeast-1
```


## Amazon S3

You can create indexes on Amazon S3. This is recommended if you have a multi-node cluster configuration.

### URI specification

```
s3://<BUCKET_NAME>/<PATH_TO_INDEX_DIR>/<INDEX_NAME>
```

- `<BUCKET_NAME>`: (Required, string) Name of the bucket where the index will be stored.
- `<PATH_TO_INDEX_DIR>`: (Optional, string) Path to the directory of the indexes. You can specify this directory if necessary.
- `<INDEX_NAME>`: (Required, string) Name of the index. Under this directory, the segment files and other files that make up the index are stored.

#### URI parameters

- `endpoint_url`: (Optional, string) AWS endpoint URL. e.g. `https://s3.us-west-2.amazonaws.com` / `https://192.168.1.21:9000`
- `profile`: (Optional, string) AWS profile name.
- `access_key_id`: (Optional, string) AWS access key ID.
- `secret_access_key`: (Optional, string) AWS secret access key.
- `session_token`: (Optional, string) AWS session token.
- `region`: (Optional, string) AWS region. e.g. `us-west-2`
- `use_path_style`:  (Optional, boolean) Use AWS path style.

### Environment variables

The following parameters can be specified in the environment variable, but the value specified in the URI parameters will take precedence.

- `AWS_ENDPOINT_URL`: (Optional, string) AWS endpoint URL. e.g. `https://s3.us-west-2.amazonaws.com` / `https://192.168.1.21:9000`
- `AWS_PROFILE`: (Optional, string) AWS profile name.
- `AWS_ACCESS_KEY_ID`: (Optional, string) AWS access key ID.
- `AWS_SECRET_ACCESS_KEY`: (Optional, string) AWS secret access key.
- `AWS_SESSION_TOKEN`: (Optional, string) AWS session token.
- `AWS_DEFAULT_REGION`: (Optional, string) AWS region. e.g. `us-west-2`
- `AWS_USE_PATH_STYLE`: (Optional, boolean) Use AWS path style.

### Example

```
s3://phalanx-indexes/wikipedia_en
```

```
s3://phalanx-indexes/wikipedia_en?endpoint=192.168.1.20:9000&region=ap-northeast-1
```
