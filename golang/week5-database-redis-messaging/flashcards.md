# Week 5 – Flashcards / Q&A – Database + Redis + Messaging

## Database

### Q: B-Tree index — tại sao dùng B-Tree mà không phải binary tree?
**A:** B-Tree có branching factor lớn (mỗi node nhiều children):
- Giảm disk I/O (mỗi node = 1 page, ~4-16KB)
- Shallow tree (3-4 levels cho millions of rows)
- Binary tree sẽ rất deep → nhiều disk reads

---

### Q: Composite index leftmost prefix rule?
**A:** Index `(A, B, C)` có thể serve queries:
- `WHERE A = ?` ✓
- `WHERE A = ? AND B = ?` ✓
- `WHERE A = ? AND B = ? AND C = ?` ✓
- `WHERE B = ?` ✗ (skip A)
- `WHERE A = ? AND C = ?` ✗ (skip B, chỉ dùng A)

---

### Q: Covering index là gì?
**A:** Index chứa tất cả columns query cần → không cần table access (Index-Only Scan). Fastest possible query vì chỉ đọc index, không đọc table.

---

### Q: MVCC hoạt động thế nào?
**A:** 
- Mỗi transaction có snapshot (visibility rules)
- Row có multiple versions (xmin, xmax in PostgreSQL)
- Readers thấy version tương thích với snapshot
- Writers tạo version mới, không overwrite
- Old versions cleaned up bởi vacuum (PostgreSQL)

---

### Q: Isolation levels — giải thích Dirty Read, Non-repeatable Read, Phantom Read?
**A:**
- **Dirty Read**: đọc data uncommitted transaction khác
- **Non-repeatable Read**: đọc row 2 lần, giá trị khác (row updated by another)
- **Phantom Read**: query 2 lần, số rows khác (new rows inserted)

---

### Q: Connection pool sizing?
**A:**
- Formula: `(CPU cores × 2) + spindles`
- Practical: start 10-20, monitor
- Too many: DB overhead, context switching
- Too few: request queuing, latency
- Monitor: wait time, active count, idle count

---

### Q: Database deadlock — detect và prevent?
**A:**
- **Detect**: DB tự detect (timeout hoặc wait-for graph), abort one transaction
- **Prevent**:
  1. Lock consistent order
  2. Short transactions
  3. Use `SELECT ... FOR UPDATE NOWAIT`
  4. Retry logic with backoff

---

### Q: N+1 query problem?
**A:** 1 query lấy list + N queries cho mỗi item:
```sql
SELECT * FROM users;              -- 1 query
SELECT * FROM orders WHERE user_id = 1;  -- N queries
SELECT * FROM orders WHERE user_id = 2;
...
```
Fix: JOIN, batch IN clause, hoặc DataLoader pattern.

---

## Redis

### Q: Cache-Aside pattern?
**A:**
- Read: check cache → hit: return. miss: query DB → write cache → return
- Write: update DB → invalidate cache (NOT update cache)
- Most common pattern, application controls caching

---

### Q: Cache Stampede — problem và solution?
**A:**
- Problem: key expires → 1000 requests simultaneously hit DB
- Solutions:
  1. **Mutex/singleflight**: only 1 request loads, others wait
  2. **Early expiration**: refresh before actual expiry
  3. **Probabilistic refresh**: randomly refresh before expiry
  4. **Never expire + background refresh**

---

### Q: Redis distributed lock — implementation?
**A:**
```
SET key value NX PX 30000
```
- NX: only if not exists (acquire)
- PX: auto-expire (prevent deadlock if holder crashes)
- Value: unique ID (only holder can release)
- Release: Lua script checking value before DEL

---

### Q: Redis vs Memcached?
**A:**
| Feature | Redis | Memcached |
|---|---|---|
| Data structures | Rich (Hash, List, Set, ZSet) | String only |
| Persistence | Yes (RDB, AOF) | No |
| Pub/Sub | Yes | No |
| Clustering | Yes (native) | Client-side sharding |
| Memory efficiency | Less | More (slab allocator) |
| Use case | Cache + data store | Pure cache |

---

### Q: Hot key problem — solutions?
**A:**
1. **Local cache** (in-process): cache phía application, short TTL
2. **Key sharding**: split `hot_key` → `hot_key_1`, `hot_key_2`, ... (random read)
3. **Read replicas**: spread reads across replicas
4. **Rate limiting**: limit requests to hot key

---

## Kafka / Messaging

### Q: Kafka — ordering guarantee?
**A:**
- **Within partition**: total order guaranteed
- **Across partitions**: no ordering guarantee
- Same key → same partition → ordered
- Use case: order events for same order_id → use order_id as key

---

### Q: At-least-once vs Exactly-once?
**A:**
- **At-least-once**: commit offset AFTER processing. If crash after process but before commit → reprocess (duplicate)
- **Exactly-once**: idempotent producer + transactional consumer. Or: at-least-once + idempotent consumer (most practical)

---

### Q: Consumer group rebalancing — problem?
**A:**
- When consumer joins/leaves → partitions reassigned
- During rebalance: all consumers pause
- Problem: frequent rebalances = downtime
- Fix: proper heartbeat config, sticky assignor, cooperative rebalance

---

### Q: DLQ pattern?
**A:** Dead Letter Queue:
1. Consumer processes message
2. If fails → retry N times
3. After N retries → send to DLQ
4. DLQ: monitor, alert, manual review/replay

```
Main Topic → Consumer
              ├── Success → commit offset
              └── Failure (after retries) → DLQ Topic
                                             └── Alert + Manual intervention
```

---

### Q: Kafka vs RabbitMQ?
**A:**
| Feature | Kafka | RabbitMQ |
|---|---|---|
| Model | Log (append-only) | Queue (message consumed) |
| Ordering | Per-partition | Per-queue |
| Retention | Configurable (days/weeks) | Until consumed |
| Throughput | Very high (100K+ msg/s) | High (50K+ msg/s) |
| Use case | Event streaming, log | Task queue, RPC |
| Replay | Yes (offset reset) | No (once consumed) |

---

### Q: Idempotent consumer — implementation?
**A:**
```go
func processMessage(ctx context.Context, msg Message) error {
    // 1. Check if already processed
    processed, err := db.IsProcessed(ctx, msg.ID)
    if processed {
        return nil // skip duplicate
    }
    
    // 2. Process + mark as processed (in same transaction)
    tx, _ := db.Begin(ctx)
    defer tx.Rollback()
    
    err = processBusinessLogic(tx, msg)
    err = tx.MarkProcessed(msg.ID)
    
    return tx.Commit()
}
```
Key: store message ID, check before processing, use transaction.
