# Metadata Store

The metadata-store stores the index metadata.  
For more information about index metadata, see the [index metadata section](index_metadata.md).

The metastore supports the following data stores:
- Local file system
- [etcd](https://etcd.io/)


## Local file system

It is possible to create a metastore on the local file system, but this is useful when running Phalanx on a single node.
You need to take care of the backup or replication of the stored metadata yourself.

### URI specification

```
file://<PATH_TO_INDEX_METADATA_DIR>
```

- `<PATH_TO_INDEX_METADATA_DIR>`: (Required, string) Path to the directory where the metadata will be stored. The metadata for each index is created under this directory when the index is created.  

### Example

```
file:///var/lib/phalanx-metastore
```

## etcd

You can create a metastore on etcd. This is recommended if you have a multi-node cluster configuration.

### URI specification

```
etcd://<PATH_TO_INDEX_METADATA_DIR>
```

- `<PATH_TO_INDEX_METADATA_DIR>`: (Required, string) Path to the directory where the index metadata is stored. The metadata for each index is created under this directory when the index is created.  

#### URI parameters

- `endpoints`: (Optional, string) Comma separated list of etcd endpoints. e.g, 192.168.1.12:2379,192.168.1.13:2379,192.168.1.14:2379

### Environment variables

The following parameters can be specified in the environment variable, but the value specified in the URI parameter will take precedence.

- `ETCD_ENDPOINTS`: (Optional, string) Comma separated list of etcd endpoints. e.g, 192.168.1.12:2379,192.168.1.13:2379,192.168.1.14:2379

### Example

```
etcd://phalanx-metastore
```

```
etcd://phalanx-metastore?endpoints=192.168.1.12:2379,192.168.1.13:2379,192.168.1.14:2379
```
