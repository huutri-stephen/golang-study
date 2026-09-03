# Week 8 – Flashcards / Final Review Q&A

## Go Internals — Must Answer Perfectly

### Q: Explain slice internals in 30 seconds.
**A:** Slice is a 3-field struct: pointer to underlying array, len, cap. Append within cap modifies shared array. Append beyond cap allocates new array (double if <256, +25% otherwise). Sub-slices share underlying array — use full slice expression `a[low:high:max]` to prevent accidental sharing.

---

### Q: Why is Go map not thread-safe? What happens with concurrent access?
**A:** Map is not designed for concurrent access because adding synchronization to every map operation would be expensive for the common single-goroutine case. Concurrent read+write or write+write causes `fatal error: concurrent map read and map write` — a runtime panic, not a data race. Fix: sync.Mutex, sync.RWMutex, or sync.Map (for stable-key read-heavy cases).

---

### Q: Explain the nil interface trap.
**A:** Interface is nil only when BOTH type and value are nil. If you assign a typed nil pointer to an interface, the interface's type field is set (non-nil) even though the value is nil → comparing to nil returns false. Common bug: returning `*MyError(nil)` where error interface is expected. Fix: return `nil` directly when no error.

---

### Q: How does Go GMP scheduler work?
**A:** G (goroutine) = unit of work. M (machine) = OS thread. P (processor) = logical CPU with local run queue (max 256 G). GOMAXPROCS = number of P. Flow: P picks G from local queue → M executes G. Work stealing: idle P steals half from other P. Syscall: M blocks, P detaches and finds new M. Network I/O: non-blocking via netpoller. Preemption: signal-based (Go 1.14+).

---

### Q: Channel internals — what happens when you send to unbuffered channel?
**A:** hchan struct contains: buffer (circular), sendq/recvq (waiting goroutines), lock. Unbuffered send: if receiver waiting → direct copy from sender stack to receiver stack (no buffer involved). If no receiver → sender goroutine parked in sendq. Reverse for receive. This is why unbuffered channels are synchronization points.

---

### Q: Escape analysis — name 5 scenarios where variable escapes to heap.
**A:**
1. Return pointer to local variable
2. Assign to interface (boxing)
3. Closure captures variable
4. Send address through channel
5. Slice/map with runtime-determined size
6. (Bonus) Too large for stack

Check: `go build -gcflags="-m"`

---

### Q: Go GC — explain tri-color marking and STW.
**A:** Objects colored White (unvisited), Gray (visited, children pending), Black (fully scanned). Start: roots → gray. Loop: pick gray → scan children → mark gray → mark self black. End: all white = garbage. STW happens twice: mark setup (~10-100μs) and mark termination (~10-100μs). Marking itself is concurrent (runs alongside program). Write barrier ensures correctness during concurrent marking.

---

## Backend — Must Explain Clearly

### Q: How would you configure Go HTTP client for production?
**A:**
```go
client := &http.Client{
    Timeout: 30*time.Second,
    Transport: &http.Transport{
        MaxIdleConns: 100,
        MaxIdleConnsPerHost: 10,
        MaxConnsPerHost: 50,
        IdleConnTimeout: 90*time.Second,
        TLSHandshakeTimeout: 5*time.Second,
    },
}
```
Key: always close response body, set all timeouts, configure pool size based on expected concurrency. Never use default client in production.

---

### Q: Graceful shutdown — step by step?
**A:**
1. Catch SIGTERM/SIGINT
2. Stop accepting new connections (health check returns 503)
3. Wait for in-flight requests to complete (with timeout)
4. Close server connections
5. Cleanup resources (DB pool, cache, flush metrics)
6. Exit

In Kubernetes: SIGTERM → preStop hook → terminationGracePeriodSeconds → SIGKILL

---

### Q: Context best practices — what are the rules?
**A:**
1. First parameter: `func(ctx context.Context, ...)`
2. Never store in struct
3. Always defer cancel
4. Values for request-scoped data only (trace ID, user ID)
5. Use unexported key types
6. Never pass nil → use Background() or TODO()
7. Propagate to all downstream calls

