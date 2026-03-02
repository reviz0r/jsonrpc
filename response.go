package jsonrpc

import "encoding/json"

// response represents a JSON-RPC response returned by the server
type response struct {
	ID      *ID             `json:"id"`
	Jsonprc string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

func responseWithResult(id *ID, result json.RawMessage) (json.RawMessage, error) {
	return json.Marshal(response{ID: id, Jsonprc: jsonrpcVersion, Result: result})
}

func responseWithError(id *ID, isNotification bool, err *Error) (json.RawMessage, error) {
	if isNotification {
		return nil, nil
	}

	return json.Marshal(response{ID: id, Jsonprc: jsonrpcVersion, Error: err})
}
