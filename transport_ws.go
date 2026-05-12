package jsonrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
)

type wsTransport struct {
	conn   Conn
	writeM sync.Mutex
	chanM  sync.RWMutex
	chans  map[ID]chan json.RawMessage
}

func newWsTransport(conn Conn) *wsTransport {
	t := &wsTransport{
		conn:  conn,
		chans: make(map[ID]chan json.RawMessage),
	}
	go t.readLoop()
	return t
}

func (t *wsTransport) readLoop() {
	for {
		msg, err := t.conn.ReadMessage()
		if err != nil {
			slog.Warn("jsonrpc: read response", "error", err.Error())
			t.closeAllChans()
			return
		}

		var resp response
		if err := json.Unmarshal(msg, &resp); err != nil {
			slog.Warn("jsonrpc: unmarshal response", "error", err.Error())
			continue
		}

		if resp.ID == nil {
			continue
		}

		ch := t.popChan(*resp.ID)
		if ch != nil {
			ch <- msg
		}
	}
}

func (t *wsTransport) call(ctx context.Context, id *ID, req []byte) ([]byte, error) {
	ch := make(chan json.RawMessage, 1)
	t.putChan(*id, ch)

	t.writeM.Lock()
	err := t.conn.WriteMessage(req)
	t.writeM.Unlock()

	if err != nil {
		t.popChan(*id)
		return nil, fmt.Errorf("jsonrpc: write request: %w", err)
	}

	select {
	case rawResp, ok := <-ch:
		if !ok {
			return nil, ErrClientClosed
		}
		return rawResp, nil
	case <-ctx.Done():
		t.popChan(*id)
		return nil, ctx.Err()
	}
}

func (t *wsTransport) notify(_ context.Context, req []byte) error {
	t.writeM.Lock()
	defer t.writeM.Unlock()

	return t.conn.WriteMessage(req)
}

func (t *wsTransport) close() error {
	return t.conn.Close()
}

func (t *wsTransport) putChan(id ID, ch chan json.RawMessage) {
	t.chanM.Lock()
	defer t.chanM.Unlock()
	t.chans[id] = ch
}

func (t *wsTransport) popChan(id ID) chan json.RawMessage {
	t.chanM.Lock()
	defer t.chanM.Unlock()
	ch := t.chans[id]
	delete(t.chans, id)
	return ch
}

func (t *wsTransport) closeAllChans() {
	t.chanM.Lock()
	defer t.chanM.Unlock()

	for id, ch := range t.chans {
		close(ch)
		delete(t.chans, id)
	}
}
