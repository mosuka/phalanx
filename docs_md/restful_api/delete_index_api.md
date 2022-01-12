# Delete Index API

This API deletes an existing index.

## Request

```
DELETE /v1/indexes/<INDEX_NAME>
```


## Path parameters

- `<INDEX_NAME>`: (Required, string) Name of the index you want to delete.


## Examples

```
% curl -XDELETE http://localhost:8000/v1/indexes/example
```
