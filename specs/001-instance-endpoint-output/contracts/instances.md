# Contracts: Instances Endpoint

This feature relies on the instance `endpoint` JSON field.

## GET /instances

Response shape (relevant subset):

```json
{
  "instances": [
    {
      "resourceId": "...",
      "name": "...",
      "status": "RUNNING",
      "endpoint": "https://example.invalid"
    }
  ]
}
```

## GET /instances/{instanceResourceId}

Response shape (relevant subset):

```json
{
  "instance": {
    "resourceId": "...",
    "name": "...",
    "status": "RUNNING",
    "endpoint": "https://example.invalid"
  }
}
```

## POST /instances

Response shape (relevant subset):

```json
{
  "instance": {
    "resourceId": "...",
    "name": "...",
    "status": "RUNNING",
    "endpoint": "https://example.invalid"
  }
}
```
