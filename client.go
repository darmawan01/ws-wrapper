package ws

import (
	"context"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	*websocket.Conn
	mu sync.Mutex

	send    chan ResponseMessage
	onClose func()
	Context ClientContext
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		Conn: conn,
		mu:   sync.Mutex{},
		send: make(chan ResponseMessage),
		Context: ClientContext{
			Context: context.Background(),
		},
	}
}

func (c *Client) Send(message ResponseMessage) {
	c.mu.Lock()
	defer c.mu.Unlock()

	message.JSONRPC = JSONRPC
	message.ID = c.Context.GetId()

	c.send <- message
}

func (c *Client) closeConnection() {
	log.Printf("Connection closed :: %d", c.Context.GetId())

	if c.onClose != nil {
		c.onClose()
	}

	c.Close()
}

func (c *Client) OnClose(fn func()) {
	c.onClose = fn
}
