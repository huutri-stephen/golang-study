# Week 4 – Backend Engineering – Study Notes

## 1. HTTP Fundamentals

### HTTP/1.1 vs HTTP/2 vs HTTP/3

| Feature | HTTP/1.1 | HTTP/2 | HTTP/3 |
|---|---|---|---|
| Protocol | TCP | TCP | QUIC (UDP) |
| Multiplexing | No (head-of-line blocking) | Yes (streams) | Yes |
| Header compression | No | HPACK | QPACK |
| Server push | No | Yes | Yes |
| Connection setup | TCP + TLS (3 RTT) | TCP + TLS (2 RTT) | 0-1 RTT |

### TCP Connection Lifecycle

```
Client              Server
  │ SYN ──────────→ │
  │ ←──── SYN+ACK  │
  │ ACK ──────────→ │
  │                  │  ← 3-way handshake (1.5 RTT)
  │ TLS ClientHello→│
  │ ←── ServerHello │
  │ ← Certificate   │
  │ Finished ──────→│  ← TLS handshake (1-2 RTT)
  │                  │
  │ HTTP Request ──→│
  │ ←── Response    │  ← Data exchange
```

### Keep-Alive

- HTTP/1.1 default: `Connection: keep-alive`
- Reuse TCP connection for multiple requests
- Saves 3-way handshake + TLS handshake per request
- Go `http.Transport` reuses connections by default

---

## 2. Go HTTP Client

### Default Client Problems

```go
// BAD: no timeout, can hang forever
resp, err := http.Get("https://api.example.com/data")

// BAD: shared default Transport may not fit needs
http.DefaultClient
```

### Proper Client Configuration

```go
client := &http.Client{
    Timeout: 30 * time.Second, // overall request timeout
    Transport: &http.Transport{
        DialContext: (&net.Dialer{
            Timeout:   5 * time.Second,  // connection timeout
            KeepAlive: 30 * time.Second, // keep-alive probe interval
        }).DialContext,
        TLSHandshakeTimeout:   5 * time.Second,
        MaxIdleConns:          100,             // total idle connections
        MaxIdleConnsPerHost:   10,              // per-host idle connections
        MaxConnsPerHost:       50,              // max connections per host
        IdleConnTimeout:       90 * time.Second,
        ResponseHeaderTimeout: 10 * time.Second,
        ExpectContinueTimeout: 1 * time.Second,
    },
}
```

### Timeout Hierarchy

```
|--- Overall Timeout (http.Client.Timeout) ---|
|                                              |
| Dial | TLS | Send | Wait | Read Response    |
```

### Response Body — MUST Close!

```go
resp, err := client.Do(req)
if err != nil {
    return err
}
defer resp.Body.Close() // CRITICAL! Prevents connection leak

body, err := io.ReadAll(resp.Body)
```

---

## 3. Go HTTP Server

### Server Configuration

```go
srv := &http.Server{
    Addr:           ":8080",
    Handler:        mux,
    ReadTimeout:    5 * time.Second,   // time to read entire request
    WriteTimeout:   10 * time.Second,  // time to write response
    IdleTimeout:    120 * time.Second, // keep-alive timeout
    MaxHeaderBytes: 1 << 20,           // 1MB max header size
}
```

### Graceful Shutdown

```go
// Start server
go func() {
    if err := srv.ListenAndServe(); err != http.ErrServerClosed {
        log.Fatalf("Server error: %v", err)
    }
}()

// Wait for interrupt signal
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

// Graceful shutdown with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := srv.Shutdown(ctx); err != nil {
    log.Fatalf("Shutdown error: %v", err)
}
log.Println("Server stopped gracefully")
```

### Middleware Pattern

```go
type Middleware func(http.Handler) http.Handler

func Logging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
    })
}

func Recovery(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                w.WriteHeader(http.StatusInternalServerError)
                log.Printf("panic: %v", err)
            }
        }()
        next.ServeHTTP(w, r)
    })
}

// Chain middlewares
handler := Recovery(Logging(mux))
```

