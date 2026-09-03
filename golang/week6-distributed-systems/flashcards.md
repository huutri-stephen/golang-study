# Week 6 – Flashcards / Q&A – Distributed Systems

## CAP Theorem

### Q: CAP theorem — giải thích và ví dụ thực tế?
**A:** Distributed system chỉ guarantee 2 trong 3:
- **C** (Consistency): every read returns latest write
- **A** (Availability): every request gets response
- **P** (Partition tolerance): works despite network failure

Reality: P is mandatory → choose CP or AP:
- CP: banking (reject requests if can't guarantee consistency)
- AP: social feed (show potentially stale data)

---

### Q: Eventual consistency — khi nào acceptable?
**A:** Khi:
- Business tolerates stale data temporarily
- User won't notice inconsistency
- Examples: social feed, analytics, recommendation, cache
- NOT acceptable: payment, inventory, authentication

---

## Distributed Transactions

### Q: 2PC problems?
**A:**
1. **Blocking**: coordinator fails → participants wait indefinitely
2. **Single point of failure**: coordinator crash
3. **Performance**: synchronous, all participants must respond
4. **Not partition-tolerant**: network split = stuck

Alternative: Saga pattern (compensating transactions)

---

### Q: Saga choreography vs orchestration?
**A:**
- **Choreography**: services emit events, each service listens and reacts
  - Pro: loosely coupled, no single point of failure
  - Con: hard to trace flow, implicit dependencies
- **Orchestration**: central coordinator directs the flow
  - Pro: easy to understand, centralized error handling
  - Con: orchestrator is single point of failure, coupling

Choose orchestration for complex flows (>3 steps).

---

### Q: Payment succeeds but Inventory fails — how to recover?
**A:** Saga compensation:
1. Inventory service fails → emit `InventoryReservationFailed` event
2. Payment service listens → execute compensating action: `RefundPayment`
3. Order service listens → update status to `cancelled`

Implementation:
- Each step has a compensating action defined
- On failure, execute compensations in reverse order
- Compensations must be idempotent!

---

### Q: Outbox pattern — tại sao cần?
**A:** Problem: DB update + message publish = not atomic
- If publish fails after DB commit → lost event
- If DB fails after publish → phantom event

Solution: 
1. Write both DB change + event to same DB in 1 transaction
2. Background process reads outbox → publishes to message queue
3. Guarantees: if DB committed, event will eventually be published

---

## Idempotency

### Q: How to prevent double charging on retry?
**A:**
1. Client sends `Idempotency-Key` header (UUID)
2. Server checks: key already processed?
   - Yes → return cached response
   - No → process, store response, return
3. Use distributed lock to prevent concurrent processing of same key
4. TTL on stored responses (24-48h)

---

### Q: Idempotent operations?
**A:**
- **Naturally idempotent**: GET, PUT (absolute), DELETE
- **NOT idempotent**: POST, PATCH (relative, e.g., `balance += 10`)
- Make non-idempotent idempotent: idempotency key, deduplication ID, unique constraint

---

## Reliability Patterns

### Q: Circuit breaker — states và transitions?
**A:**
- **Closed** (normal): count failures. If threshold exceeded → Open
- **Open** (fast-fail): reject all requests. After timeout → Half-Open
- **Half-Open** (test): allow 1 request. Success → Closed. Fail → Open

Key configs: failure threshold, timeout duration, success threshold (half-open → closed)

---

### Q: Retry — khi nào KHÔNG nên retry?
**A:**
- 4xx errors (client error, won't succeed with retry)
- 400 Bad Request (invalid input)
- 401/403 (auth problems)
- 409 Conflict (state won't change)
- Non-idempotent operations without idempotency key
- Circuit breaker is open

Always retry: 5xx, timeout, network error (transient)

---

### Q: Exponential backoff + jitter — tại sao cần jitter?
**A:** Without jitter:
- 100 clients timeout at same time
- All retry after 1s → thundering herd on server
- Server overloaded again → cycle repeats

With jitter: random spread reduces coordinated retries.
```
delay = min(base * 2^attempt + random(0, base * 2^attempt), maxDelay)
```

---

### Q: Rate limiting — Token Bucket vs Sliding Window?
**A:**
- **Token Bucket**: allows bursts up to bucket size, refills at constant rate. Simple, memory efficient.
- **Sliding Window**: counts requests in rolling time window. More accurate, no burst allowance, higher memory.

Redis implementation:
- Token bucket: DECR + TTL
- Sliding window: sorted set with timestamps

---

### Q: Graceful degradation examples?
**A:**
1. Recommendation service down → show popular items instead
2. Payment slow → queue order for later processing
3. Cache down → serve from DB (slower but works)
4. Third-party API down → use cached/stale data
5. Non-critical feature → disable and return empty response

Key: identify critical vs non-critical paths, have fallbacks.

---

## Architecture

### Q: Event Sourcing — benefits và challenges?
**A:**
Benefits:
- Complete audit trail (every change recorded)
- Time travel (rebuild state at any point)
- Event replay (fix bugs, rebuild projections)
- Natural fit for event-driven architecture

Challenges:
- Eventual consistency
- Event schema evolution
- Storage growth
- Complex querying (need projections/CQRS)

---

### Q: CQRS — khi nào dùng?
**A:**
Use when:
- Read patterns very different from write patterns
- Need separate scaling for reads vs writes
- Complex read queries that don't map to write model
- Event sourcing (natural companion)

Don't use when:
- Simple CRUD application
- Single team, simple domain
- Adding unnecessary complexity

---

### Q: Distributed lock — problems?
**A:**
1. **Clock skew**: TTL-based locks may expire early/late
2. **Network partition**: client holds lock but can't reach other nodes
3. **GC pauses**: client paused → lock expires → another client acquires
4. **Fencing**: even after lock expires, old holder may still write

Solution: Fencing tokens — monotonically increasing token, storage rejects lower tokens.

---

### Q: How to design a system that handles failures gracefully?
**A:** Defense-in-depth:
1. **Timeouts**: on every external call
2. **Retries**: with exponential backoff + jitter
3. **Circuit breaker**: prevent cascading failures
4. **Bulkhead**: isolate failure domains
5. **Rate limiting**: protect from overload
6. **Fallbacks**: graceful degradation
7. **Health checks**: detect problems early
8. **Observability**: metrics + logs + traces for debugging
