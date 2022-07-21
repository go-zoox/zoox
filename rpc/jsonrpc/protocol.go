package jsonrpc

type Request struct {
	JSONRPC string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
	ID      string                 `json:"id"`
}

type Response struct {
	JSONRPC string                 `json:"jsonrpc"`
	Result  map[string]interface{} `json:"result"`
	Error   Error                  `json:"error"`
	ID      string                 `json:"id"`
}

type Error struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}
