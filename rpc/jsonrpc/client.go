package jsonrpc

import (
	"encoding/json"
	"fmt"

	"github.com/go-zoox/fetch"
	"github.com/go-zoox/uuid"
)

// Client is a JSON-RPC client.
type Client[C any] interface {
	Call(method string, params interface{}) (interface{}, error)
}

type client[C any] struct {
	server string
	path   string
}

// NewClient creates a new JSON-RPC client.
func NewClient[C any](server string, path ...string) Client[C] {
	pathX := "/"
	if len(path) > 0 {
		if path[0] != "" {
			pathX = path[0]
		}
	}

	return &client[C]{
		server: server,
		path:   pathX,
	}
}

// Call calls a JSON-RPC method.
func (c *client[C]) Call(method string, params interface{}) (interface{}, error) {
	response, err := fetch.Post(c.server+c.path, &fetch.Config{
		Body: map[string]interface{}{
			"jsonrpc": "2.0",
			"method":  method,
			"params":  params,
			"id":      uuid.V4(),
		},
	})
	if err != nil {
		return nil, err
	}

	var res Response
	err = json.Unmarshal(response.Body, &res)
	if err != nil {
		return nil, err
	}

	if res.JSONRPC != "2.0" {
		return nil, fmt.Errorf("invalid jsonrpc version: %s", res.JSONRPC)
	}

	if res.Error.Code != 0 {
		return nil, fmt.Errorf("[%d] %s", res.Error.Code, res.Error.Message)
	}

	return res.Result, nil
}
