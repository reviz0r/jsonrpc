package jsonrpc

import (
	"encoding/json"
	"errors"
	"io"
)

// request represents a JSON-RPC request received by the server
type request struct {
	ID      *id             `json:"id,omitempty"`
	Jsonrpc string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

func (r *request) Read(p []byte) (n int, err error) {
	if r == nil || r.Params == nil {
		return 0, io.EOF
	}

	n = copy(p, r.Params)

	return n, nil
}

// isNotification Уведомление
func (r *request) isNotification() bool {
	return r.ID == nil
}

// isValidVersion Правильная ли версия
func (r *request) isValidVersion() bool {
	return r.Jsonrpc == jsonrpcVersion
}

// isMethodEmpty Пустой ли метод
func (r *request) isMethodEmpty() bool {
	return len(r.Method) == 0
}

// validate Корректный ли запрос
func (r *request) validate() error {
	if !r.isValidVersion() {
		return errors.New("invalid json-rpc version")
	}

	if r.isMethodEmpty() {
		return errors.New("method is empty")
	}

	return nil
}
