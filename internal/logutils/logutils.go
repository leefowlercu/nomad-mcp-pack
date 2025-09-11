package logutils

import (
	"context"
	"log/slog"
	"os"

	"github.com/leefowlercu/nomad-mcp-pack/internal/config"
	"github.com/leefowlercu/nomad-mcp-pack/internal/ctxutils"
)

func ToSlogLevel(l config.LogLevel) slog.Level {
	switch l {
	case config.LogLevelDebug:
		return slog.LevelDebug
	case config.LogLevelInfo:
		return slog.LevelInfo
	case config.LogLevelWarn:
		return slog.LevelWarn
	case config.LogLevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func SetupLogger(cfg *config.Config) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: ToSlogLevel(cfg.LogLevel),
	}

	var handler slog.Handler
	switch cfg.Env {
	case config.EnvDev:
		opts.AddSource = true
		handler = slog.NewTextHandler(os.Stderr, opts)
	case config.EnvNonProd:
		handler = slog.NewJSONHandler(os.Stderr, opts)
	case config.EnvProd:
		handler = slog.NewJSONHandler(os.Stderr, opts)
	default:
		handler = slog.NewJSONHandler(os.Stderr, opts)
	}

	handler = &ContextHandler{Handler: handler}

	return slog.New(handler)
}

type ContextHandler struct {
	Handler slog.Handler
}

var _ slog.Handler = (*ContextHandler)(nil)

// Custom Handle method that groups and automatically logs contextual values
func (h *ContextHandler) Handle(context context.Context, r slog.Record) error {
	var contextValues []any

	if requestID, ok := context.Value(ctxutils.RequestIDKey).(string); ok {
		contextValues = append(contextValues, "request_id", requestID)
	}
	if userID, ok := context.Value(ctxutils.UserIDKey).(string); ok {
		contextValues = append(contextValues, "user_id", userID)
	}
	if traceID, ok := context.Value(ctxutils.TraceIDKey).(string); ok {
		contextValues = append(contextValues, "trace_id", traceID)
	}

	if len(contextValues) > 0 {
		r.AddAttrs(slog.Group("context", contextValues...))
	}

	return h.Handler.Handle(context, r)
}

func (h *ContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.Handler.Enabled(ctx, level)
}

func (h *ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ContextHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *ContextHandler) WithGroup(name string) slog.Handler {
	return &ContextHandler{Handler: h.Handler.WithGroup(name)}
}
