# Senior Backend Engineer – Golang Interview Preparation Plan

## 1. Mục tiêu

Chuẩn bị cho Senior Backend Engineer – Golang interview, tập trung vào:

- Go internals và idiomatic Go
- Concurrency và Go runtime
- Memory, GC và performance
- Backend/API engineering
- Database, Redis và messaging
- Distributed systems
- Microservices
- Observability
- System Design
- Coding interview
- Production troubleshooting
- Behavioral / Senior-level discussion

---

# 2. Roadmap tổng quan – 8 tuần (+ Week 0 nền tảng)

| Tuần | Chủ đề | Priority |
|---|---|---|
| 0 | Foundations & Tooling (modules, testing, tooling, errors, json) | ⭐⭐⭐⭐ |
| 1 | Go Core + Memory | ⭐⭐⭐⭐⭐ |
| 2 | Goroutine + Concurrency | ⭐⭐⭐⭐⭐ |
| 3 | Go Runtime + Performance | ⭐⭐⭐⭐⭐ |
| 4 | Backend Engineering | ⭐⭐⭐⭐⭐ |
| 5 | Database + Redis + Messaging | ⭐⭐⭐⭐⭐ |
| 6 | Distributed Systems | ⭐⭐⭐⭐⭐ |
| 7 | System Design | ⭐⭐⭐⭐⭐ |
| 8 | Mock Interview + Revision | ⭐⭐⭐⭐⭐ |

## Tỷ lệ thời gian đề xuất

- 30% Theory
- 30% Coding
- 25% System Design / Architecture
- 15% Mock Interview / Review

---

# 2b. Week 0 – Foundations & Tooling

> Bổ sung theo roadmap.sh/golang: phần "ecosystem nền tảng" mà interviewer coi là hiển nhiên ở mức senior.
> Có thể học xen kẽ hoặc lồng vào routine hằng ngày, không nhất thiết tách tuần riêng.

## Modules & Dependency Management

- [ ] `go.mod` / `go.sum` (checksum, supply-chain integrity)
- [ ] `go mod tidy` / `download` / `vendor`
- [ ] Semantic versioning + Semantic Import Versioning (`/v2`)
- [ ] MVS (Minimal Version Selection) vs npm
- [ ] Go workspaces (`go work`) cho monorepo
- [ ] `replace` directive

## Testing (chiều sâu)

- [ ] Table-driven tests + subtests (`t.Run`)
- [ ] Mocking qua interface (gomock / testify / moq)
- [ ] Coverage (`-cover`, `-coverprofile`)
- [ ] Fuzzing (`go test -fuzz`)
- [ ] `t.Helper()`, `t.Cleanup()`, `require` vs `assert`
- [ ] Integration test (build tags, testcontainers)

## Tooling & Code Quality

- [ ] `gofmt` / `goimports`
- [ ] `go vet`
- [ ] `staticcheck`
- [ ] `golangci-lint` (CI)
- [ ] `delve` (dlv) debugger

## Error Handling hiện đại

- [ ] `fmt.Errorf` + `%w` wrapping
- [ ] `errors.Is` vs `errors.As`
- [ ] Sentinel error vs custom error type
- [ ] `errors.Join` (Go 1.20+)

## encoding/json

- [ ] Struct tags (`json:"-"`, `omitempty`)
- [ ] Marshal / Unmarshal / streaming Decoder
- [ ] `DisallowUnknownFields`, `UseNumber`
- [ ] Custom `MarshalJSON`

## Reflection & Ecosystem

- [ ] `reflect` cơ bản (nền tảng của json/ORM/validator)
- [ ] Config: flag > env > file (viper/envconfig)
- [ ] Logging: `log/slog`, zap, zerolog

---

# 3. Week 1 – Go Core + Memory

## Go Language

- [ ] Variable, constant
- [ ] Pointer
- [ ] Struct
- [ ] Interface
- [ ] Method / receiver
- [ ] Embedding
- [ ] Slice
- [ ] Array
- [ ] Map
- [ ] String / rune / byte
- [ ] Error
- [ ] `defer`
- [ ] `panic` / `recover`
- [ ] Generics

## Slice

Cần hiểu sâu:

- [ ] `len`
- [ ] `cap`
- [ ] Underlying array
- [ ] `append`
- [ ] Reallocation
- [ ] Memory sharing
- [ ] Slice header

Practice:

```go
a := []int{1, 2, 3}
b := a[:2]
b = append(b, 10)
```

Giải thích chính xác giá trị của `a` và `b`.

## Map

- [ ] Hash table
- [ ] Bucket
- [ ] Collision
- [ ] Growth
- [ ] Concurrent access
- [ ] `sync.Map`
- [ ] `Mutex`
- [ ] `RWMutex`

