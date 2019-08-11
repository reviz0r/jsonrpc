package jsonrpc

import (
	"encoding/json"
	"io"
)

// response Ответ
type response struct {
	ID      *id             `json:"id"`
	Jsonprc string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

func (r *response) Write(p []byte) (n int, err error) {
	if r == nil {
		return 0, io.EOF
	}

	r.Result = make([]byte, len(p))
	n = copy(r.Result, p)

	return n, nil
}
