package jsonrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

// Method jsonrpc
type Method = func(ctx context.Context, params io.Reader, result io.Writer) error

// Repo Репозиторий методов
type Repo struct {
	m       sync.RWMutex
	methods map[string]Method
}

// New Новый репозиторий
func New() *Repo {
	return &Repo{methods: make(map[string]Method)}
}

// RegisterMethod Зарегистрировать метод
func (repo *Repo) RegisterMethod(name string, fn Method) {
	repo.m.Lock()
	defer repo.m.Unlock()

	repo.methods[name] = fn
}

// UnregisterMethod Отменить регистрацию метода
func (repo *Repo) UnregisterMethod(name string) {
	repo.m.Lock()
	defer repo.m.Unlock()

	delete(repo.methods, name)
}

// takeMethod Получить метод
func (repo *Repo) takeMethod(methodName string) (Method, bool) {
	repo.m.RLock()
	defer repo.m.RUnlock()

	fn, exist := repo.methods[methodName]
	return fn, exist
}

// ServeHTTP Обработчик http запросов
func (repo *Repo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req request
	var res response

	w.Header().Set(contentType, contentTypeJSON)

	if !strings.HasPrefix(r.Header.Get(contentType), contentTypeJSON) {
		err := fmt.Errorf("%s must be %s", contentType, contentTypeJSON)
		sendError(w, false, req.ID, ErrParseError(err.Error()))
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if _, ok := err.(*json.UnmarshalTypeError); ok {
			sendError(w, false, req.ID, ErrInvalidRequest(err.Error()))
		} else {
			sendError(w, false, req.ID, ErrParseError(err.Error()))
		}
		return
	}
	defer r.Body.Close()

	if err := req.validate(); err != nil {
		sendError(w, req.isNotification(), req.ID, ErrInvalidRequest(err.Error()))
		return
	}

	fn, exist := repo.takeMethod(req.Method)
	if !exist {
		sendError(w, req.isNotification(), req.ID, ErrMethodNotFound(nil))
		return
	}

	res.ID = req.ID
	res.Jsonprc = jsonrpcVersion

	ctx := context.WithValue(r.Context(), requestID, req.ID)

	if fnerr := fn(ctx, &req, &res); fnerr != nil {
		if err, ok := fnerr.(*Error); ok {
			sendError(w, req.isNotification(), req.ID, err)
		} else {
			sendError(w, req.isNotification(), req.ID, ErrInternalError(fnerr.Error()))
		}
		return
	}

	if req.isNotification() {
		w.WriteHeader(http.StatusOK)
	} else {
		if err := json.NewEncoder(w).Encode(&res); err != nil {
			sendError(w, req.isNotification(), req.ID, ErrInternalError(err.Error()))
			return
		}
	}
}

// sendError Отправка ошибки jsonrpc
func sendError(w http.ResponseWriter, isNotification bool, id *id, err *Error) {
	res := response{
		ID:      id,
		Jsonprc: jsonrpcVersion,
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
