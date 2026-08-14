# Week 1 – Go Core + Memory – Study Notes

## 1. Slice Internals

### Slice Header

Slice trong Go là một struct 3 fields:

```go
type slice struct {
    array unsafe.Pointer  // pointer to underlying array
    len   int             // current length
    cap   int             // capacity
}
```

**Key insight:** Slice là reference type nhưng bản thân slice header là value type. Khi pass slice vào function, header được copy nhưng underlying array vẫn shared.

### Append Behavior

```go
a := make([]int, 3, 5)  // len=3, cap=5
a = append(a, 4)         // len=4, cap=5, NO reallocation
a = append(a, 5)         // len=5, cap=5, NO reallocation  
a = append(a, 6)         // len=6, cap=10, REALLOCATION! new array created
```

**Growth strategy:**
- cap < 256: double
- cap >= 256: grow by ~25% + some padding

### Memory Sharing Trap

```go
a := []int{1, 2, 3, 4, 5}
b := a[1:3]  // b = [2, 3], shares underlying array with a

b = append(b, 99)  // modifies a[3]! Because b has cap=4
// a = [1, 2, 3, 99, 5]
// b = [2, 3, 99]
```

**Prevention:** Use full slice expression `a[1:3:3]` to limit capacity.

### Memory Leak với Slice

```go
// BAD: giữ reference đến large underlying array
func getFirst(data []byte) []byte {
    return data[:1]  // underlying array never GC'd
}

// GOOD: copy to new slice
func getFirst(data []byte) []byte {
    result := make([]byte, 1)
    copy(result, data[:1])
    return result
}
```

---

## 2. Map Internals

### Structure

Go map sử dụng hash table với bucket-based approach:

```
hmap
├── count (number of elements)
├── B (log2 of number of buckets)
├── buckets (array of bmap)
│   ├── bucket[0]
│   │   ├── tophash[8]   (first byte of hash for quick comparison)
│   │   ├── keys[8]
│   │   └── values[8]
│   ├── bucket[1]
│   └── ...
└── oldbuckets (for incremental growth)
```

### Key Points

1. **Bucket size = 8**: Mỗi bucket chứa tối đa 8 key-value pairs
2. **Overflow bucket**: Khi bucket đầy, tạo overflow bucket (linked list)
3. **Load factor = 6.5**: Khi average > 6.5 elements/bucket → grow
4. **Incremental growth**: Không resize cùng lúc, mỗi access di chuyển 1-2 buckets

### Not Thread-Safe!

```go
// PANIC: concurrent map writes
m := make(map[string]int)
go func() { m["a"] = 1 }()
go func() { m["b"] = 2 }()
```

**Solutions:**
1. `sync.Mutex` – simple, predictable
2. `sync.RWMutex` – read-heavy workloads
3. `sync.Map` – specific use cases (cache with stable keys)

### When to use sync.Map

- Keys are stable (mostly reads, rare writes)
- Disjoint sets of keys per goroutine
- NOT good for: frequent writes, iteration

---

## 3. Interface Internals

### Two Types of Interface

```go
// eface - empty interface (interface{})
type eface struct {
    _type *_type     // type information
    data  unsafe.Pointer  // pointer to data
}

// iface - non-empty interface (has methods)
type iface struct {
    tab  *itab           // type + method table
    data unsafe.Pointer  // pointer to data
}
```

### Nil Interface vs Typed Nil

```go
var i interface{} = nil  // eface{type: nil, data: nil}
// i == nil → TRUE

var p *int = nil
var j interface{} = p    // eface{type: *int, data: nil}
// j == nil → FALSE! (type is not nil)
```

**Critical rule:** Interface is nil ONLY when both type and value are nil.

### Method Set Rules

| Receiver Type | Method Set |
|---|---|
| Value `T` | Methods with value receiver |
| Pointer `*T` | Methods with value AND pointer receiver |

```go
type Animal interface {
    Speak() string
}

type Dog struct{}
func (d *Dog) Speak() string { return "Woof" }

var a Animal = Dog{}   // COMPILE ERROR! Dog doesn't implement Animal
var a Animal = &Dog{}  // OK! *Dog implements Animal
```

