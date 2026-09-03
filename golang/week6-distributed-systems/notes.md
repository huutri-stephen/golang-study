# Week 6 – Distributed Systems – Study Notes

## 1. CAP Theorem

```
        Consistency
           /\
          /  \
         /    \
        / CP   \
       /________\
      /   CA    \
     /____________\
Availability ---- Partition Tolerance
         AP
```

**In practice:** Network partition WILL happen → must choose between C and A:
- **CP**: Sacrifice availability during partition (banking, inventory)
- **AP**: Sacrifice consistency during partition (social media, cache)
- **CA**: Only possible in single-node (not distributed)

### Real-world Examples

| System | Choice | Trade-off |
|---|---|---|
| PostgreSQL (single) | CA | Not partition-tolerant |
| MongoDB | CP | Unavailable during leader election |
| Cassandra | AP | Eventually consistent |
| Redis Cluster | AP | May lose writes during partition |
| ZooKeeper | CP | Unavailable if quorum lost |

---

## 2. Consistency Models

### Strong Consistency
- Every read sees the latest write
- Like a single server
- Expensive: requires consensus (Raft, Paxos)
- Use: banking transactions, inventory count

### Eventual Consistency
- Reads may return stale data
- All replicas converge eventually
- Cheap: no coordination needed
- Use: social feed, analytics, recommendation

### Read-After-Write Consistency
- User always sees their own writes
- Others may see stale data
- Implementation: read from leader after write, or use session affinity

### Causal Consistency
- If A causes B, all observers see A before B
- Unrelated events may be seen in any order
- Implementation: vector clocks, causal timestamps

---

## 3. Distributed Transactions

### Two-Phase Commit (2PC)

```
Coordinator              Participants
    │                    │        │
    │── Prepare ────────→│        │
    │── Prepare ─────────────────→│
    │                    │        │
    │←── Vote Yes ───────│        │
    │←── Vote Yes ────────────────│
    │                    │        │
    │── Commit ─────────→│        │
    │── Commit ──────────────────→│
    │                    │        │
    │←── Ack ────────────│        │
    │←── Ack ─────────────────────│
```

**Problems:**
- Coordinator failure = blocking (participants wait forever)
- Performance: synchronous, all participants must be available
- Not partition-tolerant

### Saga Pattern

**Choreography (event-driven):**
```
Order Created → Payment Charged → Inventory Reserved → Shipping Created
                    ↓ (failure)
              Payment Refunded ← Inventory Released
```

**Orchestration (coordinator):**
```
Saga Orchestrator
    ├── Step 1: Create Order     → Compensate: Cancel Order
    ├── Step 2: Charge Payment   → Compensate: Refund
    ├── Step 3: Reserve Inventory→ Compensate: Release
    └── Step 4: Ship             → Compensate: Cancel Shipping
```

| Aspect | Choreography | Orchestration |
|---|---|---|
| Coupling | Loose (events) | Tighter (orchestrator knows flow) |
| Complexity | Distributed (hard to trace) | Centralized (easy to manage) |
| Single point of failure | No | Orchestrator |
| Best for | Simple flows | Complex multi-step flows |

### Outbox Pattern

Solves: "How to reliably publish event AND update DB atomically?"

```
┌─────────────────────────────────────┐
│  Transaction                         │
│  1. UPDATE orders SET status='paid'  │
│  2. INSERT INTO outbox (event)       │
│  COMMIT                              │
└─────────────────────────────────────┘
            ↓
    CDC / Poller reads outbox
            ↓
    Publish to Kafka
            ↓
    Mark outbox entry as published
```

**Why not publish directly?**
- DB commit + Kafka publish = not atomic
- If Kafka fails after DB commit → inconsistency
- Outbox ensures: if DB committed, event WILL be published

### Change Data Capture (CDC)

- Read database change log (WAL/binlog)
- Stream changes to message queue
- Tools: Debezium, Maxwell
- Benefit: no application code changes needed

---

