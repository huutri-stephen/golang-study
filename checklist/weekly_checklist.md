# ✅ Weekly Progress Checklist

## Week 0 – Foundations & Tooling

### Modules & Dependency Management
- [ ] go.mod / go.sum (checksum, supply-chain)
- [ ] go mod tidy / download / vendor
- [ ] Semantic versioning + `/v2` import path
- [ ] MVS (Minimal Version Selection)
- [ ] Go workspaces (`go work`)
- [ ] `replace` directive

### Testing (chiều sâu)
- [ ] Table-driven tests + subtests (`t.Run`)
- [ ] Mocking qua interface (gomock/testify/moq)
- [ ] Coverage (`-cover`, `-coverprofile`)
- [ ] Fuzzing (`go test -fuzz`)
- [ ] `t.Helper()`, `t.Cleanup()`, require vs assert
- [ ] Integration test (build tags, testcontainers)

### Tooling
- [ ] gofmt / goimports
- [ ] go vet
- [ ] staticcheck
- [ ] golangci-lint (CI)
- [ ] delve (dlv) debugger

### Error Handling hiện đại
- [ ] `fmt.Errorf` + `%w`
- [ ] `errors.Is` vs `errors.As`
- [ ] Sentinel vs custom error type
- [ ] `errors.Join`

### encoding/json
- [ ] Struct tags (`-`, `omitempty`)
- [ ] Marshal / Unmarshal / streaming Decoder
- [ ] DisallowUnknownFields, UseNumber
- [ ] Custom MarshalJSON

### Reflection & Ecosystem
- [ ] reflect cơ bản
- [ ] Config (flag > env > file)
- [ ] Structured logging (slog/zap/zerolog)

### Code Examples Completed
- [ ] error_handling.go
- [ ] json_patterns.go
- [ ] testing_patterns_test.go

---

## Week 1 – Go Core + Memory

### Go Language Fundamentals
- [ ] Variable, constant – hiểu zero value, type inference
- [ ] Pointer – khi nào dùng, pointer receiver vs value receiver
- [ ] Struct – composition, embedding
- [ ] Interface – internals, nil interface, typed nil
- [ ] Method / receiver – method set rules
- [ ] Embedding – promoted methods
- [ ] Slice – internals, len, cap, append behavior
- [ ] Array – value type, fixed size
- [ ] Map – hash table internals, not thread-safe
- [ ] String / rune / byte – UTF-8, iteration
- [ ] Error – error interface, wrapping, sentinel errors
- [ ] `defer` – LIFO order, evaluated at declaration
- [ ] `panic` / `recover` – when to use, when NOT to use
- [ ] Generics – type constraints, type inference

### Slice Deep Dive
- [ ] Slice header (pointer, len, cap)
- [ ] Underlying array sharing
- [ ] `append` reallocation behavior
- [ ] Memory leak với slice
- [ ] Copy vs re-slice

### Map Deep Dive
- [ ] Hash table + bucket structure
- [ ] Collision handling
- [ ] Map growth / evacuation
- [ ] Concurrent access problem
- [ ] `sync.Map` use cases

### Interface Deep Dive
- [ ] iface vs eface
- [ ] Dynamic dispatch
- [ ] Nil interface vs nil concrete value
- [ ] Method set (pointer vs value receiver)
- [ ] Interface composition

### Memory
- [ ] Stack allocation
- [ ] Heap allocation
- [ ] Escape analysis rules
- [ ] Value semantics vs pointer semantics
- [ ] When pointer escapes to heap

### Code Examples Completed
- [ ] slice_internals.go
- [ ] map_internals.go
- [ ] interface_internals.go
- [ ] memory_escape.go

---

## Week 2 – Goroutine + Concurrency

### Goroutine & Scheduler
- [ ] Goroutine vs OS thread (size, cost)
- [ ] M:N scheduling model
- [ ] GMP model (G, M, P)
- [ ] Local run queue vs global run queue
- [ ] Work stealing
- [ ] Syscall handling
- [ ] Network poller (epoll/kqueue)
- [ ] `GOMAXPROCS`

