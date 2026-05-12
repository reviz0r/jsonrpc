package jsonrpc

import (
	"io"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

var _ Conn = new(WsConn)

type WsConn struct {
	conn     *websocket.Conn
	writeM   sync.Mutex
	readM    sync.Mutex
	isClosed atomic.Bool
}

func NewWsConn(conn *websocket.Conn) *WsConn {
	return &WsConn{conn: conn}
}

func (c *WsConn) Close() error {
	if c.isClosed.Swap(true) {
		return nil
	}

	return c.conn.Close()
}

func (c *WsConn) ReadMessage() ([]byte, error) {
	c.readM.Lock()
	defer c.readM.Unlock()

	if c.isClosed.Load() {
		return nil, io.EOF
	}

	for {
		msgType, msg, err := c.conn.ReadMessage()
		if err != nil {
			return nil, err
		}

		if msgType == websocket.TextMessage {
			return msg, nil
		}
	}
}

func (c *WsConn) WriteMessage(data []byte) error {
	c.writeM.Lock()
	defer c.writeM.Unlock()

	if c.isClosed.Load() {
		return io.ErrClosedPipe
	}

	return c.conn.WriteMessage(websocket.TextMessage, data)
}
