package ws

import "context"

type ClientContextKey string

const (
	ClientId ClientContextKey = "ClientId"
)

type ClientContext struct {
	Context context.Context
}

func (c *ClientContext) Get(key ClientContextKey) any {
	return c.Context.Value(key)
}

func (c *ClientContext) Set(key ClientContextKey, val any) {
	c.Context = context.WithValue(c.Context, key, val)
}

func (c *ClientContext) GetId() uint64 {
	val, ok := c.Get(ClientId).(uint64)
	if !ok {
		return 0
	}

	return val
}
