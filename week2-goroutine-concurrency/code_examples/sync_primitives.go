package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	fmt.Println("=== 1. sync.Mutex ===")
	mutexDemo()

	fmt.Println("\n=== 2. sync.RWMutex ===")
	rwMutexDemo()

	fmt.Println("\n=== 3. sync.WaitGroup ===")
	waitGroupDemo()

	fmt.Println("\n=== 4. sync.Once ===")
	onceDemo()

	fmt.Println("\n=== 5. sync.Cond ===")
	condDemo()

	fmt.Println("\n=== 6. sync/atomic ===")
	atomicDemo()

	fmt.Println("\n=== 7. sync.Pool ===")
	poolDemo()
}

// --- 1. Mutex ---

type SafeCounter struct {
	mu    sync.Mutex
	count int
}

func (c *SafeCounter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count++
}

func (c *SafeCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

func mutexDemo() {
	counter := &SafeCounter{}
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Increment()
		}()
	}
	wg.Wait()

	fmt.Printf("  Counter (1000 goroutines): %d\n", counter.Value())

	// Common mistakes:
	fmt.Println("  Common mistakes:")
	fmt.Println("  ✗ Copy mutex after first use")
	fmt.Println("  ✗ Lock twice in same goroutine (deadlock)")
	fmt.Println("  ✗ Forget to unlock (use defer)")
	fmt.Println("  ✗ Unlock without lock")
}

// --- 2. RWMutex ---

type Cache struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewCache() *Cache {
	return &Cache{data: make(map[string]string)}
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock() // multiple readers OK
	defer c.mu.RUnlock()
	val, ok := c.data[key]
	return val, ok
}

func (c *Cache) Set(key, value string) {
	c.mu.Lock() // exclusive write
	defer c.mu.Unlock()
	c.data[key] = value
}

func rwMutexDemo() {
	cache := NewCache()
	cache.Set("name", "Go")

	var wg sync.WaitGroup

	// 10 concurrent readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			val, _ := cache.Get("name")
			_ = val
		}(i)
	}

	// 2 writers
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			cache.Set("name", fmt.Sprintf("Go-%d", id))
		}(i)
	}

	wg.Wait()
	val, _ := cache.Get("name")
	fmt.Printf("  RWMutex cache final value: %s\n", val)
	fmt.Println("  RWMutex: multiple readers OR single writer")
	fmt.Println("  Use when reads >> writes")
}

// --- 3. WaitGroup ---

func waitGroupDemo() {
	var wg sync.WaitGroup

	tasks := []string{"task-1", "task-2", "task-3"}

	for _, task := range tasks {
		wg.Add(1) // MUST be before go statement
		go func(t string) {
			defer wg.Done()
			time.Sleep(10 * time.Millisecond)
			fmt.Printf("  Completed: %s\n", t)
		}(task)
	}

	wg.Wait() // blocks until counter = 0
	fmt.Println("  All tasks done")

	// Rules:
	// - Add() before go statement (not inside goroutine)
	// - Done() = Add(-1)
	// - Wait() blocks until counter reaches 0
	// - Counter must not go negative (panic)
}

// --- 4. sync.Once ---

type Singleton struct {
	Value string
}

var (
	instance *Singleton
	once     sync.Once
)

func GetInstance() *Singleton {
	once.Do(func() {
		fmt.Println("  Creating singleton (only once)")
		instance = &Singleton{Value: "initialized"}
	})
	return instance
}

func onceDemo() {
	var wg sync.WaitGroup

	// 10 goroutines try to get instance
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s := GetInstance()
			_ = s
		}()
	}

	wg.Wait()
	fmt.Printf("  Singleton value: %s\n", GetInstance().Value)
	fmt.Println("  sync.Once: exactly one execution, all others wait")
}

// --- 5. sync.Cond ---

func condDemo() {
	// Cond = condition variable for waiting/signaling
	var mu sync.Mutex
	cond := sync.NewCond(&mu)

	ready := false
	var wg sync.WaitGroup

	// Waiters
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			mu.Lock()
			for !ready { // MUST use loop (spurious wakeups)
				cond.Wait() // atomically: unlock + sleep + relock
			}
			fmt.Printf("  Worker %d: ready!\n", id)
			mu.Unlock()
		}(i)
	}

	// Signal all waiters
	time.Sleep(50 * time.Millisecond)
	mu.Lock()
	ready = true
	mu.Unlock()
	cond.Broadcast() // wake ALL waiters (Signal wakes one)

	wg.Wait()
	fmt.Println("  sync.Cond: Signal (one) vs Broadcast (all)")
	fmt.Println("  Use when: multiple goroutines wait for same condition")
}

// --- 6. Atomic ---

func atomicDemo() {
	var counter int64

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			atomic.AddInt64(&counter, 1)
		}()
	}
	wg.Wait()

	fmt.Printf("  Atomic counter: %d\n", atomic.LoadInt64(&counter))

	// Compare-and-Swap (CAS)
	var flag int32 = 0
	// Only first goroutine succeeds
	swapped := atomic.CompareAndSwapInt32(&flag, 0, 1)
	fmt.Printf("  CAS (0→1): swapped=%v, flag=%d\n", swapped, flag)

	swapped = atomic.CompareAndSwapInt32(&flag, 0, 1) // fails, flag is already 1
	fmt.Printf("  CAS (0→1 again): swapped=%v, flag=%d\n", swapped, flag)

	// atomic.Value — store/load any type atomically
	var config atomic.Value
	type Config struct {
		Host string
		Port int
	}
	config.Store(&Config{Host: "localhost", Port: 8080})
	cfg := config.Load().(*Config)
	fmt.Printf("  atomic.Value: %s:%d\n", cfg.Host, cfg.Port)

	fmt.Println("\n  When to use atomic vs mutex:")
	fmt.Println("  ✓ Atomic: single counter, flag, pointer")
	fmt.Println("  ✓ Mutex: multiple fields, complex logic")
	fmt.Println("  ✓ Atomic is lock-free → no deadlock possible")
}

// --- 7. sync.Pool ---

func poolDemo() {
	// Pool: temporary object reuse to reduce GC pressure
	pool := &sync.Pool{
		New: func() interface{} {
			fmt.Println("  Pool: creating new buffer")
			return make([]byte, 0, 1024)
		},
	}

	// Get from pool (creates new if empty)
	buf := pool.Get().([]byte)
	buf = append(buf, []byte("hello")...)
	fmt.Printf("  Got buffer: %s (len=%d, cap=%d)\n", buf, len(buf), cap(buf))

	// Put back to pool (reset first!)
	buf = buf[:0] // reset length, keep capacity
	pool.Put(buf)

	// Get again (reuses pooled buffer)
	buf2 := pool.Get().([]byte)
	fmt.Printf("  Got reused buffer: len=%d, cap=%d\n", len(buf2), cap(buf2))

	fmt.Println("\n  sync.Pool rules:")
	fmt.Println("  • Objects may be removed at any time (GC)")
	fmt.Println("  • Don't rely on Pool for caching")
	fmt.Println("  • Always reset before Put")
	fmt.Println("  • Good for: buffers, temporary objects in hot paths")
	fmt.Println("  • Example: encoding/json, fmt use sync.Pool internally")
}
