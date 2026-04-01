package jsonrpc

import "errors"

var (
	ErrClientClosed = errors.New("jsonrpc: client closed")

	errInvalidVersion = errors.New("jsonrpc: invalid version")
	errEmptyMethod    = errors.New("jsonrpc: method is empty")
)
