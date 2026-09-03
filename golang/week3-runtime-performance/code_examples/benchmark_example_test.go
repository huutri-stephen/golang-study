package main

import (
	"fmt"
	"strings"
	"testing"
)

// Run: go test -bench=. -benchmem ./benchmark_example_test.go

// --- String Concatenation Benchmarks ---

// BAD: O(n²) allocations
func concatWithPlus(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		s += "a"
	}
	return s
}

// GOOD: O(n) with strings.Builder
func concatWithBuilder(n int) string {
	var b strings.Builder
	b.Grow(n) // pre-allocate
	for i := 0; i < n; i++ {
		b.WriteString("a")
	}
	return b.String()
}

// GOOD: O(n) with []byte
func concatWithBytes(n int) string {
	buf := make([]byte, 0, n)
	for i := 0; i < n; i++ {
		buf = append(buf, 'a')
	}
	return string(buf)
}

func BenchmarkConcatPlus(b *testing.B) {
	for i := 0; i < b.N; i++ {
		concatWithPlus(100)
	}
}

func BenchmarkConcatBuilder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		concatWithBuilder(100)
	}
}

func BenchmarkConcatBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		concatWithBytes(100)
	}
}

// --- Slice Pre-allocation Benchmarks ---

func sliceNoPrealloc(n int) []int {
	var s []int // no pre-allocation
	for i := 0; i < n; i++ {
		s = append(s, i)
	}
	return s
}

func slicePrealloc(n int) []int {
	s := make([]int, 0, n) // pre-allocate capacity
	for i := 0; i < n; i++ {
		s = append(s, i)
	}
	return s
}

func BenchmarkSliceNoPrealloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sliceNoPrealloc(1000)
	}
}

func BenchmarkSlicePrealloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		slicePrealloc(1000)
	}
}

// --- Map Pre-allocation Benchmarks ---

func mapNoPrealloc(n int) map[int]int {
	m := make(map[int]int) // no hint
	for i := 0; i < n; i++ {
		m[i] = i
	}
	return m
}

func mapPrealloc(n int) map[int]int {
	m := make(map[int]int, n) // with size hint
	for i := 0; i < n; i++ {
		m[i] = i
	}
	return m
}

func BenchmarkMapNoPrealloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mapNoPrealloc(1000)
	}
}

func BenchmarkMapPrealloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mapPrealloc(1000)
	}
}

// --- Value vs Pointer Benchmarks ---

type SmallStruct struct {
	X, Y int
}

type LargeStruct struct {
	Data [256]byte
}

//go:noinline
func processSmallValue(s SmallStruct) int { return s.X + s.Y }

//go:noinline
func processSmallPointer(s *SmallStruct) int { return s.X + s.Y }

//go:noinline
func processLargeValue(s LargeStruct) byte { return s.Data[0] }

//go:noinline
func processLargePointer(s *LargeStruct) byte { return s.Data[0] }

func BenchmarkSmallStructValue(b *testing.B) {
	s := SmallStruct{X: 1, Y: 2}
	for i := 0; i < b.N; i++ {
		processSmallValue(s)
	}
}

func BenchmarkSmallStructPointer(b *testing.B) {
	s := &SmallStruct{X: 1, Y: 2}
	for i := 0; i < b.N; i++ {
		processSmallPointer(s)
	}
}

func BenchmarkLargeStructValue(b *testing.B) {
	s := LargeStruct{}
	s.Data[0] = 42
	for i := 0; i < b.N; i++ {
		processLargeValue(s)
	}
}

func BenchmarkLargeStructPointer(b *testing.B) {
	s := &LargeStruct{}
	s.Data[0] = 42
	for i := 0; i < b.N; i++ {
		processLargePointer(s)
	}
}

// --- Sprintf vs manual formatting ---

func BenchmarkSprintf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = fmt.Sprintf("user:%d:profile", 12345)
	}
}

func BenchmarkManualFormat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var buf strings.Builder
		buf.WriteString("user:")
		buf.WriteString("12345")
		buf.WriteString(":profile")
		_ = buf.String()
	}
}

// --- Preventing compiler optimization ---

// Package-level var prevents dead code elimination
var globalResult int
var globalString string

func BenchmarkCorrectPattern(b *testing.B) {
	var r int
	for i := 0; i < b.N; i++ {
		r = processSmallValue(SmallStruct{X: i, Y: i + 1})
	}
	globalResult = r // prevent optimization
}
