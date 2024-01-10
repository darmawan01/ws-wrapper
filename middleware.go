package ws

type WsHandlerFunc func(RequestMessage, *Client)

type WsMiddewareFunc func(RequestMessage, *Client) *ResponseMessage

func MiddlewaresWrapper(handler WsHandlerFunc, middlewares ...WsMiddewareFunc) WsHandlerFunc {
	return func(i RequestMessage, c *Client) {
		for _, middleware := range middlewares {
			if res := middleware(i, c); res != nil {
				c.Send(*res)

				return
			}
		}

		handler(i, c)
	}
}
