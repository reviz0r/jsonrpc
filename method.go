package jsonrpc

import (
	"context"
	"errors"
)

type Method[P, R any] interface {
	Call(context.Context, P) (R, error)
}

func CreateMethod[P, R any](method Method[P, R]) handler {
	return handlerImpl[P, R]{method: method}
}

var _ Method[any, any] = UnimplementedMethod[any, any]{}

type UnimplementedMethod[P, R any] struct{}

func (m UnimplementedMethod[P, R]) Call(ctx context.Context, params P) (result R, err error) {
	return result, errors.ErrUnsupported
}

var _ Method[any, any] = FuncMethod[any, any](nil)

type FuncMethod[P, R any] func(context.Context, P) (R, error)

func (m FuncMethod[P, R]) Call(ctx context.Context, params P) (R, error) {
	return m(ctx, params)
}
