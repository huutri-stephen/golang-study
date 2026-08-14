# Week 2 – Goroutine + Concurrency – Study Notes

## 1. Goroutine vs OS Thread

| Feature | OS Thread | Goroutine |
|---|---|---|
| Stack size | 1-8 MB (fixed) | 2-8 KB (grows dynamically) |
| Creation cost | ~1ms | ~1μs |
| Context switch | ~1-10μs (kernel) | ~200ns (user space) |
| Scheduling | OS kernel | Go runtime |
| Number | Thousands | Millions |

### Key insight
Goroutines are **multiplexed** onto OS threads (M:N scheduling). Go runtime manages scheduling in **user space**, avoiding expensive kernel context switches.

---

## 2. GMP Model

```
┌─────────────────────────────────────────────────────────┐
│                    Go Scheduler                          │
├─────────────────────────────────────────────────────────┤
│                                                         │
│   G (Goroutine)     M (Machine/Thread)    P (Processor) │
│   ┌───┐             ┌───┐                 ┌───┐        │
│   │ G │ goroutine   │ M │ OS thread       │ P │ logical│
│   │   │ stack       │   │ kernel thread   │   │ CPU    │
│   │   │ status      │   │ current G       │   │ runq   │
│   │   │ instruction │   │ bound P         │   │ cache  │
│   └───┘             └───┘                 └───┘        │
│                                                         │
│   GOMAXPROCS = number of P                              │
│   Each P has a local run queue (max 256 G)              │
│   Global run queue for overflow                         │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

### G – Goroutine
- Contains: stack, instruction pointer, status
- Status: `_Grunnable`, `_Grunning`, `_Gwaiting`, `_Gsyscall`, `_Gdead`
- Stack starts 2-8KB, grows via contiguous stack copying

### M – Machine (OS Thread)
- Actual kernel thread
- Bound to one P at a time to execute goroutines
- Can be parked when no work available
- `runtime.GOMAXPROCS` does NOT limit M count

### P – Processor (Logical CPU)
- Count = `GOMAXPROCS` (default = num CPU cores)
- Has local run queue (LRQ) with max 256 goroutines
- Has mcache for memory allocation (no lock needed)
- Coordinates work between M and G

### Scheduling Flow
```
1. New goroutine → put in P's local queue
2. P picks next G from its local queue
3. If local queue empty → steal from other P's queue
4. If all empty → check global queue
5. If still empty → check network poller
6. If nothing → park M
```

### Work Stealing
When P's local run queue is empty:
1. Check global run queue
2. Check network poller
3. **Steal half** of another P's local run queue

---

## 3. Scheduler Preemption

### Cooperative Preemption (Go < 1.14)
- Goroutine yields at function calls (preemption points)
- Problem: tight loop without function calls can't be preempted

### Asynchronous Preemption (Go >= 1.14)
- Runtime sends signal (SIGURG) to preempt goroutines
- Even tight loops can be preempted
- Solves GC STW latency issues

---

## 4. Channel Internals

### Channel Structure (hchan)
```go
type hchan struct {
    qcount   uint     // current elements in buffer
    dataqsiz uint     // buffer capacity
    buf      unsafe.Pointer  // circular buffer
    elemsize uint16
    closed   uint32
    sendx    uint     // send index
    recvx    uint     // receive index
    recvq    waitq    // list of waiting receivers
    sendq    waitq    // list of waiting senders
    lock     mutex
}
```

### Unbuffered Channel
- **Send blocks** until receiver is ready
- **Receive blocks** until sender is ready
- Direct copy from sender's stack to receiver's stack
- Synchronization point (happens-before guarantee)

### Buffered Channel
- **Send blocks** only when buffer is full
- **Receive blocks** only when buffer is empty
- Uses circular buffer (ring buffer)

### Channel Operations & Nil/Closed Behavior

| Operation | Nil Channel | Closed Channel |
|---|---|---|
| Send | Block forever | **PANIC** |
| Receive | Block forever | Returns zero value, ok=false |
| Close | **PANIC** | **PANIC** |

### Select Statement
```go
select {
case msg := <-ch1:
    // received from ch1
case ch2 <- data:
    // sent to ch2
case <-time.After(5 * time.Second):
    // timeout
default:
    // non-blocking (only if no other case ready)
}
```

- If multiple cases ready → random selection
- `default` makes select non-blocking
- Nil channel in select case → that case is ignored

---

## 5. Synchronization Primitives

### sync.Mutex
```go
var mu sync.Mutex
mu.Lock()
// critical section
mu.Unlock()
```
- Not reentrant (same goroutine can deadlock)
- Must not be copied after first use

### sync.RWMutex
```go
var rw sync.RWMutex
rw.RLock()   // multiple readers OK
rw.RUnlock()
rw.Lock()    // exclusive write lock
rw.Unlock()
```
- Multiple readers OR single writer
- Writer waits for all readers to finish
- New readers wait if writer is waiting (prevents writer starvation)

### sync.WaitGroup
```go
var wg sync.WaitGroup
for i := 0; i < n; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        // work
    }()
}
wg.Wait() // blocks until counter = 0
```

### sync.Once
```go
var once sync.Once
var instance *Singleton

