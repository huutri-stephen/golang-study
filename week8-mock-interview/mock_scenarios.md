# Mock Interview Scenarios

## Scenario 1: Go Coding – Concurrent-Safe Cache with TTL

**Task:** Implement a thread-safe in-memory cache with TTL in Go.

**Requirements:**
- Get(key) — returns value if exists and not expired
- Set(key, value, ttl) — stores with expiration
- Thread-safe for concurrent access
- Expired entries should be cleaned up

```go
package main

import (
	"sync"
	"time"
)

type Cache struct {
	mu      sync.RWMutex
	entries map[string]*entry
}

type entry struct {
	value     interface{}
	expiresAt time.Time
}

func NewCache() *Cache {
	c := &Cache{entries: make(map[string]*entry)}
	go c.cleanup()
	return c
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, ok := c.entries[key]
	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.value, true
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = &entry{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

func (c *Cache) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, e := range c.entries {
			if now.After(e.expiresAt) {
				delete(c.entries, key)
			}
		}
		c.mu.Unlock()
	}
}
```

**Follow-up questions:**
- How would you shard this cache for less lock contention?
- What about LRU eviction?
- How would you add metrics (hit rate, size)?

---

## Scenario 2: Go Coding – Worker Pool with Timeout

**Task:** Implement a bounded worker pool that processes jobs with individual timeouts.

```go
package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Job struct {
	ID      int
	Timeout time.Duration
	Work    func(ctx context.Context) (string, error)
}

type Result struct {
	JobID  int
	Output string
	Err    error
}

func WorkerPool(ctx context.Context, numWorkers int, jobs <-chan Job) <-chan Result {
	results := make(chan Result, len(jobs))
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for job := range jobs {
				// Per-job timeout
				jobCtx, cancel := context.WithTimeout(ctx, job.Timeout)
				
				resultCh := make(chan Result, 1)
				go func() {
					output, err := job.Work(jobCtx)
					resultCh <- Result{JobID: job.ID, Output: output, Err: err}
				}()

				select {
				case r := <-resultCh:
					results <- r
				case <-jobCtx.Done():
					results <- Result{JobID: job.ID, Err: jobCtx.Err()}
				}
				cancel()
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

func main() {
	jobs := make(chan Job, 5)
	ctx := context.Background()

	// Send jobs
	go func() {
		for i := 0; i < 5; i++ {
			id := i
			jobs <- Job{
				ID:      id,
				Timeout: 2 * time.Second,
				Work: func(ctx context.Context) (string, error) {
					time.Sleep(time.Duration(id) * time.Second)
					return fmt.Sprintf("result-%d", id), nil
				},
			}
		}
		close(jobs)
	}()

	results := WorkerPool(ctx, 3, jobs)
	for r := range results {
		if r.Err != nil {
			fmt.Printf("Job %d: ERROR %v\n", r.JobID, r.Err)
		} else {
			fmt.Printf("Job %d: %s\n", r.JobID, r.Output)
		}
	}
}
```

---

## Scenario 3: System Design – Payment System

**Prompt:** "Design a payment system that supports 10K TPS."

### Expected Answer Structure:

**1. Requirements:**
- Accept payments (credit card, bank transfer)
- Exactly-once processing (idempotency)
- Multi-currency support
- Audit trail for every transaction
- Reconciliation with payment providers
- P99 latency < 500ms

**2. Scale:**
- 10K TPS write (payments)
- 50K TPS read (status checks)
- ~1B transactions/year → ~1TB storage

**3. Architecture:**
```
Client
  ↓
API Gateway (rate limit, auth)
  ↓
Payment Service (idempotency check, validation)
  ↓
Payment Processor (state machine, retry)
  ↓
PSP Adapter (Stripe, PayPal, Bank)
  ↓
Ledger Service (double-entry bookkeeping)
  ↓
Reconciliation Service (daily batch)
```

**4. Key Design Points:**
- Idempotency: Redis store with 24h TTL
- State machine: CREATED → PROCESSING → SUCCEEDED/FAILED
- Double-entry: every transaction has debit + credit entries
- Outbox pattern: publish events reliably
- Retry: exponential backoff for PSP calls
- Reconciliation: daily job matches internal vs PSP records

---

## Scenario 4: Production Troubleshooting

**Prompt:** "Your Go service's memory usage grows 100MB/hour and OOMs after 12 hours."

### Expected Debugging Approach:

**Step 1: Confirm and measure**
```bash
# Check memory growth pattern
curl http://service:6060/debug/pprof/heap > heap1.prof
# Wait 10 minutes
curl http://service:6060/debug/pprof/heap > heap2.prof

go tool pprof -diff_base=heap1.prof heap2.prof
```

**Step 2: Common Go memory leak causes**
1. Goroutine leak (check goroutine count)
2. Unbounded slice/map growth
3. Unclosed resources (response body, file handles)
4. time.Ticker not stopped
5. String/slice holding reference to large backing array

**Step 3: Profile**
```bash
# Goroutine count
curl http://service:6060/debug/pprof/goroutine?debug=1

# Heap allocations
go tool pprof http://service:6060/debug/pprof/heap
(pprof) top 20 -inuse_space
(pprof) list suspectedFunction
```

**Step 4: Fix based on findings**
- Goroutine leak → add context cancellation
- Unbounded cache → add max size + eviction
- Unclosed body → `defer resp.Body.Close()`
- Large slice reference → copy needed data to new slice

**Step 5: Verify**
- Deploy fix to canary
- Monitor memory for 24h
- Confirm growth stopped

---

## Scenario 5: Behavioral – Technical Decision Making

**Prompt:** "Tell me about a time you had to make a difficult technical decision."

### Sample Answer (STAR):

**Situation:** We had a monolithic service handling 5K RPS with increasing latency. Team debated microservices vs optimization.

**Task:** Decide the approach and lead implementation.

**Action:**
1. Profiled the monolith — identified 3 independent hot paths
2. Evaluated options:
   - Option A: Optimize monolith (faster, less risk, limited ceiling)
   - Option B: Full microservices (high cost, long timeline)
   - Option C: Extract 2 critical paths, keep rest monolithic (hybrid)
3. Presented trade-offs to team with data
4. Chose Option C: extract payment and notification into services
5. Used strangler fig pattern — gradual migration

**Result:**
- Latency improved 3x for critical paths
- Monolith complexity reduced
- Team learned microservice patterns incrementally
- Could scale payment service independently during peak

**Key learning:** Don't default to trends (microservices). Measure first, solve the actual bottleneck, prefer incremental over big-bang.
