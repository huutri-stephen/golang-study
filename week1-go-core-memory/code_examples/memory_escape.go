package main

import "fmt"

// Run with: go build -gcflags="-m" ./memory_escape.go
// to see escape analysis output

func main() {
	fmt.Println("=== 1. Stack Allocation (no escape) ===")
	stackAllocation()

	fmt.Println("\n=== 2. Heap Allocation (escapes) ===")
	heapAllocation()

	fmt.Println("\n=== 3. Common Escape Scenarios ===")
	commonEscapes()

	fmt.Println("\n=== 4. Value vs Pointer Trade-offs ===")
	valueVsPointer()
}

// --- 1. Stack Allocation ---
// These variables stay on stack (cheap, no GC)

func stackAllocation() {
	// Local value types — stack
	x := 42
	y := "hello"
	arr := [3]int{1, 2, 3}

	// Struct by value — stack
	type Point struct{ X, Y int }
	p := Point{X: 1, Y: 2}

	// Passing by value — stack
	result := addStack(10, 20)

	fmt.Printf("x=%d, y=%s, arr=%v, p=%v, result=%d\n", x, y, arr, p, result)
}

func addStack(a, b int) int {
	sum := a + b // sum stays on stack
	return sum   // return by value
}

// --- 2. Heap Allocation ---
// These variables escape to heap (need GC)

func heapAllocation() {
	// Returning pointer to local → escapes
	p := newPoint(1, 2)
	fmt.Printf("Point from heap: %v\n", *p)

	// Interface boxing → may escape
	var i interface{} = 42
	fmt.Printf("Interface value: %v\n", i)
}

//go:noinline
func newPoint(x, y int) *Point2 {
	p := Point2{X: x, Y: y} // p escapes to heap
	return &p               // returning address of local
}

type Point2 struct{ X, Y int }

// --- 3. Common Escape Scenarios ---

func commonEscapes() {
	fmt.Println("Scenarios where variables escape to heap:")
	fmt.Println("")

	// Scenario 1: Return pointer
	fmt.Println("1. Return pointer to local variable")
	_ = returnPointer()

	// Scenario 2: Assign to interface
	fmt.Println("2. Assign value to interface (boxing)")
	var animal Animal2 = &Dog2{name: "Rex"}
	_ = animal

	// Scenario 3: Closure captures variable
	fmt.Println("3. Closure captures local variable")
	_ = closureCapture()

	// Scenario 4: Send to channel
	fmt.Println("4. Send pointer to channel")
	// ch := make(chan *int)
	// go func() { x := 42; ch <- &x }()

	// Scenario 5: Slice with dynamic size
	fmt.Println("5. Slice with size unknown at compile time")
	_ = dynamicSlice(10)

	// Scenario 6: Append may reallocate
	fmt.Println("6. Slice that grows beyond initial capacity")
	_ = growingSlice()

	// Scenario 7: Map values
	fmt.Println("7. Values stored in maps")
	m := make(map[string]*Point2)
	p := &Point2{X: 1, Y: 2} // escapes because stored in map
	m["origin"] = p
}

func returnPointer() *int {
	x := 42   // x escapes to heap
	return &x // because we return its address
}

type Animal2 interface {
	Speak() string
}
type Dog2 struct{ name string }

func (d *Dog2) Speak() string { return "woof" }

func closureCapture() func() int {
	x := 0 // x escapes — captured by closure
	return func() int {
		x++
		return x
	}
}

func dynamicSlice(n int) []int {
	// n is not known at compile time → heap
	return make([]int, n)
}

func growingSlice() []int {
	s := make([]int, 0, 2)
	for i := 0; i < 100; i++ {
		s = append(s, i)
	}
	return s
}

// --- 4. Value vs Pointer Trade-offs ---

type SmallStruct struct {
	X, Y int
}

type LargeStruct struct {
	Data [1024]byte
	Name string
	Tags []string
}

func valueVsPointer() {
	fmt.Println(`
Value vs Pointer Decision Guide:

┌─────────────────────────────────────────────────────────────┐
│ Use VALUE when:                                             │
├─────────────────────────────────────────────────────────────┤
│ • Small struct (≤ 3-4 fields, no slices/maps)              │
│ • Immutable / no need to modify                            │
│ • Want stack allocation (performance-critical hot path)    │
│ • Short-lived (function scope)                             │
│ • Safe for concurrent read                                 │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ Use POINTER when:                                           │
├─────────────────────────────────────────────────────────────┤
│ • Large struct (avoid copy cost)                           │
│ • Need to modify (mutation required)                       │
│ • Shared across goroutines (with sync)                     │
│ • Implementing interface with pointer receiver             │
│ • Long-lived object                                        │
│ • Optional value (nil = absent)                            │
└─────────────────────────────────────────────────────────────┘

Trade-offs:
• Pointer: 1 heap alloc + GC pressure + cache miss
• Value: copy cost (proportional to size)
• For small structs: copy is cheaper than heap alloc + GC
• Rule of thumb: if struct ≤ 64 bytes, value is usually faster
`)

	// Example: SmallStruct → use value
	s := SmallStruct{X: 1, Y: 2}
	processValue(s)

	// Example: LargeStruct → use pointer
	l := &LargeStruct{Name: "big"}
	processPointer(l)
}

//go:noinline
func processValue(s SmallStruct) int {
	return s.X + s.Y
}

//go:noinline
func processPointer(s *LargeStruct) string {
	return s.Name
}
