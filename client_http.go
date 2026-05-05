package jsonrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type HTTPClient struct {
	client      *http.Client
	endpoint    string
	idGenerator RequestIDGenerator
}

func NewHTTPClient(endpoint string, ig RequestIDGenerator, httpClient *http.Client) *HTTPClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &HTTPClient{
		client:      httpClient,
		endpoint:    endpoint,
		idGenerator: ig,
	}
}

func (c *HTTPClient) doHTTP(ctx context.Context, req request) (*http.Response, error) {
	rawReq, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(rawReq))
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: create http request: %w", err)
	}

	httpReq.Header.Set(contentType, contentTypeJSON)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: http request: %w", err)
	}

	return resp, nil
}

func CallHTTP[R, P any](c *HTTPClient, ctx context.Context, methodName string, params P) (result R, err error) {
	rawParams, err := json.Marshal(params)
	if err != nil {
		return result, fmt.Errorf("jsonrpc: marshal params: %w", err)
	}

	requestID := c.idGenerator.Generate()
	req := request{ID: &requestID, Jsonrpc: jsonrpcVersion, Method: methodName, Params: rawParams}

	resp, err := c.doHTTP(ctx, req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, fmt.Errorf("jsonrpc: read http response: %w", err)
	}

	var rpcResp response
	err = json.Unmarshal(body, &rpcResp)
	if err != nil {
		if resp.StatusCode != http.StatusOK {
			return result, fmt.Errorf("jsonrpc: http status %d: %s", resp.StatusCode, body)
		}
		return result, fmt.Errorf("jsonrpc: unmarshal response: %w", err)
	}

	if rpcResp.Error != nil {
		return result, fmt.Errorf("jsonrpc: response error: %w", rpcResp.Error)
	}

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("jsonrpc: http status %d: %s", resp.StatusCode, body)
	}

	err = json.Unmarshal(rpcResp.Result, &result)
	if err != nil {
		return result, fmt.Errorf("jsonrpc: unmarshal result: %w", err)
	}

	return result, nil
}

func CallNotifyHTTP[P any](c *HTTPClient, ctx context.Context, methodName string, params P) error {
	rawParams, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("jsonrpc: marshal params: %w", err)
	}

	req := request{ID: nil, Jsonrpc: jsonrpcVersion, Method: methodName, Params: rawParams}

	resp, err := c.doHTTP(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("jsonrpc: http status %d", resp.StatusCode)
	}

	return nil
}
