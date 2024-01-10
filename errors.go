package ws

const (
	// Parse error	Invalid JSON was received by the server An error occurred on the server while parsing the JSON text.
	InvalidJSON int64 = -32700
	// Invalid Request	The JSON sent is not a valid Request object.
	InvalidRequest int64 = -32600
	// Method not found	The method does not exist / is not available.
	MethodNotFound int64 = -32601
	// Invalid params	Invalid method parameter(s).
	Invalidparams int64 = -32602
	// Internal error	Internal JSON-RPC error.
	InternalError int64 = -32603
	// -32000 to -32099	Server error
	ServerError int64 = -32000
)
