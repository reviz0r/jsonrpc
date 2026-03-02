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

type Repo struct {
	handlers map[string]handler
	timeout  time.Duration

	closeCh chan struct{}
}

func New(requestTimeout time.Duration) Repo {
	return Repo{
		handlers: make(map[string]handler),
		timeout:  requestTimeout,
		closeCh:  make(chan struct{}),
	}
}

func (s *Repo) RegisterMethod(name string, handler handler) {
	s.handlers[name] = handler
}

var _ io.Closer = new(Repo)

func (s *Repo) Close() error {
	s.closeCh <- struct{}{}
	close(s.closeCh)
	return nil
}

func (s *Repo) Handle(ctx context.Context, in json.RawMessage) (json.RawMessage, error) {
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

	method, found := s.handlers[req.Method]
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

func (s *Repo) Listen(conn io.ReadWriteCloser) {
	for {
		select {
		case <-s.closeCh:
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
				ctx, cancel := context.WithTimeout(ctx, s.timeout)
				defer cancel()

				result, err := s.Handle(ctx, msg)
				if err != nil {
					slog.WarnContext(ctx, "jsonrpc: hanlde request", "error", err.Error())
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
