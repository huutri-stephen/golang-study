# Week 7 – Flashcards / Q&A – System Design

## Framework

### Q: System design interview — first 5 minutes nên làm gì?
**A:**
1. Clarify functional requirements (what does the system do?)
2. Clarify non-functional requirements (scale, latency, availability)
3. Identify key constraints (read/write ratio, data size, geography)
4. State assumptions and confirm with interviewer

NEVER jump into database schema or tech stack first.

---

### Q: Scale estimation — quick math?
**A:**
- 1 day ≈ 100K seconds
- 1M users = ~10 QPS (if each makes 1 request/day)
- Peak = 3-5x average
- 1 server ≈ 1K-10K QPS (depends on work)
- 1 year storage: daily_size × 365

---

## URL Shortener

### Q: URL shortening — how to generate short ID?
**A:**
- Option 1: Counter + Base62 (sequential, predictable)
- Option 2: Hash (MD5/SHA) first 7 chars (may collide, check DB)
- Option 3: Snowflake ID + Base62 (distributed, unique)

7 chars Base62 = 62^7 = 3.5 trillion combinations.

---

### Q: 301 vs 302 redirect for URL shortener?
**A:**
- **301 Permanent**: browser caches, doesn't hit our server again → better performance, lose analytics
- **302 Temporary**: browser always hits our server → more control, track clicks

Choose based on use case: analytics → 302, pure redirect → 301.

---

## Rate Limiter

### Q: Token Bucket vs Sliding Window?
**A:**
- **Token Bucket**: allows bursts (tokens accumulate), simple, low memory
- **Sliding Window**: strict rate, no bursts, higher memory (per-request tracking)

Token bucket: good for APIs that allow burst. Sliding window: strict rate enforcement.

---

### Q: Distributed rate limiting — race condition?
**A:** Multiple instances check+increment non-atomically:
```
Instance A: GET count → 99 (< 100 limit)
Instance B: GET count → 99 (< 100 limit)
Instance A: INCR → 100
Instance B: INCR → 101 (exceeded!)
```
Fix: Redis Lua script (atomic check+increment) or MULTI/EXEC.

---

## Chat System

### Q: WebSocket scaling — how to handle multi-server?
**A:**
1. User connects to any gateway server
2. Message arrives for user on Server A, but user is on Server B
3. Solution: Redis Pub/Sub or dedicated message broker
   - Publish message to channel for target user
   - Server B subscribes to that channel → delivers

Alternative: Consistent hashing (user always connects to same server).

---

### Q: Chat ordering guarantee?
**A:**
- Assign sequence number per conversation
- Server generates monotonically increasing seq for each room
- Client sorts by seq number
- Gap detection: if seq jumps, request missing messages

---

### Q: Offline messages?
**A:**
1. User disconnects
2. Messages stored in persistent queue (per user)
3. On reconnect: deliver all queued messages
4. Client acknowledges receipt → remove from queue
5. Limit: store last N messages or last 7 days

---

## Payment System

### Q: How to prevent double charging?
**A:**
1. **Idempotency key**: client sends unique key per payment intent
2. **Server checks**: key already processed? → return cached result
3. **Distributed lock**: prevent concurrent processing of same key
4. **State machine**: payment can only transition CREATED→PROCESSING once

---

### Q: Payment reconciliation?
**A:**
- Daily job compares internal records vs PSP (Stripe/etc)
- Match by transaction ID, amount, status
- Mismatches → investigation:
  - PSP succeeded but local failed → mark as succeeded
  - PSP failed but local succeeded → refund
  - Missing on either side → alert

---

## Notification System

### Q: How to handle notification at scale?
**A:**
Architecture: API → Priority Queue → Router → Channel Workers

Key components:
1. **Priority queue**: urgent (OTP) vs regular (marketing)
2. **Rate limiting**: max N notifications per user per hour
3. **Channel routing**: user preferences (push/email/SMS)
4. **Deduplication**: same notification within window → skip
5. **Retry + DLQ**: failed delivery → retry 3x → DLQ
6. **Template engine**: personalize content

---

## E-commerce

### Q: Flash sale — how to prevent overselling?
**A:**
1. **Pre-warm**: load stock count into Redis
2. **Rate limit**: queue excess requests
3. **Atomic decrement**: `DECR stock_key` (Redis atomic)
4. **If DECR < 0**: sold out, return immediately
5. **Create order**: only for users who got stock
6. **TTL on reservation**: if not paid in 10min → release stock

---

### Q: Cart to Order — consistency?
**A:**
1. User clicks "checkout"
2. **Lock inventory** (SELECT FOR UPDATE or Redis DECR)
3. Create order in DB
4. Start payment
5. If payment fails → release inventory
6. Saga pattern for distributed services

---

## General

### Q: SQL vs NoSQL — how to choose?
**A:**
| Factor | SQL | NoSQL |
|---|---|---|
| Schema | Fixed, structured | Flexible, schema-less |
| Relationships | Complex joins | Denormalized |
| Consistency | Strong (ACID) | Eventually consistent |
| Scale | Vertical (+ read replicas) | Horizontal (sharding) |
| Use case | Transactions, relational | High write, big data |

Examples:
- Users, orders, payments → SQL (relationships, ACID)
- Chat messages, logs, analytics → NoSQL (scale, flexible schema)

---

### Q: Microservice communication — sync vs async?
**A:**
- **Sync (HTTP/gRPC)**: simple, immediate response, coupling
- **Async (Message Queue)**: decoupled, resilient, eventual consistency

Use sync for: real-time response needed (API gateway → service)
Use async for: background work, cross-service events, reliability needed

---

### Q: Database sharding — key considerations?
**A:**
1. **Shard key**: determines data distribution
   - Good: user_id (even distribution)
   - Bad: country (uneven, hot shards)
2. **Cross-shard queries**: expensive (avoid or use materialized views)
3. **Resharding**: painful (virtual shards help)
4. **Join across shards**: not possible (denormalize or application-level)

---

### Q: Caching — where to cache?
**A:**
```
Browser cache → CDN → API Gateway cache → Application cache → DB
```
Each level: closer to user = faster, less control.
- CDN: static assets, public API responses
- Application: hot data, computed results
- Redis: shared across instances, session, rate limit
