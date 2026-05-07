package jsonrpc

import (
	"encoding/json"
	"errors"
	"unicode"
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
		return errInvalidVersion
	}

	if r.isMethodEmpty() {
		return errEmptyMethod
	}

	return nil
}

type batchRequest []request

func (r batchRequest) isNotification() bool {
	for _, req := range r {
		if !req.isNotification() {
			return false
		}
	}

	return true
}

func (r batchRequest) validate() error {
	var err error

	for _, req := range r {
		err = errors.Join(err, req.validate())
	}

	return err
}

func IsBatchRequest(in json.RawMessage) bool {
	for _, r := range []rune(string(in)) {
		switch {
		case unicode.IsSpace(r):
			continue
		case r == '[':
			return true
		default:
			return false
		}
	}

	return false
}
