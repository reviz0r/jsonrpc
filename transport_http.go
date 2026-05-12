package jsonrpc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

type httpTransport struct {
	client   *http.Client
	endpoint string
}

func newHTTPTransport(endpoint string, client *http.Client) *httpTransport {
	if client == nil {
		client = http.DefaultClient
	}

	return &httpTransport{client: client, endpoint: endpoint}
}

func (t *httpTransport) call(ctx context.Context, _ *ID, req []byte) ([]byte, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, t.endpoint, bytes.NewReader(req))
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: create http request: %w", err)
	}

	httpReq.Header.Set(contentType, contentTypeJSON)

	resp, err := t.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: read http response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jsonrpc: http status %d: %s", resp.StatusCode, body)
	}

	return body, nil
}

func (t *httpTransport) notify(ctx context.Context, req []byte) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, t.endpoint, bytes.NewReader(req))
	if err != nil {
		return fmt.Errorf("jsonrpc: create http request: %w", err)
	}

	httpReq.Header.Set(contentType, contentTypeJSON)

	resp, err := t.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("jsonrpc: http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("jsonrpc: http status %d", resp.StatusCode)
	}

	return nil
}

func (t *httpTransport) close() error {
	return nil
}