## 4. Idempotency

### The Problem

```
Client → POST /payment → Server processes → Response lost
Client → POST /payment (retry) → Double charge!
```

### Implementation

```go
type IdempotencyStore interface {
    // Returns stored result if key exists
    Get(ctx context.Context, key string) (*Response, error)
    // Stores result with TTL
    Set(ctx context.Context, key string, resp *Response, ttl time.Duration) error
    // Acquire processing lock
    Lock(ctx context.Context, key string) (bool, error)
}

func ProcessPayment(ctx context.Context, key string, req PaymentRequest) (*PaymentResponse, error) {
    // 1. Check if already processed
    if cached, err := store.Get(ctx, key); err == nil {
        return cached, nil
    }
    
    // 2. Acquire lock (prevent concurrent processing of same key)
    acquired, err := store.Lock(ctx, key)
    if !acquired {
        return nil, ErrProcessing // tell client to wait
    }
    
    // 3. Process
    result, err := chargePayment(ctx, req)
    if err != nil {
        return nil, err
    }
    
    // 4. Store result
    store.Set(ctx, key, result, 24*time.Hour)
    
    return result, nil
}
```

### Idempotency Key Design

- Client generates unique key per operation
- Server stores: key → response mapping
- TTL: 24-48 hours (balance storage vs safety)
- Storage: Redis (fast) or DB (durable)

---

## 5. Reliability Patterns

### Retry with Exponential Backoff + Jitter

```
Attempt 1: immediate
Attempt 2: 1s + random(0, 1s)
Attempt 3: 2s + random(0, 2s)
Attempt 4: 4s + random(0, 4s)
...
Max delay: 30s
```

**Jitter prevents thundering herd:** Without jitter, all retries happen at same time.

### Circuit Breaker

```
     ┌────────┐    failures > threshold    ┌────────┐
     │ CLOSED │ ─────────────────────────→ │  OPEN  │
     │(normal)│                            │(reject)│
     └────────┘                            └────────┘
          ↑                                     │
          │ success                   timeout   │
          │                                     ↓
          │         ┌──────────┐               │
          └─────────│HALF-OPEN │←──────────────┘
                    │(test one)│
                    └──────────┘
```

**States:**
- **Closed**: Normal operation, count failures
- **Open**: Reject immediately (fast fail), after timeout → half-open
- **Half-Open**: Allow one request, if success → closed, if fail → open

### Bulkhead

Isolate failures to prevent cascading:

```
┌──────────────────────────────────┐
│ Service                          │
│  ┌────────────┐ ┌────────────┐  │
│  │ Pool: Auth │ │ Pool: Data │  │
│  │ (10 conns) │ │ (20 conns) │  │
│  └────────────┘ └────────────┘  │
│                                  │
│  If Auth pool exhausted,         │
│  Data pool still works!          │
└──────────────────────────────────┘
```

### Rate Limiting Algorithms

**Token Bucket:**
- Bucket holds N tokens
- Each request takes 1 token
- Tokens refill at rate R/second
- Allows bursts (up to bucket size)

**Sliding Window:**
- Count requests in rolling time window
- More accurate than fixed window
- Higher memory (per-request timestamp)

---

## 6. Consistency Patterns for Microservices

### Event Sourcing

```
Events (source of truth):
  OrderCreated{id=1, items=[A,B]}
  ItemRemoved{id=1, item=B}
  OrderConfirmed{id=1}

Current State (derived):
  Order{id=1, items=[A], status=confirmed}
```

Benefits: full audit trail, temporal queries, replay
Challenges: eventual consistency, complexity, storage

### CQRS (Command Query Responsibility Segregation)

```
Commands (writes)          Queries (reads)
     │                         │
     ↓                         ↓
Write Model              Read Model
(normalized)             (denormalized, optimized)
     │                         ↑
     └── Events ──────────────┘
```

Separate write and read models:
- Write: optimized for consistency
- Read: optimized for query performance
- Sync via events (eventual consistency)
