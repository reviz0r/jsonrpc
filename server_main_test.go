package jsonrpc_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/reviz0r/jsonrpc"
	"github.com/stretchr/testify/require"
)

var (
	wsClient   *jsonrpc.Client
	httpClient *jsonrpc.HTTPClient
	httpURL    string
)

func TestMain(m *testing.M) {
	jserver := jsonrpc.NewServer()
	jserver.RegisterMethod("unimplemented_method", jsonrpc.CreateMethod(jsonrpc.UnimplementedMethod[int, int]{}))
	jserver.RegisterMethod("subtract_positional", jsonrpc.CreateMethod(SubtractPositional{}))
	jserver.RegisterMethod("subtract_named", jsonrpc.CreateMethod(SubtractNamed{}))
	jserver.RegisterMethod("panic_method", jsonrpc.CreateMethod(PanicMethod{}))

	mux := http.NewServeMux()
	mux.Handle("/rpc", jserver)
	mux.Handle("/ws", &wsHandler{server: jserver})

	server := httptest.NewServer(mux)

	httpURL = server.URL + "/rpc"
	wsURL := strings.ReplaceAll(server.URL, "http", "ws") + "/ws"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		jserver.Close()
		log.Fatalf("dial to websocket server: %s", err.Error())
	}

	wsClient = jsonrpc.NewClient(jsonrpc.NewWsConn(conn), jsonrpc.NewIntGenerator())
	httpClient = jsonrpc.NewHTTPClient(httpURL, jsonrpc.NewIntGenerator(), nil)

	exitCode := m.Run()

	wsClient.Close()
	jserver.Close()

	os.Exit(exitCode)
}

type wsHandler struct {
	upgrader websocket.Upgrader
	server   *jsonrpc.Server
}

func (h *wsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w,
			fmt.Errorf("upgrade websocket connection: %w", err).Error(),
			http.StatusInternalServerError)

		return
	}

	go h.server.Serve(jsonrpc.NewWsConn(conn))
}

type httpCallOption func(*httpCallConfig)

type httpCallConfig struct {
	contentType string
	rawBody     []byte
}

func withContentType(ct string) httpCallOption {
	return func(c *httpCallConfig) {
		c.contentType = ct
	}
}

func withRawBody(body []byte) httpCallOption {
	return func(c *httpCallConfig) {
		c.rawBody = body
	}
}

func doHTTPCall(t *testing.T, methodName string, params any, id any, opts ...httpCallOption) (response, int) {
	t.Helper()

	cfg := httpCallConfig{
		contentType: "application/json",
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	var reqBody []byte
	if cfg.rawBody != nil {
		reqBody = cfg.rawBody
	} else {
		req := map[string]any{
			"jsonrpc": "2.0",
			"method":  methodName,
		}
		if params != nil {
			req["params"] = params
		}
		if id != nil {
			req["id"] = id
		}
		var err error
		reqBody, err = json.Marshal(req)
		require.NoError(t, err)
	}

	httpReq, err := http.NewRequestWithContext(context.Background(), http.MethodPost, httpURL, bytes.NewReader(reqBody))
	require.NoError(t, err)
	httpReq.Header.Set("Content-Type", cfg.contentType)

	resp, err := http.DefaultClient.Do(httpReq)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var got response
	if len(body) > 0 {
		require.NoError(t, json.Unmarshal(body, &got))
	}
	return got, resp.StatusCode
}
