# Delete Documents API

This API deletes documents.

## Request

```
DELETE /v1/indexes/<INDEX_NAME>/documents
```


## Path parameters

- `<INDEX_NAME>`: (Required, string) Name of the index you want to delete documents.


## Request body

```
<IDS>
```

- `<IDS>`: (Required, string) Document IDs to delete.  
The document IDs are in plain text, as below.  
```text
1
2
3
```


## Examples

```
% curl -XDELETE -H 'Content-type: text/plain' http://localhost:8000/v1/indexes/example/documents --data-binary '
1
2
3
'
```