---

## Distributed Systems — Must Reason About Trade-offs

### Q: You're designing a payment system. How do you prevent double charging?
**A:** Idempotency key pattern:
1. Client generates unique key per payment intent
2. Server flow: check store → lock key → process → store result → respond
3. Duplicate request: return cached response (no re-processing)
4. Concurrent same-key: distributed lock rejects second request
5. Failed processing: remove from store (allow clean retry)
6. Store: Redis with 24h TTL (or DB for durability)

---

### Q: Saga vs 2PC — when to use which?
**A:**
- **2PC**: strong consistency, blocking, not partition-tolerant. Use for: single database across tables.
- **Saga**: eventual consistency, non-blocking, fault-tolerant. Use for: microservices with separate databases.

Saga compensations must be idempotent. Use orchestration for complex flows (>3 steps), choreography for simple flows.

---

### Q: Circuit breaker — explain the states and when you'd implement one.
**A:**
- **Closed**: normal, count failures. Threshold exceeded → Open.
- **Open**: fast-fail all requests (protect downstream). After timeout → Half-Open.
- **Half-Open**: allow 1 request. Success → Closed. Fail → Open.

Implement when: calling external service that may fail, want to prevent cascading failures, want fast-fail instead of timeout-wait.

---

## System Design — Must Structure Well

### Q: Design a real-time chat system.
**A (structured):**

**Requirements**: 1:1 chat, group chat (up to 500), online status, message history, 50M DAU.

**Scale**: 50M DAU, ~10 messages/user/day = 500M messages/day ≈ 6K msg/sec. Peak: 20K msg/sec.

**Architecture**:
- WebSocket gateway (Go) → handles persistent connections
- Message service → validates, stores, routes
- Redis Pub/Sub → cross-server message delivery
- PostgreSQL/Cassandra → message persistence
- Redis → online status, typing indicators

**Key decisions**:
- Message ordering: per-conversation sequence number
- Multi-server: Redis Pub/Sub channels per user
- Offline delivery: queue messages, deliver on reconnect
- Group chat: fan-out write (small groups) vs fan-out read (large rooms)

---

## Production Troubleshooting — Must Have Methodology

### Q: Production P99 latency jumped from 100ms to 3s. Walk me through debugging.
**A:**
1. **When?** Check metrics — sudden or gradual? Correlate with deploys, traffic spikes.
2. **Where?** Distributed tracing — which service, which endpoint?
3. **What changed?** Recent deployments, config changes, traffic pattern?
4. **Narrow down**: Is it CPU? Memory? I/O? Network?
   - CPU profile: hot functions
   - Heap profile: memory pressure → GC
   - Goroutine profile: goroutine leak → contention
5. **Database**: Connection pool stats, slow queries, lock waits
6. **Network**: DNS, TCP connections in TIME_WAIT, downstream latency
7. **Root cause** → fix → verify → post-mortem

---

## Senior-Level Thinking

### Q: How do you decide between approaches when designing a system?
**A:** Evaluate on these axes:
1. **Correctness**: Does it solve the actual problem?
2. **Simplicity**: Can the team understand and maintain it?
3. **Scalability**: Does it work at 10x load?
4. **Reliability**: What happens when components fail?
5. **Operability**: How do we deploy, monitor, debug?
6. **Cost**: Infrastructure, development time, maintenance

Start simple. Optimize when metrics prove you need to. Document trade-offs for future reference.

---

### Q: What makes a senior engineer different from a mid-level?
**A:**
- **Scope**: Think beyond the ticket — consider system-wide impact
- **Trade-offs**: Articulate why one approach over another
- **Failure thinking**: Design for what goes wrong, not just happy path
- **Communication**: Explain complex ideas simply to different audiences
- **Mentoring**: Help others grow, share knowledge proactively
- **Ownership**: Drive decisions, own outcomes, lead post-mortems
- **Pragmatism**: Ship working solutions, not perfect ones
