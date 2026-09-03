package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// --- Idempotency Implementation ---

// IdempotencyStore stores processed request results
type IdempotencyStore struct {
	mu      sync.RWMutex
	entries map[string]*IdempotencyEntry
}

type IdempotencyEntry struct {
	Key       string
	Status    string // "processing", "completed", "failed"
	Response  *CachedResponse
	CreatedAt time.Time
	ExpiresAt time.Time
}

type CachedResponse struct {
	StatusCode int
	Body       []byte
	Headers    map[string]string
}

func NewIdempotencyStore() *IdempotencyStore {
	store := &IdempotencyStore{
		entries: make(map[string]*IdempotencyEntry),
	}
	// Cleanup expired entries periodically
	go store.cleanup()
	return store
}

func (s *IdempotencyStore) Get(key string) (*IdempotencyEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.entries[key]
	if !ok || time.Now().After(entry.ExpiresAt) {
		return nil, false
	}
	return entry, true
}

func (s *IdempotencyStore) Lock(key string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if key exists
	if entry, ok := s.entries[key]; ok {
		if entry.Status == "processing" {
			return false, nil // another request is processing
		}
		if entry.Status == "completed" && time.Now().Before(entry.ExpiresAt) {
			return false, nil // already completed
		}
	}

	// Acquire lock
	s.entries[key] = &IdempotencyEntry{
		Key:       key,
		Status:    "processing",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	return true, nil
}

func (s *IdempotencyStore) Complete(key string, resp *CachedResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entry, ok := s.entries[key]; ok {
		entry.Status = "completed"
		entry.Response = resp
	}
}

func (s *IdempotencyStore) Fail(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, key) // allow retry
}

func (s *IdempotencyStore) cleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for key, entry := range s.entries {
			if now.After(entry.ExpiresAt) {
				delete(s.entries, key)
			}
		}
		s.mu.Unlock()
	}
}

// --- Payment Handler with Idempotency ---

type PaymentRequest struct {
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Description string  `json:"description"`
}

type PaymentResponse struct {
	ID          string  `json:"id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Status      string  `json:"status"`
	ProcessedAt string  `json:"processed_at"`
}

type PaymentHandler struct {
	store *IdempotencyStore
}

func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Get idempotency key
	idempotencyKey := r.Header.Get("Idempotency-Key")
	if idempotencyKey == "" {
		writeError(w, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	// 2. Check if already processed
	if entry, ok := h.store.Get(idempotencyKey); ok {
		switch entry.Status {
		case "completed":
			// Return cached response
			w.Header().Set("X-Idempotent-Replayed", "true")
			w.WriteHeader(entry.Response.StatusCode)
			w.Write(entry.Response.Body)
			return
		case "processing":
			// Another request with same key is in progress
			w.WriteHeader(http.StatusConflict)
			fmt.Fprintf(w, `{"error":"Request with this idempotency key is being processed"}`)
			return
		}
	}

	// 3. Acquire processing lock
	acquired, err := h.store.Lock(idempotencyKey)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to acquire lock")
		return
	}
	if !acquired {
		writeError(w, http.StatusConflict, "Duplicate request")
		return
	}

	// 4. Parse request
	var req PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.store.Fail(idempotencyKey) // allow retry with correct body
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 5. Process payment
	payment, err := processPayment(ctx, req)
	if err != nil {
		h.store.Fail(idempotencyKey) // allow retry
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 6. Cache response
	body, _ := json.Marshal(payment)
	h.store.Complete(idempotencyKey, &CachedResponse{
		StatusCode: http.StatusCreated,
		Body:       body,
	})

	// 7. Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(body)
}

func processPayment(ctx context.Context, req PaymentRequest) (*PaymentResponse, error) {
	// Simulate payment processing
	_ = ctx
	if req.Amount <= 0 {
		return nil, errors.New("invalid amount")
	}

	return &PaymentResponse{
		ID:          fmt.Sprintf("pay_%d", time.Now().UnixNano()),
		Amount:      req.Amount,
		Currency:    req.Currency,
		Status:      "succeeded",
		ProcessedAt: time.Now().Format(time.RFC3339),
	}, nil
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	fmt.Fprintf(w, `{"error":"%s"}`, msg)
}

// --- Demo ---

func main() {
	store := NewIdempotencyStore()
	handler := &PaymentHandler{store: store}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /payments", handler.CreatePayment)

	fmt.Println("Idempotent Payment Service on :8080")
	fmt.Println("")
	fmt.Println("Test commands:")
	fmt.Println(`  # First request`)
	fmt.Println(`  curl -X POST http://localhost:8080/payments \`)
	fmt.Println(`    -H "Idempotency-Key: key-123" \`)
	fmt.Println(`    -H "Content-Type: application/json" \`)
	fmt.Println(`    -d '{"amount": 99.99, "currency": "USD"}'`)
	fmt.Println("")
	fmt.Println(`  # Same key = same response (idempotent replay)`)
	fmt.Println(`  curl -X POST http://localhost:8080/payments \`)
	fmt.Println(`    -H "Idempotency-Key: key-123" \`)
	fmt.Println(`    -H "Content-Type: application/json" \`)
	fmt.Println(`    -d '{"amount": 99.99, "currency": "USD"}'`)
	fmt.Println("")
	fmt.Println(`Key Implementation Details:`)
	fmt.Println(`• Idempotency-Key header (client-generated UUID)`)
	fmt.Println(`• Lock mechanism prevents concurrent duplicate processing`)
	fmt.Println(`• Response cached with TTL (24h typical)`)
	fmt.Println(`• Failed requests: remove from store to allow retry`)
	fmt.Println(`• Completed requests: return cached response`)
	fmt.Println(`• Storage: Redis (production) or DB for durability`)

	http.ListenAndServe(":8080", mux)
}