Interview:

> Is Go map thread-safe?

## Interface

- [ ] Interface internals
- [ ] Dynamic type / dynamic value
- [ ] Nil interface
- [ ] Typed nil
- [ ] Method set
- [ ] Pointer receiver vs value receiver

## Memory

- [ ] Stack
- [ ] Heap
- [ ] Value semantics
- [ ] Pointer
- [ ] Escape analysis
- [ ] Allocation

---

# 4. Week 2 – Goroutine + Concurrency

## Goroutine

- [ ] Goroutine vs OS thread
- [ ] M:N scheduling
- [ ] `GOMAXPROCS`
- [ ] Go scheduler
- [ ] Run queue
- [ ] Global run queue
- [ ] Work stealing
- [ ] Syscall
- [ ] Network poller

## GMP

Must understand:

- [ ] G – Goroutine
- [ ] M – Machine / OS thread
- [ ] P – Processor
- [ ] Scheduling lifecycle
- [ ] Goroutine blocking
- [ ] Goroutine parking / wake-up

## Channel

- [ ] Unbuffered channel
- [ ] Buffered channel
- [ ] Send / receive blocking
- [ ] Channel close
- [ ] Range over channel
- [ ] Select
- [ ] Nil channel
- [ ] Closed channel

## Synchronization

- [ ] `sync.Mutex`
- [ ] `sync.RWMutex`
- [ ] `sync.WaitGroup`
- [ ] `sync.Once`
- [ ] `sync.Cond`
- [ ] `sync/atomic`
- [ ] `sync.Map`
- [ ] `errgroup`

## Concurrency problems

- [ ] Data race
- [ ] Deadlock
- [ ] Livelock
- [ ] Starvation
- [ ] Goroutine leak

Practice:

```bash
go test -race ./...
```

## Interview questions

- [ ] Mutex vs channel?
- [ ] Buffered vs unbuffered channel?
- [ ] How does scheduler handle blocking goroutines?
- [ ] How do you prevent goroutine leaks?
- [ ] How do you build a worker pool?
- [ ] How do you implement fan-in / fan-out?

---

# 5. Week 3 – Go Runtime + Performance

## Memory & Escape Analysis

- [ ] Stack growth
- [ ] Heap allocation
- [ ] Escape analysis
- [ ] Allocation pressure
- [ ] Pointer escaping
- [ ] Value vs pointer trade-offs

Practice:

```bash
go build -gcflags="-m"
```

## Garbage Collector

- [ ] Tri-color marking
- [ ] Concurrent GC
- [ ] STW
- [ ] GC cycles
- [ ] Heap growth
- [ ] Allocation rate
- [ ] GC pressure

## Profiling

- [ ] CPU profile
- [ ] Heap profile
- [ ] Goroutine profile
- [ ] Mutex profile
- [ ] Block profile
- [ ] `pprof`

## Performance workflow

```text
Metrics
   ↓
Identify bottleneck
   ↓
Profile
   ↓
Find hot path
   ↓
Optimize
   ↓
Benchmark
   ↓
Validate
```

## Benchmark

- [ ] `go test -bench`
- [ ] `-benchmem`
- [ ] Benchmark design
- [ ] Avoid misleading benchmarks

---

# 6. Week 4 – Backend Engineering

## HTTP

- [ ] HTTP/1.1
- [ ] HTTP/2
- [ ] HTTP/3
- [ ] TCP
- [ ] TLS
- [ ] Keep-Alive
- [ ] Connection pooling
- [ ] Timeout
- [ ] Retry

## Go HTTP

- [ ] `net/http`
- [ ] `http.Client`
- [ ] `Transport`
- [ ] Server
- [ ] Middleware
- [ ] Graceful shutdown

## Context

- [ ] `context.WithCancel`
- [ ] `context.WithTimeout`
- [ ] `context.WithDeadline`
- [ ] `context.WithValue`
- [ ] Cancellation propagation
- [ ] Request-scoped context

Interview:

> Why shouldn't `context.Context` be stored inside a struct?

## API Design

- [ ] REST
- [ ] HTTP status codes
- [ ] Request validation
- [ ] Error handling
- [ ] Pagination
- [ ] Filtering
- [ ] Sorting
- [ ] Versioning
- [ ] Idempotency
- [ ] API timeout
- [ ] Retry semantics

Example:

```http
POST /payments
Idempotency-Key: abc123
```

## gRPC & Protobuf

