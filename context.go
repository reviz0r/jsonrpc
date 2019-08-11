package jsonrpc

import "context"

type contextKey string

const requestID contextKey = "request_id"

// RequestID Получение id запроса из контекста
func RequestID(ctx context.Context) string {
	raw := ctx.Value(requestID)
	value, ok := raw.(*id)
	if !ok {
		return ""
	}
	return value.String()
}
