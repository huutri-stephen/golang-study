package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== 1. Simple Worker Pool ===")
	simpleWorkerPool()

	fmt.Println("\n=== 2. Worker Pool with Results ===")
	workerPoolWithResults()

	fmt.Println("\n=== 3. Worker Pool with Context Cancellation ===")
	workerPoolWithContext()

	fmt.Println("\n=== 4. Dynamic Worker Pool (Semaphore) ===")
	semaphorePool()

	fmt.Println("\n=== 5. Rate-Limited Worker Pool ===")
	rateLimitedPool()
}

// --- 1. Simple Worker Pool ---

func simpleWorkerPool() {
	const numWorkers = 3
	const numJobs = 10

	jobs := make(chan int, numJobs)
	var wg sync.WaitGroup

	// Start workers
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for job := range jobs {
				fmt.Printf("  Worker %d processing job %d\n", id, job)
				time.Sleep(20 * time.Millisecond) // simulate work
			}
		}(w)
	}

	// Send jobs
	for j := 0; j < numJobs; j++ {
		jobs <- j
	}
	close(jobs) // signal no more jobs

	wg.Wait()
	fmt.Println("  All jobs completed")
}

// --- 2. Worker Pool with Results ---

type Job struct {
	ID    int
	Input int
}

type Result struct {
	Job    Job
	Output int
	Err    error
}

func workerPoolWithResults() {
	const numWorkers = 3
	const numJobs = 8

	jobs := make(chan Job, numJobs)
	results := make(chan Result, numJobs)

	// Start workers
	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for job := range jobs {
				// Process job
				output := job.Input * job.Input
				time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
				results <- Result{Job: job, Output: output}
			}
		}(w)
	}

	// Send jobs
	for j := 0; j < numJobs; j++ {
		jobs <- Job{ID: j, Input: j + 1}
	}
	close(jobs)

	// Close results when all workers done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	for result := range results {
		fmt.Printf("  Job %d: %d² = %d\n", result.Job.ID, result.Job.Input, result.Output)
	}
}

// --- 3. Worker Pool with Context ---

func workerPoolWithContext() {
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	const numWorkers = 3
	jobs := make(chan int, 20)
	var wg sync.WaitGroup

	// Start workers that respect context
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					fmt.Printf("  Worker %d: cancelled (%v)\n", id, ctx.Err())
					return
				case job, ok := <-jobs:
					if !ok {
						return
					}
					// Simulate work that checks context
					select {
					case <-ctx.Done():
						fmt.Printf("  Worker %d: cancelled during job %d\n", id, job)
						return
					case <-time.After(time.Duration(rand.Intn(100)) * time.Millisecond):
						fmt.Printf("  Worker %d: completed job %d\n", id, job)
					}
				}
			}
		}(w)
	}

	// Send many jobs (some won't be processed due to timeout)
	for j := 0; j < 20; j++ {
		select {
		case jobs <- j:
		case <-ctx.Done():
			fmt.Printf("  Producer: context cancelled, sent %d jobs\n", j)
			break
		}
	}
	close(jobs)

	wg.Wait()
	fmt.Println("  Pool shutdown complete")
}

// --- 4. Semaphore Pattern ---

func semaphorePool() {
	// Use buffered channel as semaphore to limit concurrency
	const maxConcurrent = 3
	sem := make(chan struct{}, maxConcurrent)

	var wg sync.WaitGroup
	tasks := 10

	for i := 0; i < tasks; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			sem <- struct{}{}        // acquire (blocks if full)
			defer func() { <-sem }() // release

			fmt.Printf("  Task %d running (max %d concurrent)\n", id, maxConcurrent)
			time.Sleep(50 * time.Millisecond)
		}(i)
	}

	wg.Wait()
	fmt.Printf("  All %d tasks completed with max %d concurrency\n", tasks, maxConcurrent)
}

// --- 5. Rate-Limited Worker Pool ---

func rateLimitedPool() {
	// Process at most N requests per second
	const rateLimit = 5 // 5 per second
	const numRequests = 15

	ticker := time.NewTicker(time.Second / time.Duration(rateLimit))
	defer ticker.Stop()

	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < numRequests; i++ {
		<-ticker.C // wait for next tick (rate limit)
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			elapsed := time.Since(start)
			fmt.Printf("  Request %2d at %v\n", id, elapsed.Round(time.Millisecond))
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)
	fmt.Printf("  Processed %d requests in %v (rate: %d/sec)\n",
		numRequests, elapsed.Round(time.Millisecond), rateLimit)
}
