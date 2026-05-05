package jsonrpc

import (
	"errors"
	"io"
	"sync"

	"github.com/gorilla/websocket"
)

var _ io.ReadWriteCloser = new(WsConn)

type WsConn struct {
	conn *websocket.Conn
	m    sync.Mutex

	isClosed bool
}

func NewWsConn(conn *websocket.Conn) *WsConn {
	return &WsConn{conn: conn}
}

func (c *WsConn) Close() error {
	c.m.Lock()
	defer c.m.Unlock()

	if c.isClosed {
		return nil
	}

	c.isClosed = true
	return c.conn.Close()
}

func (c *WsConn) Read(buf []byte) (int, error) {
	if c.isClosed {
		return 0, io.EOF
	}

	for {
		msgType, msg, err := c.conn.ReadMessage()
		if err != nil {
			return 0, err
		}

		if msgType == websocket.TextMessage {
			return copy(buf, msg), io.EOF
		}
	}
}

func (c *WsConn) Write(buf []byte) (int, error) {
	c.m.Lock()
	defer c.m.Unlock()

	if c.isClosed {
		return 0, errors.New("connection is closed")
	}

	err := c.conn.WriteMessage(websocket.TextMessage, buf)
	return len(buf), err
}
