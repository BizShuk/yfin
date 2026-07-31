// metrics.go — Prometheus metrics for request, retry, backoff, decode,
// circuit-breaker, latency, and inflight HTTP state.
package obsv

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Counters
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "yfin_requests_total",
			Help: "Total number of requests.",
		},
		[]string{"endpoint", "outcome", "code"},
	)

	retriesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "yfin_retries_total",
			Help: "Total number of retries.",
		},
		[]string{"endpoint", "reason"},
	)

	backoffTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "yfin_backoff_total",
			Help: "Total number of backoffs.",
		},
		[]string{"endpoint", "reason"},
	)

	decodeFailTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "yfin_decode_fail_total",
			Help: "Total number of decode failures.",
		},
		[]string{"reason"},
	)

	cbOpenTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "yfin_cb_open_total",
			Help: "Total number of circuit breaker opens.",
		},
		[]string{"scope"},
	)

	// Gauges
	inflightRequests = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "yfin_inflight_requests",
			Help: "Number of inflight requests.",
		},
		[]string{"endpoint"},
	)

	cbState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "yfin_cb_state",
			Help: "Circuit breaker state (0=closed, 1=half-open, 2=open).",
		},
		[]string{"scope"},
	)

	// Histograms
	requestLatencyMs = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "yfin_request_latency_ms",
			Help:    "Request latency in milliseconds.",
			Buckets: []float64{10, 25, 50, 100, 200, 500, 1000, 2000},
		},
		[]string{"endpoint"},
	)

	backoffSleepMs = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "yfin_backoff_sleep_ms",
			Help:    "Backoff sleep duration in milliseconds.",
			Buckets: []float64{10, 25, 50, 100, 200, 500, 1000, 2000, 5000, 10000},
		},
		[]string{"endpoint"},
	)
)

var (
	metricsRegistered = false
	metricsServer     *http.Server
)

// initMetrics initializes Prometheus metrics and starts the HTTP server for exposition.
func initMetrics(cfg PrometheusConfig) error {
	if !cfg.Enabled {
		return nil
	}

	// Register metrics only once
	if !metricsRegistered {
		prometheus.MustRegister(
			requestsTotal,
			retriesTotal,
			backoffTotal,
			decodeFailTotal,
			cbOpenTotal,
			inflightRequests,
			cbState,
			requestLatencyMs,
			backoffSleepMs,
		)
		metricsRegistered = true
	}

	// Start HTTP server for Prometheus metrics
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	metricsServer = &http.Server{
		Addr:    cfg.Addr,
		Handler: mux,
	}

	go func() {
		Logger().Info("Prometheus metrics exporter started", "addr", cfg.Addr)
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			Logger().Error("Failed to start Prometheus metrics exporter", "error", err, "addr", cfg.Addr)
		}
	}()

	return nil
}

// shutdownMetrics shuts down the metrics server
func shutdownMetrics(ctx context.Context) error {
	if metricsServer != nil {
		return metricsServer.Shutdown(ctx)
	}
	return nil
}

func RecordRequest(endpoint, outcome, code string) {
	if globalObsv == nil || !globalObsv.config.MetricsEnabled {
		return
	}
	requestsTotal.WithLabelValues(endpoint, outcome, code).Inc()
}

func RecordRequestLatency(endpoint string, duration time.Duration) {
	if globalObsv == nil || !globalObsv.config.MetricsEnabled {
		return
	}
	requestLatencyMs.WithLabelValues(endpoint).Observe(float64(duration.Milliseconds()))
}

func RecordRetry(endpoint, reason string) {
	if globalObsv == nil || !globalObsv.config.MetricsEnabled {
		return
	}
	retriesTotal.WithLabelValues(endpoint, reason).Inc()
}

func RecordBackoff(endpoint, reason string) {
	if globalObsv == nil || !globalObsv.config.MetricsEnabled {
		return
	}
	backoffTotal.WithLabelValues(endpoint, reason).Inc()
}

func RecordBackoffSleep(endpoint string, duration time.Duration) {
	if globalObsv == nil || !globalObsv.config.MetricsEnabled {
		return
	}
	backoffSleepMs.WithLabelValues(endpoint).Observe(float64(duration.Milliseconds()))
}

func RecordCBOpen(scope string) {
	if globalObsv == nil || !globalObsv.config.MetricsEnabled {
		return
	}
	cbOpenTotal.WithLabelValues(scope).Inc()
}

func SetCBState(scope string, state int) {
	if globalObsv == nil || !globalObsv.config.MetricsEnabled {
		return
	}
	cbState.WithLabelValues(scope).Set(float64(state))
}

func RecordDecodeFail(reason string) {
	if globalObsv == nil || !globalObsv.config.MetricsEnabled {
		return
	}
	decodeFailTotal.WithLabelValues(reason).Inc()
}

func SetInflightRequests(endpoint string, count int) {
	if globalObsv == nil || !globalObsv.config.MetricsEnabled {
		return
	}
	inflightRequests.WithLabelValues(endpoint).Set(float64(count))
}