- [ ] Protobuf IDL, field number & compatibility
- [ ] 4 kiểu RPC (unary, server/client/bidi streaming)
- [ ] gRPC vs REST trade-offs
- [ ] Interceptor (auth, logging, tracing, recovery)
- [ ] Deadline propagation, status codes
- [ ] Load balancing với connection bền (HTTP/2)

## Security / Auth

- [ ] AuthN vs AuthZ (401 vs 403)
- [ ] JWT (HS256/RS256, claims, exp, refresh token)
- [ ] OAuth2 / OIDC (Authorization Code + PKCE)
- [ ] Password hashing (bcrypt/argon2), constant-time compare
- [ ] Input validation, SQL injection, CORS, body size limit

## Web Frameworks

- [ ] `net/http` + ServeMux (Go 1.22 routing)
- [ ] chi / Gin / Echo / Fiber trade-offs

---

# 7. Week 5 – Database + Redis + Messaging

## SQL / PostgreSQL / MySQL

- [ ] Index
- [ ] B-Tree
- [ ] Composite index
- [ ] Covering index
- [ ] Query plan
- [ ] `EXPLAIN`
- [ ] Transaction
- [ ] ACID
- [ ] Isolation levels
- [ ] MVCC
- [ ] Lock
- [ ] Deadlock
- [ ] Connection pool
- [ ] N+1 query

## Database troubleshooting

Scenario:

> API latency increases from 100ms to 2s.

Investigate:

```text
Application
    ↓
DB connection pool
    ↓
Slow query
    ↓
Lock
    ↓
Index
    ↓
Network
```

## Redis

- [ ] String
- [ ] Hash
- [ ] List
- [ ] Set
- [ ] Sorted Set
- [ ] TTL
- [ ] Cache
- [ ] Pub/Sub
- [ ] Stream
- [ ] Distributed lock

## Cache strategies

- [ ] Cache Aside
- [ ] Read Through
- [ ] Write Through
- [ ] Write Behind

## Cache problems

- [ ] Cache stampede
- [ ] Cache penetration
- [ ] Cache avalanche
- [ ] Hot key

## Kafka / Message Queue

- [ ] Producer
- [ ] Consumer
- [ ] Partition
- [ ] Offset
- [ ] Consumer group
- [ ] Ordering
- [ ] At-most-once
- [ ] At-least-once
- [ ] Exactly-once
- [ ] Retry
- [ ] DLQ

## Go Database Access

- [ ] `database/sql` (pool, `sql.Open` lazy, Ping)
- [ ] QueryRow/Query + Scan, `sql.ErrNoRows`
- [ ] `rows.Close()` + `rows.Err()`
- [ ] Placeholder chống SQL injection
- [ ] Transaction: `BeginTx` + `defer tx.Rollback()`
- [ ] Prepared statement
- [ ] sqlx / sqlc / GORM / ent trade-offs
- [ ] Migrations (golang-migrate / goose), backward-compatible

---

# 8. Week 6 – Distributed Systems

## CAP

- [ ] Consistency
- [ ] Availability
- [ ] Partition tolerance
- [ ] CAP under network partition

## Consistency

- [ ] Strong consistency
- [ ] Eventual consistency
- [ ] Read-after-write
- [ ] Causal consistency

## Distributed transactions

- [ ] 2PC
- [ ] Saga
- [ ] Outbox Pattern
- [ ] CDC

Example:

```text
Order Service
      ↓
Payment Service
      ↓
Inventory Service
```

Question:

> Payment succeeds but Inventory fails. How do you recover?

## Idempotency

Scenario:

```text
Client
  ↓
POST /payment
  ↓
Timeout
  ↓
Client retries
```

Question:

> How do you prevent double charging?

## Reliability patterns

- [ ] Retry
- [ ] Exponential backoff
- [ ] Circuit breaker
- [ ] Bulkhead
- [ ] Rate limiting
- [ ] Timeout
- [ ] Graceful degradation
- [ ] Dead letter queue

---

# 9. Week 7 – System Design

Practice one system per day.

## Day 1 – URL Shortener

- [ ] Requirements
- [ ] Scale estimation
- [ ] API
- [ ] Data model
- [ ] Architecture
- [ ] Cache
- [ ] Database
- [ ] Scaling

## Day 2 – Rate Limiter

- [ ] Token bucket
- [ ] Leaky bucket
- [ ] Sliding window
- [ ] Distributed rate limiting
- [ ] Redis

## Day 3 – Notification System

- [ ] Push notification
- [ ] Email
- [ ] SMS
- [ ] Queue
- [ ] Retry
- [ ] DLQ
- [ ] Deduplication

## Day 4 – Chat System

