package jsonrpc

import (
	"encoding/json"
)

// Server is a JSON-RPC server.
type Server[C any] interface {
	Path() string
	Register(method string, handler func(ctx C, params map[string]interface{}) (map[string]interface{}, error))
	Invoke(ctx C, body []byte) ([]byte, error)
}

type server[C any] struct {
	path    string
	methods map[string]func(ctx C, params map[string]interface{}) (map[string]interface{}, error)
}

// NewServer creates a new JSON-RPC server.
func NewServer[C any](path string) Server[C] {
	return &server[C]{
		path:    path,
		methods: make(map[string]func(ctx C, params map[string]interface{}) (map[string]interface{}, error)),
	}
}

func (s *server[C]) Path() string {
	return s.path
}

func (s *server[C]) Register(method string, handler func(ctx C, params map[string]interface{}) (map[string]interface{}, error)) {
	s.methods[method] = handler
}

func (s *server[C]) Invoke(ctx C, body []byte) ([]byte, error) {
	response := &Response{
		JSONRPC: "2.0",
	}

	var request Request
	err := json.Unmarshal(body, &request)
	if err != nil {
		response.Error = Error{
			Code:    -32700,
			Message: "Parse error",
		}
		return json.Marshal(response)
	}

	if request.JSONRPC != "2.0" {
		response.Error = Error{
			Code:    -32600,
			Message: "Invalid Request",
		}
		return json.Marshal(response)
	}

	if request.Method == "" {
		response.Error = Error{
			Code:    -32600,
			Message: "Invalid Request",
		}
		return json.Marshal(response)
	}

	if request.ID == "" {
		response.Error = Error{
			Code:    -32600,
			Message: "Invalid Request",
		}
		return json.Marshal(response)
	}

	response.ID = request.ID

	handler, ok := s.methods[request.Method]
	if !ok {
		response.Error = Error{
			Code:    -32601,
			Message: "Method not found",
		}

		return json.Marshal(response)
	}

	result, err := handler(ctx, request.Params)
	if err != nil {
		response.Error = Error{
			Code:    -32603,
			Message: err.Error(),
		}

		return json.Marshal(response)
	}

	response.Result = result
	return json.Marshal(response)
}
