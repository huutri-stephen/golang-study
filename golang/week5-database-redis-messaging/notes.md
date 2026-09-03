# Week 5 – Database + Redis + Messaging – Study Notes

## 1. Database Indexing

### B-Tree Index

```
                    [50]
                   /    \
          [20, 35]       [70, 85]
         /   |   \      /   |   \
    [10,15] [25,30] [40,45] [60,65] [75,80] [90,95]
                            (leaf nodes contain row pointers)
```

**Key properties:**
- Balanced: all leaves same depth
- Sorted: enables range queries
- O(log n) lookup, insert, delete
- Leaf nodes linked: efficient range scans

### Composite Index

```sql
CREATE INDEX idx_user_status_created ON orders(user_id, status, created_at);
```

**Leftmost Prefix Rule:**
- ✓ `WHERE user_id = 1`
- ✓ `WHERE user_id = 1 AND status = 'active'`
- ✓ `WHERE user_id = 1 AND status = 'active' AND created_at > '2024-01-01'`
- ✗ `WHERE status = 'active'` (skips user_id)
- ✗ `WHERE created_at > '2024-01-01'` (skips user_id, status)

### Covering Index

Index chứa tất cả columns cần cho query → không cần access table (Index-Only Scan):

```sql
-- Index: (user_id, status, created_at)
-- Covering query (all columns in index):
SELECT status, created_at FROM orders WHERE user_id = 1;
-- → Index-Only Scan (fast!)

-- NOT covering (needs table access):
SELECT * FROM orders WHERE user_id = 1;
-- → Index Scan + Table Lookup
```

### EXPLAIN ANALYZE

```sql
EXPLAIN ANALYZE SELECT * FROM orders WHERE user_id = 1 AND status = 'active';

-- Output interpretation:
-- Seq Scan: full table scan (BAD for large tables)
-- Index Scan: uses index, fetches rows from table
-- Index Only Scan: all data from index (BEST)
-- Bitmap Index Scan: index → bitmap → table (good for many rows)
-- Nested Loop: for joins (good for small result sets)
-- Hash Join: for joins (good for large sets)
```

---

## 2. Transactions & Isolation

### ACID

- **Atomicity**: All or nothing
- **Consistency**: Valid state → valid state
- **Isolation**: Concurrent transactions don't interfere
- **Durability**: Committed = permanent

### Isolation Levels

| Level | Dirty Read | Non-repeatable Read | Phantom Read |
|---|---|---|---|
| Read Uncommitted | ✓ | ✓ | ✓ |
| Read Committed | ✗ | ✓ | ✓ |
| Repeatable Read | ✗ | ✗ | ✓ (MySQL: ✗) |
| Serializable | ✗ | ✗ | ✗ |

**PostgreSQL default:** Read Committed
**MySQL default:** Repeatable Read

### MVCC (Multi-Version Concurrency Control)

- Each transaction sees a **snapshot** of data
- Writers don't block readers
- Readers don't block writers
- Old versions kept until no transaction needs them

```
Transaction 1 starts (snapshot at T1)
Transaction 2 updates row X → creates new version
Transaction 1 reads row X → sees old version (snapshot)
Transaction 2 commits
Transaction 1 still sees old version until it commits
```

### Deadlock

```
T1: UPDATE accounts SET balance=100 WHERE id=1;  (locks row 1)
T2: UPDATE accounts SET balance=200 WHERE id=2;  (locks row 2)
T1: UPDATE accounts SET balance=300 WHERE id=2;  (waits for T2)
T2: UPDATE accounts SET balance=400 WHERE id=1;  (waits for T1) → DEADLOCK!
```

**Prevention:**
1. Lock in consistent order
2. Keep transactions short
3. Use `SELECT ... FOR UPDATE` explicitly
4. Set lock timeout

---

## 3. Connection Pool

### Why Pool?

- TCP connection: ~1-3ms (handshake)
- TLS: ~5-10ms additional
- Authentication: ~1-5ms
- Without pool: every query pays this cost
- With pool: reuse existing connections

### Sizing Formula

```
Pool Size = (Core Count * 2) + Effective Spindle Count

Practical:
- Start with 10-20 connections
- Monitor wait time and active count
- Never set MaxOpen too high (DB has connection limits!)
```

### Go Database Pool

```go
db, _ := sql.Open("postgres", connStr)

// Pool configuration
db.SetMaxOpenConns(25)          // max active connections
db.SetMaxIdleConns(10)          // max idle connections
db.SetConnMaxLifetime(5 * time.Minute)  // recycle connections
db.SetConnMaxIdleTime(1 * time.Minute)  // close idle connections
```

