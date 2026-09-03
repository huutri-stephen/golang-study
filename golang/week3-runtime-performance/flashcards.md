# Week 3 – Flashcards / Q&A – Runtime + Performance

## Escape Analysis

### Q: Escape analysis là gì và tại sao quan trọng?
**A:** Compiler analysis tại compile-time quyết định variable được allocate trên stack hay heap. Quan trọng vì:
- Stack allocation = free (no GC)
- Heap allocation = GC pressure, slower
- Biết escape rules → viết code ít allocate hơn → better performance

---

### Q: Cách check escape analysis?
**A:**
```bash
go build -gcflags="-m" ./...
go build -gcflags="-m -m" ./...  # more verbose
```
Output: `moved to heap: x` hoặc `x escapes to heap`

---

### Q: Tại sao `fmt.Println(x)` gây heap allocation?
**A:** `fmt.Println` nhận `interface{}`. Khi pass value vào interface → boxing → heap allocation. Đây là lý do fmt functions appear trong allocation profiles.

---

## Garbage Collector

### Q: Go GC algorithm?
**A:** Concurrent tri-color mark-and-sweep:
- Tri-color: White (garbage candidate), Gray (visited, children pending), Black (reachable)
- Concurrent: marking runs alongside program (mostly non-STW)
- STW pauses: chỉ ~10-100μs (mark setup + mark termination)

---

### Q: GC phases?
**A:**
1. **Mark Setup (STW ~10-100μs):** Enable write barrier
2. **Marking (Concurrent):** Traverse heap, mark reachable
3. **Mark Termination (STW ~10-100μs):** Finalize marking
4. **Sweep (Concurrent):** Reclaim white (unreachable) objects

---

### Q: Write barrier là gì? Tại sao cần?
**A:** Mechanism đảm bảo GC concurrent marking đúng. Khi program modify pointer trong khi GC đang mark:
- Without write barrier: GC có thể miss newly-reachable objects → memory corruption
- Write barrier: shade both old and new target gray → GC sẽ scan them

---

### Q: GOGC là gì? Tune thế nào?
**A:** 
- `GOGC=100` (default): GC triggers khi heap doubles
- Formula: trigger = live_heap × (1 + GOGC/100)
- `GOGC=50`: GC more frequent, less memory
- `GOGC=200`: GC less frequent, more memory
- `GOGC=off`: disable GC
- Go 1.19+: `SetMemoryLimit` cho hard limit

---

### Q: STW pause trong Go typical bao lâu?
**A:** ~10-100 microseconds. Rất ngắn vì:
- Marking là concurrent
- STW chỉ cho setup/termination
- Go 1.8+ đạt sub-millisecond pauses consistently

---

### Q: Cách giảm GC pressure?
**A:**
1. Reduce allocations (pre-allocate, reuse)
2. Use sync.Pool for temporary objects
3. Avoid interface boxing in hot paths
4. Use value types for small structs
5. Increase GOGC (trade memory for CPU)
6. Set memory limit (Go 1.19+)
7. strings.Builder thay vì string concatenation

---

## Profiling

### Q: Các loại profile trong Go?
**A:**
| Profile | Measures |
|---|---|
| CPU | Time spent in functions |
| Heap | Current live heap allocations |
| Allocs | All allocations (including freed) |
| Goroutine | Stack traces of all goroutines |
| Mutex | Lock contention time |
| Block | Time goroutines spend blocked |

---

### Q: Cách lấy CPU profile production service?
**A:**
```go
import _ "net/http/pprof"
// GET /debug/pprof/profile?seconds=30
```
```bash
go tool pprof http://host:port/debug/pprof/profile?seconds=30
```
Trong pprof: `top`, `list funcName`, `web` (flamegraph)

---

### Q: Heap profile — "inuse" vs "alloc"?
**A:**
- `inuse_space/inuse_objects`: currently allocated (find memory leaks)
- `alloc_space/alloc_objects`: total ever allocated (find allocation hotspots)
- Use inuse for memory leaks, alloc for GC pressure

---

### Q: Khi nào dùng mutex profile?
**A:** Khi nghi ngờ lock contention (high CPU but low throughput). Enable:
```go
runtime.SetMutexProfileFraction(1)
```
Profile shows which mutexes have high contention time.

---

## Benchmarking

### Q: Cách viết benchmark đúng?
**A:**
```go
func BenchmarkX(b *testing.B) {
    // setup (not timed)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        functionUnderTest()
    }
}
```
Key: loop `b.N` times, runtime decides N for stable measurement.

---

### Q: `-benchmem` output có ý nghĩa gì?
**A:**
```
BenchmarkX-8  1000000  1052 ns/op  256 B/op  3 allocs/op
```
- `1052 ns/op`: thời gian mỗi operation
- `256 B/op`: bytes allocated mỗi operation
- `3 allocs/op`: số allocations mỗi operation

---

### Q: Compiler optimization có thể sai benchmark?
**A:** Có. Compiler có thể eliminate dead code:
```go
// BAD
_ = compute()  // compiler might remove

// GOOD: assign to package-level var
var result int
func BenchmarkX(b *testing.B) {
    var r int
    for i := 0; i < b.N; i++ {
        r = compute()
    }
    result = r
}
```

---

### Q: Cách compare benchmark trước/sau optimization?
**A:**
```bash
go test -bench=. -count=5 > old.txt
# make changes
go test -bench=. -count=5 > new.txt
benchstat old.txt new.txt
```
benchstat cho statistical comparison (p-value, confidence interval).

---

## Performance General

### Q: Performance optimization workflow?
**A:**
1. Benchmark (establish baseline)
2. Profile (identify bottleneck)
3. Optimize (specific hot path)
4. Benchmark again (measure improvement)
5. Validate in production

Never optimize without measuring first.

---

### Q: Top Go performance anti-patterns?
**A:**
1. String concatenation in loop (`+` → `strings.Builder`)
2. Not pre-allocating slices/maps
3. Interface boxing in hot paths
4. Excessive goroutine creation
5. Lock contention (use sharding/RWMutex)
6. Reflection in hot paths
7. Not reusing allocations (missing sync.Pool)

---

### Q: sync.Pool khi nào dùng?
**A:**
- Temporary objects allocated frequently on hot path
- Example: buffers in HTTP handlers, JSON encoder buffers
- Rules: objects may be collected at any GC, always reset before Put
- NOT a cache (no guarantee of persistence)

---

### Q: Goroutine stack growth có impact performance không?
**A:** Có nhưng minimal:
- Growth: allocate new stack 2x → copy → update pointers
- Mỗi growth ~microseconds
- Vấn đề: deep recursion = many growths
- Fix: avoid deep recursion, use iterative when possible
- Note: stack can also shrink (halved when < 25% used)
