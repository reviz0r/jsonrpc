package jsonrpc

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"sync"
	"time"
)

type handler interface {
	handle(ctx context.Context, in json.RawMessage) (json.RawMessage, error)
}

type Server struct {
	m        sync.Mutex
	handlers map[string]handler
	timeout  time.Duration

	closeCh chan struct{}
}

func NewServer() *Server {
	return &Server{
		handlers: make(map[string]handler),
		closeCh:  make(chan struct{}),
	}
}

func (s *Server) SetTimeout(timeout time.Duration) {
	s.timeout = timeout
}

func (s *Server) RegisterMethod(name string, handler handler) {
	s.m.Lock()
	defer s.m.Unlock()

	s.handlers[name] = handler
}

var _ io.Closer = new(Server)

func (s *Server) Close() error {
	close(s.closeCh)

	return nil
}

func (s *Server) Handle(ctx context.Context, in json.RawMessage) (json.RawMessage, error) {
	var req request

	err := json.Unmarshal(in, &req)
	if err != nil {
		var errUnmarshal *json.UnmarshalTypeError

		if errors.As(err, &errUnmarshal) {
			return responseWithError(req.ID, false, ErrInvalidRequest(err.Error()))
		}

		return responseWithError(req.ID, false, ErrParseError(err.Error()))
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
		var err *Error

		if errors.As(methoderr, &err) {
			return responseWithError(req.ID, req.isNotification(), err)
		}

		return responseWithError(req.ID, req.isNotification(), ErrInternalError(methoderr.Error()))
	}

	if req.isNotification() {
		return nil, nil
	}

	return responseWithResult(req.ID, result)
}

func (s *Server) Serve(conn io.ReadWriteCloser) {
	defer s.Close()

	go func(c io.ReadWriteCloser) {
		<-s.closeCh
		c.Close()
	}(conn)

	for {
		ctx := context.Background()

		msg, err := io.ReadAll(conn)
		if err != nil {
			slog.WarnContext(ctx, "jsonrpc: read request", "error", err.Error())

			return
		}

		go func(ctx context.Context) {
			defer func() {
				if r := recover(); r != nil {
					slog.ErrorContext(ctx, "jsonrpc: handler got panic", "recover_message", r)
				}
			}()

			var cancel context.CancelFunc

			if s.timeout > 0 {
				ctx, cancel = context.WithTimeout(ctx, s.timeout)
			} else {
				ctx, cancel = context.WithCancel(ctx)
			}

			defer cancel()

			result, err := s.Handle(ctx, msg)
			if err != nil {
				slog.WarnContext(ctx, "jsonrpc: handle request", "error", err.Error())

				return
			}

			if result == nil {
				return // No need response for notification
			}

			_, err = conn.Write(result)
			if err != nil {
				slog.WarnContext(ctx, "jsonrpc: write response", "error", err.Error())

				return
			}
		}(ctx)
	}
}
