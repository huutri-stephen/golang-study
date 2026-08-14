package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof" // register pprof endpoints
	"runtime"
	"sync"
	"time"
)

// This example demonstrates how to use pprof for profiling.
// Run this program and use the commands below to profile it.
//
// Start: go run pprof_example.go
//
// Profile commands (while program is running):
//   CPU:       go tool pprof http://localhost:6060/debug/pprof/profile?seconds=10
//   Heap:      go tool pprof http://localhost:6060/debug/pprof/heap
//   Goroutine: go tool pprof http://localhost:6060/debug/pprof/goroutine
//   Mutex:     go tool pprof http://localhost:6060/debug/pprof/mutex
//   Block:     go tool pprof http://localhost:6060/debug/pprof/block
//
// Inside pprof interactive mode:
//   top 10         → top functions by resource usage
//   list funcName  → annotated source code
//   web            → open flamegraph in browser
//   svg            → save flamegraph as SVG
//   peek funcName  → callers and callees

func main() {
	// Enable mutex and block profiling
	runtime.SetMutexProfileFraction(1)
	runtime.SetBlockProfileRate(1)

	// Start pprof HTTP server
	go func() {
		log.Println("pprof server: http://localhost:6060/debug/pprof/")
		log.Fatal(http.ListenAndServe(":6060", nil))
	}()

	// Simulate various workloads for profiling
	fmt.Println("Starting workloads for profiling...")
	fmt.Println("Use go tool pprof to analyze")
	fmt.Println("")

	go cpuIntensiveWork()
	go memoryIntensiveWork()
	go mutexContentionWork()
	go goroutineWork()

	// Print stats periodically
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		printStats()
	}
}

// --- CPU Intensive ---
func cpuIntensiveWork() {
	for {
		// Simulate CPU work
		fibonacci(30)
		time.Sleep(10 * time.Millisecond)
	}
}

func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

// --- Memory Intensive ---
func memoryIntensiveWork() {
	var data [][]byte
	for {
		// Allocate memory
		chunk := make([]byte, 1024*10) // 10KB
		for i := range chunk {
			chunk[i] = byte(rand.Intn(256))
		}
		data = append(data, chunk)

		// Keep only last 1000 chunks (simulate working set)
		if len(data) > 1000 {
			data = data[len(data)-500:]
		}

		time.Sleep(5 * time.Millisecond)
	}
}

// --- Mutex Contention ---
func mutexContentionWork() {
	var mu sync.Mutex
	counter := 0

	for i := 0; i < 10; i++ {
		go func() {
			for {
				mu.Lock()
				counter++
				time.Sleep(time.Millisecond) // hold lock
				mu.Unlock()
				time.Sleep(time.Millisecond)
			}
		}()
	}
}

// --- Goroutine Work ---
func goroutineWork() {
	for {
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			}()
		}
		wg.Wait()
		time.Sleep(100 * time.Millisecond)
	}
}

func printStats() {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	fmt.Printf("--- Stats ---\n")
	fmt.Printf("Goroutines: %d\n", runtime.NumGoroutine())
	fmt.Printf("Heap Alloc: %d MB\n", stats.Alloc/1024/1024)
	fmt.Printf("Heap Sys:   %d MB\n", stats.HeapSys/1024/1024)
	fmt.Printf("Heap Objects: %d\n", stats.HeapObjects)
	fmt.Printf("GC Cycles:  %d\n", stats.NumGC)
	fmt.Printf("GC Pause Total: %d ms\n", stats.PauseTotalNs/1_000_000)
	if stats.NumGC > 0 {
		lastPause := stats.PauseNs[(stats.NumGC+255)%256]
		fmt.Printf("Last GC Pause: %d μs\n", lastPause/1000)
	}
	fmt.Println("")
}
