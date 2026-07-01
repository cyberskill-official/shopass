package obs

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// httpRequestsTotal tracks the total number of HTTP requests.
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"service", "route", "status"},
	)

	// httpRequestDurationMs tracks the duration of HTTP requests in milliseconds.
	httpRequestDurationMs = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_ms",
			Help:    "Duration of HTTP requests in milliseconds",
			Buckets: []float64{10, 50, 100, 200, 300, 500, 1000, 2000, 5000},
		},
		[]string{"service", "route"},
	)
)

// HTTPObserve records the metrics for an HTTP request.
func HTTPObserve(serviceName, route, status string, duration time.Duration) {
	httpRequestsTotal.WithLabelValues(serviceName, route, status).Inc()
	httpRequestDurationMs.WithLabelValues(serviceName, route).Observe(float64(duration.Milliseconds()))
}

// MetricsHandler returns an http.Handler that exposes the Prometheus metrics.
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
