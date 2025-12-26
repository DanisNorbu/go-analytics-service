package monitoring

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path"},
	)

	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	AnomaliesTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "anomalies_total",
			Help: "Total number of detected anomalies (RPS)",
		},
	)

	MetricsProcessedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "metrics_processed_total",
			Help: "Total number of processed metrics",
		},
	)
)

func init() {
	prometheus.MustRegister(HTTPRequestsTotal)
	prometheus.MustRegister(HTTPRequestDuration)
	prometheus.MustRegister(AnomaliesTotal)
	prometheus.MustRegister(MetricsProcessedTotal)
}

// Middleware для измерения RPS и latency.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		HTTPRequestsTotal.WithLabelValues(r.Method, r.URL.Path).Inc()

		next.ServeHTTP(w, r)

		duration := time.Since(start).Seconds()
		HTTPRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
	})
}

// Handler для /metrics.
func Handler() http.Handler {
	return promhttp.Handler()
}
