package jsonrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	transport   transport
	idGenerator RequestIDGenerator
}

func NewClient(conn Conn, ig RequestIDGenerator) *Client {
	return &Client{
		transport:   newWsTransport(conn),
		idGenerator: ig,
	}
}

func NewHTTPClient(endpoint string, ig RequestIDGenerator, httpClient *http.Client) *Client {
	return &Client{
		transport:   newHTTPTransport(endpoint, httpClient),
		idGenerator: ig,
	}
}

func (c *Client) Close() error {
	return c.transport.close()
}

func Call[R, P any](c *Client, ctx context.Context, methodName string, params P) (result R, err error) {
	rawParams, err := json.Marshal(params)
	if err != nil {
		return result, fmt.Errorf("jsonrpc: marshal params: %w", err)
	}

	requestID := c.idGenerator.Generate()
	req := request{ID: &requestID, Jsonrpc: jsonrpcVersion, Method: methodName, Params: rawParams}

	rawReq, err := json.Marshal(req)
	if err != nil {
		return result, fmt.Errorf("jsonrpc: marshal request: %w", err)
	}

	rawResp, err := c.transport.call(ctx, &requestID, rawReq)
	if err != nil {
		return result, err
	}

	var resp response
	if err := json.Unmarshal(rawResp, &resp); err != nil {
		return result, fmt.Errorf("jsonrpc: unmarshal response: %w", err)
	}

	if resp.Error != nil {
		return result, fmt.Errorf("jsonrpc: response error: %w", resp.Error)
	}

	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return result, fmt.Errorf("jsonrpc: unmarshal result: %w", err)
	}

	return result, nil
}

func CallNotify[P any](c *Client, ctx context.Context, methodName string, params P) error {
	rawParams, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("jsonrpc: marshal params: %w", err)
	}

	req := request{ID: nil, Jsonrpc: jsonrpcVersion, Method: methodName, Params: rawParams}

	rawReq, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("jsonrpc: marshal request: %w", err)
	}

	return c.transport.notify(ctx, rawReq)
}
