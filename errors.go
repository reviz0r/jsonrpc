package jsonrpc

import "errors"

var (
	errInvalidVersion = errors.New("jsonrpc: invalid version")
	errEmptyMethod    = errors.New("jsonrpc: method is empty")
)
