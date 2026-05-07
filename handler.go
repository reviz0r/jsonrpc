package jsonrpc

import (
	"context"
	"encoding/json"
	"fmt"
)

var _ handler = handlerImpl[any, any]{}

type handlerImpl[P, R any] struct {
	method Method[P, R]
}

func (h handlerImpl[P, R]) handle(ctx context.Context, in json.RawMessage) (json.RawMessage, error) {
	var req P

	if len(in) != 0 {
		err := json.Unmarshal(in, &req)
		if err != nil {
			return nil, fmt.Errorf("jsonrpc: unmarshal request failed: %w", err)
		}
	}

	res, err := h.method.Call(ctx, req)
	if err != nil {
		return nil, err
	}

	out, err := json.Marshal(res)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: marshal response failed: %w", err)
	}

	return out, nil
}
