package utils

import "context"

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	UserIDKey    contextKey = "user_id"
	TraceIDKey   contextKey = "trace_id"
)

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

func GetRequestID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(RequestIDKey).(string)
	return id, ok
}

func GetUserID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(UserIDKey).(string)
	return id, ok
}

func GetTraceID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(TraceIDKey).(string)
	return id, ok
}
