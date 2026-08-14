package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// GracefulServer demonstrates proper shutdown handling
type GracefulServer struct {
	server     *http.Server
	wg         sync.WaitGroup // track in-flight requests
	shutdownCh chan struct{}
}

func NewGracefulServer(addr string, handler http.Handler) *GracefulServer {
	gs := &GracefulServer{
		shutdownCh: make(chan struct{}),
	}

	gs.server = &http.Server{
		Addr:         addr,
		Handler:      gs.trackRequests(handler),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return gs
}

// trackRequests middleware counts active requests
func (gs *GracefulServer) trackRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gs.wg.Add(1)
		defer gs.wg.Done()

		// Check if shutting down
		select {
		case <-gs.shutdownCh:
			http.Error(w, "Server is shutting down", http.StatusServiceUnavailable)
			return
		default:
		}

		next.ServeHTTP(w, r)
	})
}

func (gs *GracefulServer) Start() error {
	log.Printf("Server starting on %s", gs.server.Addr)

	if err := gs.server.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

func (gs *GracefulServer) Shutdown(ctx context.Context) error {
	log.Println("Initiating graceful shutdown...")

	// Signal that we're shutting down (reject new requests)
	close(gs.shutdownCh)

	// Stop accepting new connections & wait for in-flight requests
	if err := gs.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown error: %w", err)
	}

	// Wait for all tracked requests to complete
	done := make(chan struct{})
	go func() {
		gs.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("All requests completed")
	case <-ctx.Done():
		log.Println("Shutdown timeout — forcing close")
		return ctx.Err()
	}

	return nil
}

func main() {
	// --- Setup routes ---
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy"}`)
	})

	mux.HandleFunc("GET /slow", func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow request
		log.Println("Slow request started")
		select {
		case <-time.After(5 * time.Second):
			fmt.Fprintf(w, `{"message":"completed"}`)
			log.Println("Slow request completed")
		case <-r.Context().Done():
			log.Println("Slow request cancelled")
			return
		}
	})

	// --- Create server ---
	server := NewGracefulServer(":8080", mux)

	// --- Start server in background ---
	go func() {
		if err := server.Start(); err != nil {
			log.Fatal(err)
		}
	}()

	// --- Wait for shutdown signal ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Printf("Received signal: %v", sig)

	// --- Graceful shutdown with timeout ---
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Shutdown error: %v", err)
		os.Exit(1)
	}

	// --- Cleanup resources ---
	log.Println("Cleaning up resources...")
	// Close DB connections, flush metrics, etc.
	time.Sleep(100 * time.Millisecond) // simulate cleanup

	log.Println("Server stopped gracefully")
}

/*
Shutdown Flow:
1. Receive SIGTERM/SIGINT
2. Stop accepting new connections
3. Reject new requests with 503
4. Wait for in-flight requests (with timeout)
5. Close server
6. Cleanup resources (DB, cache, metrics)
7. Exit

Kubernetes Graceful Shutdown:
- Pod receives SIGTERM
- preStop hook runs (if configured)
- Kubelet waits terminationGracePeriodSeconds (default 30s)
- SIGKILL sent if still running

Best Practices:
- Set shutdown timeout < terminationGracePeriodSeconds
- Implement /health endpoint for readiness probe
- Return 503 after shutdown signal (before connections closed)
- Flush metrics/logs before exit
- Close DB connections gracefully
*/
