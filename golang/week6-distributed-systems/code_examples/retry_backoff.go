package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"time"
)

// --- Retry with Exponential Backoff + Jitter ---

type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Multiplier  float64
}

var DefaultRetryConfig = RetryConfig{
	MaxAttempts: 5,
	BaseDelay:   100 * time.Millisecond,
	MaxDelay:    30 * time.Second,
	Multiplier:  2.0,
}

// RetryableError indicates the operation can be retried
type RetryableError struct {
	Err error
}

func (e *RetryableError) Error() string { return e.Err.Error() }
func (e *RetryableError) Unwrap() error { return e.Err }

func IsRetryable(err error) bool {
	var retryable *RetryableError
	return errors.As(err, &retryable)
}

// --- Strategy 1: Simple Retry ---

func Retry(ctx context.Context, cfg RetryConfig, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		// Check context before attempt
		if ctx.Err() != nil {
			return ctx.Err()
		}

		lastErr = fn()
		if lastErr == nil {
			return nil // success
		}

		// Check if error is retryable
		if !IsRetryable(lastErr) {
			return lastErr // non-retryable, fail immediately
		}

		// Don't sleep after last attempt
		if attempt < cfg.MaxAttempts-1 {
			delay := calculateDelay(attempt, cfg)
			fmt.Printf("  Attempt %d failed: %v. Retrying in %v...\n",
				attempt+1, lastErr, delay.Round(time.Millisecond))

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return fmt.Errorf("max retries (%d) exceeded: %w", cfg.MaxAttempts, lastErr)
}

// calculateDelay computes exponential backoff with full jitter
func calculateDelay(attempt int, cfg RetryConfig) time.Duration {
	// Exponential: base * multiplier^attempt
	exp := math.Pow(cfg.Multiplier, float64(attempt))
	delay := time.Duration(float64(cfg.BaseDelay) * exp)

	// Cap at max delay
	if delay > cfg.MaxDelay {
		delay = cfg.MaxDelay
	}

	// Full jitter: random between 0 and calculated delay
	jitter := time.Duration(rand.Int63n(int64(delay)))

	return jitter
}

// --- Strategy 2: Retry with Result ---

func RetryWithResult[T any](ctx context.Context, cfg RetryConfig, fn func() (T, error)) (T, error) {
	var lastErr error
	var zero T

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		if ctx.Err() != nil {
			return zero, ctx.Err()
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err
		if !IsRetryable(err) {
			return zero, err
		}

		if attempt < cfg.MaxAttempts-1 {
			delay := calculateDelay(attempt, cfg)
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return zero, ctx.Err()
			}
		}
	}

	return zero, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// --- Strategy 3: HTTP-specific Retry ---

func IsHTTPRetryable(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests, // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return true
	default:
		return false
	}
}

func RetryHTTP(ctx context.Context, cfg RetryConfig, fn func() (*http.Response, error)) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		resp, err := fn()
		if err != nil {
			lastErr = err
			// Network error → retry
			if attempt < cfg.MaxAttempts-1 {
				delay := calculateDelay(attempt, cfg)
				select {
				case <-time.After(delay):
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}
			continue
		}

		// Check if HTTP status is retryable
		if !IsHTTPRetryable(resp.StatusCode) {
			return resp, nil // success or non-retryable error
		}

		// Check Retry-After header
		resp.Body.Close()
		lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)

		if attempt < cfg.MaxAttempts-1 {
			delay := calculateDelay(attempt, cfg)
			// Respect Retry-After if present
			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if d, err := time.ParseDuration(retryAfter + "s"); err == nil {
					delay = d
				}
			}
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// --- Demo ---

func main() {
	fmt.Println("=== Retry with Exponential Backoff Demo ===\n")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Simulate flaky service
	callCount := 0
	flakyService := func() error {
		callCount++
		if callCount < 4 {
			return &RetryableError{Err: fmt.Errorf("connection timeout (attempt %d)", callCount)}
		}
		return nil // succeeds on 4th attempt
	}

	fmt.Println("--- Retrying flaky service ---")
	err := Retry(ctx, DefaultRetryConfig, flakyService)
	if err != nil {
		fmt.Printf("  Final error: %v\n", err)
	} else {
		fmt.Printf("  Success after %d attempts!\n", callCount)
	}

	fmt.Println("\n--- Non-retryable error (fails immediately) ---")
	err = Retry(ctx, DefaultRetryConfig, func() error {
		return errors.New("validation error") // NOT RetryableError
	})
	fmt.Printf("  Error (no retry): %v\n", err)

	fmt.Println("\n--- With timeout context ---")
	shortCtx, shortCancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer shortCancel()

	err = Retry(shortCtx, RetryConfig{
		MaxAttempts: 10,
		BaseDelay:   200 * time.Millisecond,
		MaxDelay:    5 * time.Second,
		Multiplier:  2.0,
	}, func() error {
		return &RetryableError{Err: errors.New("always fails")}
	})
	fmt.Printf("  Error (context timeout): %v\n", err)

	fmt.Println(`
Key Points:
• Exponential backoff: delay grows with each attempt
• Jitter: randomize to avoid thundering herd
• Max delay cap: prevent unbounded wait
• Context respect: cancel retry loop when context done
• Retryable check: only retry transient errors
• HTTP: retry 429, 5xx; don't retry 4xx
• Respect Retry-After header from server
`)
}
