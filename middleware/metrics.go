package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

// MetricsCollector holds Prometheus metrics
type MetricsCollector struct {
	RequestsTotal   *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
	ActiveRequests  prometheus.Gauge
}

// NewMetricsCollector creates a new metrics collector with default metrics
func NewMetricsCollector(namespace string) *MetricsCollector {
	return &MetricsCollector{
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		ActiveRequests: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "active_requests",
				Help:      "Number of active HTTP requests",
			},
		),
	}
}

// Register registers all metrics with Prometheus
func (mc *MetricsCollector) Register() error {
	if err := prometheus.Register(mc.RequestsTotal); err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
			return err
		}
	}
	if err := prometheus.Register(mc.RequestDuration); err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
			return err
		}
	}
	if err := prometheus.Register(mc.ActiveRequests); err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
			return err
		}
	}
	return nil
}

// Metrics creates a middleware that collects Prometheus metrics
func Metrics(collector *MetricsCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Increment active requests
		collector.ActiveRequests.Inc()
		defer collector.ActiveRequests.Dec()

		c.Next()

		// Record metrics after request completes
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}

		collector.RequestsTotal.WithLabelValues(
			c.Request.Method,
			endpoint,
			status,
		).Inc()

		collector.RequestDuration.WithLabelValues(
			c.Request.Method,
			endpoint,
		).Observe(duration)
	}
}

// DefaultMetrics creates a metrics middleware with default collector
// The namespace parameter is used to prefix all metric names
func DefaultMetrics(namespace string) gin.HandlerFunc {
	collector := NewMetricsCollector(namespace)
	collector.Register()
	return Metrics(collector)
}