- [ ] WebSocket
- [ ] Connection management
- [ ] Redis Pub/Sub
- [ ] Message persistence
- [ ] Ordering
- [ ] Offline messages
- [ ] Horizontal scaling

## Day 5 – Payment System

- [ ] Idempotency
- [ ] Transaction
- [ ] Payment state machine
- [ ] Retry
- [ ] Reconciliation
- [ ] Audit log
- [ ] Distributed transaction

## Day 6 – E-commerce

- [ ] Product
- [ ] Cart
- [ ] Order
- [ ] Inventory
- [ ] Payment
- [ ] Shipping
- [ ] Event-driven architecture

## Day 7 – Live Streaming / Chat

Highly relevant to current Go backend experience.

- [ ] WebSocket / Socket.IO
- [ ] Connection manager
- [ ] Redis Pub/Sub
- [ ] Multi-instance architecture
- [ ] Message ordering
- [ ] Fan-out
- [ ] Backpressure
- [ ] Failure recovery

---

# 10. System Design Answer Framework

When interviewer asks:

> Design a notification system.

Use:

```text
1. Requirements
       ↓
2. Scale estimation
       ↓
3. API
       ↓
4. Data model
       ↓
5. High-level architecture
       ↓
6. Communication
       ↓
7. Storage
       ↓
8. Cache
       ↓
9. Failure handling
       ↓
10. Scaling
       ↓
11. Observability
```

Do not jump immediately into database/schema.

---

# 11. Week 8 – Mock Interview

## Mock 1 – Go

60 minutes:

- [ ] Go internals
- [ ] Concurrency
- [ ] Memory
- [ ] GC
- [ ] Error handling
- [ ] Context
- [ ] Coding

## Mock 2 – Backend

- [ ] REST API
- [ ] Database
- [ ] Redis
- [ ] Kafka
- [ ] Microservices
- [ ] Observability

## Mock 3 – System Design

Example:

> Design a payment system supporting 10K TPS.

Must cover:

- [ ] Requirements
- [ ] Capacity
- [ ] Architecture
- [ ] Data
- [ ] Consistency
- [ ] Idempotency
- [ ] Failure handling
- [ ] Scaling
- [ ] Monitoring

## Mock 4 – Production Troubleshooting

Scenario:

> Production latency increased from 100ms to 3s.

Approach:

```text
Metrics
   ↓
Logs
   ↓
Tracing
   ↓
Profile
   ↓
Database
   ↓
Network
   ↓
Root cause
   ↓
Fix
   ↓
Prevention
```

---

# 12. Senior Go Interview Question Bank

## Go Fundamentals

- [ ] Array vs Slice?
- [ ] Slice internals?
- [ ] Map internals?
- [ ] Interface internals?
- [ ] Nil interface?
- [ ] Pointer vs value?
- [ ] Method set?
- [ ] `defer` execution order?
- [ ] `panic/recover`?
- [ ] Generics?
- [ ] `errors.Is` vs `errors.As`? `%w` wrapping?
- [ ] `json:"-"` vs `omitempty`? Vì sao field phải exported?

## Foundations & Tooling

- [ ] go.mod vs go.sum?
- [ ] MVS (Minimal Version Selection)?
- [ ] Semantic Import Versioning (`/v2`)?
- [ ] Table-driven test + subtest?
- [ ] Mock trong Go (không có mock magic)?
- [ ] Fuzzing giải quyết gì?
- [ ] `go vet` vs `golangci-lint`?

## Concurrency

- [ ] What is a goroutine?
- [ ] How does GMP work?
- [ ] How does Go scheduler work?
- [ ] Channel vs Mutex?
- [ ] Buffered vs unbuffered channel?
- [ ] Race condition?
- [ ] Deadlock?
- [ ] Goroutine leak?
- [ ] Worker pool?
- [ ] Fan-in / fan-out?

## Runtime

- [ ] Stack growth?
- [ ] Heap allocation?
- [ ] Escape analysis?
- [ ] GC?
- [ ] STW?
- [ ] Allocation pressure?
- [ ] pprof?

## Networking

- [ ] HTTP keep-alive?
- [ ] Connection pool?
- [ ] TCP?
- [ ] HTTP/2?
- [ ] Context cancellation?
- [ ] Timeout?
- [ ] Retry?

## Database

- [ ] Index?
- [ ] Composite index?
- [ ] Query plan?
- [ ] Transaction?
- [ ] Isolation?
- [ ] MVCC?
- [ ] Deadlock?
- [ ] Connection pool?
- [ ] `sql.DB` là pool hay connection?
- [ ] `defer tx.Rollback()` pattern?
- [ ] sqlx vs GORM vs sqlc?
- [ ] Migration backward-compatible?

