package jsonrpc

import "context"

type contextKey string

const requestID contextKey = "request_id"

// RequestID takes request id from context
func RequestID(ctx context.Context) string {
	raw := ctx.Value(requestID)
	value, ok := raw.(id)
	if !ok {
		return ""
	}
	return value.String()
}

func requestIDToContext(ctx context.Context, reqID id) context.Context {
	return context.WithValue(ctx, requestID, reqID)
}
