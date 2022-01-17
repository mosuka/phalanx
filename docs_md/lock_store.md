# Lock Store

The lock-store stores the lock objects.

The lock-store supports the following data stores:
- [etcd](https://etcd.io/)
- [Amazon DynamoDB](https://aws.amazon.com/DynamoDB/)


## etcd

You can create locks on etcd to avoid index write conflicts. This is recommended if you have a multi-node cluster configuration.

### URI specification

```
etcd://<PATH_TO_LOCK_DIR>
```

- `<PATH_TO_LOCK_DIR>`: (Required, string) Path to the directory where the lock will be stored. The lock for each index is created under this directory when the index writer opened the index.

#### URI parameters

- `endpoints`: (Optional, string) Comma separated list of etcd endpoints. e.g, 192.168.1.12:2379,192.168.1.13:2379,192.168.1.14:2379

### Environment variables

The following parameters can be specified in the environment variable, but the value specified in the URI parameter will take precedence.

- `ETCD_ENDPOINTS`: (Optional, string) Comma separated list of etcd endpoints. e.g, 192.168.1.12:2379,192.168.1.13:2379,192.168.1.14:2379

### Example

```
etcd://phalanx-locks
```

```
etcd://phalanx-locks?endpoints=192.168.1.12:2379,192.168.1.13:2379,192.168.1.14:2379
```


## Amazon DynamoDB

You can create locks on Amazon DynamoDB to avoid index write conflicts. This is recommended if you have a multi-node cluster configuration.

### URI specification

```
dynamodb://<TABLE_NAME>/<PREFIX>/<INDEX_NAME>
```

- `<TABLE_NAME>`: (Required, string) Name of the table where the index will be stored.
- `<PREFIX>`: (Optional, string) Prefix of the indexes. You can specify this prefix if necessary.
- `<INDEX_NAME>`: (Required, string) The name of the index. Create a lock using a key with this prefix.

#### URI parameters

- `endpoint_url`: (Optional, string) AWS endpoint URL. e.g. `https://dynamodb.us-west-2.amazonaws.com` / `https://192.168.1.21:9000`
- `profile`: (Optional, string) AWS profile name.
- `access_key_id`: (Optional, string) AWS access key ID.
- `secret_access_key`: (Optional, string) AWS secret access key.
- `session_token`: (Optional, string) AWS session token.
- `region`: (Optional, string) AWS region. e.g. `us-west-2`
- `use_path_style`:  (Optional, boolean) Use AWS path style.

### Environment variables

The following parameters can be specified in the environment variable, but the value specified in the URI parameters will take precedence.

- `AWS_ENDPOINT_URL`: (Optional, string) AWS endpoint URL. e.g. `https://dynamodb.us-west-2.amazonaws.com` / `https://192.168.1.21:9000`
- `AWS_PROFILE`: (Optional, string) AWS profile name.
- `AWS_ACCESS_KEY_ID`: (Optional, string) AWS access key ID.
- `AWS_SECRET_ACCESS_KEY`: (Optional, string) AWS secret access key.
- `AWS_SESSION_TOKEN`: (Optional, string) AWS session token.
- `AWS_DEFAULT_REGION`: (Optional, string) AWS region. e.g. `us-west-2`
- `AWS_USE_PATH_STYLE`: (Optional, boolean) Use AWS path style.

### Example

```
dynamodb://phalanx-locks/wikipedia_en
```

```
dynamodb://phalanx-locks/wikipedia_en?endpoint=192.168.1.20:9000&region=ap-northeast-1
```