---

## 4. Context

### Purpose
- Cancellation propagation
- Deadline/timeout propagation
- Request-scoped values (trace ID, auth info)

### Rules

1. Context is first parameter: `func DoSomething(ctx context.Context, ...)`
2. Never store context in struct
3. Never pass nil context — use `context.TODO()` or `context.Background()`
4. Context values are for request-scoped data only, NOT for function parameters
5. Key type should be unexported to prevent collision

### Why not store in struct?

```go
// BAD
type Service struct {
    ctx context.Context // WRONG! Context has request lifecycle
}

// GOOD
func (s *Service) Process(ctx context.Context, data Data) error {
    // ctx lives for this request only
}
```

Reason: Struct may outlive request. Context represents a single operation/request. Storing it creates lifecycle mismatch.

### Cancellation Flow

```
                    Background
                        │
            WithCancel (cancel func)
                        │
            ┌───────────┼───────────┐
            │           │           │
    WithTimeout(5s)  WithValue   Direct child
            │
        handler()
            │
    if ctx.Err() != nil → cancelled!
```

---

## 5. API Design Best Practices

### Error Response Format

```json
{
    "error": {
        "code": "VALIDATION_ERROR",
        "message": "Invalid request parameters",
        "details": [
            {
                "field": "email",
                "message": "must be valid email format"
            }
        ],
        "request_id": "req-abc123"
    }
}
```

### Pagination

```
# Offset-based (simple, but skip problem)
GET /users?page=2&page_size=20

# Cursor-based (efficient, consistent)
GET /users?cursor=eyJpZCI6MTAwfQ&limit=20

Response:
{
    "data": [...],
    "pagination": {
        "next_cursor": "eyJpZCI6MTIwfQ",
        "has_more": true
    }
}
```

### Idempotency

```
POST /payments
Idempotency-Key: abc-123-def

Server behavior:
1. Check if key exists in store
2. If exists → return cached response
3. If not → process, store response, return
```

Implementation:
```go
func (h *Handler) CreatePayment(w http.ResponseWriter, r *http.Request) {
    key := r.Header.Get("Idempotency-Key")
    if key == "" {
        http.Error(w, "Idempotency-Key required", 400)
        return
    }
    
    // Check cache
    if cached, ok := h.idempotencyStore.Get(key); ok {
        writeJSON(w, cached.StatusCode, cached.Body)
        return
    }
    
    // Process payment
    result, err := h.paymentService.Create(r.Context(), req)
    
    // Store result
    h.idempotencyStore.Set(key, result, 24*time.Hour)
    
    writeJSON(w, http.StatusCreated, result)
}
```

### Retry Semantics

- **Safe to retry**: GET, PUT, DELETE (idempotent)
- **Dangerous to retry**: POST (non-idempotent without idempotency key)
- **Client retry strategy**: exponential backoff + jitter
- **Retry-After header**: server tells client when to retry

---

## 6. HTTP Status Codes (Senior Level)

| Code | Meaning | When to use |
|---|---|---|
| 200 | OK | Successful GET, PUT |
| 201 | Created | Successful POST creating resource |
| 202 | Accepted | Async operation queued |
| 204 | No Content | Successful DELETE |
| 400 | Bad Request | Validation error |
| 401 | Unauthorized | Missing/invalid authentication |
| 403 | Forbidden | Authenticated but not authorized |
| 404 | Not Found | Resource doesn't exist |
| 409 | Conflict | Resource state conflict (duplicate) |
| 422 | Unprocessable | Valid syntax but semantic error |
| 429 | Too Many Requests | Rate limited |
| 500 | Internal Error | Unexpected server error |
| 502 | Bad Gateway | Upstream service error |
| 503 | Service Unavailable | Server overloaded/maintenance |
| 504 | Gateway Timeout | Upstream timeout |
