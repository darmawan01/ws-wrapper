package ws

const (
	PING    = "PING"
	PONG    = "PONG"
	JSONRPC = "2.0"
)

type ResponseMessage struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      uint64        `json:"id,omitempty"`
	Method  string        `json:"method,omitempty"`
	Result  any           `json:"result,omitempty"`
	Params  any           `json:"params,omitempty"`
	UsIn    uint64        `json:"usIn,omitempty"`
	UsOut   uint64        `json:"usOut,omitempty"`
	UsDiff  uint64        `json:"usDiff,omitempty"`
	Error   *ErrorMessage `json:"error,omitempty"`
}

type ErrorMessage struct {
	Message string         `json:"message"`
	Data    *ReasonMessage `json:"data"`
	Code    int64          `json:"code"`
}

type ReasonMessage struct {
	Reason string `json:"reason"`
}

type RequestMessage struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	ID      *uint64     `json:"id,omitempty"`
	Params  interface{} `json:"params"`
}
