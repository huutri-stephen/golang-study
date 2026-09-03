# Week 3 – Go Runtime + Performance – Study Notes

## 1. Memory & Escape Analysis

### Stack in Go

**Contiguous Stack Model:**
- Goroutine bắt đầu với stack nhỏ (2-8KB)
- Khi stack đầy → allocate stack mới lớn hơn (2x)
- Copy toàn bộ old stack sang new stack
- Update tất cả pointers pointing into stack
- Free old stack

```
Before growth:         After growth:
┌──────────┐           ┌──────────────────────┐
│  Stack   │    →      │     New Stack (2x)    │
│  (4KB)   │           │  (copy old content)   │
└──────────┘           └──────────────────────┘
```

### Escape Analysis Rules

Compiler quyết định tại **compile time** — không phải runtime:

| Scenario | Stack/Heap |
|---|---|
| Local variable, no pointer returned | Stack |
| Return pointer to local | Heap |
| Interface boxing | Heap (usually) |
| Closure captures pointer | Heap |
| `make([]T, n)` where n is variable | Heap |
| Slice too large | Heap |
| Sent to channel | Heap |
| Stored in global variable | Heap |

### Check Escape Analysis

```bash
# Show escape decisions
go build -gcflags="-m" ./...

# More verbose
go build -gcflags="-m -m" ./...

# Output examples:
# ./main.go:10:6: can inline add
# ./main.go:15:2: moved to heap: x
# ./main.go:20:9: &Point{} escapes to heap
```

### Optimization Tips

```go
// BAD: allocates on heap
func newBuffer() *bytes.Buffer {
    buf := &bytes.Buffer{}  // escapes
    return buf
}

// BETTER: caller provides buffer (stays on caller's stack)
func fillBuffer(buf *bytes.Buffer) {
    buf.WriteString("hello")
}

// BAD: interface boxing forces heap
func process(v interface{}) { ... }
process(42)  // 42 boxed to heap

// BETTER: use concrete type when possible
func processInt(v int) { ... }
```

---

## 2. Garbage Collector

### Tri-Color Marking

Go GC uses concurrent tri-color mark-and-sweep:

```
White: Not yet visited (potentially garbage)
Gray:  Visited, but children not yet scanned
Black: Visited, all children scanned (reachable)

Algorithm:
1. Start: all objects White
2. Mark roots → Gray
3. Pick Gray object:
   - Scan its children → mark them Gray
   - Mark itself → Black
4. Repeat until no Gray objects remain
5. All remaining White objects = garbage → sweep
```

### GC Phases

```
┌─────────┐     ┌──────────┐     ┌─────────┐     ┌─────────┐
│ STW     │ →   │ Marking  │ →   │ STW     │ →   │ Sweep   │
│ Mark    │     │ (concur) │     │ Mark    │     │ (concur)│
│ Setup   │     │          │     │ Term.   │     │         │
└─────────┘     └──────────┘     └─────────┘     └─────────┘
 ~10-100μs       concurrent        ~10-100μs      concurrent
```

1. **Mark Setup (STW):** Enable write barrier, prepare
2. **Marking (Concurrent):** Scan heap, mark reachable objects
3. **Mark Termination (STW):** Finish marking, disable write barrier
4. **Sweeping (Concurrent):** Reclaim unmarked objects

### Write Barrier

During concurrent marking, mutator (program) might modify pointers. Write barrier ensures GC doesn't miss objects:

```go
// Pseudo-code: write barrier
func writePointer(slot *unsafe.Pointer, ptr unsafe.Pointer) {
    shade(ptr)       // mark new target gray
    shade(*slot)     // mark old target gray
    *slot = ptr      // actual write
}
```

### GC Tuning

```go
// GOGC: target heap growth ratio (default=100)
// New heap trigger = live heap * (1 + GOGC/100)
// GOGC=100 → GC when heap doubles
// GOGC=50  → GC when heap grows 50% (more frequent, less memory)
// GOGC=200 → GC when heap triples (less frequent, more memory)
// GOGC=off → disable GC

import "runtime/debug"
debug.SetGCPercent(50) // or GOGC=50 env var

// Go 1.19+: Memory limit
debug.SetMemoryLimit(1 << 30) // 1GB hard limit
// GC becomes more aggressive near limit
```

### GC Metrics

```go
var stats runtime.MemStats
runtime.ReadMemStats(&stats)

fmt.Printf("Alloc: %d MB\n", stats.Alloc/1024/1024)      // current heap
fmt.Printf("TotalAlloc: %d MB\n", stats.TotalAlloc/1024/1024) // cumulative
fmt.Printf("NumGC: %d\n", stats.NumGC)                     // GC cycles
fmt.Printf("PauseTotalNs: %d ms\n", stats.PauseTotalNs/1e6) // total STW
```

