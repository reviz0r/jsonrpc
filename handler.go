package jsonrpc

import (
	"context"
	"encoding/json"
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
			return nil, ErrInvalidParams(err.Error())
		}
	}

	res, err := h.method.Call(ctx, req)
	if err != nil {
		return nil, err
	}

	out, err := json.Marshal(res)
	if err != nil {
		return nil, ErrInternalError(err.Error())
	}

	return out, nil
}
