package obs

import (
	"context"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// HTTP creates a middleware that traces HTTP requests and records metrics.
func HTTP(serviceName string) func(http.Handler) http.Handler {
	tr := otel.Tracer(serviceName)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Extract trace context from the incoming headers (W3C traceparent).
			ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			// Get route pattern for tracing (e.g. if using a router, we might extract the pattern)
			route := routePattern(r)

			ctx, span := tr.Start(ctx, r.Method+" "+route)
			defer span.End()

			// Extract or generate request ID
			reqID := requestIDFromHeader(r)
			ctx = withRequestID(ctx, reqID)

			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r.WithContext(ctx))

			span.SetAttributes(semconv.HTTPResponseStatusCode(sw.status))
			HTTPObserve(serviceName, route, http.StatusText(sw.status), time.Since(start))
		})
	}
}

// statusWriter wraps http.ResponseWriter to capture the status code.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// requestIDFromHeader gets the X-Request-Id header.
func requestIDFromHeader(r *http.Request) string {
	id := r.Header.Get("X-Request-Id")
	if id == "" {
		// In a real app we would generate one here if missing, but gateway usually sets it.
		return "unknown-req-id"
	}
	return id
}

// routePattern attempts to get a generic route pattern rather than full raw path
// to avoid high cardinality in Prometheus metrics.
func routePattern(r *http.Request) string {
	// A simple heuristic for this example. In a real application, you would extract
	// the pattern from the router context (e.g. chi.RouteContext(r.Context()).RoutePattern()).
	// We'll just strip the last segment if it looks like an ID for now, or just return the path.
	// We'll return the raw path for simplicity in tests unless we have a specific pattern.

	// Fast path check to avoid cardinality explosion for common patterns
	path := r.URL.Path
	if strings.HasPrefix(path, "/api/") {
		// Just a dummy simplification
		return "/api/..."
	}
	return path
}

// InjectTraceContext injects the tracing context into an outgoing HTTP request.
// Used when making downstream calls.
func InjectTraceContext(ctx context.Context, req *http.Request) {
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
}
