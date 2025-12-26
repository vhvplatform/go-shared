// Package metrics provides utilities for collecting and exposing application metrics
// with support for Prometheus and other monitoring systems
package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Collector provides methods for collecting various types of metrics
// Note: Collector is thread-safe and can be used concurrently
type Collector struct {
	namespace string
	subsystem string

	// Counters
	counters map[string]prometheus.Counter

	// Gauges
	gauges map[string]prometheus.Gauge

	// Histograms
	histograms map[string]prometheus.Histogram

	// Summaries
	summaries map[string]prometheus.Summary
}

// CollectorConfig contains configuration for metrics collector
type CollectorConfig struct {
	Namespace string // Metrics namespace (e.g., "myapp")
	Subsystem string // Metrics subsystem (e.g., "http")
}

// NewCollector creates a new metrics collector
func NewCollector(config CollectorConfig) *Collector {
	return &Collector{
		namespace:  config.Namespace,
		subsystem:  config.Subsystem,
		counters:   make(map[string]prometheus.Counter),
		gauges:     make(map[string]prometheus.Gauge),
		histograms: make(map[string]prometheus.Histogram),
		summaries:  make(map[string]prometheus.Summary),
	}
}

// Counter creates or retrieves a counter metric
func (c *Collector) Counter(name, help string) prometheus.Counter {
	key := c.makeKey(name)
	if counter, exists := c.counters[key]; exists {
		return counter
	}

	counter := promauto.NewCounter(prometheus.CounterOpts{
		Namespace: c.namespace,
		Subsystem: c.subsystem,
		Name:      name,
		Help:      help,
	})

	c.counters[key] = counter
	return counter
}

// CounterVec creates a counter vector metric
func (c *Collector) CounterVec(name, help string, labels []string) *prometheus.CounterVec {
	return promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: c.namespace,
			Subsystem: c.subsystem,
			Name:      name,
			Help:      help,
		},
		labels,
	)
}

// Gauge creates or retrieves a gauge metric
func (c *Collector) Gauge(name, help string) prometheus.Gauge {
	key := c.makeKey(name)
	if gauge, exists := c.gauges[key]; exists {
		return gauge
	}

	gauge := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: c.namespace,
		Subsystem: c.subsystem,
		Name:      name,
		Help:      help,
	})

	c.gauges[key] = gauge
	return gauge
}

// GaugeVec creates a gauge vector metric
func (c *Collector) GaugeVec(name, help string, labels []string) *prometheus.GaugeVec {
	return promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: c.namespace,
			Subsystem: c.subsystem,
			Name:      name,
			Help:      help,
		},
		labels,
	)
}

// Histogram creates or retrieves a histogram metric
func (c *Collector) Histogram(name, help string, buckets []float64) prometheus.Histogram {
	key := c.makeKey(name)
	if histogram, exists := c.histograms[key]; exists {
		return histogram
	}

	if buckets == nil {
		buckets = prometheus.DefBuckets
	}

	histogram := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: c.namespace,
		Subsystem: c.subsystem,
		Name:      name,
		Help:      help,
		Buckets:   buckets,
	})

	c.histograms[key] = histogram
	return histogram
}

// HistogramVec creates a histogram vector metric
func (c *Collector) HistogramVec(name, help string, labels []string, buckets []float64) *prometheus.HistogramVec {
	if buckets == nil {
		buckets = prometheus.DefBuckets
	}

	return promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: c.namespace,
			Subsystem: c.subsystem,
			Name:      name,
			Help:      help,
			Buckets:   buckets,
		},
		labels,
	)
}

// Summary creates or retrieves a summary metric
func (c *Collector) Summary(name, help string, objectives map[float64]float64) prometheus.Summary {
	key := c.makeKey(name)
	if summary, exists := c.summaries[key]; exists {
		return summary
	}

	if objectives == nil {
		objectives = map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}
	}

	summary := promauto.NewSummary(prometheus.SummaryOpts{
		Namespace:  c.namespace,
		Subsystem:  c.subsystem,
		Name:       name,
		Help:       help,
		Objectives: objectives,
	})

	c.summaries[key] = summary
	return summary
}

// SummaryVec creates a summary vector metric
func (c *Collector) SummaryVec(name, help string, labels []string, objectives map[float64]float64) *prometheus.SummaryVec {
	if objectives == nil {
		objectives = map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}
	}

	return promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  c.namespace,
			Subsystem:  c.subsystem,
			Name:       name,
			Help:       help,
			Objectives: objectives,
		},
		labels,
	)
}

// Timer provides a simple way to time operations
type Timer struct {
	start time.Time
	obs   prometheus.Observer
}

// NewTimer creates a new timer
func NewTimer(obs prometheus.Observer) *Timer {
	return &Timer{
		start: time.Now(),
		obs:   obs,
	}
}

// ObserveDuration records the duration since timer creation
func (t *Timer) ObserveDuration() time.Duration {
	duration := time.Since(t.start)
	if t.obs != nil {
		t.obs.Observe(duration.Seconds())
	}
	return duration
}

// makeKey creates a unique key for a metric
func (c *Collector) makeKey(name string) string {
	return c.namespace + "_" + c.subsystem + "_" + name
}

// Common metric buckets
var (
	// DurationBuckets for measuring request/operation durations in seconds
	DurationBuckets = []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}

	// SizeBuckets for measuring response/data sizes in bytes
	SizeBuckets = []float64{100, 500, 1000, 5000, 10000, 50000, 100000, 500000, 1000000, 5000000}

	// CountBuckets for measuring counts/quantities
	CountBuckets = []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000}
)

// CommonMetrics provides commonly used metrics
type CommonMetrics struct {
	RequestsTotal     *prometheus.CounterVec
	RequestDuration   *prometheus.HistogramVec
	RequestsInFlight  *prometheus.GaugeVec
	ResponseSizeBytes *prometheus.HistogramVec
	ErrorsTotal       *prometheus.CounterVec
}

// NewCommonMetrics creates a set of common metrics
func NewCommonMetrics(namespace string) *CommonMetrics {
	return &CommonMetrics{
		RequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "requests_total",
				Help:      "Total number of requests",
			},
			[]string{"method", "path", "status"},
		),
		RequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "request_duration_seconds",
				Help:      "Request duration in seconds",
				Buckets:   DurationBuckets,
			},
			[]string{"method", "path"},
		),
		RequestsInFlight: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "requests_in_flight",
				Help:      "Number of requests currently being processed",
			},
			[]string{"method"},
		),
		ResponseSizeBytes: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "response_size_bytes",
				Help:      "Response size in bytes",
				Buckets:   SizeBuckets,
			},
			[]string{"method", "path"},
		),
		ErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "errors_total",
				Help:      "Total number of errors",
			},
			[]string{"type", "operation"},
		),
	}
}
