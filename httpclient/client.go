package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is a standardized HTTP client with retry and circuit breaker support
type Client struct {
	httpClient *http.Client
	baseURL    string
	headers    map[string]string
	retry      *RetryConfig
	breaker    *CircuitBreaker
}

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxRetries        int
	RetryDelay        time.Duration
	RetryableStatuses []int
}

// NewClient creates a new HTTP client with optional configuration
func NewClient(opts ...Option) *Client {
	client := &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		headers: make(map[string]string),
		retry: &RetryConfig{
			MaxRetries:        3,
			RetryDelay:        time.Second,
			RetryableStatuses: []int{502, 503, 504},
		},
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, path string, result interface{}) error {
	url := c.buildURL(path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	return c.doWithRetry(req, result)
}

// Post performs a POST request
func (c *Client) Post(ctx context.Context, path string, body, result interface{}) error {
	return c.doRequest(ctx, http.MethodPost, path, body, result)
}

// Put performs a PUT request
func (c *Client) Put(ctx context.Context, path string, body, result interface{}) error {
	return c.doRequest(ctx, http.MethodPut, path, body, result)
}

// Delete performs a DELETE request
func (c *Client) Delete(ctx context.Context, path string) error {
	url := c.buildURL(path)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	return c.doWithRetry(req, nil)
}

// Do performs a request with retry logic
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	// Apply headers
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	// Use circuit breaker if configured
	if c.breaker != nil {
		return c.breaker.Execute(func() (*http.Response, error) {
			return c.executeWithRetry(req)
		})
	}

	return c.executeWithRetry(req)
}

// doRequest is a helper for POST/PUT requests
func (c *Client) doRequest(ctx context.Context, method, path string, body, result interface{}) error {
	url := c.buildURL(path)

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.doWithRetry(req, result)
}

// doWithRetry performs a request and parses the response
func (c *Client) doWithRetry(req *http.Request, result interface{}) error {
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response if result is provided
	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// executeWithRetry executes the request with retry logic
func (c *Client) executeWithRetry(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= c.retry.MaxRetries; attempt++ {
		// Clone request for retry (body needs to be reset)
		reqClone := req.Clone(req.Context())
		if req.Body != nil {
			// Read original body
			bodyBytes, readErr := io.ReadAll(req.Body)
			if readErr != nil {
				return nil, fmt.Errorf("failed to read request body: %w", readErr)
			}
			// Reset original body
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			// Set clone body
			reqClone.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		// Apply headers
		for k, v := range c.headers {
			reqClone.Header.Set(k, v)
		}

		resp, err = c.httpClient.Do(reqClone)
		if err != nil {
			// Retry on network errors
			if attempt < c.retry.MaxRetries {
				time.Sleep(c.retry.RetryDelay * time.Duration(attempt+1))
				continue
			}
			return nil, fmt.Errorf("request failed after %d attempts: %w", attempt+1, err)
		}

		// Check if status code is retryable
		if c.isRetryableStatus(resp.StatusCode) && attempt < c.retry.MaxRetries {
			resp.Body.Close()
			time.Sleep(c.retry.RetryDelay * time.Duration(attempt+1))
			continue
		}

		// Success or non-retryable error
		return resp, nil
	}

	return resp, err
}

// isRetryableStatus checks if a status code is retryable
func (c *Client) isRetryableStatus(statusCode int) bool {
	for _, s := range c.retry.RetryableStatuses {
		if s == statusCode {
			return true
		}
	}
	return false
}

// buildURL constructs the full URL
func (c *Client) buildURL(path string) string {
	if c.baseURL == "" {
		return path
	}

	// Ensure baseURL doesn't end with / and path starts with /
	baseURL := c.baseURL
	if len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}

	if len(path) == 0 || path[0] != '/' {
		path = "/" + path
	}

	return baseURL + path
}
