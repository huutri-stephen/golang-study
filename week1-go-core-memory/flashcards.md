# Week 1 – Flashcards / Q&A

## Slice

### Q: Slice header gồm những gì?
**A:** 3 fields: pointer to underlying array, len (int), cap (int). Total 24 bytes on 64-bit system.

---

### Q: Khi nào `append` tạo array mới?
**A:** Khi `len == cap`, append allocates new array. Growth: double if cap < 256, else ~25% increase.

---

### Q: Giải thích output:
```go
a := []int{1, 2, 3}
b := a[:2]
b = append(b, 10)
// a = ? b = ?
```
**A:** 
- `a = [1, 2, 10]` — vì b share underlying array với a, và `cap(b) = 3 > len(b) = 2`, nên append ghi đè `a[2]`
- `b = [1, 2, 10]`

---

### Q: Làm sao tránh memory leak khi slice large array?
**A:** Copy data sang slice mới thay vì re-slice. Re-slice giữ reference tới toàn bộ underlying array, prevent GC.

---

### Q: Full slice expression `a[low:high:max]` dùng khi nào?
**A:** Khi muốn limit capacity của sub-slice, tránh append ảnh hưởng original slice. `cap = max - low`.

---

## Map

### Q: Go map có thread-safe không?
**A:** Không. Concurrent read+write hoặc write+write gây panic. Cần `sync.Mutex`, `sync.RWMutex`, hoặc `sync.Map`.

---

### Q: Map bucket chứa bao nhiêu elements?
**A:** 8 key-value pairs per bucket. Overflow buckets được tạo khi bucket đầy.

---

### Q: Khi nào map grow?
**A:** Khi load factor > 6.5 (average elements/bucket) hoặc quá nhiều overflow buckets. Growth là incremental, không resize cùng lúc.

---

### Q: Khi nào dùng `sync.Map` thay vì `Mutex + map`?
**A:** Khi keys stable (mostly reads), hoặc disjoint key sets per goroutine. Không tốt cho frequent writes hay iteration.

---

### Q: Tại sao không thể lấy address of map element?
```go
m := map[string]Point{"a": {1, 2}}
// &m["a"] → compile error
```
**A:** Vì map có thể relocate elements khi grow. Address sẽ invalid sau growth.

---

## Interface

### Q: Nil interface vs typed nil – khác nhau thế nào?
**A:**
- Nil interface: cả type và value đều nil → `== nil` returns true
- Typed nil: type không nil, value là nil → `== nil` returns FALSE

```go
var err error = (*MyError)(nil)  // typed nil
err == nil  // false!
```

---

### Q: Method set của `T` vs `*T`?
**A:**
- `T`: chỉ có value receiver methods
- `*T`: có cả value receiver + pointer receiver methods

→ Nếu interface yêu cầu pointer receiver method, phải dùng `*T` để implement.

---

### Q: Interface internals – iface vs eface?
**A:**
- `eface`: empty interface `interface{}` — chứa `_type` + `data`
- `iface`: non-empty interface — chứa `itab` (type + method table) + `data`

---

### Q: Cost of interface method call?
**A:** ~2ns overhead vs direct call. Indirect call through itab function table. Không thể inline.

---

## Memory

### Q: Escape analysis là gì?
**A:** Compiler analysis tại compile-time quyết định variable allocate trên stack hay heap. Nếu variable "escapes" function scope → heap.

---

### Q: Liệt kê các trường hợp variable escape to heap?
**A:**
1. Return pointer to local variable
2. Send pointer to channel
3. Store in interface value
4. Closure capture
5. Too large for stack
6. Unknown size at compile time (e.g., `make([]int, n)`)

---

### Q: Stack vs Heap – trade-offs?
**A:**
- Stack: fast (pointer bump), no GC, automatic cleanup, limited to function scope
- Heap: slower (allocator), GC pressure, shared across goroutines, longer lifetime

---

### Q: Goroutine stack size?
**A:** Starts at 2KB-8KB (depends on Go version), grows dynamically. Go uses contiguous stacks — khi grow, copy toàn bộ stack sang vùng nhớ lớn hơn.

---

### Q: Khi nào dùng value, khi nào dùng pointer?
**A:**
- Value: small structs, immutable, want stack allocation
- Pointer: large structs, need mutation, shared state, implementing interface with pointer receiver

---

## Error Handling

### Q: `errors.Is` vs `errors.As`?
**A:**
- `errors.Is(err, target)`: check nếu err chain chứa target (value comparison)
- `errors.As(err, &target)`: check nếu err chain chứa type matching target, và assign vào target

---

### Q: Error wrapping – `%w` vs `%v`?
**A:**
- `%w`: wraps error, preserves chain → `errors.Is`/`errors.As` hoạt động
- `%v`: formats error string, BREAKS chain → không thể unwrap

---

## Defer

### Q: Defer execution order?
**A:** LIFO (Last In, First Out). Deferred functions execute in reverse order.

---

### Q: Output là gì?
```go
func f() int {
    x := 0
    defer func() { x++ }()
    return x
}
```
**A:** Returns 0. Vì return value không phải named return, defer modify local `x` nhưng return value đã được set.

---

### Q: Output là gì?
```go
func f() (x int) {
    defer func() { x++ }()
    return 0
}
```
**A:** Returns 1. Named return `x` được set = 0 bởi return, rồi defer tăng lên 1.

---

## Generics

### Q: `~int` nghĩa là gì trong type constraint?
**A:** Underlying type approximation. `~int` match tất cả types có underlying type là `int`, bao gồm `type MyInt int`.

---

### Q: `any` vs `comparable`?
**A:**
- `any` = `interface{}`, cho phép mọi type
- `comparable`: types hỗ trợ `==` và `!=` (dùng được làm map key)

---

### Q: Khi nào KHÔNG nên dùng generics?
**A:**
- Khi `interface{}` + type assertion đủ tốt
- Khi chỉ có 1-2 types cụ thể
- Khi code đơn giản hơn không có generics
- Idiomatic Go ưu tiên simplicity over abstraction
