package obs

import (
	"context"
	"io"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

var (
	baseLogger *slog.Logger
)

func init() {
	// Initialize default logger to stdout using JSON handler
	SetOutput(os.Stdout)
}

// SetOutput allows changing the output destination of the logger, useful for tests.
func SetOutput(w io.Writer) {
	baseLogger = slog.New(slog.NewJSONHandler(w, nil))
}

// FromContext creates an slog.Logger that binds trace_id and request_id from the context.
func FromContext(ctx context.Context) *slog.Logger {
	sc := trace.SpanContextFromContext(ctx)
	
	traceID := ""
	if sc.HasTraceID() {
		traceID = sc.TraceID().String()
	}

	reqID := requestIDFrom(ctx)

	return baseLogger.With(
		slog.String("trace_id", traceID),
		slog.String("request_id", reqID),
	)
}

// Info logs an informational message along with the redacted attributes.
func Info(ctx context.Context, msg string, fields ...slog.Attr) {
	logger := FromContext(ctx)
	// We need to convert []slog.Attr to []any for logger.Log
	redacted := redactAttrs(fields)
	args := make([]any, len(redacted))
	for i, v := range redacted {
		args[i] = v
	}
	logger.Log(ctx, slog.LevelInfo, msg, args...)
}

// Error logs an error message along with the redacted attributes.
func Error(ctx context.Context, msg string, fields ...slog.Attr) {
	logger := FromContext(ctx)
	redacted := redactAttrs(fields)
	args := make([]any, len(redacted))
	for i, v := range redacted {
		args[i] = v
	}
	logger.Log(ctx, slog.LevelError, msg, args...)
}

// Request ID context key logic
type reqIDKey struct{}

// requestIDFrom retrieves the request ID from the context if it exists.
func requestIDFrom(ctx context.Context) string {
	if val, ok := ctx.Value(reqIDKey{}).(string); ok {
		return val
	}
	return ""
}

// withRequestID injects a request ID into the context.
func withRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, reqIDKey{}, id)
}
