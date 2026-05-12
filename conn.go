package jsonrpc

type Conn interface {
	ReadMessage() ([]byte, error)
	WriteMessage([]byte) error
	Close() error
}
