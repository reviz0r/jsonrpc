package jsonrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"sync"
)

type Client struct {
	conn        io.ReadWriteCloser
	idGenerator RequestIDGenerator

	m  sync.Mutex
	ch map[ID]chan json.RawMessage
}

func NewClient(conn io.ReadWriteCloser, ig RequestIDGenerator) *Client {
	cl := &Client{
		conn:        conn,
		idGenerator: ig,
		ch:          make(map[ID]chan json.RawMessage),
	}
	go cl.readResponses()
	return cl
}

func (c *Client) createChan(id ID) chan json.RawMessage {
	c.m.Lock()
	defer c.m.Unlock()

	ch := make(chan json.RawMessage, 1)
	c.ch[id] = ch
	return ch
}

func (c *Client) getChan(id ID) chan json.RawMessage {
	return c.ch[id]
}

func (c *Client) dropChan(id ID) chan json.RawMessage {
	c.m.Lock()
	defer c.m.Unlock()

	ch := c.ch[id]
	delete(c.ch, id)
	return ch
}

func (c *Client) Close() error {
	err := c.conn.Close()

	for id := range c.ch {
		ch := c.dropChan(id)
		close(ch)
	}

	if err != nil {
		return fmt.Errorf("jsonrpc: close client conn: %w", err)
	}

	return nil
}

func (c *Client) readResponses() {
	defer c.Close()

	for {
		msg, err := io.ReadAll(c.conn)
		if err != nil {
			slog.Warn("jsonrpc: read response", "error", err.Error())

			return
		}

		var response response

		err = json.Unmarshal(msg, &response)
		if err != nil {
			slog.Warn("jsonrpc: unmarshal response", "error", err.Error())

			continue
		}

		if response.ID == nil {
			continue
		}

		ch := c.getChan(*response.ID)
		if ch != nil {
			ch <- msg
			close(ch)
		}
	}
}

func Call[R, P any](c *Client, ctx context.Context, methodName string, params P) (result R, err error) {
	rawParams, err := json.Marshal(params)
	if err != nil {
		return result, fmt.Errorf("jsonrpc: marshal params: %w", err)
	}

	requestID := c.idGenerator.Generate()
	request := request{ID: &requestID, Jsonrpc: jsonrpcVersion, Method: methodName, Params: rawParams}

	ch := c.createChan(requestID)
	defer c.dropChan(requestID) // убираем канал из ожидания ответа

	err = json.NewEncoder(c.conn).Encode(request)
	if err != nil {
		return result, fmt.Errorf("jsonrpc: marshal request to conn: %w", err)
	}

	select {
	case rawResponse, isSuccess := <-ch:
		if !isSuccess {
			return result, ErrClientClosed
		}

		var response response
		err := json.Unmarshal(rawResponse, &response)
		if err != nil {
			return result, fmt.Errorf("jsonrpc: unmarshal response: %w", err)
		}

		if response.Error != nil {
			return result, fmt.Errorf("jsonrpc: response error: %w", response.Error)
		}

		err = json.Unmarshal(response.Result, &result)
		if err != nil {
			return result, fmt.Errorf("jsonrpc: unmarshal result: %w", err)
		}

		return result, nil
	case <-ctx.Done():
		return result, ctx.Err()
	}
}
