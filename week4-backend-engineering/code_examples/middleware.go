package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"crypto/rand"
	"encoding/hex"
)

// Middleware type definition
type Middleware func(http.Handler) http.Handler

// --- Middleware Chain ---

func Chain(middlewares ...Middleware) Middleware {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}

// --- 1. Request ID Middleware ---

type contextKey string

const requestIDKey contextKey = "request_id"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client sent one
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = generateID()
		}

		// Add to context
		ctx := context.WithValue(r.Context(), requestIDKey, id)

		// Add to response header
		w.Header().Set("X-Request-ID", id)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return "unknown"
}

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// --- 2. Logging Middleware ---

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status
		wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		// Log after request
		duration := time.Since(start)
		reqID := GetRequestID(r.Context())

		log.Printf("[%s] %s %s %d %d %v",
			reqID,
			r.Method,
			r.URL.Path,
			wrapped.status,
			wrapped.size,
			duration,
		)
	})
}

// --- 3. Recovery Middleware ---

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log stack trace
				reqID := GetRequestID(r.Context())
				log.Printf("[%s] PANIC: %v\n%s", reqID, err, debug.Stack())

				// Return 500
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, `{"error":{"code":"INTERNAL_ERROR","message":"Internal server error"}}`)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// --- 4. CORS Middleware ---

func CORS(allowedOrigins []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			for _, o := range allowedOrigins {
				if o == "*" || o == origin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Max-Age", "86400")
			}

			// Handle preflight
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// --- 5. Rate Limiting Middleware ---

func RateLimit(requestsPerSecond int) Middleware {
	// Simple token bucket using buffered channel
	tokens := make(chan struct{}, requestsPerSecond)

	// Refill tokens
	go func() {
		ticker := time.NewTicker(time.Second / time.Duration(requestsPerSecond))
		defer ticker.Stop()
		for range ticker.C {
			select {
			case tokens <- struct{}{}:
			default: // bucket full
			}
		}
	}()

	// Fill initial tokens
	for i := 0; i < requestsPerSecond; i++ {
		tokens <- struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-tokens:
				next.ServeHTTP(w, r)
			default:
				w.Header().Set("Retry-After", "1")
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			}
		})
	}
}

// --- 6. Timeout Middleware ---

func Timeout(duration time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, duration, `{"error":{"code":"TIMEOUT","message":"Request timeout"}}`)
	}
}

// --- Usage Example ---

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"message":"Hello, World!"}`)
	})

	mux.HandleFunc("GET /panic", func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong!")
	})

	// Chain middlewares (order matters!)
	// Request flow: RequestID → Recovery → Logging → CORS → RateLimit → Timeout → Handler
	middleware := Chain(
		RequestID,
		Recovery,
		Logging,
		CORS([]string{"*"}),
		RateLimit(100),
		Timeout(30*time.Second),
	)

	handler := middleware(mux)

	fmt.Println("Server with middleware chain on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