---

## 4. Redis

### Data Structures

| Type | Use Case | Complexity |
|---|---|---|
| String | Cache, counter, session | O(1) |
| Hash | Object storage | O(1) per field |
| List | Queue, recent items | O(1) push/pop |
| Set | Tags, unique items | O(1) add/check |
| Sorted Set | Leaderboard, scheduling | O(log n) |
| Stream | Event log, message queue | O(1) append |

### Cache Strategies

**Cache-Aside (most common):**
```
Read:  App → Cache hit? → Return
       App → Cache miss → DB → Write cache → Return

Write: App → Write DB → Invalidate cache
```

**Read-Through:**
```
Read:  App → Cache (cache loads from DB if miss)
```

**Write-Through:**
```
Write: App → Cache → DB (synchronous)
```

**Write-Behind (Write-Back):**
```
Write: App → Cache (async batch write to DB)
```

### Cache Problems

**Cache Stampede (Thundering Herd):**
- Many requests hit expired key simultaneously
- All go to DB → DB overloaded
- Fix: mutex lock, early expiration, probabilistic refresh

**Cache Penetration:**
- Requests for keys that don't exist in DB
- Always miss cache → hit DB
- Fix: cache null values (short TTL), bloom filter

**Cache Avalanche:**
- Many keys expire at same time
- Massive DB load spike
- Fix: random TTL jitter, never expire hot keys

**Hot Key:**
- Single key receives extreme traffic
- Fix: local cache, key sharding, read replicas

### Distributed Lock (Redis)

```
SET lock_key unique_value NX PX 30000
// NX = only if not exists
// PX = expire in 30 seconds

// Release (Lua script for atomicity):
if redis.call("get", KEYS[1]) == ARGV[1] then
    return redis.call("del", KEYS[1])
else
    return 0
end
```

---

## 5. Kafka / Message Queue

### Core Concepts

```
Producer → Topic → Partition → Consumer Group
                      │
              [msg1, msg2, msg3, ...]  (ordered within partition)
```

### Partition & Ordering

- Messages with same key → same partition → ordered
- Different keys → different partitions → parallel but unordered
- Number of partitions = max parallelism

### Consumer Group

```
Topic: orders (4 partitions)
Consumer Group: order-processor

Consumer A: [Partition 0, Partition 1]
Consumer B: [Partition 2, Partition 3]

If Consumer B dies → rebalance:
Consumer A: [Partition 0, Partition 1, Partition 2, Partition 3]
```

### Delivery Semantics

| Semantic | Producer | Consumer |
|---|---|---|
| At-most-once | acks=0, no retry | Auto-commit before process |
| At-least-once | acks=all, retry | Commit after process |
| Exactly-once | Idempotent producer | Transactional consume+produce |

### Dead Letter Queue (DLQ)

```
Main Topic → Consumer → Process
                ↓ (failure after N retries)
           DLQ Topic → Alert/Manual Review
```

---

## 6. N+1 Query Problem

```go
// BAD: N+1 queries
users, _ := db.Query("SELECT * FROM users")
for users.Next() {
    var user User
    users.Scan(&user)
    // +1 query for each user!
    orders, _ := db.Query("SELECT * FROM orders WHERE user_id = ?", user.ID)
}

// GOOD: JOIN or batch
rows, _ := db.Query(`
    SELECT u.*, o.*
    FROM users u
    LEFT JOIN orders o ON o.user_id = u.id
`)

// GOOD: Batch load
userIDs := getIDs(users)
orders, _ := db.Query("SELECT * FROM orders WHERE user_id IN (?)", userIDs)
```

---

## 7. Go Database Access (database/sql, sqlx, GORM, sqlc)

> Plan gốc phủ lý thuyết SQL nhưng thiếu phần **code Go tương tác DB** — thứ trực tiếp bị hỏi/coding.

### database/sql (stdlib) — nền tảng

`database/sql` là lớp trừu tượng chung; cắm driver cụ thể (đăng ký qua `import _`):

```go
import (
    "database/sql"
    _ "github.com/jackc/pgx/v5/stdlib" // hoặc lib/pq, go-sql-driver/mysql
)

db, err := sql.Open("pgx", dsn) // KHÔNG mở connection ngay — chỉ tạo pool (lazy)
defer db.Close()
db.PingContext(ctx)             // ping để thực sự kiểm tra kết nối
```

**`sql.DB` là một connection POOL** (safe cho concurrent), không phải 1 connection. Cấu hình pool (xem mục 3).

### Query & Scan

