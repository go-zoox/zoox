package jsonrpc

// Request is a JSON-RPC request.
type Request struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      string      `json:"id"`
}

// Response is a JSON-RPC response.
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Error   *Error      `json:"error"`
	ID      string      `json:"id"`
}

// Error is a JSON-RPC error.
type Error struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}
