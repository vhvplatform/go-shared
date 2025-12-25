package httpclient

import (
	"errors"
	"net/http"
	"sync"
	"time"
)

// CircuitState represents the state of the circuit breaker
type CircuitState int

const (
	// StateClosed means the circuit is closed and requests are allowed
	StateClosed CircuitState = iota
	// StateOpen means the circuit is open and requests are blocked
	StateOpen
	// StateHalfOpen means the circuit is testing if the service has recovered
	StateHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	maxFailures  int
	timeout      time.Duration
	resetTimeout time.Duration
	state        CircuitState
	failures     int
	lastFailTime time.Time
	mu           sync.RWMutex
}

var (
	// ErrCircuitOpen is returned when the circuit breaker is open
	ErrCircuitOpen = errors.New("circuit breaker is open")
)

// NewCircuitBreaker creates a new circuit breaker
// maxFailures: number of consecutive failures before opening the circuit
// timeout: how long to wait before attempting recovery (half-open state)
// resetTimeout: how long to keep the circuit open before moving to half-open
func NewCircuitBreaker(maxFailures int, timeout, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		timeout:      timeout,
		resetTimeout: resetTimeout,
		state:        StateClosed,
	}
}

// Execute runs the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() (*http.Response, error)) (*http.Response, error) {
	cb.mu.Lock()

	// Check if we should transition from open to half-open
	if cb.state == StateOpen {
		if time.Since(cb.lastFailTime) > cb.resetTimeout {
			cb.state = StateHalfOpen
		} else {
			cb.mu.Unlock()
			return nil, ErrCircuitOpen
		}
	}

	cb.mu.Unlock()

	// Execute the function
	resp, err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil || (resp != nil && resp.StatusCode >= 500) {
		// Request failed
		cb.failures++
		cb.lastFailTime = time.Now()

		if cb.state == StateHalfOpen {
			// Failed in half-open state, reopen the circuit
			cb.state = StateOpen
		} else if cb.failures >= cb.maxFailures {
			// Too many failures, open the circuit
			cb.state = StateOpen
		}

		return resp, err
	}

	// Request succeeded
	if cb.state == StateHalfOpen {
		// Success in half-open state, close the circuit
		cb.state = StateClosed
		cb.failures = 0
	} else if cb.state == StateClosed {
		// Reset failure count on success
		cb.failures = 0
	}

	return resp, nil
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetFailures returns the current number of failures
func (cb *CircuitBreaker) GetFailures() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failures
}

// Reset manually resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = StateClosed
	cb.failures = 0
}
