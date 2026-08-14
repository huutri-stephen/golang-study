package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== 1. Data Race ===")
	dataRaceDemo()

	fmt.Println("\n=== 2. Deadlock Examples ===")
	deadlockExamples()

	fmt.Println("\n=== 3. Goroutine Leak ===")
	goroutineLeakDemo()

	fmt.Println("\n=== 4. Goroutine Leak Prevention ===")
	leakPrevention()

	fmt.Println("\n=== 5. Starvation ===")
	starvationDemo()
}

// --- 1. Data Race ---

func dataRaceDemo() {
	// BUG: data race - run with `go run -race` to detect
	fmt.Println("  --- BUG (data race) ---")
	fmt.Println("  var counter int")
	fmt.Println("  go func() { counter++ }()  // write")
	fmt.Println("  go func() { _ = counter }() // read")
	fmt.Println("  → Undefined behavior!")
	fmt.Println("")

	// FIX 1: Mutex
	fmt.Println("  --- FIX 1: Mutex ---")
	var mu sync.Mutex
	counter := 0
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}
	wg.Wait()
	fmt.Printf("  Mutex counter: %d\n", counter)

	// FIX 2: Channel
	fmt.Println("  --- FIX 2: Channel ---")
	counterCh := make(chan int, 1)
	counterCh <- 0

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			val := <-counterCh
			val++
			counterCh <- val
		}()
	}
	wg.Wait()
	fmt.Printf("  Channel counter: %d\n", <-counterCh)

	fmt.Println("\n  Detection: go test -race ./...")
	fmt.Println("  Race detector uses ThreadSanitizer")
	fmt.Println("  Overhead: 2-10x slower, 5-10x more memory")
}

// --- 2. Deadlock ---

func deadlockExamples() {
	// Example 1: Self-deadlock (non-reentrant mutex)
	fmt.Println("  --- Example 1: Self-deadlock ---")
	fmt.Println("  var mu sync.Mutex")
	fmt.Println("  mu.Lock()")
	fmt.Println("  mu.Lock() // DEADLOCK! same goroutine")
	fmt.Println("")

	// Example 2: Lock ordering deadlock
	fmt.Println("  --- Example 2: Lock ordering ---")
	fmt.Println("  Goroutine 1: Lock(A) → Lock(B)")
	fmt.Println("  Goroutine 2: Lock(B) → Lock(A)")
	fmt.Println("  → Circular wait = DEADLOCK")
	fmt.Println("")
	fmt.Println("  FIX: Always lock in same order (e.g., alphabetical)")
	fmt.Println("")

	// Example 3: Channel deadlock
	fmt.Println("  --- Example 3: Channel deadlock ---")
	fmt.Println("  ch := make(chan int)")
	fmt.Println("  ch <- 1  // blocks forever (no receiver)")
	fmt.Println("  → All goroutines asleep = DEADLOCK")
	fmt.Println("")

	// Safe lock ordering demo
	fmt.Println("  --- Safe: Consistent lock ordering ---")
	var muA, muB sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			muA.Lock()
			muB.Lock()
			// work
			muB.Unlock()
			muA.Unlock()
		}()
		go func() {
			defer wg.Done()
			muA.Lock() // same order as above!
			muB.Lock()
			// work
			muB.Unlock()
			muA.Unlock()
		}()
	}
	wg.Wait()
	fmt.Println("  Safe lock ordering completed without deadlock")
}

// --- 3. Goroutine Leak ---

