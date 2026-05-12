package jsonrpc

import "context"

type transport interface {
	call(ctx context.Context, id *ID, req []byte) ([]byte, error)
	notify(ctx context.Context, req []byte) error
	close() error
}