func GetInstance() *Singleton {
    once.Do(func() {
        instance = &Singleton{}
    })
    return instance
}
```
- Guaranteed to execute function exactly once
- Even if concurrent goroutines call it simultaneously
- All goroutines observe the result after Do completes

### sync/atomic
```go
var counter int64
atomic.AddInt64(&counter, 1)
val := atomic.LoadInt64(&counter)
atomic.StoreInt64(&counter, 0)

// Compare-and-swap
atomic.CompareAndSwapInt64(&counter, oldVal, newVal)
```
- Lock-free operations
- Fastest for simple counters/flags
- No complex operations (only load/store/add/swap/cas)

### errgroup
```go
g, ctx := errgroup.WithContext(ctx)
g.Go(func() error {
    return task1(ctx)
})
g.Go(func() error {
    return task2(ctx)
})
if err := g.Wait(); err != nil {
    // first error from any goroutine
}
```
- WaitGroup + error handling + context cancellation
- Cancels context on first error

---

## 6. Concurrency Patterns

### Worker Pool
```
         ┌─────────────┐
jobs ──→ │   Channel   │ ──→ Worker 1
         │  (buffered) │ ──→ Worker 2
         │             │ ──→ Worker 3
         └─────────────┘ ──→ Worker N
```

### Fan-out / Fan-in
```
              ┌─→ Worker 1 ─┐
Input ──→ ───┼─→ Worker 2 ─┼──→ Merge ──→ Output
              └─→ Worker 3 ─┘
```
- Fan-out: distribute work to multiple goroutines
- Fan-in: merge results from multiple channels into one

### Pipeline
```
Stage 1 ──→ Stage 2 ──→ Stage 3 ──→ Output
(generate)   (process)   (filter)
```
- Each stage: receive from input channel, process, send to output channel
- Cancellation via context or done channel

---

## 7. Concurrency Problems

### Data Race
- Two goroutines access same variable, at least one writes, no synchronization
- Detection: `go test -race ./...`
- Fix: mutex, channel, or atomic

### Deadlock
Four conditions (all must be present):
1. Mutual exclusion
2. Hold and wait
3. No preemption
4. Circular wait

### Goroutine Leak
Common causes:
1. Channel with no receiver
2. Channel with no sender (blocked receive)
3. Infinite loop without exit
4. Missing context cancellation

Prevention:
- Always have exit path (context, done channel)
- Use `defer close(ch)` pattern
- Monitor goroutine count in production

---

## 8. Mutex vs Channel – When to Use Which

### Use Mutex when:
- Protecting shared state (cache, map, counter)
- Simple lock/unlock pattern
- Performance critical (mutex is faster)
- State doesn't flow between goroutines

### Use Channel when:
- Communicating between goroutines
- Signaling (done, cancel)
- Implementing pipeline/fan-out
- Coordinating work (worker pool)
- Ownership transfer of data

### Go Proverb
> "Don't communicate by sharing memory; share memory by communicating."

But pragmatically: use whichever is simpler and clearer for the specific case.