### Channel
- [ ] Unbuffered – synchronous communication
- [ ] Buffered – async with capacity
- [ ] Send/receive blocking rules
- [ ] Close semantics
- [ ] Range over channel
- [ ] Select statement
- [ ] Nil channel behavior
- [ ] Directional channels

### Synchronization
- [ ] `sync.Mutex` / `sync.RWMutex`
- [ ] `sync.WaitGroup`
- [ ] `sync.Once`
- [ ] `sync.Cond`
- [ ] `sync/atomic`
- [ ] `sync.Map`
- [ ] `errgroup`

### Concurrency Problems
- [ ] Data race – detection, prevention
- [ ] Deadlock – conditions, avoidance
- [ ] Livelock
- [ ] Starvation
- [ ] Goroutine leak – causes, prevention

### Patterns
- [ ] Worker pool
- [ ] Fan-in / fan-out
- [ ] Pipeline
- [ ] Semaphore
- [ ] Rate limiting with channels

### Code Examples Completed
- [ ] goroutine_basics.go
- [ ] channel_patterns.go
- [ ] sync_primitives.go
- [ ] worker_pool.go
- [ ] concurrency_problems.go

---

## Week 3 – Go Runtime + Performance

### Memory & Escape Analysis
- [ ] Stack growth (contiguous stack, copying)
- [ ] Heap allocation triggers
- [ ] Escape analysis rules
- [ ] `go build -gcflags="-m"`
- [ ] Pointer escaping scenarios
- [ ] Value vs pointer trade-offs

### Garbage Collector
- [ ] Tri-color marking algorithm
- [ ] Concurrent GC phases
- [ ] STW pauses (when, how long)
- [ ] GC pacing
- [ ] GOGC setting
- [ ] Write barrier

### Profiling
- [ ] CPU profiling
- [ ] Heap profiling
- [ ] Goroutine profiling
- [ ] Mutex contention profiling
- [ ] Block profiling
- [ ] `pprof` HTTP endpoint
- [ ] `go tool pprof`

### Benchmarking
- [ ] `go test -bench`
- [ ] `-benchmem`
- [ ] `b.ResetTimer()`
- [ ] `b.ReportAllocs()`
- [ ] Avoiding compiler optimizations

### Code Examples Completed
- [ ] escape_analysis.go
- [ ] benchmark_example_test.go
- [ ] pprof_example.go

---

## Week 4 – Backend Engineering

### HTTP & Networking
- [ ] HTTP/1.1 vs HTTP/2 vs HTTP/3
- [ ] TCP handshake, keep-alive
- [ ] TLS handshake
- [ ] Connection pooling
- [ ] Timeout hierarchy (dial, TLS, response, idle)

### Go HTTP
- [ ] `net/http` server internals
- [ ] `http.Client` + `Transport`
- [ ] Custom transport configuration
- [ ] Middleware pattern
- [ ] Graceful shutdown

### Context
- [ ] `context.WithCancel`
- [ ] `context.WithTimeout`
- [ ] `context.WithDeadline`
- [ ] `context.WithValue`
- [ ] Cancellation propagation
- [ ] Best practices

### API Design
- [ ] RESTful design principles
- [ ] Status code usage
- [ ] Error response format
- [ ] Pagination (cursor vs offset)
- [ ] Idempotency keys
- [ ] Versioning strategies
- [ ] Rate limiting headers

### gRPC & Protobuf
- [ ] Protobuf IDL, field number compatibility
- [ ] 4 kiểu RPC (unary, streaming)
- [ ] gRPC vs REST trade-offs
- [ ] Interceptor (auth, logging, tracing)
- [ ] Deadline propagation, status codes

### Security / Auth
- [ ] AuthN vs AuthZ (401 vs 403)
- [ ] JWT (HS256/RS256, exp, refresh token)
- [ ] OAuth2 / OIDC (Auth Code + PKCE)
- [ ] Password hashing (bcrypt/argon2), constant-time
- [ ] Input validation, CORS, body size limit

### Web Frameworks
- [ ] net/http + ServeMux (Go 1.22)
- [ ] chi / Gin / Echo / Fiber trade-offs

