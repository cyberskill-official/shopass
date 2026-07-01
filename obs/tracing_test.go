package obs_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"shopass/obs"
)

func TestTrace_ContinuesAcrossServices(t *testing.T) {
	// Setup trace propagation and tracer for test
	shutdown, err := obs.InitTracer(context.Background(), "test-tracing")
	require.NoError(t, err)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		shutdown(ctx)
	}()

	var aTrace, bTrace string

	// Service B: extracts traceparent and continues
	srvB := httptest.NewServer(obs.HTTP("svc-b")(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			sc := trace.SpanContextFromContext(r.Context())
			bTrace = sc.TraceID().String()
		})))
	defer srvB.Close()

	// Service A: starts trace, calls Service B
	srvA := httptest.NewServer(obs.HTTP("svc-a")(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			sc := trace.SpanContextFromContext(r.Context())
			aTrace = sc.TraceID().String()

			req, _ := http.NewRequestWithContext(r.Context(), "GET", srvB.URL, nil)
			obs.InjectTraceContext(r.Context(), req) // Inject traceparent

			_, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
		})))
	defer srvA.Close()

	// Initial request to A
	resp, err := http.Get(srvA.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.NotEmpty(t, aTrace)
	require.NotEmpty(t, bTrace)
	require.Equal(t, aTrace, bTrace)
}
