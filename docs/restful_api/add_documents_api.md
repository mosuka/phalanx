# Add Documents API

This API adds or updates documents.

## Request

```
PUT /v1/indexes/<INDEX_NAME>/documents
```


## Path parameters

- `<INDEX_NAME>`: (Required, string) Name of the index you want to add or update documents.


## Request body

```
<DOCUMENTS>
```

- `<DOCUMENTS>`: (Required, string) Documents to add or update.  
The documents are in JSONL format, as below.  
```json
{"_id":"1", "id":1, "text":"This is an example document 1."}
{"_id":"2", "id":2, "text":"This is an example document 2."}
{"_id":"3", "id":3, "text":"This is an example document 3."}
```
Each document must have an `_id` field representing a unique key.


## Examples

```
% curl -XPUT -H 'Content-type: application/x-ndjson' http://localhost:8000/v1/indexes/example/documents --data-binary '
{"_id":"1", "id":1, "text":"This is an example document 1."}
{"_id":"2", "id":2, "text":"This is an example document 2."}
{"_id":"3", "id":3, "text":"This is an example document 3."}
'
```
