package httpclient

import (
	"net/http"
	"time"
)

// Option is a function that configures the Client
type Option func(*Client)

// WithBaseURL sets the base URL for all requests
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithTimeout sets the HTTP client timeout
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithHeader adds a default header to all requests
func WithHeader(key, value string) Option {
	return func(c *Client) {
		c.headers[key] = value
	}
}

// WithHeaders adds multiple default headers to all requests
func WithHeaders(headers map[string]string) Option {
	return func(c *Client) {
		for k, v := range headers {
			c.headers[k] = v
		}
	}
}

// WithRetry configures retry behavior
func WithRetry(maxRetries int, delay time.Duration) Option {
	return func(c *Client) {
		c.retry.MaxRetries = maxRetries
		c.retry.RetryDelay = delay
	}
}

// WithRetryableStatuses sets which HTTP status codes should trigger a retry
func WithRetryableStatuses(statuses []int) Option {
	return func(c *Client) {
		c.retry.RetryableStatuses = statuses
	}
}

// WithCircuitBreaker enables circuit breaker with default settings
func WithCircuitBreaker() Option {
	return func(c *Client) {
		c.breaker = NewCircuitBreaker(5, 10*time.Second, 30*time.Second)
	}
}

// WithCircuitBreakerConfig enables circuit breaker with custom settings
func WithCircuitBreakerConfig(maxFailures int, timeout, resetTimeout time.Duration) Option {
	return func(c *Client) {
		c.breaker = NewCircuitBreaker(maxFailures, timeout, resetTimeout)
	}
}

// WithHTTPClient allows using a custom HTTP client
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithTransport sets a custom HTTP transport
func WithTransport(transport http.RoundTripper) Option {
	return func(c *Client) {
		c.httpClient.Transport = transport
	}
}
