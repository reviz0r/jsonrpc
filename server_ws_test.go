package jsonrpc_test

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/reviz0r/jsonrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestServer_Serve_panic(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	a := rand.Int()

	result, err := jsonrpc.Call[int](wsClient, ctx, "panic_method", a)
	require.ErrorIs(t, err, context.DeadlineExceeded)

	assert.Equal(t, 0, result)
}

func TestServer_Serve_method_error(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	a := rand.Int()

	result, err := jsonrpc.Call[int](wsClient, ctx, "unimplemented_method", a)
	require.EqualError(t, err, "jsonrpc: response error: Internal error (unsupported operation)")

	assert.Equal(t, 0, result)

	err = jsonrpc.CallNotify(wsClient, ctx, "unimplemented_method", a)
	require.NoError(t, err)
}

func TestServer_Serve_notification(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	a := rand.Int()
	b := rand.Int()

	err := jsonrpc.CallNotify(wsClient, ctx, "subtract_positional", [2]int{a, b})
	require.NoError(t, err)
}

func TestServer_Serve(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	var group errgroup.Group

	for range 10 {
		group.Go(func() error {
			a := rand.Int()
			b := rand.Int()

			result, err := jsonrpc.Call[int](wsClient, ctx, "subtract_positional", [2]int{a, b})
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

			result, err := jsonrpc.Call[int](wsClient, ctx, "subtract", [2]int{a, b})
			require.NoError(t, err)

			assert.Equal(t, a-b, result)

			return nil
		})
	}

	err := group.Wait()
	require.NoError(t, err)
}
