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

---

## 7. gRPC & Protobuf

> roadmap.sh liệt kê gRPC như kỹ năng cốt lõi của Go backend. Plan gốc chỉ nói chung "service-to-service".
> Đây là phần rất hay hỏi cho microservices Go.

### Protobuf (Protocol Buffers)

- IDL (Interface Definition Language) + binary serialization format của Google.
- Nhỏ hơn và nhanh hơn JSON nhiều (binary, có schema, không lặp field name).
- Định nghĩa `.proto` → sinh code Go bằng `protoc` + `protoc-gen-go` / `protoc-gen-go-grpc`.

```protobuf
syntax = "proto3";
package payment.v1;
option go_package = "github.com/acme/payment/gen/paymentv1";

message ChargeRequest {
  string idempotency_key = 1;   // field number: định danh trên wire, KHÔNG đổi
  int64  amount_cents    = 2;
  string currency        = 3;
}
message ChargeResponse {
  string payment_id = 1;
  string status     = 2;
}

service PaymentService {
  rpc Charge(ChargeRequest) returns (ChargeResponse);            // unary
  rpc StreamStatus(ChargeRequest) returns (stream ChargeResponse); // server streaming
}
```

**Quy tắc tương thích:** field number là hợp đồng trên wire — **không đổi/không tái dùng**. Thêm field mới với number mới → backward compatible. Đây là lý do protobuf tốt cho API tiến hoá dần.

### 4 kiểu RPC

| Kiểu | Mô tả | Ví dụ |
|---|---|---|
| Unary | 1 request → 1 response | Charge |
| Server streaming | 1 request → nhiều response | tail log, progress |
| Client streaming | nhiều request → 1 response | upload chunk |
| Bidirectional streaming | nhiều ↔ nhiều | chat, real-time |

### gRPC vs REST

| | gRPC | REST/JSON |
|---|---|---|
| Payload | Protobuf (binary, nhỏ) | JSON (text, lớn hơn) |
| Transport | HTTP/2 (multiplexing, stream) | thường HTTP/1.1 |
| Contract | `.proto` strict, sinh code | OpenAPI (tùy chọn) |
| Streaming | Native (4 kiểu) | Hạn chế (SSE/WebSocket) |
| Browser | Cần grpc-web/proxy | Native |
| Dùng khi | internal service-to-service, low latency | public API, browser, đơn giản |

### Interceptor (≈ middleware của gRPC)

```go
func UnaryLogging(ctx context.Context, req any, info *grpc.UnaryServerInfo,
    handler grpc.UnaryHandler) (any, error) {
    start := time.Now()
    resp, err := handler(ctx, req)              // gọi handler thật
    log.Printf("%s took %v err=%v", info.FullMethod, time.Since(start), err)
    return resp, err
}
// server := grpc.NewServer(grpc.ChainUnaryInterceptor(UnaryLogging, UnaryAuth, UnaryRecovery))
```

Dùng interceptor cho: logging, auth, metrics, tracing (OpenTelemetry), panic recovery, rate limit — giống middleware HTTP.

### Điểm cần nắm cho interview

- gRPC dùng **HTTP/2** → multiplexing nhiều RPC trên 1 connection, không HOL blocking ở tầng ứng dụng.
- **Deadline propagation**: client set deadline → truyền qua context xuống server và downstream. Luôn dùng `context.WithTimeout` khi gọi.
- Status code riêng (`codes.NotFound`, `codes.DeadlineExceeded`, `codes.Unavailable`...) — map sang HTTP khi cần.
- Load balancing: gRPC connection bền (long-lived) → cần L7 LB (client-side LB, hoặc proxy như Envoy) chứ L4 LB sẽ dồn tải một pod.

---

## 8. Security / Authentication & Authorization

> Plan gốc mới có TLS. Auth là chủ đề gần như luôn xuất hiện trong backend interview.

### Authentication vs Authorization

- **Authentication (AuthN):** "Bạn là ai?" → verify danh tính (login, token).
- **Authorization (AuthZ):** "Bạn được làm gì?" → kiểm tra quyền (RBAC/ABAC).
- Nhắc lại: **401** = chưa/không xác thực được; **403** = xác thực rồi nhưng không đủ quyền.

