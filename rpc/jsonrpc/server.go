package jsonrpc

import (
	"encoding/json"

	"github.com/go-zoox/logger"
)

// Server is a JSON-RPC server.
type Server[C any] interface {
	Register(method string, handler func(ctx C, params interface{}) (interface{}, error))
	Invoke(ctx C, body []byte) ([]byte, error)
}

type server[C any] struct {
	methods map[string]func(ctx C, params interface{}) (interface{}, error)
}

// NewServer creates a new JSON-RPC server.
func NewServer[C any]() Server[C] {
	return &server[C]{
		methods: make(map[string]func(ctx C, params interface{}) (interface{}, error)),
	}
}

func (s *server[C]) Register(method string, handler func(ctx C, params interface{}) (interface{}, error)) {
	s.methods[method] = handler
}

func (s *server[C]) Invoke(ctx C, body []byte) ([]byte, error) {
	response := &Response{
		JSONRPC: "2.0",
	}

	var request Request
	err := json.Unmarshal(body, &request)
	if err != nil {
		logger.Info("jsonrpc: invalid request: %s(%s)", err, string(body))

		response.Error = &Error{
			Code:    -32700,
			Message: "Parse error",
		}

		return json.Marshal(response)
	}

	if request.JSONRPC != "2.0" {
		response.Error = &Error{
			Code:    -32600,
			Message: "Invalid Request (invlid JSON-RPC version)",
		}
		return json.Marshal(response)
	}

	if request.Method == "" {
		response.Error = &Error{
			Code:    -32600,
			Message: "Invalid Request (method is required)",
		}
		return json.Marshal(response)
	}

	if request.ID == "" {
		response.Error = &Error{
			Code:    -32600,
			Message: "Invalid Request (id is required)",
		}
		return json.Marshal(response)
	}

	// fmt.Println("request.ID", request.ID)
	// fmt.Println("request.Method", request.Method)
	// fmt.Println("request.Params", request.Params)

	response.ID = request.ID

	handler, ok := s.methods[request.Method]
	if !ok {
		response.Error = &Error{
			Code:    -32601,
			Message: "Method not found",
		}

		return json.Marshal(response)
	}

	result, err := handler(ctx, request.Params)
	if err != nil {
		response.Error = &Error{
			Code:    -32603,
			Message: err.Error(),
		}

		return json.Marshal(response)
	}

	response.Result = result
	return json.Marshal(response)
}
