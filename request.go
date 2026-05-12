package jsonrpc

import (
	"bytes"
	"encoding/json"
)

// request represents a JSON-RPC request received by the server
type request struct {
	ID      *ID             `json:"id,omitempty"`
	Jsonrpc string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// isNotification Уведомление
func (r *request) isNotification() bool {
	return r.ID == nil && r.Jsonrpc == jsonrpcVersion
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
		return errInvalidVersion
	}

	if r.isMethodEmpty() {
		return errEmptyMethod
	}

	return nil
}

func IsBatch(in json.RawMessage) bool {
	trimmed := bytes.TrimSpace(in)
	return len(trimmed) > 0 && trimmed[0] == '['
}

func extractID(msg json.RawMessage) *ID {
	var partial struct {
		ID *ID `json:"id"`
	}
	json.Unmarshal(msg, &partial)
	return partial.ID
}
