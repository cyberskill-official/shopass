package obs_test

import (
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	
	"shopass/obs"
)

func TestMetrics_HTTPObserve(t *testing.T) {
	// Emit a metric
	obs.HTTPObserve("svc-a", "/v1/track", "200", 42*time.Millisecond)

	// Scrape the metrics handler
	ts := httptest.NewServer(obs.MetricsHandler())
	defer ts.Close()

	res, err := ts.Client().Get(ts.URL)
	require.NoError(t, err)
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	body := string(bodyBytes)

	// In Prometheus text format, we expect something like:
	// http_requests_total{route="/v1/track",service="svc-a",status="200"} 1
	require.Contains(t, body, `http_requests_total`)
	require.Contains(t, body, `route="/v1/track"`)
	require.Contains(t, body, `service="svc-a"`)
	require.Contains(t, body, `status="200"`)
}
