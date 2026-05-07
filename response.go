package jsonrpc

import (
	"encoding/json"
	"log/slog"
)

// response represents a JSON-RPC response returned by the server
type response struct {
	ID      *ID             `json:"id"`
	Jsonrpc string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

func responseWithResult(id *ID, result json.RawMessage) (json.RawMessage, error) {
	data, marshalErr := json.Marshal(response{ID: id, Jsonrpc: jsonrpcVersion, Result: result})
	if marshalErr != nil {
		return nil, &marshalError{id: id, err: marshalErr}
	}
	return data, nil
}

func responseWithError(id *ID, isNotification bool, err *Error) (json.RawMessage, error) {
	if isNotification {
		slog.Warn("jsonrpc: error in notification response", "error", err)
		return nil, nil
	}

	data, marshalErr := json.Marshal(response{ID: id, Jsonrpc: jsonrpcVersion, Error: err})
	if marshalErr != nil {
		return nil, &marshalError{id: id, err: marshalErr}
	}
	return data, nil
}

func batchResponseWithResult(result []json.RawMessage) (json.RawMessage, error) {
	data, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		return nil, &marshalError{err: marshalErr}
	}
	return data, nil
}
