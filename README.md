# Websocket Wrapper

Write your Websocket message routing like http route handler.

[Use the JSON RPC Specification](https://www.jsonrpc.org/specification)

# Usage

## Get dependencies

```bash
go get github.com/darmawan01/ws-wrapper
```

## Register websocket wrapper handler

```go
engine := gin.Default()

router := engine.Group("/api")

/* Register websocket route */
router.GET("/ws", func(ctx *gin.Context) {
    /* Websocket wrapper handler */
    ws.WebsocketHandler(ctx.Writer, ctx.Request)
})
```

### Registering your method handler

```go
ws.RegisterChannelHandler("public/health", h.healthCheck)
```

