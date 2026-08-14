package main

import (
	"fmt"
	"sync"
)

func main() {
	fmt.Println("=== 1. Map Basics ===")
	mapBasics()

	fmt.Println("\n=== 2. Map Iteration Order ===")
	mapIterationOrder()

	fmt.Println("\n=== 3. Map Concurrent Access Problem ===")
	mapConcurrentProblem()

	fmt.Println("\n=== 4. Safe Concurrent Map with Mutex ===")
	safeConcurrentMap()

	fmt.Println("\n=== 5. sync.Map Usage ===")
	syncMapDemo()

	fmt.Println("\n=== 6. Map Internals Explanation ===")
	mapInternals()
}

func mapBasics() {
	// Zero value of map is nil
	var m map[string]int
	fmt.Printf("nil map: %v, len=%d\n", m, len(m))
	// m["key"] = 1 → PANIC! assignment to nil map

	// Must initialize
	m = make(map[string]int)
	m["one"] = 1
	m["two"] = 2

	// Check existence
	val, ok := m["three"]
	fmt.Printf("m[\"three\"]: val=%d, ok=%v\n", val, ok)

	// Delete
	delete(m, "one")
	fmt.Printf("After delete: %v\n", m)

	// Cannot take address of map element
	// p := &m["two"] → compile error
	// Reason: map có thể relocate khi grow
}

func mapIterationOrder() {
	m := map[int]string{
		1: "one",
		2: "two",
		3: "three",
		4: "four",
		5: "five",
	}

	fmt.Println("Iteration 1:")
	for k, v := range m {
		fmt.Printf("  %d: %s\n", k, v)
	}

	fmt.Println("Iteration 2 (order may differ!):")
	for k, v := range m {
		fmt.Printf("  %d: %s\n", k, v)
	}

	fmt.Println("→ Map iteration order is NOT guaranteed and intentionally randomized")
}

func mapConcurrentProblem() {
	// WARNING: This demonstrates the problem — do NOT use in production
	fmt.Println("Concurrent map access WITHOUT protection → PANIC")
	fmt.Println("(Skipping actual execution to avoid crash)")
	fmt.Println("")
	fmt.Println("Code that panics:")
	fmt.Println(`
  m := make(map[int]int)
  go func() {
      for i := 0; i < 1000; i++ {
          m[i] = i  // concurrent write
      }
  }()
  go func() {
      for i := 0; i < 1000; i++ {
          _ = m[i]  // concurrent read
      }
  }()
  // → fatal error: concurrent map read and map write
`)
}

func safeConcurrentMap() {
	// Solution 1: sync.Mutex
	type SafeMap struct {
		mu sync.RWMutex
		m  map[string]int
	}

	sm := &SafeMap{m: make(map[string]int)}

	// Write with Lock
	write := func(key string, val int) {
		sm.mu.Lock()
		defer sm.mu.Unlock()
		sm.m[key] = val
	}

	// Read with RLock
	read := func(key string) (int, bool) {
		sm.mu.RLock()
		defer sm.mu.RUnlock()
		val, ok := sm.m[key]
		return val, ok
	}

	// Concurrent usage
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			write(fmt.Sprintf("key%d", i), i)
		}(i)
	}
	wg.Wait()

	val, ok := read("key5")
	fmt.Printf("SafeMap: key5=%d, ok=%v\n", val, ok)
}

func syncMapDemo() {
	// sync.Map — optimized for specific use cases:
	// 1. Keys are stable (mostly reads)
	// 2. Disjoint key sets per goroutine
	var m sync.Map

	// Store
	m.Store("name", "Go")
	m.Store("version", "1.21")

	// Load
	val, ok := m.Load("name")
	fmt.Printf("Load(\"name\"): %v, ok=%v\n", val, ok)

	// LoadOrStore — load existing or store new
	actual, loaded := m.LoadOrStore("name", "Rust")
	fmt.Printf("LoadOrStore(\"name\", \"Rust\"): actual=%v, loaded=%v\n", actual, loaded)

	// Delete
	m.Delete("version")

	// Range — iterate
	m.Store("a", 1)
	m.Store("b", 2)
	m.Range(func(key, value interface{}) bool {
		fmt.Printf("  %v: %v\n", key, value)
		return true // continue iteration
	})

	fmt.Println("\nWhen to use sync.Map:")
	fmt.Println("  ✓ Cache with stable keys (read-heavy)")
	fmt.Println("  ✓ Each goroutine writes to disjoint keys")
	fmt.Println("  ✗ Frequent writes to same keys")
	fmt.Println("  ✗ Need to iterate frequently")
	fmt.Println("  ✗ Need len() — sync.Map has no Len()")
}

func mapInternals() {
	fmt.Println(`
Map Internal Structure (hmap):
┌─────────────────────────────┐
│ hmap                        │
├─────────────────────────────┤
│ count    int                │ ← number of elements
│ B        uint8              │ ← log2(number of buckets)
│ hash0    uint32             │ ← hash seed (randomized)
│ buckets  unsafe.Pointer     │ ← → array of bmap
│ oldbuckets unsafe.Pointer   │ ← → old buckets (during growth)
│ nevacuate uintptr           │ ← evacuation progress
└─────────────────────────────┘

Each Bucket (bmap):
┌─────────────────────────────┐
│ tophash  [8]uint8           │ ← top byte of hash (fast compare)
│ keys     [8]keyType         │
│ values   [8]valueType       │
│ overflow *bmap              │ ← pointer to overflow bucket
└─────────────────────────────┘

Key Points:
• 8 key-value pairs per bucket
• tophash enables fast rejection without full key compare
• Load factor threshold: 6.5
• Growth is incremental (not all-at-once)
• Hash function is randomized per map instance (hash0)
`)
}
