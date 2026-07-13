// news_metrics.go — Prometheus counters/histograms for the news scrape
// path. Split out of extract_news.go so observability wiring is separable
// from parsing.
package scrape

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// newsMetrics is the news-specific subset of the old scrape.Metrics —
// outcome counter + parse latency. They have no use outside ParseNews.
//
// IMPORTANT (test-isolation constraint): newNewsMetrics registers
// `yfin_scrape_news_total` and `yfin_scrape_news_parse_latency_ms`
// against `prometheus.DefaultRegisterer` via `promauto`, guarded by a
// package-level `sync.Once`. Tests that reset `DefaultRegisterer` (e.g.
// `prometheus.DefaultRegisterer = prometheus.NewRegistry()`) MUST
// invoke `ParseNews` (or otherwise call `newNewsMetrics`) BEFORE the
// reset, or the once-fired registration will continue pointing at the
// old registry and the new one will not have these metrics. Changing
// this contract to accept a `prometheus.Registerer` is deliberately
// deferred — see progress.md `Minor findings ledger` Task 3 entry.
type newsMetrics struct{}

var (
	newsTotal        *prometheus.CounterVec
	newsParseLatency *prometheus.HistogramVec
	newsMetricsOnce  sync.Once
)

func newNewsMetrics() *newsMetrics {
	newsMetricsOnce.Do(func() {
		newsTotal = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "yfin_scrape_news_total",
				Help: "Total number of news parsing operations",
			},
			[]string{"outcome"},
		)
		newsParseLatency = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "yfin_scrape_news_parse_latency_ms",
				Help:    "News parsing latency in milliseconds",
				Buckets: prometheus.ExponentialBuckets(1, 2, 12), // 1ms to ~4s
			},
			[]string{},
		)
	})
	return &newsMetrics{}
}

func (m *newsMetrics) recordNews(outcome string) {
	newsTotal.WithLabelValues(outcome).Inc()
}

func (m *newsMetrics) recordNewsParseLatency(duration time.Duration) {
	newsParseLatency.WithLabelValues().Observe(float64(duration.Milliseconds()))
}
