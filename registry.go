package jsonrpc

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"time"
)

type handler interface {
	handle(ctx context.Context, in json.RawMessage) (json.RawMessage, error)
}

type Registry struct {
	handlers map[string]handler
	timeout  time.Duration

	closeCh chan struct{}
}

func New(requestTimeout time.Duration) Registry {
	return Registry{
		handlers: make(map[string]handler),
		timeout:  requestTimeout,
		closeCh:  make(chan struct{}),
	}
}

func (r *Registry) RegisterMethod(name string, handler handler) {
	r.handlers[name] = handler
}

var _ io.Closer = new(Registry)

func (r *Registry) Close() error {
	r.closeCh <- struct{}{}
	close(r.closeCh)
	return nil
}

func (r *Registry) Handle(ctx context.Context, in json.RawMessage) (json.RawMessage, error) {
	var req request

	err := json.Unmarshal(in, &req)
	if err != nil {
		if _, ok := err.(*json.UnmarshalTypeError); ok {
			return responseWithError(req.ID, false, ErrInvalidRequest(err.Error()))
		} else {
			return responseWithError(req.ID, false, ErrParseError(err.Error()))
		}
	}

	err = req.validate()
	if err != nil {
		return responseWithError(req.ID, req.isNotification(), ErrInvalidRequest(err.Error()))
	}

	method, found := r.handlers[req.Method]
	if !found {
		return responseWithError(req.ID, req.isNotification(), ErrMethodNotFound(nil))
	}

	if req.ID != nil {
		ctx = requestIDToContext(ctx, *req.ID)
	}

	result, methoderr := method.handle(ctx, req.Params)
	if methoderr != nil {
		if err, ok := methoderr.(*Error); ok {
			return responseWithError(req.ID, req.isNotification(), err)
		} else {
			return responseWithError(req.ID, req.isNotification(), ErrInternalError(methoderr.Error()))
		}
	}

	if req.isNotification() {
		return nil, nil
	}

	return responseWithResult(req.ID, result)
}

func (r *Registry) Listen(conn io.ReadWriteCloser) {
	for {
		select {
		case <-r.closeCh:
			err := conn.Close()
			if err != nil {
				slog.Warn("jsonrpc: close connection", "error", err.Error())
			}

			return

		default:
			ctx := context.Background()

			msg, err := io.ReadAll(conn)
			if err != nil {
				slog.WarnContext(ctx, "jsonrpc: read request", "error", err.Error())
				continue
			}

			go func(ctx context.Context) {
				ctx, cancel := context.WithTimeout(ctx, r.timeout)
				defer cancel()

				result, err := r.Handle(ctx, msg)
				if err != nil {
					slog.WarnContext(ctx, "jsonrpc: handle request", "error", err.Error())
					return
				}

				_, err = conn.Write(result)
				if err != nil {
					slog.WarnContext(ctx, "jsonrpc: write response", "error", err.Error())
					return
				}
			}(ctx)
		}
	}
}
