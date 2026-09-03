# Week 5 – Flashcards / Q&A (Deep Dive)

## Class Loading

### Q: 5 bước class loading?
**A:** Loading (đọc .class, tạo Class object) → Linking gồm Verification (kiểm bytecode an toàn), Preparation (cấp static field giá trị mặc định), Resolution (symbolic→direct reference) → Initialization (chạy <clinit>, gán static field thật). Lazy: init khi lần đầu chủ động dùng.

---

### Q: Truy cập `static final int B = 10` có kích hoạt class init không?
**A:** Không. Hằng compile-time (static final với giá trị hằng) được inline vào call site lúc compile, không kích hoạt <clinit>. Truy cập static field non-final hoặc gọi static method thì có.

---

### Q: <clinit> có thread-safe không? Ứng dụng?
**A:** Có, JVM đảm bảo <clinit> chạy đúng 1 lần, lock trên Class. Nền của initialization-on-demand holder idiom: lazy singleton qua static nested Holder class, không cần volatile/synchronized.

---

### Q: Parent delegation model làm gì?
**A:** ClassLoader hỏi cha trước khi tự load: App→Platform→Bootstrap. Đảm bảo class core (java.lang.String) luôn do bootstrap load (chống giả mạo), tránh trùng class. Hai class cùng tên khác loader = hai class khác nhau.

---

## Memory areas

### Q: Metaspace khác PermGen thế nào?
**A:** Metaspace (Java 8+) ở native memory, tự grow (giới hạn MaxMetaspaceSize). PermGen (cũ) ở trong heap, kích thước cố định, hay gây OutOfMemoryError: PermGen. String pool cũng chuyển từ PermGen sang heap ở Java 7.

---

### Q: StackOverflowError vs OutOfMemoryError?
**A:** SOF: JVM stack của một thread đầy (đệ quy quá sâu). OOM: heap/metaspace/native đầy. Stack là per-thread (frame biến local); heap chia sẻ (object).

---

## Allocation

### Q: TLAB là gì, giải quyết vấn đề gì?
**A:** Thread-Local Allocation Buffer: mỗi thread có vùng riêng trong Eden để cấp phát object bằng bump-pointer KHÔNG cần đồng bộ. Tránh nhiều thread tranh con trỏ cấp phát chung. Hết TLAB mới xin block mới.

---

### Q: Escape analysis + scalar replacement?
**A:** JIT phân tích nếu object không thoát khỏi method (không leak ra ngoài) thì có thể không cấp trên heap mà tách thành các biến trên stack (scalar replacement) → giảm áp lực GC, đôi khi bỏ luôn lock (lock elision).

---

## GC

### Q: GC xác định rác thế nào? Vì sao không reference counting?
**A:** Reachability: từ GC roots (biến local stack, static field, JNI, active thread) đánh dấu object reachable; không reachable là rác. Không dùng reference counting vì không xử lý được tham chiếu vòng (A↔B).

---

### Q: Weak generational hypothesis?
**A:** Hầu hết object chết trẻ. Nên chia young (Eden+Survivor) và old. Minor GC dọn young thường xuyên và rẻ (ít object sống); object sống sót qua nhiều lần được promote lên old. Major/full GC dọn old, đắt hơn.

---

### Q: Card table dùng để làm gì?
**A:** Giải quyết reference old→young. Chia old thành card 512 byte; write barrier đánh dấu card "dirty" khi ghi reference. Minor GC chỉ quét card dirty thay vì cả old gen → nhanh. G1 dùng remembered set per-region tương tự.

---

### Q: Vì sao G1 tên "Garbage First"?
**A:** Heap chia ~2048 region kích thước đều, vai trò linh hoạt (Eden/Survivor/Old/Humongous). G1 ưu tiên thu region NHIỀU RÁC NHẤT trước để tối đa bộ nhớ thu về trong pause target (MaxGCPauseMillis).

---

### Q: G1 có mấy loại thu gom?
**A:** Young GC (STW, thu Eden+Survivor), Concurrent marking (đánh dấu old song song app, SATB), Mixed GC (STW, thu young + một số old region rác nhất). Cân bằng throughput/latency.

---

### Q: ZGC đạt pause <1ms bằng cách nào?
**A:** Colored pointers (nhét metadata GC vào bit con trỏ 64-bit) + load barrier (kiểm/relocate object khi đọc reference). Hầu hết công việc GC chạy concurrent với app, chỉ vài pha STW cực ngắn, gần như không phụ thuộc heap size.

---

## Reference & Leak

### Q: Strong/soft/weak/phantom reference?
**A:** Strong: không thu khi còn. Soft: thu khi sắp thiếu bộ nhớ (cache). Weak: thu ở GC kế nếu chỉ còn weak (WeakHashMap). Phantom: enqueue vào ReferenceQueue sau finalize để cleanup off-heap (thay finalize).

---

### Q: ThreadLocal leak trong thread pool vì sao?
**A:** ThreadLocalMap dùng weak key nhưng value là STRONG ref. Thread pool sống mãi → nếu không remove(), value bị giữ qua thread → leak. Luôn remove() trong finally sau khi dùng.

---

### Q: ClassLoader leak là gì?
**A:** Giữ reference tới class/instance của một classloader (thường khi redeploy app) → cả classloader và mọi class nó nạp không được unload → Metaspace phình dần tới OOM Metaspace.

---

## Troubleshooting

### Q: Các loại OutOfMemoryError?
**A:** Java heap space (leak/object nhiều), GC overhead limit exceeded (GC chạy mãi thu ít), Metaspace (quá nhiều class), unable to create native thread (quá nhiều thread), Direct buffer memory (NIO off-heap).

---

### Q: Quy trình debug heap tăng dần dẫn tới OOM?
**A:** Chụp heap dump (jmap -dump:live hoặc HeapDumpOnOutOfMemoryError) → mở Eclipse MAT → xem Leak Suspects + dominator tree tìm object giữ nhiều RAM và đường reference giữ nó (GC root path).

---

### Q: App CPU cao/treo, debug thế nào?
**A:** jstack <pid> nhiều lần tìm thread RUNNABLE chiếm CPU hoặc BLOCKED (deadlock). top -H tìm thread OS nóng rồi map nid (hex) sang thread trong jstack. async-profiler cho flame graph CPU.

---

### Q: Flag nào nên đặt cho production?
**A:** -Xms=-Xmx (tránh resize), chọn GC (-XX:+UseG1GC), -Xlog:gc* để có GC log, -XX:+HeapDumpOnOutOfMemoryError + HeapDumpPath để tự chụp dump khi OOM phục vụ điều tra.
