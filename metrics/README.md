# Metrics Package

Package `metrics` cung cấp utilities để thu thập và expose application metrics với hỗ trợ Prometheus.

## Tính Năng

- Counter, Gauge, Histogram, Summary metrics
- Vector metrics với labels
- Timer utilities
- Common metrics presets
- Thread-safe
- Tích hợp Prometheus

## Sử Dụng Cơ Bản

```go
import (
    "github.com/vhvcorp/go-shared/metrics"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// Tạo collector
collector := metrics.NewCollector(metrics.CollectorConfig{
    Namespace: "myapp",
    Subsystem: "http",
})

// Tạo counter
requestCounter := collector.Counter(
    "requests_total",
    "Total number of HTTP requests",
)

// Sử dụng
requestCounter.Inc()

// Expose metrics endpoint
http.Handle("/metrics", promhttp.Handler())
```

## Counter Metrics

```go
// Simple counter
counter := collector.Counter("orders_total", "Total orders")
counter.Inc()
counter.Add(5)

// Counter with labels
orderCounter := collector.CounterVec(
    "orders_total",
    "Total orders by status",
    []string{"status", "region"},
)

orderCounter.WithLabelValues("completed", "us-east-1").Inc()
orderCounter.WithLabelValues("pending", "us-west-2").Add(3)
```

## Gauge Metrics

```go
// Simple gauge
activeUsers := collector.Gauge("active_users", "Number of active users")
activeUsers.Set(150)
activeUsers.Inc()
activeUsers.Dec()
activeUsers.Add(10)
activeUsers.Sub(5)

// Gauge with labels
cacheSize := collector.GaugeVec(
    "cache_size_bytes",
    "Cache size in bytes",
    []string{"cache_name"},
)

cacheSize.WithLabelValues("users").Set(1024000)
cacheSize.WithLabelValues("sessions").Set(512000)
```

## Histogram Metrics

```go
// Request duration histogram
requestDuration := collector.Histogram(
    "request_duration_seconds",
    "HTTP request duration",
    metrics.DurationBuckets,
)

requestDuration.Observe(0.25) // 250ms

// Histogram với labels
queryDuration := collector.HistogramVec(
    "db_query_duration_seconds",
    "Database query duration",
    []string{"operation", "table"},
    metrics.DurationBuckets,
)

queryDuration.WithLabelValues("select", "users").Observe(0.015)
queryDuration.WithLabelValues("insert", "orders").Observe(0.032)
```

## Summary Metrics

```go
// Response size summary
responseSize := collector.Summary(
    "response_size_bytes",
    "HTTP response size",
    map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
)

responseSize.Observe(1024)

// Summary với labels
apiLatency := collector.SummaryVec(
    "api_latency_seconds",
    "API latency",
    []string{"endpoint", "method"},
    nil, // Use default objectives
)

apiLatency.WithLabelValues("/users", "GET").Observe(0.123)
```

## Timer Utility

```go
// Đo thời gian operation
histogram := collector.Histogram(
    "operation_duration_seconds",
    "Operation duration",
    metrics.DurationBuckets,
)

func ProcessData() {
    timer := metrics.NewTimer(histogram)
    defer timer.ObserveDuration()
    
    // Your operation here
    time.Sleep(100 * time.Millisecond)
    // Duration được tự động record khi hàm return
}
```

## Common Metrics

```go
// Tạo bộ metrics thường dùng cho HTTP service
commonMetrics := metrics.NewCommonMetrics("myapp")

// Trong HTTP handler
func handler(w http.ResponseWriter, r *http.Request) {
    // Track request
    commonMetrics.RequestsTotal.
        WithLabelValues(r.Method, r.URL.Path, "200").
        Inc()
    
    // Track in-flight requests
    commonMetrics.RequestsInFlight.
        WithLabelValues(r.Method).
        Inc()
    defer commonMetrics.RequestsInFlight.
        WithLabelValues(r.Method).
        Dec()
    
    // Time request
    timer := metrics.NewTimer(
        commonMetrics.RequestDuration.
            WithLabelValues(r.Method, r.URL.Path),
    )
    defer timer.ObserveDuration()
    
    // Process request
    // ...
    
    // Track response size
    commonMetrics.ResponseSizeBytes.
        WithLabelValues(r.Method, r.URL.Path).
        Observe(float64(responseSize))
}
```

## Gin Middleware

```go
func MetricsMiddleware(metrics *metrics.CommonMetrics) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.FullPath()
        method := c.Request.Method
        
        // In-flight
        metrics.RequestsInFlight.WithLabelValues(method).Inc()
        defer metrics.RequestsInFlight.WithLabelValues(method).Dec()
        
        // Process request
        c.Next()
        
        // Record metrics
        duration := time.Since(start).Seconds()
        status := fmt.Sprintf("%d", c.Writer.Status())
        
        metrics.RequestsTotal.
            WithLabelValues(method, path, status).
            Inc()
        
        metrics.RequestDuration.
            WithLabelValues(method, path).
            Observe(duration)
        
        metrics.ResponseSizeBytes.
            WithLabelValues(method, path).
            Observe(float64(c.Writer.Size()))
    }
}

// Sử dụng
router := gin.Default()
commonMetrics := metrics.NewCommonMetrics("myapp")
router.Use(MetricsMiddleware(commonMetrics))
```

## Custom Buckets

```go
// Duration buckets (seconds)
metrics.DurationBuckets // 0.001 to 10 seconds

// Size buckets (bytes)
metrics.SizeBuckets // 100 bytes to 5MB

// Count buckets
metrics.CountBuckets // 1 to 10,000

// Custom buckets
customBuckets := []float64{0.1, 0.5, 1.0, 5.0, 10.0, 30.0, 60.0}
histogram := collector.Histogram(
    "custom_metric",
    "Custom metric with custom buckets",
    customBuckets,
)
```

## Best Practices

1. **Sử dụng labels hiệu quả**: Tránh high cardinality labels
2. **Chọn đúng metric type**: 
   - Counter: cho giá trị tăng dần (requests, errors)
   - Gauge: cho giá trị lên xuống (memory, connections)
   - Histogram: cho phân phối (duration, size)
   - Summary: cho percentiles
3. **Buckets phù hợp**: Chọn buckets phù hợp với phạm vi dữ liệu
4. **Namespace**: Sử dụng namespace để tránh xung đột

## Status

✅ **Production Ready** - Package đã sẵn sàng sử dụng trong production.
