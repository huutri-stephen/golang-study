package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	fmt.Println("=== 1. Context Cancellation ===")
	cancellationDemo()

	fmt.Println("\n=== 2. Context Timeout ===")
	timeoutDemo()

	fmt.Println("\n=== 3. Context Values ===")
	valuesDemo()

	fmt.Println("\n=== 4. Context Propagation in HTTP ===")
	httpContextDemo()

	fmt.Println("\n=== 5. Best Practices ===")
	bestPractices()
}

// --- 1. Cancellation ---

func cancellationDemo() {
	ctx, cancel := context.WithCancel(context.Background())

	// Worker that respects cancellation
	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Printf("  Worker stopped: %v\n", ctx.Err())
				return
			default:
				fmt.Println("  Worker: doing work...")
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	// Let it work for a bit
	time.Sleep(300 * time.Millisecond)

	// Cancel
	cancel()
	time.Sleep(50 * time.Millisecond) // let worker exit

	// Key points:
	// - cancel() is idempotent (safe to call multiple times)
	// - Always defer cancel() to prevent context leak
	// - ctx.Done() channel is closed when cancelled
	// - ctx.Err() returns Canceled or DeadlineExceeded
}

// --- 2. Timeout ---

func timeoutDemo() {
	// WithTimeout = WithDeadline(time.Now().Add(duration))
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel() // always defer cancel!

	// Simulate slow operation
	result := make(chan string, 1)
	go func() {
		time.Sleep(500 * time.Millisecond) // slower than timeout
		result <- "completed"
	}()

	select {
	case r := <-result:
		fmt.Printf("  Result: %s\n", r)
	case <-ctx.Done():
		fmt.Printf("  Timeout! Error: %v\n", ctx.Err())
	}

	// Nested timeouts — child cannot exceed parent
	parentCtx, parentCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer parentCancel()

	// Child timeout is 1s (effective: min(1s, parent remaining))
	childCtx, childCancel := context.WithTimeout(parentCtx, 1*time.Second)
	defer childCancel()

	deadline, ok := childCtx.Deadline()
	fmt.Printf("  Child deadline: %v, has deadline: %v\n",
		time.Until(deadline).Round(time.Millisecond), ok)
}

// --- 3. Values ---

type ctxKey string

const (
	userIDKey    ctxKey = "user_id"
	traceIDKey   ctxKey = "trace_id"
	requestIDKey ctxKey = "request_id"
)

func valuesDemo() {
	// Build context with values
	ctx := context.Background()
	ctx = context.WithValue(ctx, userIDKey, "user-123")
	ctx = context.WithValue(ctx, traceIDKey, "trace-abc")
	ctx = context.WithValue(ctx, requestIDKey, "req-xyz")

	// Retrieve values
	userID := ctx.Value(userIDKey).(string)
	traceID := ctx.Value(traceIDKey).(string)
	fmt.Printf("  UserID: %s, TraceID: %s\n", userID, traceID)

	// Safe retrieval (check existence)
	if v, ok := ctx.Value(userIDKey).(string); ok {
		fmt.Printf("  Found user: %s\n", v)
	}

	// Value not found → returns nil
	val := ctx.Value(ctxKey("nonexistent"))
	fmt.Printf("  Missing key: %v (nil)\n", val)

	// Pass to function
	processRequest(ctx)
}

func processRequest(ctx context.Context) {
	userID, _ := ctx.Value(userIDKey).(string)
	traceID, _ := ctx.Value(traceIDKey).(string)
	fmt.Printf("  processRequest: user=%s, trace=%s\n", userID, traceID)
}

// --- 4. HTTP Context ---

func httpContextDemo() {
	// In real HTTP handler, request carries context
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Add values (middleware would do this)
		ctx = context.WithValue(ctx, userIDKey, "user-456")

		// Call service with context (propagates cancellation)
		result, err := callService(ctx)
		if err != nil {
			if ctx.Err() == context.Canceled {
				// Client disconnected
				log.Println("Client cancelled request")
				return
			}
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Fprintf(w, "Result: %s", result)
	})

	// Simulate request with context
	req, _ := http.NewRequest("GET", "/test", nil)

	// Add timeout to request context
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	// Execute (simulated)
	_ = handler
	_ = req
	fmt.Println("  HTTP handler demonstrates context from request")
	fmt.Println("  r.Context() carries: cancellation, deadline, values")
}

func callService(ctx context.Context) (string, error) {
	// Simulate service call that respects context
	select {
	case <-time.After(100 * time.Millisecond):
		return "service response", nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// --- 5. Best Practices ---

func bestPractices() {
	fmt.Println(`
Context Best Practices:

1. FIRST PARAMETER:
   func Process(ctx context.Context, id string) error { ... }

2. NEVER STORE IN STRUCT:
   // BAD
   type Service struct { ctx context.Context }
   // GOOD  
   func (s *Service) Do(ctx context.Context) error { ... }

3. ALWAYS DEFER CANCEL:
   ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
   defer cancel()  // prevents context leak

4. CHECK CANCELLATION IN LOOPS:
   for _, item := range items {
       select {
       case <-ctx.Done():
           return ctx.Err()
       default:
       }
       process(item)
   }

5. USE UNEXPORTED KEY TYPES:
   type ctxKey string  // unexported prevents collision
   const myKey ctxKey = "my_key"

6. VALUES ARE FOR REQUEST-SCOPED DATA ONLY:
   ✓ Trace ID, Request ID, User ID
   ✗ Database connection, Logger, Config

7. NEVER PASS NIL CONTEXT:
   // Use context.Background() or context.TODO()

8. PROPAGATE TO DOWNSTREAM CALLS:
   resp, err := http.DefaultClient.Do(req.WithContext(ctx))
   rows, err := db.QueryContext(ctx, query)
   result, err := redis.Get(ctx, key).Result()
`)
}
