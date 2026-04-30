package jsonrpc

import "errors"

var (
	ErrClientClosed = errors.New("jsonrpc: client closed")

	errInvalidVersion = errors.New("jsonrpc: invalid version")
	errEmptyMethod    = errors.New("jsonrpc: method is empty")
)

// marshalError wraps a json.Marshal error with the request ID,
// so callers can recover the ID when the marshal fails.
type marshalError struct {
	id  *ID
	err error
}

func (e *marshalError) Error() string {
	return e.err.Error()
}

func (e *marshalError) ID() *ID {
	return e.id
}

func (e *marshalError) Unwrap() error {
	return e.err
}