```go
// Query một dòng
var u User
err := db.QueryRowContext(ctx,
    `SELECT id, email FROM users WHERE id = $1`, id).
    Scan(&u.ID, &u.Email)
if errors.Is(err, sql.ErrNoRows) {   // "not found" -> dùng errors.Is
    return ErrNotFound
}

// Query nhiều dòng — PHẢI đóng rows để trả connection về pool
rows, err := db.QueryContext(ctx, `SELECT id, email FROM users WHERE active = $1`, true)
if err != nil { return err }
defer rows.Close()                    // BẮT BUỘC (giống resp.Body.Close)
for rows.Next() {
    var u User
    if err := rows.Scan(&u.ID, &u.Email); err != nil { return err }
    users = append(users, u)
}
return rows.Err()                     // kiểm tra lỗi xảy ra giữa chừng
```

### Placeholder & SQL Injection

- **LUÔN dùng placeholder** (`$1` Postgres, `?` MySQL) — driver tự escape → chống SQL injection.
- **KHÔNG BAO GIỜ** nối chuỗi: `"...WHERE id = " + userInput` ← lỗ hổng nghiêm trọng.
- `Exec` cho INSERT/UPDATE/DELETE; `Query`/`QueryRow` cho SELECT.

### Prepared Statement

```go
stmt, _ := db.PrepareContext(ctx, `INSERT INTO logs(msg) VALUES($1)`)
defer stmt.Close()
for _, m := range msgs { stmt.ExecContext(ctx, m) } // tái dùng plan, tránh parse lại
```

### Transaction trong Go

```go
tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
if err != nil { return err }
defer tx.Rollback() // no-op nếu đã Commit — pattern an toàn chống quên rollback

if _, err = tx.ExecContext(ctx, `UPDATE accounts SET balance=balance-$1 WHERE id=$2`, amt, from); err != nil {
    return err // defer Rollback() chạy
}
if _, err = tx.ExecContext(ctx, `UPDATE accounts SET balance=balance+$1 WHERE id=$2`, amt, to); err != nil {
    return err
}
return tx.Commit()
```

**Pattern quan trọng:** `defer tx.Rollback()` ngay sau BeginTx. Nếu Commit thành công thì Rollback là no-op; nếu return sớm vì lỗi thì tự rollback → không bao giờ để tx treo (leak connection).

### So sánh các lựa chọn

| Thư viện | Kiểu | Ưu | Nhược |
|---|---|---|---|
| `database/sql` | thuần stdlib | không magic, kiểm soát tối đa | scan thủ công, nhiều boilerplate |
| **sqlx** | mở rộng database/sql | `StructScan`, `Get`, `Select` map vào struct | vẫn viết SQL tay (đây thường là ưu điểm) |
| **sqlc** | sinh code từ SQL | viết SQL → sinh Go **type-safe**, không reflection | thêm bước codegen |
| **GORM** | ORM đầy đủ | nhanh viết, migration, association | reflection (chậm), query ẩn khó tối ưu, dễ N+1 |
| **ent** | ORM/graph (Meta) | type-safe, schema-as-code, sinh code | learning curve |

**Quan điểm senior:** dự án cần kiểm soát hiệu năng → `sqlx`/`sqlc` (SQL tường minh). GORM phù hợp CRUD nhanh nhưng cẩn thận N+1 và query sinh tự động. Tránh phụ thuộc ORM che giấu SQL trong hot path.

### Migrations

- Công cụ: **golang-migrate**, **goose**, **atlas**. Migration = file SQL có version (up/down).

```
migrations/
  001_create_users.up.sql     001_create_users.down.sql
  002_add_index.up.sql        002_add_index.down.sql
```

- Chạy tự động lúc deploy hoặc bằng job riêng. **Quy tắc:** migration phải **backward-compatible** khi rolling deploy (thêm cột nullable trước, backfill, rồi mới enforce NOT NULL) để tránh downtime.
- Tránh `AutoMigrate` của GORM ở production — khó kiểm soát, dễ khoá bảng lớn.

### Bẫy hay gặp (interview)

- Quên `rows.Close()` / `rows.Err()` → connection leak, bỏ sót lỗi.
- Dùng `db.Query` trong vòng lặp → N+1 (xem mục 6). Dùng JOIN hoặc `IN (...)`.
- `sql.Open` không kết nối ngay → phải `Ping` để phát hiện config sai sớm.
- Không set `SetConnMaxLifetime` → connection cũ bị DB/LB đóng ngầm gây lỗi "connection reset".
- Nhầm `sql.DB` là 1 connection — thực ra là pool, không cần tự pool thêm.