---

## 3. Profiling with pprof

### Types of Profiles

| Profile | What it measures |
|---|---|
| CPU | Where time is spent |
| Heap | Current heap allocations |
| Allocs | All allocations (including freed) |
| Goroutine | Goroutine stack traces |
| Mutex | Lock contention |
| Block | Where goroutines block |

### HTTP Endpoint

```go
import _ "net/http/pprof"

// Endpoints available at:
// /debug/pprof/
// /debug/pprof/profile    (CPU, 30s default)
// /debug/pprof/heap       (heap allocations)
// /debug/pprof/goroutine  (goroutine stacks)
// /debug/pprof/mutex      (mutex contention)
// /debug/pprof/block      (blocking)
```

### CLI Usage

```bash
# CPU profile (30 seconds)
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Heap profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine profile
go tool pprof http://localhost:6060/debug/pprof/goroutine

# Inside pprof:
(pprof) top 10          # top functions by CPU/memory
(pprof) list funcName   # source annotated with profile data
(pprof) web             # open flamegraph in browser
(pprof) svg             # generate SVG flamegraph
```

### Programmatic Profiling

```go
import "runtime/pprof"

// CPU
f, _ := os.Create("cpu.prof")
pprof.StartCPUProfile(f)
defer pprof.StopCPUProfile()

// Heap
f, _ := os.Create("heap.prof")
pprof.WriteHeapProfile(f)
```

---

## 4. Benchmarking

### Writing Benchmarks

```go
// File: xxx_test.go
func BenchmarkFoo(b *testing.B) {
    // Setup (not timed)
    data := setupTestData()
    
    b.ResetTimer()  // reset after setup
    b.ReportAllocs() // report allocations
    
    for i := 0; i < b.N; i++ {
        foo(data) // code under test
    }
}
```

### Running Benchmarks

```bash
# Run all benchmarks
go test -bench=. -benchmem ./...

# Run specific benchmark
go test -bench=BenchmarkFoo -benchmem

# With count (for statistical significance)
go test -bench=. -benchmem -count=5

# Compare results
go install golang.org/x/perf/cmd/benchstat@latest
go test -bench=. -count=5 > old.txt
# make changes
go test -bench=. -count=5 > new.txt
benchstat old.txt new.txt
```

### Output Format

```
BenchmarkFoo-8    1000000    1052 ns/op    256 B/op    3 allocs/op
│             │          │              │            │
│             │          │              │            └─ allocations per op
│             │          │              └─ bytes allocated per op
│             │          └─ time per operation
│             └─ number of iterations
└─ GOMAXPROCS
```

### Common Pitfalls

```go
// BAD: compiler might optimize away
func BenchmarkBad(b *testing.B) {
    for i := 0; i < b.N; i++ {
        _ = compute() // compiler might eliminate
    }
}

// GOOD: use package-level var to prevent elimination
var result int
func BenchmarkGood(b *testing.B) {
    var r int
    for i := 0; i < b.N; i++ {
        r = compute()
    }
    result = r // prevent compiler optimization
}
```

---

## 5. Performance Optimization Workflow

```
1. Establish baseline (benchmark)
      ↓
2. Identify bottleneck (metrics, profiling)
      ↓
3. Profile (CPU? Memory? Lock? I/O?)
      ↓
4. Find hot path (top functions)
      ↓
5. Optimize (specific technique)
      ↓
6. Benchmark again (compare)
      ↓
7. Validate in production
```

### Common Optimizations

| Problem | Solution |
|---|---|
| Too many allocations | Reuse objects, sync.Pool, pre-allocate |
| String concatenation | strings.Builder, []byte |
| Interface overhead | Concrete types in hot paths |
| Mutex contention | Sharding, RWMutex, atomic |
| GC pressure | Reduce allocations, larger GOGC |
| Slice growing | Pre-allocate with make([]T, 0, cap) |
| Map growing | Pre-allocate with make(map[K]V, size) |
| Reflection | Code generation, avoid reflect in hot paths |

### Memory Optimization Checklist

- [ ] Pre-allocate slices/maps with known capacity
- [ ] Use `sync.Pool` for frequently allocated temp objects
- [ ] Avoid `interface{}` in hot paths (forces heap allocation)
- [ ] Use `strings.Builder` instead of `+` for string building
- [ ] Pass large structs by pointer (avoid copy)
- [ ] Use value types for small structs (avoid heap)
- [ ] Batch allocations
- [ ] Reuse buffers
