# Mock Interview Scenarios – Senior Java Backend

> Bộ kịch bản luyện tập theo 4 vòng. Với mỗi câu: tự đặt giờ, nói to suy nghĩ, rồi đối chiếu gợi ý.

---

## Vòng 1 – Coding (45 phút)

### 1.1 LRU Cache
Thiết kế LRU cache dung lượng cố định, `get`/`put` O(1).
- **Gợi ý**: `LinkedHashMap` với access-order + override `removeEldestEntry`; hoặc HashMap + doubly linked list tự cài.

### 1.2 Rate limiter
Cài rate limiter cho phép tối đa N request/giây mỗi user.
- **Gợi ý**: token bucket hoặc sliding window; thread-safe bằng `AtomicLong`/`ConcurrentHashMap`; bàn phân tán (Redis) nếu nhiều instance.

### 1.3 Producer-Consumer
N producer, M consumer chia sẻ bounded buffer.
- **Gợi ý**: `BlockingQueue` (put/take tự block); hoặc `wait/notifyAll` + `synchronized`. Bàn back-pressure.

### 1.4 Group anagrams / word count
Xử lý collection với Stream API.
- **Gợi ý**: `Collectors.groupingBy`; phân tích độ phức tạp; khi nào parallel stream.

---

## Vòng 2 – Java / JVM Deep Dive (45 phút)

### 2.1 equals/hashCode
"Điều gì xảy ra nếu override equals mà quên hashCode?"
- **Điểm chốt**: object bằng nhau rơi vào bucket khác trong HashMap → contains/get sai. Nêu 5 quy tắc equals.

### 2.2 Concurrency
"`count++` trên nhiều thread bị sai vì sao? volatile có cứu được không?"
- **Điểm chốt**: read-modify-write không atomic; volatile chỉ visibility. Fix: AtomicInteger/synchronized/LongAdder. Giải thích happens-before.

### 2.3 GC & memory leak
"App heap tăng dần rồi OOM dù có GC. Bạn debug thế nào?"
- **Điểm chốt**: leak = reachable nhưng không cần (static collection, ThreadLocal trong pool, listener). Công cụ: heap dump + MAT, jstat, GC log. Nêu GC roots.

### 2.4 Collections
"HashMap hoạt động bên trong thế nào? Treeify khi nào?"
- **Điểm chốt**: bucket + hash + collision list → cây đỏ-đen khi bucket ≥8 (table ≥64); load factor 0.75, resize gấp đôi.

### 2.5 Modern Java
"Record khác class thường? Khi nào dùng sealed?"
- **Điểm chốt**: record = immutable data carrier tự sinh equals/hashCode; sealed giới hạn subtype → exhaustive switch.

---

## Vòng 3 – System Design (45–60 phút)

### 3.1 URL Shortener
Thiết kế dịch vụ rút gọn URL.
- **Khung**: yêu cầu (QPS đọc/ghi, độ dài, TTL) → API (POST /shorten, GET /{code}) → sinh code (base62 từ ID, hoặc hash) → data model (KV store) → cache đọc → scale (read replica, CDN).

### 3.2 Payment / Order service
Thiết kế service xử lý thanh toán.
- **Khung**: idempotency key (tránh double charge), transaction & consistency, saga cho phân tán, retry + circuit breaker, audit log, exactly-once với message queue.

### 3.3 News feed / notification
Thiết kế hệ thống feed.
- **Khung**: fan-out on write vs read, cache, message queue, pagination, ranking.

### 3.4 Rate limiter phân tán
Mở rộng bài coding 1.2 cho nhiều instance.
- **Khung**: Redis (INCR + EXPIRE, hoặc token bucket bằng Lua script) để state tập trung; đánh đổi latency vs độ chính xác.

---

## Vòng 4 – Behavioral / Senior (30–45 phút)

Trả lời theo **STAR** (Situation, Task, Action, Result):

- Kể về một sự cố production nghiêm trọng bạn xử lý. Root cause? Phòng ngừa sau đó?
- Một quyết định kỹ thuật khó (chọn công nghệ, refactor lớn) — bạn cân nhắc trade-off thế nào?
- Bất đồng với đồng nghiệp/team lead về thiết kế — bạn giải quyết ra sao?
- Bạn mentor junior thế nào? Ví dụ cụ thể.
- Cách bạn cân bằng nợ kỹ thuật với deadline.

