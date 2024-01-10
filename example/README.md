# Example Request

URL: `ws://localhost:9898/api/ws`

request:

```json
{
    "jsonrpc": "2.0",
    "id": 1234,
    "method": "public/health"
}
```

response:

```json
{
    "jsonrpc":"2.0",
    "id":1234,
    "result":"online"
}
```
