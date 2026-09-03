# Week 4 – Flashcards / Q&A (Deep Dive)

## Java Memory Model

### Q: JMM giải quyết vấn đề gì?
**A:** Định nghĩa khi nào một thread thấy được ghi của thread khác, và cho phép compiler/CPU reorder lệnh miễn không đổi ngữ nghĩa đơn luồng. Không có đồng bộ → thread có thể thấy ghi theo thứ tự khác hoặc giá trị cũ (do cache/reorder).

### Q: happens-before, kể các quy tắc chính?
**A:** Program order (trong 1 thread), monitor (unlock hb lock cùng monitor), volatile (ghi hb đọc), thread start (start hb hành động trong thread), thread join (hành động trong thread hb join), và transitivity. Nếu A hb B thì mọi ghi trước A thấy được ở B.

### Q: Vì sao `while(!ready){} print(x)` có thể in giá trị cũ của x?
**A:** Không có happens-before giữa ghi của thread A và đọc của B. B có thể thấy ready=true nhưng x chưa được flush/reorder → in 0. Fix: khai báo ready volatile để tạo happens-before.

## volatile

### Q: volatile đảm bảo gì, không đảm bảo gì?
**A:** Đảm bảo visibility (đọc thấy ghi mới nhất) + cấm reorder (memory barrier). KHÔNG đảm bảo atomicity: count++ (read-modify-write) vẫn race. Dùng cho flag, publish reference, double-checked locking.

### Q: Vì sao double-checked locking cần volatile?
**A:** `instance = new Singleton()` gồm cấp bộ nhớ, chạy constructor, gán reference. CPU có thể reorder gán reference TRƯỚC khi constructor xong → thread khác thấy instance != null nhưng chưa init. volatile cấm reorder này.

## CAS & Atomic

### Q: CAS hoạt động thế nào?
**A:** Compare-And-Swap: thao tác phần cứng nguyên tử, nếu giá trị hiện tại == kỳ vọng thì ghi giá trị mới. AtomicInteger.incrementAndGet lặp: đọc v, tính v+1, CAS(v,v+1), fail thì retry. Lock-free, nhanh khi tranh chấp thấp.

### Q: Bài toán ABA là gì? Khắc phục?
**A:** CAS chỉ so giá trị, không biết A→B→A đã xảy ra. Thread thấy vẫn "A" nên CAS thành công dù trạng thái đã đổi. Vấn đề với cấu trúc lock-free (con trỏ node). Khắc phục: AtomicStampedReference (gắn version/stamp).

### Q: Khi nào LongAdder tốt hơn AtomicLong?
**A:** Contention cao. AtomicLong nhiều thread cùng CAS một biến → nhiều retry (spin lãng phí) + false sharing. LongAdder chia thành nhiều cell, mỗi thread cộng vào cell riêng, tổng khi cần → giảm tranh chấp.

## synchronized & AQS

### Q: Lock upgrade của synchronized?
**A:** biased → lightweight (CAS spin trên mark word) → heavyweight (OS mutex, park thread). Biased locking đã bỏ mặc định từ Java 15. Heavyweight tốn context switch vào kernel.

### Q: AQS là gì?
**A:** AbstractQueuedSynchronizer: khung nền cho ReentrantLock/Semaphore/CountDownLatch. Giữ volatile int state (ý nghĩa tùy lock) + hàng đợi FIFO thread chờ. Thread không giành được thì LockSupport.park (không spin lãng phí).

### Q: synchronized vs ReentrantLock?
**A:** synchronized tự nhả khoá, đơn giản. ReentrantLock (AQS): tryLock/timeout, lockInterruptibly, fairness, nhiều Condition, nhưng phải unlock trong finally. Cả hai reentrant.

### Q: CountDownLatch vs CyclicBarrier?
**A:** CountDownLatch: chờ n sự kiện (countDown → await), dùng 1 lần. CyclicBarrier: n thread chờ nhau tại điểm hẹn, tái sử dụng được (reset sau mỗi vòng).

## ThreadPool

### Q: ThreadPoolExecutor xử lý task theo thứ tự nào?
**A:** 1) < corePoolSize → tạo thread mới. 2) đủ core → vào queue. 3) queue đầy & < max → tạo thread tới max. 4) queue đầy & đạt max → rejection handler.

### Q: Vì sao tránh Executors.newFixedThreadPool trong production?
**A:** Dùng LinkedBlockingQueue KHÔNG giới hạn → task dồn vô hạn → OOM thay vì back-pressure (bước tạo thread tới max/reject không bao giờ xảy ra). Nên tự tạo ThreadPoolExecutor với bounded queue + rejection policy.

### Q: CallerRunsPolicy làm gì?
**A:** Khi pool quá tải, task bị từ chối sẽ chạy trên thread gọi submit → làm chậm producer → tạo back-pressure tự nhiên, tránh dồn task vô hạn.

## CompletableFuture

### Q: thenApply vs thenApplyAsync?
**A:** thenApply chạy trên thread hoàn thành stage trước (có thể chiếm thread đó). thenApplyAsync chạy trên pool khác. Production nên truyền executor riêng thay vì mặc định ForkJoinPool.commonPool.

### Q: CompletableFuture xử lý exception thế nào?
**A:** Exception lan dọc chain, stage thenApply bị skip cho tới khi gặp exceptionally(fn) (fallback) hoặc handle((r,e)->...) (cả 2 nhánh). thenCompose = flatMap nối future, thenCombine gộp 2 future song song.

## False sharing & Virtual threads

### Q: False sharing là gì?
**A:** CPU cache theo cache line (~64 byte). 2 biến khác nhau cùng cache line: ghi biến X invalidate cache line ở core đang đọc biến Y (dù Y không đổi) → chậm trong hot loop đa luồng. Fix: padding hoặc @Contended.

### Q: Virtual thread vs platform thread?
**A:** Platform map 1-1 OS thread (~1MB, vài nghìn). Virtual (Loom) nhẹ, JVM lập lịch trên carrier thread, unmount khi block IO → triệu cái, hợp IO-bound. Không tăng tốc CPU-bound.

### Q: Virtual thread pinning là gì?
**A:** Nếu virtual thread block trong synchronized hoặc native call, nó bị ghim vào carrier thread (không unmount) → giảm lợi ích. Khuyến nghị dùng ReentrantLock thay synchronized trong đoạn có blocking.

## Deadlock

### Q: 4 điều kiện deadlock và cách tránh?
**A:** mutual exclusion, hold-and-wait, no preemption, circular wait. Tránh: lock ordering (thứ tự cố định phá circular wait), tryLock timeout, giảm phạm vi/số khoá. Chẩn đoán: jstack phát hiện deadlock.

### Q: ThreadLocal bẫy gì trong thread pool?
**A:** Thread tái dùng nên phải remove() sau khi xong (trong finally), nếu không dữ liệu rò rỉ sang task sau và giữ object không cho GC → memory leak.
