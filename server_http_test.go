package jsonrpc_test

import (
	"context"
	"errors"
	"math/rand"
	"testing"

	"github.com/reviz0r/jsonrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_ServeHTTP_successful_call(t *testing.T) {
	a := rand.Int()
	b := rand.Int()

	result, err := jsonrpc.CallHTTP[int](httpClient, context.Background(), "subtract_positional", [2]int{a, b})

	require.NoError(t, err)
	assert.Equal(t, a-b, result)
}

func TestServer_ServeHTTP_method_not_found(t *testing.T) {
	_, err := jsonrpc.CallHTTP[int](httpClient, context.Background(), "nonexistent", 0)

	require.Error(t, err)

	var rpcErr *jsonrpc.Error
	require.True(t, errors.As(err, &rpcErr))
	assert.Equal(t, -32601, rpcErr.Code)
}

func TestServer_ServeHTTP_method_error(t *testing.T) {
	_, err := jsonrpc.CallHTTP[int](httpClient, context.Background(), "unimplemented_method", 0)

	require.Error(t, err)

	var rpcErr *jsonrpc.Error
	require.True(t, errors.As(err, &rpcErr))
	assert.Equal(t, -32603, rpcErr.Code)
}

func TestServer_ServeHTTP_wrong_content_type(t *testing.T) {
	got, _ := doHTTPCall(t, "subtract_positional", [2]int{1, 2}, 1, withContentType("text/plain"))

	assert.Equal(t, -32700, got.Error.Code)
}

func TestServer_ServeHTTP_invalid_json(t *testing.T) {
	got, _ := doHTTPCall(t, "", nil, nil, withRawBody([]byte("{invalid}")))

	assert.Equal(t, -32700, got.Error.Code)
}

func TestServer_ServeHTTP_notification(t *testing.T) {
	err := jsonrpc.CallNotifyHTTP(httpClient, context.Background(), "subtract_positional", [2]int{42, 23})

	require.NoError(t, err)
}

func TestServer_ServeHTTP_concurrent(t *testing.T) {
	for range 10 {
		t.Run("", func(t *testing.T) {
			t.Parallel()
			a := rand.Int()
			b := rand.Int()

			result, err := jsonrpc.CallHTTP[int](httpClient, context.Background(), "subtract_positional", [2]int{a, b})

			require.NoError(t, err)
			assert.Equal(t, a-b, result)
		})
	}
}
