package jsonrpc

import (
	"context"
	"io"
)

// Method jsonrpc
type Method interface {
	Name() string
	Handle(ctx context.Context, params io.Reader, result io.Writer) error
}

type method struct {
	name string
	fn   func(ctx context.Context, params io.Reader, result io.Writer) error
}

func (m *method) Name() string {
	return m.name
}

func (m *method) Handle(ctx context.Context, params io.Reader, result io.Writer) error {
	return m.fn(ctx, params, result)
}

// MethodFunc wrap func to implement interface Method
func MethodFunc(name string, fn func(ctx context.Context, params io.Reader, result io.Writer) error) Method {
	return &method{name, fn}
}