**Tiêu chí senior**: ownership, tác động đo được, tư duy trade-off, giao tiếp rõ ràng, nâng đội lên.

---

## Tự đánh giá sau mock

- [ ] Trả lời có cấu trúc (không lan man)?
- [ ] Giải thích được "vì sao", nêu trade-off?
- [ ] Chủ động nêu edge case / cách kiểm chứng?
- [ ] Phân tích được độ phức tạp / bottleneck?
- [ ] Thành thật khi không chắc, nêu hướng tìm hiểu?


---

# Bổ sung: Deep Technical Drill (chuỗi follow-up)

> Interviewer senior hiếm khi dừng ở câu đầu. Mỗi mục dưới là một **chuỗi đào sâu** —
> luyện trả lời tới đáy. Đáp án tóm tắt xem notes/flashcards từng tuần.

## Drill 1 — "Giải thích HashMap"
1. HashMap hoạt động thế nào? → bucket + hash + collision.
2. Index bucket tính ra sao? → `(n-1) & hash`, n lũy thừa 2.
3. Vì sao cần `h ^ (h >>> 16)`? → trộn bit cao xuống thấp.
4. Khi nào treeify, vì sao ngưỡng 8? → bucket ≥8 & table ≥64; Poisson.
5. Resize tối ưu rehash thế nào? → tách lo/hi bằng `hash & oldCap`.
6. Nếu key sai equals/hashCode thì sao? → mất entry / không tìm ra.
7. ConcurrentHashMap khác gì? → CAS + synchronized per-bin, không null.

## Drill 2 — "count++ trên nhiều thread bị sai"
1. Vì sao sai? → read-modify-write không atomic.
2. volatile có cứu được không? → không, chỉ visibility.
3. Sửa bằng gì? → AtomicInteger (CAS) / synchronized / LongAdder.
4. CAS là gì? → compare-and-swap, lock-free, spin retry.
5. ABA là gì? → giá trị A→B→A qua mắt CAS; AtomicStampedReference.
6. Contention cao thì AtomicLong vs LongAdder? → LongAdder tách cell.
7. happens-before liên quan thế nào? → đảm bảo visibility giữa thread.

## Drill 3 — "@Transactional không rollback / không chạy"
1. @Transactional hoạt động thế nào? → proxy + TransactionInterceptor.
2. Proxy tạo lúc nào? → BeanPostProcessor.after trong bean lifecycle.
3. Self-invocation vì sao hỏng? → this.method() bỏ qua proxy.
4. Checked exception có rollback không? → không, cần rollbackFor.
5. Method private có được proxy không? → không.
6. JDK proxy vs CGLIB? → interface vs subclass; final không proxy được.
7. Fix self-invocation? → tách bean, self-inject, AopContext.

## Drill 4 — "App production OOM / GC pause cao"
1. Bước đầu điều tra? → xem GC log, metrics, xác nhận triệu chứng.
2. Full GC liên tục nghi gì? → memory leak hoặc heap nhỏ.
3. Java có GC sao vẫn leak? → reachable nhưng không cần.
4. Ví dụ leak điển hình? → static collection, ThreadLocal trong pool, listener, classloader.
5. Công cụ tìm leak? → heap dump + Eclipse MAT (dominator tree, leak suspects).
6. GC roots là gì? → điểm sống chắc chắn (stack, static, JNI, thread).
7. Chọn GC nào cho low-latency? → ZGC/Shenandoah (pause <1ms).

## Drill 5 — "Thiết kế payment service"
1. Yêu cầu & scale? → QPS, tính nhất quán tiền, retry.
2. Chống double charge? → idempotency key.
3. Cross-service consistency? → saga / outbox pattern, eventual consistency.
4. Downstream lỗi? → circuit breaker + retry + timeout.
5. Exactly-once với queue? → idempotent consumer + dedup key.
6. Audit độc lập transaction chính? → @Transactional(REQUIRES_NEW).
7. Quan sát & phục hồi? → tracing, DLQ, reconciliation job.

---

# Tự đánh giá sau mỗi drill

- [ ] Trả lời được tới câu số mấy mà không bí?
- [ ] Có giải thích "vì sao" ở mỗi bước, không chỉ "là gì"?
- [ ] Nêu được trade-off và edge case?
- [ ] Liên hệ được sang cơ chế JVM/JMM khi phù hợp?
