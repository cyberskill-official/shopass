package obs_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"shopass/obs"
)

func TestLog_HasTraceAndRequestID(t *testing.T) {
	// Initialize a tracer provider so trace IDs get generated
	shutdown, err := obs.InitTracer(context.Background(), "test-svc")
	require.NoError(t, err)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		shutdown(ctx)
	}()

	var buf bytes.Buffer
	obs.SetOutput(&buf)

	// We use the middleware to properly set trace and request ID in the context
	srv := httptest.NewServer(obs.HTTP("test-svc")(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			obs.Info(r.Context(), "tracked product")
		})))
	defer srv.Close()

	req, _ := http.NewRequest("GET", srv.URL, nil)
	req.Header.Set("X-Request-Id", "req-123")
	
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var m map[string]any
	err = json.Unmarshal(buf.Bytes(), &m)
	require.NoError(t, err)

	require.NotEmpty(t, m["trace_id"])
	require.Equal(t, "req-123", m["request_id"])
	require.Equal(t, "tracked product", m["msg"])
}
