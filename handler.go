package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var channels map[string]func(RequestMessage, *Client)

func WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Error during connection upgrade :: ", err.Error())

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}

	con := NewClient(conn)
	con.SetCloseHandler(closeHandler(con))

	go readPump(con)
	go writePump(con)
}

func RegisterChannelHandler(channel string, fn func(RequestMessage, *Client)) {
	if channel == "" {
		log.Fatal("channel can not be an empty string")
	}

	if fn == nil {
		log.Fatal("handler should not be nil")
	}

	ch := GetChannels()
	if ch[channel] != nil {
		log.Fatal("channel already registered")
	}

	ch[channel] = fn
}

func GetChannels() map[string]func(RequestMessage, *Client) {
	if channels == nil {
		channels = make(map[string]func(RequestMessage, *Client))
	}

	return channels
}

func readPump(c *Client) {
	defer func() {
		c.closeConnection()
	}()

	c.SetReadLimit(maxMessageSize)
	c.SetReadDeadline(time.Now().Add(pongWait))
	c.SetPongHandler(func(string) error { c.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		msgType, payload, err := c.ReadMessage()
		if err != nil {
			log.Println("Error reading message :: ", err.Error())

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
				log.Println("Error reading message :: ", err.Error())

				c.Send(ResponseMessage{
					Error: &ErrorMessage{
						Message: "error reading message",
						Data:    nil,
						Code:    InvalidJSON,
					},
				})

				return
			}

			if msg.ID == nil {
				c.Send(ResponseMessage{
					Error: &ErrorMessage{
						Message: "id should not be empty",
						Data:    nil,
						Code:    InvalidRequest,
					},
				})
			}

			handler, ok := channels[msg.Method]
			if !ok {
				c.Send(ResponseMessage{
					Error: &ErrorMessage{
						Message: "method not found",
						Data:    nil,
						Code:    MethodNotFound,
					},
				})
			} else {
				go handler(msg, c)
			}

		}

	}
}

func writePump(c *Client) {
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
				log.Println("Error writing message :: ", err.Error())

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
				log.Println("Error writing message :: ", err.Error())
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
	return func(int, string) error {
		c.closeConnection()
		return nil
	}
}