### Code Examples Completed
- [ ] http_server.go
- [ ] middleware.go
- [ ] graceful_shutdown.go
- [ ] context_usage.go
- [ ] jwt_auth.go
- [ ] payment.proto

---

## Week 5 – Database + Redis + Messaging

### SQL / PostgreSQL
- [ ] B-Tree index internals
- [ ] Composite index (leftmost prefix)
- [ ] Covering index
- [ ] `EXPLAIN ANALYZE`
- [ ] Transaction isolation levels
- [ ] MVCC
- [ ] Lock types (row, table, advisory)
- [ ] Deadlock detection
- [ ] Connection pool sizing
- [ ] N+1 query problem

### Redis
- [ ] Data structures (String, Hash, List, Set, ZSet)
- [ ] TTL / expiration
- [ ] Cache strategies
- [ ] Pub/Sub
- [ ] Stream
- [ ] Distributed lock (Redlock)
- [ ] Cache problems (stampede, penetration, avalanche)

### Kafka
- [ ] Producer (acks, retries)
- [ ] Consumer (offset, commit)
- [ ] Partition & ordering
- [ ] Consumer group rebalancing
- [ ] Delivery semantics
- [ ] DLQ pattern

### Go Database Access
- [ ] database/sql (pool, sql.Open lazy, Ping)
- [ ] QueryRow/Query + Scan, sql.ErrNoRows
- [ ] rows.Close() + rows.Err()
- [ ] Placeholder chống SQL injection
- [ ] Transaction: BeginTx + defer Rollback
- [ ] Prepared statement
- [ ] sqlx / sqlc / GORM / ent trade-offs
- [ ] Migrations (golang-migrate/goose)

### Code Examples Completed
- [ ] db_transaction.go
- [ ] redis_cache.go
- [ ] kafka_consumer.go
- [ ] db_access_patterns.go

---

## Week 6 – Distributed Systems

### Fundamentals
- [ ] CAP theorem
- [ ] Consistency models
- [ ] Distributed consensus
- [ ] Clock synchronization

### Patterns
- [ ] Saga pattern (choreography vs orchestration)
- [ ] Outbox pattern
- [ ] CDC (Change Data Capture)
- [ ] 2PC (Two-Phase Commit)
- [ ] Idempotency implementation

### Reliability
- [ ] Retry + exponential backoff + jitter
- [ ] Circuit breaker (states, thresholds)
- [ ] Bulkhead
- [ ] Rate limiting (token bucket, sliding window)
- [ ] Timeout cascading
- [ ] Graceful degradation

### Code Examples Completed
- [ ] circuit_breaker.go
- [ ] retry_backoff.go
- [ ] idempotency.go
- [ ] saga_pattern.go

---

## Week 7 – System Design

- [ ] Day 1: URL Shortener
- [ ] Day 2: Rate Limiter
- [ ] Day 3: Notification System
- [ ] Day 4: Chat System
- [ ] Day 5: Payment System
- [ ] Day 6: E-commerce
- [ ] Day 7: Live Streaming / Chat

### Each Design Must Cover
- [ ] Requirements (functional + non-functional)
- [ ] Scale estimation
- [ ] API design
- [ ] Data model
- [ ] High-level architecture
- [ ] Deep dive on critical component
- [ ] Failure handling
- [ ] Scaling strategy
- [ ] Observability

---

## Week 8 – Mock Interview + Revision

### Mock Sessions
- [ ] Mock 1: Go internals (60 min)
- [ ] Mock 2: Backend engineering (60 min)
- [ ] Mock 3: System design (45 min)
- [ ] Mock 4: Production troubleshooting (30 min)

### Final Review
- [ ] Go GMP scheduler
- [ ] Channel internals
- [ ] Escape analysis
- [ ] GC
- [ ] HTTP connection pooling
- [ ] Database indexing + isolation
- [ ] Redis caching strategies
- [ ] Kafka delivery semantics
- [ ] Idempotency
- [ ] Saga / Outbox
- [ ] Circuit breaker
- [ ] System Design framework
- [ ] Production troubleshooting approach
