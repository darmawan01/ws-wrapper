package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 60 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var socketChannels map[string]func(RequestMessage, *Client)

func WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("error upgrading connection")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))

		return
	}

	con := NewClient(conn)
	con.SetCloseHandler(closeHandler(con))

	go requestHandler(con)
	go responseHandler(con)
}

func RegisterChannelHandler(channel string, fn func(RequestMessage, *Client)) error {
	if channel == "" {
		return errors.New("channel can not be an empty string")
	}

	if fn == nil {
		return errors.New("handler should not be nil")
	}

	ch := getChannels()
	if ch[channel] != nil {
		return fmt.Errorf("channel already registered")
	}

	ch[channel] = fn

	return nil
}

func getChannels() map[string]func(RequestMessage, *Client) {
	if socketChannels == nil {
		socketChannels = make(map[string]func(RequestMessage, *Client))
	}

	return socketChannels
}

func requestHandler(c *Client) {
	defer func() {
		log.Println("closing connection")

		c.closeConnection()
	}()

	c.SetReadDeadline(time.Now().Add(pongWait))
	c.SetPongHandler(func(string) error {
		c.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		msgType, payload, err := c.ReadMessage()
		if err != nil {
			log.Println(err)

			c.Send(ResponseMessage{
				Error: &ErrorMessage{
					Message: "error reading message",
					Data:    nil,
					Code:    InvalidRequest,
				},
			})

			return
		}

		if msgType != websocket.TextMessage {
			c.Send(ResponseMessage{
				Error: &ErrorMessage{
					Message: "invalid message type",
					Data:    nil,
					Code:    InvalidRequest,
				},
			})

			return
		}

		switch string(payload) {
		case PING:
			c.Send(ResponseMessage{Result: PONG})
		default:
			msg := RequestMessage{}

			if err := json.Unmarshal(payload, &msg); err != nil {
				log.Println(err)

				c.Send(ResponseMessage{
					Error: &ErrorMessage{
						Message: "error reading message",
						Data:    nil,
						Code:    InvalidJSON,
					},
				})

				return
			}

			if socketChannels[msg.Method] == nil {
				c.Send(ResponseMessage{
					Error: &ErrorMessage{
						Message: "method not found",
						Data:    nil,
						Code:    MethodNotFound,
					},
				})
			} else if msg.ID == nil {
				c.Send(ResponseMessage{
					Error: &ErrorMessage{
						Message: "id should not be empty",
						Data:    nil,
						Code:    InvalidRequest,
					},
				})
			} else {
				c.id = *msg.ID

				go socketChannels[msg.Method](msg, c)
			}
		}

	}
}

func responseHandler(c *Client) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.closeConnection()
	}()

	for {
		select {
		case <-ticker.C:
			c.SetWriteDeadline(time.Now().Add(writeWait))

			if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println(err)

				c.Send(ResponseMessage{
					Error: &ErrorMessage{
						Message: "something went wrong",
						Data:    nil,
						Code:    InternalError,
					},
				})

				return
			}

		case m, ok := <-c.send:
			c.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.WriteMessage(websocket.CloseMessage, []byte{})
			}

			if err := c.WriteJSON(m); err != nil {
				log.Println(err)

				c.Send(ResponseMessage{
					Error: &ErrorMessage{
						Message: "something went wrong",
						Data:    nil,
						Code:    InternalError,
					},
				})
				return
			}
		}
	}
}

func closeHandler(c *Client) func(code int, text string) error {
	return func(code int, text string) error {
		c.closeConnection()
		return nil
	}
}
