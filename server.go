package jsonrpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
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

// ServeHTTP implement http.Handler for handling JSON-RPC requests over HTTP
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(contentType, contentTypeJSON)

	if !strings.HasPrefix(r.Header.Get(contentType), contentTypeJSON) {
		err := fmt.Errorf("%s must be %s", contentType, contentTypeJSON)
		sendError(w, false, nil, ErrParseError(err.Error()))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendError(w, false, nil, ErrParseError(err.Error()))
		return
	}
	defer r.Body.Close()

	res, err := s.Handle(r.Context(), body)
	if err != nil {
		var mErr *marshalError
		if errors.As(err, &mErr) {
			sendError(w, false, mErr.ID(), ErrInternalError(err.Error()))
			return
		} else {
			sendError(w, false, nil, ErrInternalError(err.Error()))
			return
		}
	}

	if res == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	n, err := w.Write(res)
	if err != nil {
		if n == 0 {
			sendError(w, false, nil, ErrInternalError(err.Error()))
		} else {
			slog.WarnContext(r.Context(), "jsonrpc: write response", "error", err.Error())
		}
	}
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

// sendError in response
func sendError(w http.ResponseWriter, isNotification bool, id *ID, err *Error) {
	res := response{
		ID:      id,
		Jsonrpc: jsonrpcVersion,
		Error:   err,
	}

	if isNotification {
		w.WriteHeader(http.StatusOK)
	} else {
		encodeErr := json.NewEncoder(w).Encode(res)
		if encodeErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