### Dynamic Dispatch

Interface method call = indirect call through itab:
1. Load itab from interface
2. Look up method in itab's function table
3. Call with data pointer as receiver

Cost: ~2ns overhead per call (vs direct call)

---

## 4. Memory – Stack vs Heap

### Stack

- **Fast**: Just move stack pointer
- **Automatic**: No GC needed
- **Limited**: Goroutine starts with 2KB-8KB, grows as needed
- **Contiguous**: Go uses contiguous stacks (copy on growth)

### Heap

- **Slower**: Needs allocator (tcmalloc-like)
- **GC managed**: Must be traced and collected
- **Shared**: Accessible across goroutines

### Escape Analysis

Compiler decides stack vs heap at compile time:

```go
// Stack allocated (doesn't escape)
func sum(a, b int) int {
    result := a + b
    return result
}

// Heap allocated (escapes!)
func newInt(x int) *int {
    v := x       // v escapes to heap
    return &v    // returning pointer to local variable
}
```

**Common escape scenarios:**
1. Returning pointer to local variable
2. Sending pointer to channel
3. Storing pointer in interface
4. Closure capturing local variable
5. Variable too large for stack
6. Slice with unknown size at compile time

### Check escape analysis:

```bash
go build -gcflags="-m" ./...
# Output:
# ./main.go:5:2: v escapes to heap
```

### Value vs Pointer Guidelines

**Use value:**
- Small structs (< 3-4 fields)
- Immutable data
- When you want stack allocation

**Use pointer:**
- Large structs
- Need mutation
- Implementing interface with pointer receiver
- Shared state

**Trade-off:** Pointer = 1 heap allocation + GC pressure vs Value = copy cost

---

## 5. String / Rune / Byte

### String Header

```go
type stringHeader struct {
    Data unsafe.Pointer  // pointer to byte array
    Len  int             // length in bytes
}
```

**Key facts:**
- Strings are immutable
- Strings are UTF-8 encoded
- `len(s)` returns bytes, not characters
- `[]rune(s)` converts to Unicode code points

```go
s := "Hello, 世界"
len(s)          // 13 (bytes)
len([]rune(s))  // 9 (characters)

// Iteration
for i, r := range s {
    // i = byte index, r = rune (Unicode code point)
}
```

---

## 6. Error Handling

### Error Interface

```go
type error interface {
    Error() string
}
```

### Best Practices

```go
// Sentinel errors
var ErrNotFound = errors.New("not found")

// Custom error types
type ValidationError struct {
    Field   string
    Message string
}
func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation: %s - %s", e.Field, e.Message)
}

// Wrapping
func doSomething() error {
    err := db.Query(...)
    if err != nil {
        return fmt.Errorf("doSomething: %w", err)
    }
    return nil
}

// Checking
if errors.Is(err, ErrNotFound) { ... }
var ve *ValidationError
if errors.As(err, &ve) { ... }
```

---

## 7. Defer

### Rules

1. **LIFO order**: Last deferred, first executed
2. **Arguments evaluated immediately**: At defer statement, not at execution
3. **Can modify named return values**

```go
func example() (result int) {
    defer func() { result++ }()  // modifies return value
    return 0  // result = 1 (not 0!)
}

// Argument evaluation
func main() {
    x := 0
    defer fmt.Println(x)  // prints 0, not 1
    x = 1
}
```

---

## 8. Generics (Go 1.18+)

```go
// Type constraint
type Number interface {
    ~int | ~float64 | ~int64
}

// Generic function
func Sum[T Number](numbers []T) T {
    var total T
    for _, n := range numbers {
        total += n
    }
    return total
}

// Generic type
type Stack[T any] struct {
    items []T
}

func (s *Stack[T]) Push(item T) {
    s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() (T, bool) {
    if len(s.items) == 0 {
        var zero T
        return zero, false
    }
    item := s.items[len(s.items)-1]
    s.items = s.items[:len(s.items)-1]
    return item, true
}
```
