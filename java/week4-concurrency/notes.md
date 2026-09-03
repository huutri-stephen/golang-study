# Week 4 – Concurrency & Multithreading – Deep Dive Notes

> Phần **quan trọng nhất** cho senior backend (fintech đặc biệt). Không dừng ở "dùng synchronized".
> Phải giải thích **Java Memory Model** (reordering, happens-before), **false sharing**,
> **ThreadPoolExecutor** hoạt động thế nào, **AQS** đứng sau các lock, và **CAS/ABA**.

## Mục lục
1. [Java Memory Model: reordering & happens-before](#1-java-memory-model)
2. [volatile: visibility, không atomic](#2-volatile)
3. [CAS, Atomic & bài toán ABA](#3-cas-atomic--aba)
4. [synchronized: monitor, lock upgrade](#4-synchronized)
5. [AQS: nền của ReentrantLock/Semaphore](#5-aqs)
6. [ThreadPoolExecutor internals](#6-threadpoolexecutor-internals)
7. [CompletableFuture](#7-completablefuture)
8. [False sharing & cache line](#8-false-sharing)
9. [Virtual threads (Loom)](#9-virtual-threads)
10. [Deadlock & chẩn đoán](#10-deadlock--chẩn-đoán)
11. [Follow-up questions](#11-follow-up-questions)

---

## 1. Java Memory Model

> JMM định nghĩa **khi nào một thread thấy được ghi của thread khác**, và cho phép compiler/CPU
> reorder lệnh miễn là không đổi ngữ nghĩa **đơn luồng**.

### Vấn đề reordering + cache

Mỗi core có cache riêng. Compiler + CPU có thể sắp xếp lại lệnh để tối ưu. Kết quả: thread A ghi 2 biến theo thứ tự này, thread B có thể "thấy" theo thứ tự khác.

```java
// Ví dụ kinh điển thiếu đồng bộ
int x = 0; boolean ready = false;
// Thread A:            // Thread B:
x = 42;                 while (!ready) {}
ready = true;           System.out.println(x);  // có thể in 0 (thấy ready=true trước x=42)!
```

Không có happens-before giữa ghi của A và đọc của B → B có thể thấy `ready=true` nhưng `x=0` (do reorder hoặc cache chưa flush).

### happens-before — các quy tắc quan trọng

Nếu A **happens-before** B thì mọi ghi trước A đều thấy được ở B:

1. **Program order**: trong một thread, lệnh trước happens-before lệnh sau.
2. **Monitor lock**: unlock một monitor happens-before mọi lock sau đó trên monitor đó.
3. **volatile**: ghi một biến volatile happens-before mọi đọc sau đó của biến đó.
4. **Thread start**: `thread.start()` happens-before mọi hành động trong thread.
5. **Thread join**: mọi hành động trong thread happens-before `thread.join()` trả về.
6. **Transitivity**: A hb B, B hb C ⇒ A hb C.

Sửa ví dụ trên: khai báo `volatile boolean ready` → ghi `x=42` (program order) hb ghi `ready=true` (volatile) hb đọc `ready` hb đọc `x` → B luôn thấy `x=42`.

---

## 2. volatile

- Đảm bảo **visibility** (đọc thấy ghi mới nhất) + **cấm reorder** quanh biến volatile (memory barrier/fence).
- **KHÔNG** đảm bảo **atomicity** cho thao tác phức hợp:

```java
volatile int count = 0;
count++;   // 3 thao tác: đọc, +1, ghi -> 2 thread có thể cùng đọc giá trị cũ -> mất update
```

- Dùng volatile cho: **flag** (`volatile boolean running`), **double-checked locking** (biến instance phải volatile), publish reference an toàn.

### Double-checked locking (vì sao cần volatile)

```java
class Singleton {
    private static volatile Singleton instance;   // volatile BẮT BUỘC
    static Singleton get() {
        if (instance == null) {                    // check 1 (không lock)
            synchronized (Singleton.class) {
                if (instance == null)              // check 2 (có lock)
                    instance = new Singleton();
            }
        }
        return instance;
    }
}
```

Không volatile: `instance = new Singleton()` gồm (1) cấp bộ nhớ, (2) chạy constructor, (3) gán reference. CPU có thể reorder thành 1→3→2 → thread khác thấy `instance != null` nhưng object **chưa init xong**. volatile cấm reorder này.

---

## 3. CAS, Atomic & ABA

### CAS (Compare-And-Swap)

Thao tác phần cứng nguyên tử: `CAS(địa chỉ, kỳ vọng, mới)` — nếu giá trị hiện tại == kỳ vọng thì ghi giá trị mới, trả true; ngược lại trả false.

```java
AtomicInteger c = new AtomicInteger(0);
c.incrementAndGet();   // vòng lặp: đọc v -> tính v+1 -> CAS(v, v+1); fail thì retry (spin)
```

- **Lock-free**: không block thread, không context switch. Nhanh khi **tranh chấp thấp**.
- Tranh chấp **cao** → nhiều retry (spin lãng phí CPU) → cân nhắc `LongAdder` (chia thành nhiều cell, cộng dồn khi cần) thay `AtomicLong`.

### Bài toán ABA

CAS chỉ so **giá trị**, không biết giá trị đã đổi A→B→A. Thread thấy vẫn "A" nên CAS thành công, dù trạng thái đã thay đổi và quay lại.

- Vấn đề với cấu trúc lock-free (stack/queue dùng con trỏ): node bị pop rồi push lại cùng địa chỉ.
- **Giải pháp**: `AtomicStampedReference` (gắn thêm version/stamp) hoặc `AtomicMarkableReference`.

---

## 4. synchronized

### Monitor & bytecode

- `synchronized` block → bytecode `monitorenter` / `monitorexit`. Method synchronized → flag `ACC_SYNCHRONIZED`.
- Mỗi object có một **monitor** (gắn với mark word). Reentrant: thread đang giữ vào lại được (đếm reentry).

### Lock upgrade (HotSpot, tới Java 15)

`biased → lightweight (CAS spin) → heavyweight (OS mutex, park thread)`.

- **Biased locking**: tối ưu khi chỉ 1 thread dùng (đã **bỏ mặc định từ Java 15**).
- **Lightweight**: tranh chấp nhẹ, CAS trên mark word (spin), không vào kernel.
- **Heavyweight**: tranh chấp cao, thread bị park (chờ ở kernel) → tốn context switch.

### synchronized vs volatile

| | volatile | synchronized |
|---|---|---|
| Visibility | có | có |
| Atomicity | không | có |
| Block thread | không | có (khi tranh chấp) |
| Dùng cho | flag, publish | vùng tới hạn, thao tác phức hợp |

---

## 5. AQS

**AbstractQueuedSynchronizer** là khung nền cho `ReentrantLock`, `Semaphore`, `CountDownLatch`, `ReentrantReadWriteLock`.

- Giữ một `volatile int state` (ý nghĩa tùy lock: ReentrantLock = số lần giữ; Semaphore = số permit; CountDownLatch = count) + hàng đợi FIFO các thread chờ (CLH queue).
- Thread không giành được → bị `LockSupport.park()` (không spin lãng phí), khi được nhả thì `unpark`.

### ReentrantLock vs synchronized

| | synchronized | ReentrantLock (AQS) |
|---|---|---|
| Nhả khoá | tự động | `unlock()` trong `finally` |
| tryLock / timeout | không | có |
| Interruptible | không | `lockInterruptibly()` |
| Fairness | không | tùy chọn (FIFO) |
| Condition | 1 (wait/notify) | nhiều `newCondition()` |

```java
Lock lock = new ReentrantLock();
lock.lock();
try { /* critical */ } finally { lock.unlock(); }   // BẮT BUỘC unlock trong finally
```

### Sync primitives (đều dựa AQS)

- `Semaphore(n)` — n permit; `acquire`/`release`. Dùng: giới hạn concurrency (connection pool, rate limit).
- `CountDownLatch(n)` — chờ n sự kiện (`countDown` → `await`), **1 lần dùng**.
- `CyclicBarrier(n)` — n thread chờ nhau tại điểm hẹn, **tái sử dụng**.
- `ReadWriteLock` — nhiều reader song song, writer độc quyền. `StampedLock` (Java 8) thêm optimistic read.

---

## 6. ThreadPoolExecutor internals

> "ExecutorService xử lý task thế nào?" — cần biết luồng: core → queue → max → reject.

```java
new ThreadPoolExecutor(
    corePoolSize,      // số thread giữ thường trực
    maximumPoolSize,   // trần số thread
    keepAliveTime,     // thread vượt core idle bao lâu thì thu hồi
    unit,
    workQueue,         // hàng đợi task
    threadFactory,
    rejectedExecutionHandler);
```

### Thứ tự xử lý khi submit task

1. Nếu **< corePoolSize** thread → tạo thread mới chạy ngay.
2. Nếu đủ core → **đưa vào queue**.
3. Nếu queue **đầy** và **< maximumPoolSize** → tạo thread mới (tới max).
4. Nếu queue đầy và đã đạt max → **rejection handler**.

### Bẫy với Executors factory

- `newFixedThreadPool` / `newSingleThreadExecutor` dùng **`LinkedBlockingQueue` không giới hạn** → task dồn vô hạn → **OOM** thay vì reject. Bước 3-4 không bao giờ xảy ra.
- `newCachedThreadPool` dùng `SynchronousQueue` + max = `Integer.MAX_VALUE` → tạo thread không giới hạn → có thể cạn thread OS.
- **Khuyến nghị**: tự tạo `ThreadPoolExecutor` với **bounded queue** + rejection policy rõ ràng (CallerRunsPolicy/AbortPolicy).

### Rejection policies

- `AbortPolicy` (mặc định): ném `RejectedExecutionException`.
- `CallerRunsPolicy`: chạy task trên thread gọi → tạo back-pressure tự nhiên.
- `DiscardPolicy` / `DiscardOldestPolicy`: bỏ task.

---

## 7. CompletableFuture

```java
CompletableFuture
    .supplyAsync(() -> fetchUser(id), executor)   // nên truyền executor riêng
    .thenApply(User::name)                         // map (đồng bộ trên kết quả)
    .thenCompose(n -> fetchOrdersAsync(n))         // flatMap (nối future)
    .thenCombine(otherFuture, (a, b) -> merge(a, b))// gộp 2 future song song
    .exceptionally(ex -> fallback())               // xử lý lỗi
    .orTimeout(2, TimeUnit.SECONDS);               // Java 9+ timeout
```

- `thenApply` vs `thenApplyAsync`: bản Async chạy trên pool khác (không dùng thread hoàn thành stage trước) → tránh chiếm thread callback.
- **Mặc định dùng ForkJoinPool.commonPool** nếu không truyền executor → nên **truyền executor riêng** cho production.
- Exception lan dọc chain; stage `thenApply` bị **skip** cho tới khi gặp `exceptionally`/`handle`. `handle((r, e) -> ...)` xử lý cả 2 nhánh.

---

## 8. False sharing

> Bug hiệu năng tinh vi mà senior nên biết.

- CPU cache theo **cache line** (thường 64 byte). Hai biến khác nhau nằm cùng một cache line → khi thread A ghi biến X, cache line bị invalidate ở core khác đang đọc biến Y (dù Y không đổi) → gọi là **false sharing**, giảm hiệu năng nghiêm trọng trong hot loop đa luồng.
- Giải pháp: padding để tách biến ra cache line riêng, hoặc annotation **`@Contended`** (`jdk.internal.vm.annotation.Contended`, cần `-XX:-RestrictContended`).
- `LongAdder` thắng `AtomicLong` khi contention cao một phần nhờ tách các cell tránh false sharing.

---

## 9. Virtual threads

- **Platform thread**: map 1-1 OS thread (~1MB stack, giới hạn vài nghìn, tạo/switch tốn kém).
- **Virtual thread** (Java 21, Loom): thread nhẹ do JVM lập lịch trên một tập nhỏ **carrier thread** (ForkJoinPool). Khi virtual thread **block IO**, nó **unmount** khỏi carrier → carrier chạy virtual thread khác.
- Tạo được **hàng triệu**; hợp workload **IO-bound nhiều kết nối** — cho phép viết code blocking đơn giản thay vì async callback phức tạp.
- **Không** tăng tốc CPU-bound (vẫn giới hạn core).
- **Pinning**: nếu virtual thread block trong `synchronized` hoặc native call, nó bị "ghim" vào carrier (không unmount) → giảm lợi ích. Java 21 khuyến nghị dùng `ReentrantLock` thay `synchronized` trong đoạn có blocking.

```java
try (var exec = Executors.newVirtualThreadPerTaskExecutor()) {
    for (int i = 0; i < 1_000_000; i++)
        exec.submit(() -> { Thread.sleep(1000); return null; });   // 1 triệu vt — bất khả thi với platform thread
}
```

---

## 10. Deadlock & chẩn đoán

### 4 điều kiện Coffman

mutual exclusion + hold-and-wait + no preemption + circular wait. Phá 1 điều kiện → hết deadlock.

### Cách tránh

- **Lock ordering**: luôn giữ khoá theo cùng thứ tự (vd theo id tăng dần) → phá circular wait.
- `tryLock(timeout)` → nhả và thử lại thay vì chờ vô hạn.
- Giảm phạm vi & số khoá giữ đồng thời.

### Chẩn đoán

- `jstack <pid>` → phát hiện "Found one Java-level deadlock" + chỉ ra chu trình khoá.
- Thread dump cho thấy thread `BLOCKED` chờ lock nào, đang giữ lock nào.

### ThreadLocal + pool = leak

Thread trong pool được **tái dùng** → phải `threadLocal.remove()` sau khi xong (thường trong `finally`), nếu không dữ liệu rò rỉ sang task sau và giữ object không cho GC.

---

## 11. Follow-up questions

1. Vì sao double-checked locking cần volatile? → cấm reorder "gán reference trước khi constructor xong".
2. volatile có làm count++ an toàn không? → không; read-modify-write không atomic.
3. CAS là gì, ABA là gì, khắc phục thế nào? → compare-and-swap; A→B→A qua mắt CAS; dùng AtomicStampedReference.
4. ThreadPoolExecutor xử lý task theo thứ tự nào? → core → queue → max → reject.
5. Vì sao tránh Executors.newFixedThreadPool trong production? → queue không giới hạn → OOM thay vì back-pressure.
6. AQS là gì? → khung state + hàng đợi FIFO cho lock; park/unpark thread.
7. False sharing là gì? → 2 biến cùng cache line, ghi 1 biến invalidate biến kia → chậm; padding/@Contended.
8. Virtual thread pinning? → block trong synchronized/native ghim vt vào carrier; dùng ReentrantLock.
9. happens-before của volatile và lock? → ghi volatile hb đọc volatile; unlock hb lock cùng monitor.
10. LongAdder thắng AtomicLong khi nào? → contention cao (tách cell, giảm CAS retry + false sharing).
