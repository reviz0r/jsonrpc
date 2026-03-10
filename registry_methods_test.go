package jsonrpc_test

import "context"

type SubtractPositional struct{}

func (SubtractPositional) Call(ctx context.Context, req [2]int) (int, error) {
	return req[0] - req[1], nil
}

type SubtractNamed struct{}

func (SubtractNamed) Call(ctx context.Context, req struct{ Minuend, Subtrahend int }) (int, error) {
	return req.Minuend - req.Subtrahend, nil
}
