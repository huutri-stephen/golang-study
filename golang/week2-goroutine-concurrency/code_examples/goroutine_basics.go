package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== 1. Goroutine Basics ===")
	goroutineBasics()

	fmt.Println("\n=== 2. GOMAXPROCS ===")
	gomaxprocsDemo()

	fmt.Println("\n=== 3. Goroutine Lifecycle ===")
	goroutineLifecycle()

	fmt.Println("\n=== 4. Goroutine vs Thread Cost ===")
	goroutineCost()

	fmt.Println("\n=== 5. Goroutine Closure Trap ===")
	closureTrap()
}

func goroutineBasics() {
	// Goroutine = lightweight thread managed by Go runtime
	// Created with `go` keyword
	// No return value, no join (use channels/waitgroup instead)

	var wg sync.WaitGroup

	wg.Add(3)
	go func() {
		defer wg.Done()
		fmt.Println("  Goroutine 1")
	}()

	go func() {
		defer wg.Done()
		fmt.Println("  Goroutine 2")
	}()

	go func() {
		defer wg.Done()
		fmt.Println("  Goroutine 3")
	}()

	wg.Wait()
	fmt.Println("  All goroutines completed")
}

func gomaxprocsDemo() {
	// GOMAXPROCS = number of P (logical processors)
	// = max goroutines running in parallel
	numCPU := runtime.NumCPU()
	currentProcs := runtime.GOMAXPROCS(0) // 0 = query without changing

	fmt.Printf("  NumCPU: %d\n", numCPU)
	fmt.Printf("  GOMAXPROCS: %d\n", currentProcs)
	fmt.Printf("  NumGoroutine: %d\n", runtime.NumGoroutine())

	// Note: GOMAXPROCS limits parallelism, NOT concurrency
	// You can have millions of goroutines with GOMAXPROCS=1
	// They just won't run in parallel (only concurrent)
}

func goroutineLifecycle() {
	fmt.Println("  Goroutine states:")
	fmt.Println("  _Grunnable → _Grunning → _Gwaiting → _Grunnable → ...")
	fmt.Println("  _Grunning → _Gsyscall → _Grunnable")
	fmt.Println("  _Grunning → _Gdead (finished)")

	// Demo: goroutine count changes
	before := runtime.NumGoroutine()

	done := make(chan struct{})
	go func() {
		time.Sleep(100 * time.Millisecond)
		close(done)
	}()

	during := runtime.NumGoroutine()
	<-done
	time.Sleep(10 * time.Millisecond) // let goroutine finish
	after := runtime.NumGoroutine()

	fmt.Printf("  Goroutines — before: %d, during: %d, after: %d\n", before, during, after)
}

func goroutineCost() {
	// Demonstrate creating many goroutines
	const n = 100_000
	start := time.Now()

	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			// minimal work
			runtime.Gosched()
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	fmt.Printf("  Created and completed %d goroutines in %v\n", n, elapsed)
	fmt.Printf("  Average: %v per goroutine\n", elapsed/time.Duration(n))
}

func closureTrap() {
	// COMMON BUG: closure captures loop variable by reference

	// BUG: all goroutines may print same value
	fmt.Println("  BUG version (may print same i):")
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Printf("    i=%d\n", i) // captures &i, not value of i
		}()
	}
	wg.Wait()

	// FIX 1: pass as parameter
	fmt.Println("  FIX 1 - pass as parameter:")
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(i int) { // i is copied
			defer wg.Done()
			fmt.Printf("    i=%d\n", i)
		}(i)
	}
	wg.Wait()

	// FIX 2: local variable (Go 1.22+ fixes this automatically in for loops)
	fmt.Println("  FIX 2 - local copy:")
	for i := 0; i < 3; i++ {
		i := i // shadow with local copy
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Printf("    i=%d\n", i)
		}()
	}
	wg.Wait()

	// Note: Go 1.22+ changes loop variable semantics
	// Each iteration gets its own variable copy
	fmt.Println("  Note: Go 1.22+ fixes loop variable capture automatically")
}
