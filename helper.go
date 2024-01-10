package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func SendMessageAndWaitForResponse(conn *websocket.Conn, req any, response chan ResponseMessage) {
	var msg []byte

	val, ok := req.(string)
	if ok {
		msg = []byte(val)
	} else {
		msg, _ = json.Marshal(req)
	}

	err := conn.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		log.Fatal(err)
	}

	timeout := time.After(writeWait)
	for {
		select {
		case <-timeout:
			log.Fatal("timeout waiting for server response")
			return
		default:
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Fatal(err)
			}

			var res ResponseMessage
			err = json.Unmarshal(msg, &res)
			if err != nil {
				log.Fatal(err)
				return
			}

			response <- res
			return
		}
	}
}