## gRPC & Security

- [ ] gRPC vs REST?
- [ ] Protobuf field number compatibility?
- [ ] 4 kiểu RPC?
- [ ] gRPC interceptor?
- [ ] JWT stateless — ưu/nhược, cách revoke?
- [ ] HS256 vs RS256?
- [ ] 401 vs 403?
- [ ] Password hashing (vì sao không SHA-256 trần)?

## Distributed Systems

- [ ] CAP?
- [ ] Consistency?
- [ ] Idempotency?
- [ ] Retry?
- [ ] Circuit breaker?
- [ ] Saga?
- [ ] Outbox?
- [ ] Distributed lock?

## Microservices

- [ ] API Gateway?
- [ ] Service discovery?
- [ ] Configuration?
- [ ] Health check?
- [ ] Graceful shutdown?
- [ ] Rate limiting?
- [ ] Bulkhead?
- [ ] Service-to-service communication?

## Observability

- [ ] Structured logging
- [ ] Metrics
- [ ] Distributed tracing
- [ ] Profiling
- [ ] OpenTelemetry
- [ ] Prometheus
- [ ] Grafana
- [ ] Jaeger
- [ ] ELK

---

# 13. Java → Go Mapping

| Java / Spring | Go |
|---|---|
| Thread | Goroutine |
| ExecutorService | Goroutine + Worker Pool |
| CompletableFuture | Goroutine + Channel |
| `synchronized` | Mutex |
| ConcurrentHashMap | Mutex / `sync.Map` |
| AtomicInteger | `sync/atomic` |
| Future | Channel |
| Spring DI | Explicit dependency injection |
| Filter / Interceptor | Middleware |
| `@Transactional` | DB transaction |
| ThreadLocal | Context / explicit passing |
| JVM GC | Go GC |
| JVM profiling | pprof |
| Spring Actuator | Metrics / pprof / health endpoints |

Important:

> Do not write Go as if it were Java. Understand idiomatic Go and explicit dependency management.

---

# 14. Priority

```text
Go Concurrency       ██████████
Go Runtime           ██████████
System Design        ██████████
Distributed System   ██████████
Database             █████████
Performance          █████████
Microservices        ████████
Observability        ████████
Go Core              ███████
Redis/Kafka          ███████
Coding               ███████
```

---

# 15. Final Preparation Checklist

Before interview, make sure you can explain without notes:

- [ ] Go GMP scheduler
- [ ] Goroutine lifecycle
- [ ] Channel internals and semantics
- [ ] Mutex vs channel
- [ ] Go memory model
- [ ] Escape analysis
- [ ] Stack vs heap
- [ ] GC
- [ ] pprof
- [ ] Context cancellation
- [ ] HTTP connection pooling
- [ ] Database indexing
- [ ] Transaction isolation
- [ ] Redis caching
- [ ] Kafka delivery semantics
- [ ] Idempotency
- [ ] Saga / Outbox
- [ ] Circuit breaker
- [ ] Rate limiting
- [ ] Distributed tracing
- [ ] System Design
- [ ] Production troubleshooting
- [ ] Go modules + MVS
- [ ] Table-driven test + mocking + fuzzing
- [ ] `errors.Is` / `errors.As` / `%w`
- [ ] encoding/json (tags, streaming, custom marshal)
- [ ] gRPC vs REST + protobuf compatibility
- [ ] JWT / OAuth2 + password hashing
- [ ] database/sql (pool, transaction pattern, ErrNoRows)

---

# 16. Recommended Interview Mindset

Senior interview is not only about:

> "Do you know the answer?"

It is about:

> "Can you reason about trade-offs and make a good engineering decision?"

For every design, practice answering:

1. Why this approach?
2. What are the alternatives?
3. What are the trade-offs?
4. What happens under failure?
5. How does it scale?
6. How do we monitor it?
7. How would you debug it in production?
8. What would you change at 10x scale?

---

# 17. Suggested Daily Study Format

Each study session:

```text
30 min  — Theory
30 min  — Deep dive / source code
30 min  — Coding
30 min  — Interview questions
30 min  — System design / troubleshooting
```

At the end of each day:

- [ ] Write down 5 key concepts
- [ ] Answer 5 interview questions without notes
- [ ] Implement at least 1 small coding exercise
- [ ] Explain one concept aloud as if interviewing

---

# 18. Final Goal

By the end of the roadmap, you should be able to:

```text
Understand Go deeply
        +
Write idiomatic concurrent Go
        +
Design scalable backend services
        +
Reason about distributed systems
        +
Debug production problems
        +
Explain engineering trade-offs
        +
Pass Senior Backend Engineer interview
```
