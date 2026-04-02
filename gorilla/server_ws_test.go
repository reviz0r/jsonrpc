package gorilla_test

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/reviz0r/jsonrpc"
	"github.com/reviz0r/jsonrpc/gorilla"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

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

	go h.server.Serve(gorilla.NewWsConn(conn))
}

var client *jsonrpc.Client

func TestMain(m *testing.M) {
	jserver := jsonrpc.NewServer()
	jserver.RegisterMethod("subtract_positional", jsonrpc.CreateMethod(SubtractPositional{}))
	jserver.RegisterMethod("subtract_named", jsonrpc.CreateMethod(SubtractNamed{}))
	jserver.RegisterMethod("panic_method", jsonrpc.CreateMethod(PanicMethod{}))

	handler := &wsHandler{server: jserver}
	server := httptest.NewServer(handler)
	defer server.Close()

	wsUrl := strings.ReplaceAll(server.URL, "http", "ws")

	conn, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if err != nil {
		log.Fatalf("dial to websocket server: %s", err.Error())
	}

	jclient := jsonrpc.NewClient(gorilla.NewWsConn(conn), jsonrpc.NewIntGenerator())
	defer jclient.Close()

	client = jclient

	os.Exit(m.Run())
}

func TestServer_Serve_panic(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	a := rand.Int()

	result, err := jsonrpc.Call[int](client, ctx, "panic_method", a)
	require.ErrorIs(t, err, context.DeadlineExceeded)

	assert.Equal(t, 0, result)
}

func TestServer_Serve_notification(t *testing.T) {
	ctx := context.Background()

	a := rand.Int()
	b := rand.Int()

	err := jsonrpc.CallNotify(client, ctx, "subtract_positional", [2]int{a, b})
	require.NoError(t, err)
}

func TestServer_Serve(t *testing.T) {
	ctx := context.Background()

	var group errgroup.Group

	for range 10 {
		group.Go(func() error {
			a := rand.Int()
			b := rand.Int()

			result, err := jsonrpc.Call[int](client, ctx, "subtract_positional", [2]int{a, b})
			require.NoError(t, err)

			assert.Equal(t, a-b, result)

			return nil
		})
	}

	err := group.Wait()
	require.NoError(t, err)
}

func BenchmarkServer_Serve(t *testing.B) {
	ctx := context.Background()

	var group errgroup.Group

	t.ResetTimer()
	for t.Loop() {
		group.Go(func() error {
			a := rand.Int()
			b := rand.Int()

			result, err := jsonrpc.Call[int](client, ctx, "subtract", [2]int{a, b})
			require.NoError(t, err)

			assert.Equal(t, a-b, result)

			return nil
		})
	}

	err := group.Wait()
	require.NoError(t, err)
}
