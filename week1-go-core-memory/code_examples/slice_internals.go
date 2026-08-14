package main

import (
	"fmt"
	"unsafe"
)

// SliceHeader minh hoạ cấu trúc bên trong của slice
// Thực tế: reflect.SliceHeader (deprecated) hoặc unsafe.Pointer
type SliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

func main() {
	fmt.Println("=== 1. Slice Header ===")
	sliceHeaderDemo()

	fmt.Println("\n=== 2. Append & Reallocation ===")
	appendDemo()

	fmt.Println("\n=== 3. Memory Sharing Trap ===")
	memorySharingTrap()

	fmt.Println("\n=== 4. Full Slice Expression ===")
	fullSliceExpression()

	fmt.Println("\n=== 5. Slice Memory Leak ===")
	sliceMemoryLeak()

	fmt.Println("\n=== 6. Interview Question ===")
	interviewQuestion()
}

func sliceHeaderDemo() {
	s := make([]int, 3, 5)
	fmt.Printf("len=%d, cap=%d\n", len(s), cap(s))
	fmt.Printf("Size of slice header: %d bytes\n", unsafe.Sizeof(s))
	// 24 bytes on 64-bit: 8 (pointer) + 8 (len) + 8 (cap)

	// Slice là value type (header), nhưng share underlying array
	s2 := s
	s2[0] = 99
	fmt.Printf("s[0]=%d (shared underlying array)\n", s[0]) // 99
}

func appendDemo() {
	s := make([]int, 0, 4)
	fmt.Printf("Initial: len=%d, cap=%d, ptr=%p\n", len(s), cap(s), s)

	// Append trong capacity — không realloc
	for i := 0; i < 4; i++ {
		s = append(s, i)
		fmt.Printf("After append(%d): len=%d, cap=%d, ptr=%p\n", i, len(s), cap(s), s)
	}

	// Append vượt capacity — realloc!
	s = append(s, 100)
	fmt.Printf("After append(100): len=%d, cap=%d, ptr=%p (NEW ARRAY!)\n", len(s), cap(s), s)

	// Growth strategy demo
	fmt.Println("\n--- Growth Strategy ---")
	var g []int
	prevCap := 0
	for i := 0; i < 20; i++ {
		g = append(g, i)
		if cap(g) != prevCap {
			fmt.Printf("len=%2d, cap=%2d (grew from %d)\n", len(g), cap(g), prevCap)
			prevCap = cap(g)
		}
	}
}

func memorySharingTrap() {
	a := []int{1, 2, 3, 4, 5}
	b := a[1:3] // b = [2, 3], len=2, cap=4

	fmt.Printf("Before: a=%v, b=%v\n", a, b)
	fmt.Printf("b: len=%d, cap=%d\n", len(b), cap(b))

	// Append to b — overwrites a[3] because cap(b) > len(b)!
	b = append(b, 99)
	fmt.Printf("After append(b, 99):\n")
	fmt.Printf("  a=%v (a[3] bị ghi đè!)\n", a)
	fmt.Printf("  b=%v\n", b)

	// Thêm append — vẫn ghi đè a[4]
	b = append(b, 88)
	fmt.Printf("After append(b, 88):\n")
	fmt.Printf("  a=%v\n", a)
	fmt.Printf("  b=%v\n", b)

	// Append vượt cap — tạo array mới, không ảnh hưởng a nữa
	b = append(b, 77)
	fmt.Printf("After append(b, 77) - exceeds cap:\n")
	fmt.Printf("  a=%v (không đổi)\n", a)
	fmt.Printf("  b=%v (new underlying array)\n", b)
}

func fullSliceExpression() {
	a := []int{1, 2, 3, 4, 5}

	// Không dùng full slice expression → dangerous
	b := a[1:3]
	fmt.Printf("b = a[1:3]: len=%d, cap=%d\n", len(b), cap(b)) // cap=4

	// Dùng full slice expression → safe
	c := a[1:3:3] // cap = 3-1 = 2, chỉ bằng len
	fmt.Printf("c = a[1:3:3]: len=%d, cap=%d\n", len(c), cap(c)) // cap=2

	// Append to c → forces reallocation, không ảnh hưởng a
	c = append(c, 999)
	fmt.Printf("After append(c, 999): a=%v (safe!)\n", a)
}

func sliceMemoryLeak() {
	// Simulate: đọc large data, chỉ cần phần nhỏ
	largeData := make([]byte, 1024*1024) // 1MB
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	// BAD: giữ reference tới 1MB array
	bad := largeData[:10]
	_ = bad

	// GOOD: copy ra slice mới
	good := make([]byte, 10)
	copy(good, largeData[:10])
	_ = good

	fmt.Println("BAD: giữ 1MB underlying array chỉ vì 10 bytes")
	fmt.Println("GOOD: copy 10 bytes, cho phép GC 1MB array")
}

func interviewQuestion() {
	// Classic interview question
	fmt.Println("--- Interview: Predict the output ---")

	a := []int{1, 2, 3}
	b := a[:2]
	b = append(b, 10)

	fmt.Printf("a = %v\n", a) // [1, 2, 10] — vì cap(b)=3, append ghi đè a[2]
	fmt.Printf("b = %v\n", b) // [1, 2, 10]

	fmt.Println("\n--- Explanation ---")
	fmt.Println("a := []int{1,2,3} → len=3, cap=3")
	fmt.Println("b := a[:2]        → len=2, cap=3 (shared array)")
	fmt.Println("b = append(b, 10) → len=3, cap=3 (NO realloc, overwrites a[2])")
}
