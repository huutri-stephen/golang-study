# Week 5 – JVM, Memory & Garbage Collection – Deep Dive Notes

> Điểm phân biệt senior rõ rệt nhất. Phải giải thích được **class loading + linking chi tiết**,
> **TLAB**, **card table / remembered set**, **các pha của G1 và ZGC**, và một **quy trình chẩn đoán**
> khi production OOM/GC pause cao.

## Mục lục
1. [JVM architecture tổng thể](#1-jvm-architecture)
2. [Class loading: 5 bước chi tiết](#2-class-loading)
3. [ClassLoader & parent delegation](#3-classloader--parent-delegation)
4. [Runtime data areas](#4-runtime-data-areas)
5. [Object allocation: TLAB & bump pointer](#5-object-allocation-tlab)
6. [GC: reachability, generational, card table](#6-gc-nền-tảng)
7. [G1 chi tiết (region, phases)](#7-g1-chi-tiết)
8. [ZGC & low-latency collectors](#8-zgc)
9. [Reference types & memory leak](#9-reference-types--memory-leak)
10. [Tuning flags & troubleshooting workflow](#10-tuning--troubleshooting)
11. [Follow-up questions](#11-follow-up-questions)

---

## 1. JVM architecture

```
.java --javac--> .class (bytecode) 
        │
        ▼
   ClassLoader  ──► Runtime Data Areas (heap, stacks, metaspace, ...)
        │
        ▼
   Execution Engine
     ├── Interpreter (chạy bytecode ngay)
     ├── JIT (C1/C2 biên dịch hot code -> native)
     └── GC
```

- JVM là **stack-based** (khác register-based): mỗi frame có operand stack; bytecode như `iadd` lấy toán hạng từ stack.
- **JIT tiered compilation**: level 0 interpreter → C1 (level 1-3, biên dịch nhanh + profiling) → C2 (level 4, tối ưu sâu: inlining, escape analysis, loop unrolling). Method "nóng" (đếm invocation/loop back-edge vượt ngưỡng) được nâng cấp dần.

---

## 2. Class loading

### 5 bước (Loading → Linking(3) → Initialization)

1. **Loading** — tìm & đọc `.class`, tạo `Class` object trong heap + metadata trong Metaspace.
2. **Linking**:
   - **Verification** — kiểm tra bytecode hợp lệ, an toàn (không tràn stack, type đúng). Bảo mật quan trọng.
   - **Preparation** — cấp phát static field với **giá trị mặc định** (0/false/null), CHƯA gán giá trị code.
   - **Resolution** — phân giải symbolic reference (tên) thành direct reference (con trỏ). Có thể lazy.
3. **Initialization** — chạy `<clinit>` (static initializer + gán static field). **Lazy** — chỉ khi lần đầu chủ động dùng class (new, gọi static method/field non-final, reflection).

```java
class Config {
    static int a = compute();   // Preparation: a=0; Initialization: a=compute()
    static final int B = 10;    // constant -> inline lúc compile, không kích hoạt init
}
```

> **Follow-up:** "Truy cập `Config.B` (static final constant) có kích hoạt class init không?" → Không, hằng compile-time được inline vào call site. Truy cập static field non-final thì có.

### `<clinit>` thread-safe

JVM đảm bảo `<clinit>` chạy **đúng một lần**, thread-safe (lock trên Class). Đây là nền của **initialization-on-demand holder** idiom (lazy singleton không cần volatile):

```java
class Singleton {
    private Singleton() {}
    private static class Holder { static final Singleton INSTANCE = new Singleton(); }
    static Singleton get() { return Holder.INSTANCE; }  // Holder chỉ init khi get() lần đầu
}
```

---

## 3. ClassLoader & parent delegation

```
Bootstrap (C++, load java.base: java.lang.*)
    ▲ delegate lên
Platform ClassLoader (JDK modules)
    ▲
Application ClassLoader (classpath của bạn)
    ▲
[Custom ClassLoader]
```

- **Parent delegation**: khi load class, loader hỏi **cha trước**; chỉ tự load nếu cha không tìm thấy. Đảm bảo `java.lang.String` luôn do bootstrap load → không thể bị giả mạo (security + tránh trùng class).
- Hai class **cùng tên nhưng khác ClassLoader → là hai class khác nhau** (không cast/equals được). Đây là nền của isolation trong app server/plugin, và cũng là nguồn `ClassCastException` khó hiểu + **ClassLoader leak**.

---

## 4. Runtime data areas

| Vùng | Phạm vi | Lưu | Lỗi khi đầy |
|---|---|---|---|
| **Heap** | toàn JVM | object, mảng | `OutOfMemoryError: Java heap space` |
| **JVM Stack** | mỗi thread | frame: biến local, operand stack | `StackOverflowError` |
| **Metaspace** | toàn JVM | metadata class (native memory, không phải heap) | `OutOfMemoryError: Metaspace` |
| **PC Register** | mỗi thread | địa chỉ lệnh đang chạy | — |
| **Native Method Stack** | mỗi thread | cho JNI | — |

- **Metaspace** (Java 8+) thay **PermGen** — nằm ở **native memory**, tự grow (giới hạn `-XX:MaxMetaspaceSize`). PermGen cũ ở trong heap, hay gây `OutOfMemoryError: PermGen`.
- String pool chuyển từ PermGen (≤ Java 6) sang **heap** (Java 7+).

---

## 5. Object allocation: TLAB

> "Cấp phát object có phải lock heap không?" — câu hỏi tinh.

- Heap được cấp phát bằng **bump-the-pointer** (chỉ tăng con trỏ trong Eden) — rất nhanh.
- Nhưng nhiều thread cùng cấp phát sẽ tranh chấp con trỏ. Giải: **TLAB (Thread-Local Allocation Buffer)** — mỗi thread có một vùng riêng trong Eden, cấp phát trong đó **không cần đồng bộ**. Hết TLAB thì xin block mới (mới cần đồng bộ, hiếm).
- Object lớn hơn TLAB hoặc quá lớn → cấp thẳng vào heap/old gen (humongous trong G1).
- **Escape analysis** (JIT): nếu object không "thoát" khỏi method, JVM có thể **scalar replacement** (không cấp trên heap, tách thành biến trên stack) → giảm áp lực GC.

---

## 6. GC nền tảng

### Xác định rác: reachability, không phải reference counting

- Từ **GC roots** (biến local trên stack các thread, static field, JNI ref, active thread, monitor đang giữ) đi theo reference đánh dấu object **reachable**.
- Object không reachable → rác. Reference counting không dùng vì không xử lý được **tham chiếu vòng** (A↔B).

### Generational hypothesis

- **Weak generational hypothesis**: hầu hết object **chết trẻ**. Nên chia heap:

```
┌───────── Young Gen ─────────┐   ┌──── Old (Tenured) ────┐
│ Eden │ Survivor S0 │ S1     │   │   object sống lâu     │
└──────┴─────────────┴────────┘   └───────────────────────┘
```

- Object mới → Eden. **Minor GC** (young đầy): copy object sống sang survivor, tăng **age**; qua ngưỡng `MaxTenuringThreshold` → promote lên old. Minor GC nhanh vì young nhỏ + ít object sống.
- **Major/Full GC**: dọn old (+young). Đắt, pause lâu. Mục tiêu tuning: giảm tần suất/độ dài full GC.

### Card table & remembered set (vì sao minor GC không quét cả old gen)

- Vấn đề: object **old** trỏ tới object **young** → khi minor GC (chỉ quét young), làm sao biết object young đó còn sống?
- Giải: **card table** — chia old gen thành "card" (512 byte). Khi ghi reference (old→young), **write barrier** đánh dấu card "dirty". Minor GC chỉ quét các card dirty thay vì cả old gen → nhanh.
- G1 dùng **remembered set (RSet)** per-region cho mục đích tương tự giữa các region.

---

## 7. G1 chi tiết

> Mặc định từ Java 9. Cần biết: region, humongous, mixed GC, phases.

### Cấu trúc region

- Heap chia thành **~2048 region** kích thước đều (1-32MB). Mỗi region đóng vai Eden / Survivor / Old / **Humongous** (object ≥ 1/2 region) tùy thời điểm — không cố định vùng liên tục như GC cũ.
- "**Garbage First**": ưu tiên thu region **nhiều rác nhất** trước → tối đa hoá bộ nhớ thu về / pause.

### Các loại GC trong G1

- **Young GC** (STW): thu Eden + Survivor, copy object sống.
- **Concurrent marking**: đánh dấu liveness ở old gen **song song** với app (SATB — Snapshot-At-The-Beginning).
- **Mixed GC** (STW): thu young + một số old region rác nhất.

### Pause target

- `-XX:MaxGCPauseMillis=200` (mặc định 200ms): G1 cố **đạt target** bằng cách chọn số region thu vừa đủ. Đây là collector **cân bằng throughput/latency**.

---

## 8. ZGC

- Mục tiêu: **pause < 1ms**, gần như **không phụ thuộc heap size** (kể cả nhiều TB). Từ Java 15 GA (Java 21 có generational ZGC).
- Kỹ thuật: **colored pointers** (nhét metadata GC vào bit của con trỏ 64-bit) + **load barrier** (khi đọc reference, kiểm tra/relocate object đang di chuyển) → hầu hết công việc GC chạy **concurrent** với app, chỉ vài pha STW cực ngắn.
- Đánh đổi: throughput hơi thấp hơn G1/Parallel, tốn thêm bộ nhớ cho barrier. Dùng khi **latency là ưu tiên số 1** (trading, low-latency service).
- **Shenandoah** (Red Hat) cùng nhóm low-pause, dùng Brooks pointer / load barrier.

| Collector | Ưu tiên | Pause | Dùng khi |
|---|---|---|---|
| Parallel | throughput | cao (STW) | batch job |
| G1 | cân bằng | ~vài chục-trăm ms | mặc định, đa số service |
| ZGC / Shenandoah | latency | <1-10ms | heap lớn, low-latency |

---

## 9. Reference types & memory leak

| Loại | GC thu khi | Dùng |
|---|---|---|
| Strong | không (khi còn ref) | mặc định |
| Soft | khi sắp thiếu bộ nhớ | cache nhạy bộ nhớ |
| Weak | lần GC kế nếu chỉ còn weak ref | `WeakHashMap`, canonicalizing map |
| Phantom | sau khi finalize, enqueue vào ReferenceQueue | cleanup tài nguyên off-heap (thay `finalize`) |

### Memory leak dù có GC (leak = reachable nhưng không cần)

- **Static collection** giữ mãi (cache không evict).
- **Listener/callback** không hủy đăng ký → object bị giữ qua reference.
- **ThreadLocal** không `remove()` trong thread pool (thread sống mãi → giữ value). Bản thân `ThreadLocalMap` dùng weak key nhưng **value là strong** → value leak.
- Key HashMap **sai equals/hashCode** hoặc mutable key đổi hash sau khi put.
- **ClassLoader leak**: giữ reference tới class/instance của một classloader (redeploy) → cả classloader + mọi class của nó không được unload → Metaspace phình.

---

## 10. Tuning & troubleshooting

### Flags cơ bản

```
-Xms4g -Xmx4g                 # heap min=max (tránh resize runtime, ổn định latency)
-XX:+UseG1GC                  # chọn collector (G1 mặc định J9+)
-XX:MaxGCPauseMillis=200      # target pause cho G1
-XX:MaxMetaspaceSize=256m     # trần metaspace
-Xlog:gc*:file=gc.log:time    # GC log (Java 9+ unified logging)
-XX:+HeapDumpOnOutOfMemoryError -XX:HeapDumpPath=/var/dumps
```

### Các loại OutOfMemoryError & hướng xử lý

| OOM | Nguyên nhân | Bước đầu |
|---|---|---|
| Java heap space | object quá nhiều / leak | heap dump + MAT tìm dominator |
| GC overhead limit exceeded | GC chạy mãi thu được ít | như trên (thường là leak) |
| Metaspace | nạp quá nhiều class (proxy, redeploy) | đếm class loaded, kiểm classloader leak |
| unable to create native thread | quá nhiều thread | giảm thread / tăng ulimit |
| Direct buffer memory | NIO off-heap | kiểm `-XX:MaxDirectMemorySize`, buffer leak |

### Quy trình chẩn đoán production (workflow)

1. **Xác nhận triệu chứng**: latency tăng? OOM? CPU cao? → xem GC log (`-Xlog:gc*`) và metrics.
2. **GC pause cao** → phân tích GC log (tần suất, thời lượng, young vs full). Full GC liên tục ⇒ nghi leak hoặc heap nhỏ.
3. **Heap tăng dần** → chụp **heap dump** (`jmap -dump:live,format=b,file=h.hprof <pid>` hoặc tự động khi OOM) → mở **Eclipse MAT** → xem "Leak Suspects" + dominator tree tìm object giữ nhiều RAM.
4. **CPU cao / treo** → `jstack <pid>` nhiều lần → tìm thread `RUNNABLE` chiếm CPU hoặc `BLOCKED` (deadlock). `top -H` để tìm thread OS nóng, map sang nid trong jstack.
5. **Alloc rate cao** → async-profiler alloc mode → tìm điểm cấp phát nóng.
6. Công cụ: `jps`, `jstat -gcutil <pid> 1s`, `jcmd <pid> GC.heap_info`, VisualVM/JMC.

---

## 11. Follow-up questions

1. 5 bước class loading? → Loading, Verification, Preparation (default value), Resolution, Initialization.
2. Truy cập static final constant có init class không? → không, inline compile-time.
3. Parent delegation để làm gì? → tránh giả mạo class core + tránh trùng; hai loader → hai class khác nhau.
4. TLAB giải quyết gì? → cấp phát không đồng bộ trong Eden per-thread; escape analysis + scalar replacement giảm alloc.
5. Card table dùng làm gì? → theo dõi reference old→young qua write barrier để minor GC không quét cả old.
6. G1 vì sao tên "Garbage First"? → thu region nhiều rác nhất trước để tối đa bộ nhớ thu / pause target.
7. ZGC đạt pause <1ms bằng cách nào? → colored pointer + load barrier, GC chủ yếu concurrent.
8. Java có GC sao vẫn leak? → reachable nhưng không cần (static, ThreadLocal, listener, classloader).
9. ThreadLocal leak trong pool vì sao? → ThreadLocalMap value là strong ref, thread pool sống mãi → không remove thì giữ value.
10. Metaspace khác PermGen? → Metaspace ở native memory, tự grow; PermGen ở heap, cố định, hay OOM.
11. Quy trình debug heap tăng dần? → heap dump → MAT → dominator tree / leak suspects.
