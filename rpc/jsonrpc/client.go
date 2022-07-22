package jsonrpc

import (
	"encoding/json"
	"fmt"

	"github.com/go-zoox/fetch"
	"github.com/go-zoox/uuid"
)

type Client[C any] interface {
	Call(method string, params map[string]interface{}) (map[string]interface{}, error)
}

type client[C any] struct {
	server string
	path   string
}

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

func (c *client[C]) Call(method string, params map[string]interface{}) (map[string]interface{}, error) {
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