func goroutineLeakDemo() {
	before := runtime.NumGoroutine()

	// LEAK 1: Channel with no receiver
	leakyFunc1 := func() {
		ch := make(chan int)
		go func() {
			ch <- 42 // blocks forever — no one reads ch
			// This goroutine is leaked!
		}()
		// function returns without reading from ch
	}

	// LEAK 2: Infinite loop without exit
	leakyFunc2 := func() {
		go func() {
			for {
				// no exit condition, no context check
				time.Sleep(time.Second)
			}
		}()
	}

	// LEAK 3: Blocked on channel receive
	leakyFunc3 := func() {
		ch := make(chan int)
		go func() {
			val := <-ch // blocks forever — no one sends
			_ = val
		}()
		// function returns without sending to ch
	}

	leakyFunc1()
	leakyFunc2()
	leakyFunc3()

	time.Sleep(50 * time.Millisecond)
	after := runtime.NumGoroutine()

	fmt.Printf("  Goroutines before: %d, after: %d (leaked: %d)\n",
		before, after, after-before)
	fmt.Println("  Leaked goroutines consume memory and never get collected!")
	fmt.Println("")
	fmt.Println("  Common causes:")
	fmt.Println("  1. Send to channel with no receiver")
	fmt.Println("  2. Receive from channel with no sender")
	fmt.Println("  3. Infinite loop without exit")
	fmt.Println("  4. Missing context.Done() check")
}

// --- 4. Leak Prevention ---

func leakPrevention() {
	before := runtime.NumGoroutine()

	// PATTERN 1: Context cancellation
	fmt.Println("  --- Pattern 1: Context cancellation ---")
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("  Worker: stopped via context")
				return
			default:
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	time.Sleep(30 * time.Millisecond)
	cancel() // cleanup: cancel the goroutine
	time.Sleep(20 * time.Millisecond)

	// PATTERN 2: Done channel
	fmt.Println("  --- Pattern 2: Done channel ---")
	done := make(chan struct{})
	result := make(chan int, 1)

	go func() {
		select {
		case <-done:
			fmt.Println("  Worker: stopped via done channel")
			return
		case result <- 42:
			return
		}
	}()

	// Either consume result or cancel
	select {
	case v := <-result:
		fmt.Printf("  Got result: %d\n", v)
	case <-time.After(50 * time.Millisecond):
		close(done) // cancel if timeout
	}

	// PATTERN 3: Buffered channel (non-blocking send)
	fmt.Println("  --- Pattern 3: Buffered channel ---")
	ch := make(chan int, 1) // buffer=1, send won't block even if no receiver

	go func() {
		ch <- 42 // won't block because buffer=1
		fmt.Println("  Worker: sent to buffered channel (non-blocking)")
	}()
	time.Sleep(20 * time.Millisecond)

	time.Sleep(50 * time.Millisecond)
	after := runtime.NumGoroutine()
	fmt.Printf("\n  Goroutines before: %d, after: %d (no leaks!)\n", before, after)

	fmt.Println("\n  Prevention checklist:")
	fmt.Println("  ✓ Always provide exit path (context/done channel)")
	fmt.Println("  ✓ Use buffered channels when sender might not have receiver")
	fmt.Println("  ✓ Set timeouts on blocking operations")
	fmt.Println("  ✓ Monitor runtime.NumGoroutine() in production")
	fmt.Println("  ✓ Use goleak in tests: go.uber.org/goleak")
}

// --- 5. Starvation ---

func starvationDemo() {
	// Starvation: goroutine rarely gets to execute because others dominate

	var mu sync.Mutex
	var wg sync.WaitGroup

	// Greedy worker: holds lock for a long time
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			mu.Lock()
			time.Sleep(10 * time.Millisecond) // holds lock long
			mu.Unlock()
			// tiny gap — other goroutines might not get a chance
		}
	}()

	// Starved workers: try to acquire lock
	starvedCount := 0
	var starvedMu sync.Mutex

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				mu.Lock()
				starvedMu.Lock()
				starvedCount++
				starvedMu.Unlock()
				mu.Unlock()
				runtime.Gosched() // yield to let others run
			}
		}(i)
	}

	wg.Wait()
	fmt.Printf("  Starved workers completed %d operations\n", starvedCount)
	fmt.Println("  Starvation happens when one goroutine dominates lock acquisition")
	fmt.Println("  Prevention: keep critical sections short, use fair scheduling")
}
