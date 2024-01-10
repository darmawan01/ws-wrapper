package ws

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func createTestServer() *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(WebsocketHandler))
	return server
}

func createTestWebSocketConnection(server *httptest.Server) (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(strings.Replace(server.URL, "http", "ws", 1), nil)
	return conn, err
}

func TestPingPongHandler(t *testing.T) {
	server := createTestServer()

	conn, err := createTestWebSocketConnection(server)
	assert.Nil(t, err, err)

	response := make(chan ResponseMessage)

	go SendMessageAndWaitForResponse(conn, PING, response)

	result := <-response
	assert.Equal(t, PONG, result.Result)

}

func TestRegisterChannel(t *testing.T) {
	server := createTestServer()

	conn, err := createTestWebSocketConnection(server)
	assert.Nil(t, err, err)

	RegisterChannelHandler("handler-channel", func(msg RequestMessage, c *Client) {
		c.Send(ResponseMessage{
			Result: "hello",
		})
	})

	id := uint64(1234)
	req := RequestMessage{
		JSONRPC: JSONRPC,
		Method:  "handler-channel",
		ID:      &id,
	}

	response := make(chan ResponseMessage)
	go SendMessageAndWaitForResponse(conn, req, response)

	result := <-response
	assert.Equal(t, "hello", result.Result)

	req = RequestMessage{
		JSONRPC: JSONRPC,
		Method:  "invalid-handler-channel",
		ID:      &id,
	}

	go SendMessageAndWaitForResponse(conn, req, response)

	result = <-response
	assert.Nil(t, result.Result)
	assert.Equal(t, MethodNotFound, result.Error.Code)
	assert.Equal(t, "method not found", result.Error.Message)

}

func TestRegisterWithMiddlewareChannel(t *testing.T) {
	server := createTestServer()

	conn, err := createTestWebSocketConnection(server)
	assert.Nil(t, err, err)

	middleware := func(msg RequestMessage, c *Client) *ResponseMessage {
		res := ResponseMessage{
			Error: &ErrorMessage{
				Message: "blocked by middleware",
				Data:    nil,
				Code:    InvalidRequest,
			},
		}

		return &res
	}

	handler := func(msg RequestMessage, c *Client) {
		c.Send(ResponseMessage{
			Result: "hello",
		})
	}

	RegisterChannelHandler("handler-private-channel", MiddlewaresWrapper(handler, middleware))

	id := uint64(1234)
	req := RequestMessage{
		JSONRPC: JSONRPC,
		Method:  "handler-private-channel",
		ID:      &id,
	}

	response := make(chan ResponseMessage)
	go SendMessageAndWaitForResponse(conn, req, response)

	result := <-response
	assert.Equal(t, InvalidRequest, result.Error.Code)
	assert.Equal(t, "blocked by middleware", result.Error.Message)
}
