package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Circuit Breaker Implementation
// States: Closed → Open → Half-Open → Closed

type State int

const (
	StateClosed   State = iota // normal operation
	StateOpen                  // rejecting requests
	StateHalfOpen              // testing with single request
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF-OPEN"
	default:
		return "UNKNOWN"
	}
}

var ErrCircuitOpen = errors.New("circuit breaker is open")

type CircuitBreaker struct {
	mu sync.Mutex

	// Configuration
	failureThreshold int           // failures before opening
	successThreshold int           // successes in half-open before closing
	timeout          time.Duration // time before half-open

	// State
	state           State
	failureCount    int
	successCount    int
	lastFailureTime time.Time
	lastStateChange time.Time
}

type CircuitBreakerConfig struct {
	FailureThreshold int
	SuccessThreshold int
	Timeout          time.Duration
}

func NewCircuitBreaker(cfg CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		failureThreshold: cfg.FailureThreshold,
		successThreshold: cfg.SuccessThreshold,
		timeout:          cfg.Timeout,
		state:            StateClosed,
		lastStateChange:  time.Now(),
	}
}

// Execute runs the given function through the circuit breaker
func (cb *CircuitBreaker) Execute(fn func() error) error {
	if !cb.allowRequest() {
		return ErrCircuitOpen
	}

	err := fn()

	cb.recordResult(err)
	return err
}

func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true

	case StateOpen:
		// Check if timeout has passed → transition to half-open
		if time.Since(cb.lastStateChange) > cb.timeout {
			cb.setState(StateHalfOpen)
			return true
		}
		return false

	case StateHalfOpen:
		// Allow single request for testing
		return true

	default:
		return false
	}
}

func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.onFailure()
	} else {
		cb.onSuccess()
	}
}

func (cb *CircuitBreaker) onFailure() {
	cb.failureCount++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failureCount >= cb.failureThreshold {
			cb.setState(StateOpen)
		}
	case StateHalfOpen:
		// Single failure in half-open → back to open
		cb.setState(StateOpen)
	}
}

func (cb *CircuitBreaker) onSuccess() {
	switch cb.state {
	case StateClosed:
		cb.failureCount = 0 // reset on success
	case StateHalfOpen:
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			cb.setState(StateClosed)
		}
	}
}

func (cb *CircuitBreaker) setState(state State) {
	if cb.state == state {
		return
	}
	prev := cb.state
	cb.state = state
	cb.lastStateChange = time.Now()
	cb.failureCount = 0
	cb.successCount = 0

	fmt.Printf("  Circuit breaker: %s → %s\n", prev, state)
}

func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// --- Demo ---

func main() {
	fmt.Println("=== Circuit Breaker Demo ===\n")

	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 3,               // open after 3 failures
		SuccessThreshold: 2,               // close after 2 successes in half-open
		Timeout:          1 * time.Second, // try half-open after 1s
	})

	// Simulate external service calls
	callCount := 0
	externalService := func() error {
		callCount++
		if callCount <= 5 {
			return errors.New("service unavailable")
		}
		return nil // recovers after 5 calls
	}

	fmt.Println("--- Phase 1: Failures (circuit opens) ---")
	for i := 0; i < 5; i++ {
		err := cb.Execute(externalService)
		if err != nil {
			fmt.Printf("  Call %d: error=%v, state=%s\n", i+1, err, cb.State())
		}
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("\n--- Phase 2: Wait for timeout ---")
	fmt.Printf("  Waiting %v for circuit to try half-open...\n", 1*time.Second)
	time.Sleep(1100 * time.Millisecond)

	fmt.Println("\n--- Phase 3: Recovery (half-open → closed) ---")
	for i := 0; i < 5; i++ {
		err := cb.Execute(externalService)
		if err != nil {
			fmt.Printf("  Call: error=%v, state=%s\n", err, cb.State())
		} else {
			fmt.Printf("  Call: success! state=%s\n", cb.State())
		}
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("\n  Final state: %s\n", cb.State())

	fmt.Println(`
Circuit Breaker Key Points:
• Prevents cascading failures
• Gives failing service time to recover
• Fast-fails requests (better UX than timeout)
• Metrics: track state changes, rejected requests
• Libraries: sony/gobreaker, afex/hystrix-go
`)
}
