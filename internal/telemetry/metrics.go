package telemetry

import (
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	once sync.Once

	crawlerDocs = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "crawler_documents_total",
		Help: "Total number of documents successfully crawled.",
	})

	crawlerErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "crawler_errors_total",
		Help: "Total number of crawl or parse errors.",
	})

	indexUpdates = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "index_updates_total",
		Help: "Number of documents ingested into the index.",
	})

	searchRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "search_requests_total",
		Help: "Total search requests processed by the API.",
	}, []string{"status"})

	searchLatency = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "search_latency_seconds",
		Help:    "Latency distribution for search requests.",
		Buckets: prometheus.DefBuckets,
	})
)

// RegisterMetrics registers the Prometheus collectors once per process.
func RegisterMetrics() {
	once.Do(func() {
		prometheus.MustRegister(crawlerDocs, crawlerErrors, indexUpdates, searchRequests, searchLatency)
	})
}

// MetricsHandler returns an HTTP handler exposing Prometheus metrics.
func MetricsHandler() http.Handler {
	RegisterMetrics()
	return promhttp.Handler()
}

// IncCrawlerDocuments increments the crawler document counter.
func IncCrawlerDocuments() {
	RegisterMetrics()
	crawlerDocs.Inc()
}

// IncCrawlerErrors increments the crawler error counter.
func IncCrawlerErrors() {
	RegisterMetrics()
	crawlerErrors.Inc()
}

// IncIndexUpdates increments the index update counter.
func IncIndexUpdates() {
	RegisterMetrics()
	indexUpdates.Inc()
}

// ObserveSearch records request latency and status.
func ObserveSearch(status string, latency time.Duration) {
	RegisterMetrics()
	searchRequests.WithLabelValues(status).Inc()
	searchLatency.Observe(latency.Seconds())
}
