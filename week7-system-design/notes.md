# Week 7 – System Design – Study Notes

## System Design Answer Framework

### Step-by-Step Process (ALWAYS follow this)

```
1. Requirements (5 min)
   ├── Functional requirements (what)
   └── Non-functional requirements (how well)

2. Scale Estimation (3 min)
   ├── Users (DAU, peak)
   ├── Traffic (QPS, peak QPS)
   ├── Storage (per day, per year)
   └── Bandwidth

3. API Design (3 min)
   └── REST/gRPC endpoints

4. Data Model (5 min)
   ├── Entities and relationships
   └── Database choice (SQL vs NoSQL)

5. High-Level Architecture (10 min)
   ├── Components
   ├── Communication
   └── Data flow

6. Deep Dive (10-15 min)
   ├── Critical component details
   ├── Algorithms/data structures
   └── Trade-offs

7. Scaling & Reliability (5 min)
   ├── Horizontal scaling
   ├── Caching
   ├── Replication
   └── Failure handling

8. Observability (2 min)
   ├── Metrics
   ├── Logging
   └── Alerting
```

---

## Quick Scale Reference

| Metric | Calculation |
|---|---|
| 1 million users | ~100 QPS (1M / 10K seconds/day) |
| 10 million users | ~1,000 QPS |
| 100 million users | ~10,000 QPS |
| 1 KB * 1M users/day | 1 GB/day, 365 GB/year |
| 1 MB * 1M uploads/day | 1 TB/day |

**Useful numbers:**
- 1 day = 86,400 sec ≈ 100K sec
- 1 server handles ~1000-10000 QPS (depending on work)
- Read-heavy: cache hit ratio 90-99%
- Write-heavy: consider eventual consistency

---

## Design 1: URL Shortener

### Requirements
- Shorten URL: long → short (7-8 chars)
- Redirect: short → long (301/302)
- Scale: 100M URLs/month, 10:1 read/write ratio
- TTL: optional expiration

### Scale
- Write: 100M/month = ~40 QPS
- Read: 400 QPS (10:1 ratio)
- Storage: 100M * 1KB = 100GB/month

### Key Decisions
- **ID generation**: Base62 encoding of auto-increment/snowflake ID
- **Hash collision**: MD5/SHA → take first 7 chars → check collision → retry
- **Database**: SQL (simple, ACID) or Redis (fast reads)
- **Cache**: Redis for hot URLs (90%+ hit ratio)
- **Redirect**: 301 (permanent) vs 302 (temporary) — affects analytics

---

## Design 2: Rate Limiter

### Algorithms
1. **Token Bucket**: burst-friendly, simple
2. **Leaky Bucket**: smooth rate, no bursts
3. **Fixed Window**: simple, boundary spike problem
4. **Sliding Window Log**: accurate, memory heavy
5. **Sliding Window Counter**: balanced

### Distributed Rate Limiting
- Redis-based (atomic operations)
- Race condition: MULTI/EXEC or Lua script
- Per-user, per-IP, per-API key

```lua
-- Redis Lua for sliding window
local key = KEYS[1]
local window = ARGV[1]
local limit = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

redis.call('ZREMRANGEBYSCORE', key, 0, now - window)
local count = redis.call('ZCARD', key)
if count < limit then
    redis.call('ZADD', key, now, now .. math.random())
    redis.call('EXPIRE', key, window)
    return 1
end
return 0
```

---

## Design 3: Notification System

### Architecture
```
API → Priority Queue → Router → Channel Workers → Delivery
                                  ├── Push (FCM/APNS)
                                  ├── Email (SES/SendGrid)
                                  ├── SMS (Twilio)
                                  └── In-App (WebSocket)
```

### Key Components
- **Rate limiting**: per-user, per-channel
- **Template engine**: personalized content
- **Deduplication**: prevent spam
- **Retry with DLQ**: failed deliveries
- **User preferences**: opt-in/opt-out per channel
- **Priority queue**: urgent vs regular

---

## Design 4: Chat System

### Architecture
```
Client ←→ WebSocket Gateway ←→ Message Service → Database
                ↕                      ↕
           Redis Pub/Sub         Message Queue
```

### Key Decisions
- **Protocol**: WebSocket (bidirectional, persistent)
- **Connection management**: gateway per region
- **Multi-device**: fan-out to all user devices
- **Ordering**: per-conversation sequence numbers
- **Offline messages**: store & forward on reconnect
- **Read receipts**: async via message queue
- **Group chat**: fan-out vs pull model

### Scaling WebSocket
- Each server handles ~100K connections
- Redis Pub/Sub for cross-server messaging
- Consistent hashing for user → server mapping

---

## Design 5: Payment System

### Key Requirements
- **Exactly-once processing** (idempotency)
- **ACID transactions** (double-entry bookkeeping)
- **Audit trail** (every state change logged)
- **Reconciliation** (match with payment provider)

### Payment State Machine
```
CREATED → PROCESSING → SUCCEEDED
                    → FAILED
                    → TIMEOUT → RECONCILIATION → SUCCEEDED/FAILED
```

### Architecture
```
API → Idempotency Check → Payment Processor → PSP (Stripe/etc)
         ↕                       ↕
   Idempotency Store      Transaction DB
                                 ↕
                          Reconciliation Job
```

### Critical Design Points
- Double-entry bookkeeping (debit + credit)
- Idempotency key on every write operation
- Outbox pattern for event publishing
- Reconciliation job matches records with PSP
- Distributed lock per payment (prevent double processing)

---

## Design 6: E-commerce (Order System)

### Services
```
Product Service → Cart Service → Order Service → Payment Service
                                       ↕
                               Inventory Service
                                       ↕
                               Shipping Service
```

### Key Challenges
- **Inventory**: optimistic locking / reserved stock
- **Cart → Order**: atomic transition (prevent overselling)
- **Distributed transaction**: Saga pattern
- **Flash sale**: rate limiting + inventory lock + queue

---

## Design 7: Live Streaming Chat

### Architecture (highly relevant for Go backend)
```
Client ←→ WebSocket Gateway (Go)
                ↕
          Message Router (Go)
                ↕
          Redis Pub/Sub (cross-instance)
                ↕
          Message Store (persistence)
```

### Key Design Points
- **Connection manager**: track user → connection mapping
- **Room-based fan-out**: send to all users in room
- **Backpressure**: slow consumers → buffer → drop oldest
- **Horizontal scaling**: multiple gateway instances
- **Hot room**: single room with millions of users → shard

### Go-specific Considerations
- Goroutine per connection (lightweight)
- Buffered channel per connection (write queue)
- Graceful connection cleanup (detect disconnects)
- Zero-copy message broadcasting

---

## Common Patterns Across All Designs

### Caching Layer
```
Client → CDN → API Gateway → Application Cache → Database
                                    ↕
                              Redis Cluster
```

### Message Queue Pattern
```
Producer → Queue → Consumer → Store
              ↓
         DLQ (failures)
```

### Database Sharding
```
Shard Key: user_id % N
Shard 1: users 0-999
Shard 2: users 1000-1999
...
```

### Load Balancing
- L4 (TCP): fast, no HTTP awareness
- L7 (HTTP): path routing, health checks, sticky sessions

---

## Interview Tips

1. **Start broad, go deep** — don't jump into details
2. **State assumptions** — "Assuming 10M DAU..."
3. **Discuss trade-offs** — "We could use X but Y is better because..."
4. **Draw diagrams** — boxes and arrows
5. **Address failure** — "What if this component fails?"
6. **Show senior thinking** — monitoring, deployment, team considerations
