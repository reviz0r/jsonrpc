package jsonrpc_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/reviz0r/jsonrpc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// response for tests
type response struct {
	ID      json.RawMessage `json:"id"`
	Jsonrpc string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   jsonrpc.Error   `json:"error,omitempty"`
}

func TestRegistry_Handle(t *testing.T) {
	testCases := []struct {
		desc           string
		register       func(jsonrpc.Registry)
		req            string
		wantRes        string
		isNotification bool
	}{
		// Test cases take from https://www.jsonrpc.org/specification
		{
			desc: "1. rpc call with positional parameters",
			register: func(r jsonrpc.Registry) {
				r.RegisterMethod("subtract", jsonrpc.CreateMethod(SubtractPositional{}))
			},
			req:     `{"jsonrpc": "2.0", "method": "subtract", "params": [42, 23], "id": 1}`,
			wantRes: `{"jsonrpc": "2.0", "result": 19, "id": 1}`,
		},
		{
			desc: "2. rpc call with positional parameters",
			register: func(r jsonrpc.Registry) {
				r.RegisterMethod("subtract", jsonrpc.CreateMethod(SubtractPositional{}))
			},
			req:     `{"jsonrpc": "2.0", "method": "subtract", "params": [23, 42], "id": 2}`,
			wantRes: `{"jsonrpc": "2.0", "result": -19, "id": 2}`,
		},
		{
			desc: "3. rpc call with named parameters",
			register: func(r jsonrpc.Registry) {
				r.RegisterMethod("subtract", jsonrpc.CreateMethod(SubtractNamed{}))
			},
			req:     `{"jsonrpc": "2.0", "method": "subtract", "params": {"subtrahend": 23, "minuend": 42}, "id": 3}`,
			wantRes: `{"jsonrpc": "2.0", "result": 19, "id": 3}`,
		},
		{
			desc: "4. rpc call with named parameters",
			register: func(r jsonrpc.Registry) {
				r.RegisterMethod("subtract", jsonrpc.CreateMethod(SubtractNamed{}))
			},
			req:     `{"jsonrpc": "2.0", "method": "subtract", "params": {"minuend": 42, "subtrahend": 23}, "id": 4}`,
			wantRes: `{"jsonrpc": "2.0", "result": 19, "id": 4}`,
		},
		{
			desc:           "5. a Notification",
			req:            `{"jsonrpc": "2.0", "method": "update", "params": [1,2,3,4,5]}`,
			isNotification: true,
		},
		{
			desc:           "6. a Notification",
			req:            `{"jsonrpc": "2.0", "method": "foobar"}`,
			isNotification: true,
		},
		{
			desc:    "7. rpc call of non-existent method",
			req:     `{"jsonrpc": "2.0", "method": "foobar", "id": "1"}`,
			wantRes: `{"jsonrpc": "2.0", "error": {"code": -32601, "message": "Method not found"}, "id": "1"}`,
		},
		{
			desc:    "8. rpc call with invalid JSON",
			req:     `{"jsonrpc": "2.0", "method": "foobar, "params": "bar", "baz]`,
			wantRes: `{"jsonrpc": "2.0", "error": {"code": -32700, "message": "Parse error"}, "id": null}`,
		},
		{
			desc:    "9. rpc call with invalid Request object",
			req:     `{"jsonrpc": "2.0", "method": 1, "params": "bar"}`,
			wantRes: `{"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null}`,
		},

		// Custom test cases
		{
			desc:           "rpc call with invalid Request id (null)",
			req:            `{"jsonrpc": "2.0", "method": "foobar", "params": "bar", "id": null}`,
			isNotification: true,
		},
		{
			desc:    "rpc call with invalid Request id (bool)",
			req:     `{"jsonrpc": "2.0", "method": "foobar", "params": "bar", "id": true}`,
			wantRes: `{"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null}`,
		},
		{
			desc:    "rpc call with invalid Request id (float)",
			req:     `{"jsonrpc": "2.0", "method": "foobar", "params": "bar", "id": 1.1}`,
			wantRes: `{"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null}`,
		},
		{
			desc:    "rpc call with invalid Request id (array)",
			req:     `{"jsonrpc": "2.0", "method": "foobar", "params": "bar", "id": ["foo", "bar"]}`,
			wantRes: `{"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null}`,
		},
		{
			desc:    "rpc call with invalid Request id (object)",
			req:     `{"jsonrpc": "2.0", "method": "foobar", "params": "bar", "id": {"foo": "bar"}}`,
			wantRes: `{"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null}`,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			registry := jsonrpc.New()
			registry.SetTimeout(time.Second)

			if tC.register != nil {
				tC.register(registry)
			}

			res, err := registry.Handle(context.Background(), json.RawMessage(tC.req))
			require.NoError(t, err)

			if tC.isNotification {
				assert.Nil(t, res)
				return
			}

			var wantRes, gotRes response

			err = json.Unmarshal(res, &gotRes)
			require.NoError(t, err)

			err = json.Unmarshal([]byte(tC.wantRes), &wantRes)
			require.NoError(t, err)

			if gotRes.Error != (jsonrpc.Error{}) {
				t.Logf("error data: %v", gotRes.Error.Data)
				gotRes.Error.Data = nil // error data is empty in test cases
			}

			assert.Equal(t, wantRes, gotRes)
		})
	}
}
