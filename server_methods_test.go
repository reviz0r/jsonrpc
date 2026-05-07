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

type Sum struct{}

func (Sum) Call(ctx context.Context, req []int) (int, error) {
	var res int

	for _, n := range req {
		res += n
	}

	return res, nil
}

type GetData struct{}

func (GetData) Call(ctx context.Context, req struct{}) ([]any, error) {
	return []any{"hello", 5}, nil
}

type PanicMethod struct{}

func (PanicMethod) Call(ctx context.Context, req int) (int, error) {
	panic("just panic")
}
