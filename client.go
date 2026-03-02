package jsonrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"sync"
)

type RequestIDGenerator interface {
	Generate() ID
}

type Client struct {
	conn io.ReadWriteCloser

	m  sync.Mutex
	ch map[ID]chan json.RawMessage

	idGenerator RequestIDGenerator
	closeCh     chan struct{}
}

func NewClient(conn io.ReadWriteCloser, ig RequestIDGenerator) *Client {
	cl := &Client{
		conn:        conn,
		ch:          make(map[ID]chan json.RawMessage),
		idGenerator: ig,
		closeCh:     make(chan struct{}),
	}
	go cl.listenResponses()
	return cl
}

func (c *Client) addChan(id ID, ch chan json.RawMessage) {
	c.m.Lock()
	defer c.m.Unlock()

	c.ch[id] = ch
}

func (c *Client) getChan(id ID) chan json.RawMessage {
	c.m.Lock()
	defer c.m.Unlock()

	ch := c.ch[id]
	delete(c.ch, id)
	return ch
}

func (c *Client) listenResponses() {
	for {
		select {
		case <-c.closeCh:
			return
		default:
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
			ch <- msg
			close(ch)
		}
	}
}

func (c *Client) Close() error {
	c.closeCh <- struct{}{}
	close(c.closeCh)

	err := c.conn.Close()
	if err != nil {
		return fmt.Errorf("jsonrpc: close connection: %w", err)
	}

	return nil
}

func Call[R, P any](c *Client, ctx context.Context, methodName string, params P) (result R, err error) {
	rawParams, err := json.Marshal(params)
	if err != nil {
		return result, fmt.Errorf("jsonrpc: marshal params: %w", err)
	}

	requestID := c.idGenerator.Generate()
	request := request{ID: &requestID, Jsonrpc: jsonrpcVersion, Method: methodName, Params: rawParams}

	rawRequest, err := json.Marshal(request)
	if err != nil {
		return result, fmt.Errorf("jsonrpc: marshal request: %w", err)
	}

	ch := make(chan json.RawMessage)
	c.addChan(requestID, ch)

	_, err = c.conn.Write(rawRequest)
	if err != nil {
		_ = c.getChan(requestID) // убираем канал из ожидания ответа
		return result, fmt.Errorf("jsonrpc: send request: %w", err)
	}

	select {
	case rawResponse := <-ch:
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
