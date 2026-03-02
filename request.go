package jsonrpc

import (
	"encoding/json"
	"errors"
)

const jsonrpcVersion = "2.0"

// request represents a JSON-RPC request received by the server
type request struct {
	ID      *ID             `json:"id,omitempty"`
	Jsonrpc string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
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
		return errors.New("jsonrpc: invalid version")
	}

	if r.isMethodEmpty() {
		return errors.New("jsonrpc: method is empty")
	}

	return nil
}