### JWT (JSON Web Token)

```
header.payload.signature   (base64url, phân cách bằng dấu chấm)
```

- **Stateless**: server không lưu session, verify bằng chữ ký. Scale ngang tốt.
- Payload (claims) chứa `sub`, `exp`, `iat`, `roles`... — **KHÔNG chứa dữ liệu nhạy cảm** vì chỉ base64, ai cũng đọc được (không mã hoá, chỉ ký).
- Ký bằng **HS256** (HMAC, secret chung) hoặc **RS256** (RSA, private ký / public verify — phù hợp nhiều service).
- **Nhược điểm:** khó thu hồi (revoke) trước khi hết hạn → dùng TTL ngắn + **refresh token** (lưu server, revoke được).

```go
// Ký
tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
    "sub": userID, "exp": time.Now().Add(15 * time.Minute).Unix(),
})
signed, _ := tok.SignedString([]byte(secret))

// Verify (LUÔN kiểm tra signing method để chống thuật toán "none" / confusion attack)
parsed, err := jwt.Parse(signed, func(t *jwt.Token) (any, error) {
    if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, errors.New("unexpected signing method")
    }
    return []byte(secret), nil
})
```

### OAuth2 / OIDC (tóm tắt)

- **OAuth2** = ủy quyền truy cập (authorization framework); **OIDC** = lớp authentication trên OAuth2 (thêm ID token).
- Flow phổ biến: **Authorization Code + PKCE** (web/mobile), Client Credentials (service-to-service).
- Đừng tự viết OAuth server — dùng provider (Auth0/Keycloak/Cognito) hoặc lib chuẩn.

### Password & Secrets

- Băm mật khẩu bằng **bcrypt / argon2 / scrypt** (có salt, chậm có chủ đích). **KHÔNG** dùng MD5/SHA-256 trần.
- So sánh token/HMAC bằng **`hmac.Equal` / `subtle.ConstantTimeCompare`** để chống timing attack.
- Secret qua env/secret manager (Vault, AWS Secrets Manager), không commit.

### Checklist bảo mật API

- [ ] TLS bắt buộc (HSTS), redirect HTTP→HTTPS.
- [ ] Validate & sanitize input; parameterized query (chống SQL injection).
- [ ] Giới hạn body size (`http.MaxBytesReader`), rate limit, timeout.
- [ ] CORS đúng origin, không `*` cho endpoint có credential.
- [ ] Không log secret/token/PII; header bảo mật (CSP, X-Content-Type-Options).
- [ ] Auth middleware đặt sát handler, sau RequestID/Recovery/RateLimit.

---

## 9. Web Frameworks (net/http vs Gin/Echo/chi/Fiber)

> Interview hay hỏi "dùng framework nào, vì sao". Cần giải thích trade-off, không cần thuộc API.

| | Đặc điểm | Khi chọn |
|---|---|---|
| `net/http` (+ ServeMux 1.22) | stdlib, không dependency; Go 1.22 thêm routing method+path param | thích tối giản, ít phụ thuộc |
| **chi** | router nhẹ, 100% tương thích `http.Handler`, middleware chuẩn stdlib | muốn nhẹ + idiomatic, dễ test |
| **Gin** | nhanh, hệ sinh thái lớn, context riêng, binding/validation sẵn | REST API phổ biến, cần năng suất |
| **Echo** | tương tự Gin, API gọn, nhiều middleware | tương đương Gin |
| **Fiber** | trên fasthttp (không phải net/http) → rất nhanh nhưng **không tương thích** ecosystem net/http | cần throughput cực cao, chấp nhận đánh đổi |

**Quan điểm senior:** Go 1.22 `net/http.ServeMux` đã hỗ trợ `GET /users/{id}` → nhiều project không cần framework. Chọn `chi` khi muốn nhẹ và giữ tương thích `http.Handler` (mọi middleware tái dùng được). Tránh chọn framework chỉ vì quen — cân nhắc coupling, khả năng test, và ecosystem.
