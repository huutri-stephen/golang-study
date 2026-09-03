# Week 2 – Flashcards / Q&A – Goroutine + Concurrency

## Goroutine & Scheduler

### Q: Goroutine khác OS thread thế nào?
**A:**
- Stack: Goroutine 2-8KB (growable) vs Thread 1-8MB (fixed)
- Creation: ~1μs vs ~1ms
- Context switch: ~200ns (user space) vs ~1-10μs (kernel)
- Scheduling: Go runtime vs OS kernel
- Count: millions vs thousands

---

### Q: GMP model là gì? Giải thích từng component.
**A:**
- **G** (Goroutine): unit of work, chứa stack + instruction pointer + status
- **M** (Machine): OS thread, thực thi goroutine
- **P** (Processor): logical CPU, chứa local run queue, count = GOMAXPROCS
- Flow: M phải bind P để chạy G. P chọn G từ local queue.

---

### Q: Work stealing hoạt động thế nào?
**A:** Khi P hết goroutine trong local queue:
1. Check global run queue
2. Check network poller
3. Steal **half** of another random P's local queue

---

### Q: GOMAXPROCS ảnh hưởng gì?
**A:** Quyết định số lượng P (logical processors). Default = num CPU cores. Limit số goroutine chạy **đồng thời** (parallel). Không limit tổng số goroutine hay OS threads.

---

### Q: Goroutine bị block khi syscall — chuyện gì xảy ra?
**A:**
1. M hiện tại bị block cùng G
2. P được detach khỏi M
3. P tìm M khác (hoặc tạo mới) để tiếp tục chạy goroutines khác
4. Khi syscall hoàn tất, G được đưa lại vào run queue

---

### Q: Network I/O có block M không?
**A:** Không. Go dùng **network poller** (epoll/kqueue). Goroutine waiting for I/O được park, M không bị block. Khi I/O ready, goroutine được wake up và đưa vào run queue.

---

### Q: Preemption trong Go hoạt động thế nào?
**A:**
- Go < 1.14: Cooperative — yield tại function calls
- Go >= 1.14: Asynchronous — runtime gửi SIGURG signal, preempt ngay cả tight loop
- Giải quyết vấn đề GC STW latency

---

## Channel

### Q: Unbuffered vs Buffered channel?
**A:**
- **Unbuffered**: Send blocks until receiver ready. Synchronization point. Direct copy sender→receiver.
- **Buffered**: Send blocks only when full. Receive blocks only when empty. Ring buffer.

---

### Q: Send to closed channel?
**A:** **PANIC** — `panic: send on closed channel`

---

### Q: Receive from closed channel?
**A:** Returns zero value immediately, `ok = false`. Không block, không panic.

---

### Q: Send/Receive on nil channel?
**A:** **Block forever**. Useful trong select để disable một case.

---

### Q: Khi nào dùng buffered channel?
**A:**
- Producer/consumer rate khác nhau
- Batch processing
- Rate limiting
- Khi muốn decouple sender/receiver speed
- Rule: nếu không biết chọn gì → dùng unbuffered (an toàn hơn)

---

### Q: Select với multiple cases ready?
**A:** Go chọn **random** (pseudo-random). Không có priority. Đảm bảo fairness.

---

### Q: Nil channel trong select?
**A:** Case đó bị **ignored** (never selected). Dùng để dynamically enable/disable channels.

```go
var ch1, ch2 chan int
if condition {
    ch1 = realChannel // enable this case
}
select {
case v := <-ch1: // skipped if ch1 is nil
case v := <-ch2: // skipped if ch2 is nil
}
```

---

## Synchronization

### Q: sync.Mutex có reentrant không?
**A:** Không. Same goroutine Lock() 2 lần → deadlock. Go không có reentrant mutex by design (keeps things simple and explicit).

---

### Q: RWMutex — writer starvation?
**A:** Không. Khi writer waiting, new readers cũng bị block. Go's RWMutex ưu tiên writer — prevents writer starvation.

---

### Q: sync.Once — nếu function panic?
**A:** Once coi như đã done. Function sẽ KHÔNG được retry. Subsequent calls return immediately without calling function.

---

### Q: Khi nào dùng atomic vs mutex?
**A:**
- **Atomic**: simple counter, flag, pointer swap. Lock-free, faster.
- **Mutex**: complex operations, multiple variables, read-modify-write on complex state.
- Rule: nếu operation là single word (int64, pointer) → atomic. Otherwise → mutex.

---

### Q: errgroup vs WaitGroup?
**A:**
- WaitGroup: chỉ wait, không handle errors
- errgroup: wait + return first error + optional context cancellation
- Use errgroup when goroutines can fail and you need error propagation

---

## Concurrency Problems

### Q: Goroutine leak xảy ra khi nào?
**A:**
1. Send to channel với no receiver
2. Receive from channel với no sender
3. Infinite loop without exit condition
4. Missing context cancellation check
5. Leaked background goroutine after parent returns

---

### Q: Cách prevent goroutine leak?
**A:**
1. Dùng context with cancellation
2. Done channel pattern
3. Defer close(channel)
4. Set timeout cho mọi blocking operation
5. Monitor goroutine count (`runtime.NumGoroutine()`)

---

### Q: Data race detection?
**A:**
```bash
go test -race ./...
go run -race main.go
```
Race detector dùng ThreadSanitizer algorithm. Detect at runtime, NOT compile time. Có overhead ~2-10x slowdown.

---

### Q: Deadlock — 4 conditions?
**A:**
1. **Mutual exclusion**: resource held exclusively
2. **Hold and wait**: hold one, wait for another
3. **No preemption**: can't force release
4. **Circular wait**: A waits B, B waits A

Break any one → no deadlock.

---

## Patterns

### Q: Worker Pool pattern?
**A:**
```go
jobs := make(chan Job, 100)
results := make(chan Result, 100)

// Start workers
for i := 0; i < numWorkers; i++ {
    go worker(jobs, results)
}

// Send jobs
for _, j := range allJobs {
    jobs <- j
}
close(jobs)
```
- Fixed number of workers
- Bounded concurrency
- Backpressure via buffered channel

---

### Q: Fan-in pattern?
**A:** Merge multiple channels into one:
```go
func fanIn(channels ...<-chan int) <-chan int {
    out := make(chan int)
    var wg sync.WaitGroup
    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan int) {
            defer wg.Done()
            for v := range c {
                out <- v
            }
        }(ch)
    }
    go func() {
        wg.Wait()
        close(out)
    }()
    return out
}
```

---

### Q: Mutex vs Channel — khi nào dùng cái nào?
**A:**
| Use Case | Choice |
|---|---|
| Protect shared state (cache, map) | Mutex |
| Simple counter | Atomic |
| Pass data between goroutines | Channel |
| Signal completion | Channel / WaitGroup |
| Pipeline processing | Channel |
| Worker pool | Channel |
| Rate limiting | Channel (ticker) |

Pragmatic rule: dùng cái nào code **rõ ràng hơn** cho case cụ thể.
