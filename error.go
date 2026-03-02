package jsonrpc

import "fmt"

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (e *Error) Error() string {
	if e.Data == nil {
		return e.Message
	}

	switch d := e.Data.(type) {
	case nil:
		return e.Message
	case error:
		return fmt.Sprintf("%s (%v)", e.Message, d.Error())
	case fmt.Stringer:
		return fmt.Sprintf("%s (%v)", e.Message, d.String())
	default:
		return fmt.Sprintf("%s (%v)", e.Message, e.Data)
	}
}
