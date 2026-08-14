# Week 4 – Flashcards / Q&A – Backend Engineering

## HTTP & Networking

### Q: HTTP/1.1 head-of-line blocking là gì?
**A:** Trên 1 TCP connection, request phải xử lý tuần tự. Request B phải đợi response A hoàn tất trước. HTTP/2 giải quyết bằng multiplexing (multiple streams trên 1 connection).

---

### Q: Connection pooling trong Go HTTP client?
**A:** `http.Transport` tự động pool connections:
- `MaxIdleConns`: tổng idle connections
- `MaxIdleConnsPerHost`: idle connections per host
- `MaxConnsPerHost`: max connections per host
- `IdleConnTimeout`: thời gian giữ idle connection

CRITICAL: phải `resp.Body.Close()` để connection được return về pool!

---

### Q: Tại sao phải close response body?
**A:** Nếu không close:
1. Connection KHÔNG được return về pool
2. TCP connection bị leak
3. Cuối cùng hết connections → requests timeout
```go
defer resp.Body.Close() // ALWAYS
```

---

### Q: Timeout hierarchy trong Go HTTP?
**A:**
```
Client.Timeout = overall deadline (bao gồm tất cả)
├── Dialer.Timeout = TCP connection
├── TLSHandshakeTimeout = TLS negotiation
├── ResponseHeaderTimeout = wait for response headers
└── Read response body (remaining time)
```
Nếu chỉ set `Client.Timeout` = bao trùm hết. Nhưng nên set granular timeouts để biết lỗi ở đâu.

---

## Go HTTP Server

### Q: Graceful shutdown flow?
**A:**
1. Receive SIGTERM/SIGINT
2. Stop accepting new connections
3. Wait for in-flight requests to complete (with timeout)
4. Close server
5. Cleanup resources (DB connections, etc.)

Key: `srv.Shutdown(ctx)` — graceful. `srv.Close()` — immediate (drops connections).

---

### Q: Middleware pattern trong Go?
**A:** Function nhận `http.Handler`, return `http.Handler`:
```go
func Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // before
        next.ServeHTTP(w, r)
        // after
    })
}
```
Chain: `Recovery(Auth(Logging(handler)))`

---

### Q: Server timeout recommendations?
**A:**
- `ReadTimeout: 5s` — prevent slowloris attack
- `WriteTimeout: 10s` — prevent stuck handlers
- `IdleTimeout: 120s` — close idle keep-alive connections
- Handler-level timeout: `http.TimeoutHandler(h, 30s, "timeout")`

---

## Context

### Q: Tại sao không store context trong struct?
**A:** Context has request lifecycle (short-lived). Struct may outlive request (long-lived). Storing creates:
1. Lifecycle mismatch
2. Stale cancellation signals
3. Memory leaks (context tree not GC'd)
4. Incorrect propagation

Always pass as first parameter: `func(ctx context.Context, ...)`

---

### Q: context.WithValue — best practices?
**A:**
- Key phải là unexported type (prevent collision)
- Value phải là request-scoped (trace ID, auth user)
- KHÔNG dùng cho function parameters
- KHÔNG dùng cho optional args

```go
type contextKey string
const userKey contextKey = "user"
ctx = context.WithValue(ctx, userKey, user)
```

---

### Q: Context cancellation propagation?
**A:** Child context cancelled khi:
1. Parent cancelled
2. Timeout/deadline reached
3. Cancel function called

Check: `ctx.Err()` returns `context.Canceled` or `context.DeadlineExceeded`
Listen: `<-ctx.Done()` channel closes khi cancelled

---

### Q: context.TODO() vs context.Background()?
**A:**
- `Background()`: top-level context (main, init, tests)
- `TODO()`: placeholder khi chưa biết dùng context nào (refactoring)

Both are non-nil, never cancelled. Difference is semantic only.

---

## API Design

### Q: Offset vs Cursor pagination?
**A:**
| Feature | Offset | Cursor |
|---|---|---|
| Implementation | Simple | Complex |
| Performance | Slow for large offsets | Consistent O(1) |
| Insert/Delete | Inconsistent (skip/duplicate) | Consistent |
| Use case | Admin panels | User-facing feeds |

Cursor = encrypted/encoded reference to last item. Server decodes to build query.

---

### Q: Idempotency key implementation?
**A:**
1. Client sends unique `Idempotency-Key` header
2. Server checks key in store (Redis/DB)
3. If found → return cached response (no re-processing)
4. If not → process, store response with TTL (24h), return
5. Handle concurrent requests: use distributed lock on key

---

### Q: 401 vs 403?
**A:**
- **401 Unauthorized**: Identity unknown (missing/invalid token)
- **403 Forbidden**: Identity known but lacks permission

401 = "Who are you?" / 403 = "I know who you are, but no."

---

### Q: Khi nào dùng 202 Accepted?
**A:** Khi request được nhận nhưng xử lý async:
- Payment processing
- Email sending
- Report generation
- Any long-running task

Response include location to check status:
```json
{
    "status": "processing",
    "status_url": "/tasks/abc123"
}
```

---

### Q: Rate limiting — server-side headers?
**A:**
```http
HTTP/1.1 429 Too Many Requests
Retry-After: 30
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1625097600
```

---

### Q: API versioning strategies?
**A:**
1. **URL path**: `/api/v1/users` (most common, explicit)
2. **Header**: `Accept: application/vnd.api.v2+json`
3. **Query param**: `/users?version=2`

Recommendation: URL path for major versions, backward-compatible changes within same version.

---

## Production Concerns

### Q: Cách handle downstream service timeout?
**A:**
1. Set proper timeouts (connect + read)
2. Use context with deadline
3. Implement circuit breaker
4. Return partial response / fallback
5. Log timeout với request details for debugging

```go
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()
resp, err := client.Do(req.WithContext(ctx))
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        // timeout → circuit breaker may open
    }
}
```

---

### Q: Connection pool exhaustion — symptoms và fix?
**A:**
Symptoms:
- Increasing latency
- "connection refused" errors
- Timeout errors
- TCP connections in TIME_WAIT

Fix:
1. Close response bodies! (`defer resp.Body.Close()`)
2. Increase `MaxConnsPerHost`
3. Set proper `IdleConnTimeout`
4. Monitor connection pool metrics
5. Fix slow downstream services (they hold connections)
