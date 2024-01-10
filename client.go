package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	*websocket.Conn
	mu sync.Mutex

	send chan ResponseMessage
	id   uint64
}

func NewClient(c *websocket.Conn) *Client {

	conn := &Client{
		Conn: c,
		mu:   sync.Mutex{},
		send: make(chan ResponseMessage),
	}

	return conn
}

func (c *Client) Send(message ResponseMessage) {
	c.mu.Lock()
	defer c.mu.Unlock()

	message.JSONRPC = JSONRPC
	message.ID = c.id

	c.send <- message
}

func (c *Client) closeConnection() {
	c.Close()
}
