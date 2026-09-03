package main

// Run: go build -gcflags="-m" ./escape_analysis.go
// to see which variables escape to heap

import "fmt"

// --- Examples that DON'T escape (stack allocated) ---

//go:noinline
func stackOnly() int {
	x := 42    // stays on stack
	y := x + 1 // stays on stack
	return y   // return by value
}

//go:noinline
func stackStruct() int {
	type Point struct{ X, Y int }
	p := Point{X: 1, Y: 2} // stays on stack (small, no pointer returned)
	return p.X + p.Y
}

//go:noinline
func stackSlice() int {
	// Known size at compile time → may stay on stack
	s := make([]int, 3) // small, size known → stack (compiler may optimize)
	s[0] = 1
	s[1] = 2
	s[2] = 3
	return s[0] + s[1] + s[2]
}

// --- Examples that ESCAPE (heap allocated) ---

//go:noinline
func escapesReturnPointer() *int {
	x := 42   // x escapes to heap
	return &x // returning address of local variable
}

//go:noinline
func escapesInterface() {
	x := 42
	var i interface{} = x // x escapes: boxed into interface
	_ = i
}

//go:noinline
func escapesClosure() func() int {
	x := 0 // x escapes: captured by returned closure
	return func() int {
		x++
		return x
	}
}

//go:noinline
func escapesDynamicSlice(n int) []int {
	// n unknown at compile time → heap
	return make([]int, n)
}

//go:noinline
func escapesChannel(ch chan *int) {
	x := 42 // x escapes: sent via channel
	ch <- &x
}

type Config struct {
	Host string
	Port int
}

//go:noinline
func escapesLargeStruct() *Config {
	// Even if pointer isn't strictly needed,
	// returning pointer → escapes
	c := Config{Host: "localhost", Port: 8080}
	return &c
}

// --- Optimization examples ---

// BAD: unnecessary heap allocation
//
//go:noinline
func badAllocate() *Config {
	c := &Config{Host: "localhost", Port: 8080} // escapes
	return c
}

// GOOD: let caller control allocation
//
//go:noinline
func goodFillConfig(c *Config) {
	c.Host = "localhost"
	c.Port = 8080
	// c was allocated by caller, may be on caller's stack
}

// BAD: fmt.Sprintf allocates
//
//go:noinline
func badFormat(name string, age int) string {
	return fmt.Sprintf("%s is %d years old", name, age) // allocates
}

// GOOD: pre-sized builder for hot paths
//
//go:noinline
func goodFormat(name string, age int) string {
	// For very hot paths, manual building avoids fmt overhead
	// But readability matters — only optimize when profiling shows it's needed
	buf := make([]byte, 0, 64)
	buf = append(buf, name...)
	buf = append(buf, " is "...)
	buf = append(buf, fmt.Sprintf("%d", age)...)
	buf = append(buf, " years old"...)
	return string(buf)
}

func main() {
	fmt.Println("=== Escape Analysis Demo ===")
	fmt.Println("Run: go build -gcflags=\"-m\" ./escape_analysis.go")
	fmt.Println("")

	// Stack examples
	fmt.Printf("stackOnly: %d\n", stackOnly())
	fmt.Printf("stackStruct: %d\n", stackStruct())
	fmt.Printf("stackSlice: %d\n", stackSlice())

	// Escape examples
	fmt.Printf("escapesReturnPointer: %d\n", *escapesReturnPointer())
	escapesInterface()
	f := escapesClosure()
	fmt.Printf("escapesClosure: %d, %d\n", f(), f())
	s := escapesDynamicSlice(5)
	fmt.Printf("escapesDynamicSlice(5): len=%d\n", len(s))
	c := escapesLargeStruct()
	fmt.Printf("escapesLargeStruct: %s:%d\n", c.Host, c.Port)

	fmt.Println("")
	fmt.Println("Key Rules:")
	fmt.Println("1. Return pointer to local → escapes")
	fmt.Println("2. Store in interface → escapes")
	fmt.Println("3. Closure captures → escapes")
	fmt.Println("4. Channel send → escapes")
	fmt.Println("5. Dynamic size slice → escapes")
	fmt.Println("6. fmt functions → arguments escape (interface{})")
}
