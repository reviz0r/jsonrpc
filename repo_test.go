package jsonrpc

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepoHandler(t *testing.T) {
	testCases := []struct {
		desc           string
		methods        map[string]Method
		req            string
		wantRes        string
		isNotification bool
	}{
		// Test cases take from https://www.jsonrpc.org/specification
		{
			desc:    "1. rpc call with positional parameters",
			methods: map[string]Method{"subtract": SubstractPositional},
			req:     `{"jsonrpc": "2.0", "method": "subtract", "params": [42, 23], "id": 1}`,
			wantRes: `{"jsonrpc": "2.0", "result": 19, "id": 1}`,
		},
		{
			desc:    "2. rpc call with positional parameters",
			methods: map[string]Method{"subtract": SubstractPositional},
			req:     `{"jsonrpc": "2.0", "method": "subtract", "params": [23, 42], "id": 2}`,
			wantRes: `{"jsonrpc": "2.0", "result": -19, "id": 2}`,
		},
		{
			desc:    "3. rpc call with named parameters",
			methods: map[string]Method{"subtract": SubstractNamed},
			req:     `{"jsonrpc": "2.0", "method": "subtract", "params": {"subtrahend": 23, "minuend": 42}, "id": 3}`,
			wantRes: `{"jsonrpc": "2.0", "result": 19, "id": 3}`,
		},
		{
			desc:    "4. rpc call with named parameters",
			methods: map[string]Method{"subtract": SubstractNamed},
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
			wantRes:        `{"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null}`,
			isNotification: true,
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
			// init server
			repo := New()
			for name, method := range tC.methods {
				repo.RegisterMethod(name, method)
			}
			server := httptest.NewServer(repo)
			defer server.Close()

			// make http request
			res, err := http.Post(server.URL, "application/json", bytes.NewBufferString(tC.req))
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, res.StatusCode)
			defer res.Body.Close()

			// decode result
			var got response
			err = json.NewDecoder(res.Body).Decode(&got)
			if tC.isNotification {
				require.EqualError(t, err, "EOF") // no response body for notifications
				return
			}
			require.NoError(t, err)
			if got.Error != nil {
				t.Logf("error data: %v", got.Error.Data)
				got.Error.Data = nil // error data is empty in test cases
			}

			// decode wanted result
			var want response
			err = json.Unmarshal([]byte(tC.wantRes), &want)
			require.NoError(t, err)

			// compare results
			assert.Equal(t, want, got)
		})
	}
}
