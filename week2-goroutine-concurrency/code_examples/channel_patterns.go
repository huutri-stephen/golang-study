package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== 1. Unbuffered Channel ===")
	unbufferedDemo()

	fmt.Println("\n=== 2. Buffered Channel ===")
	bufferedDemo()

	fmt.Println("\n=== 3. Channel Direction ===")
	directionDemo()

	fmt.Println("\n=== 4. Range over Channel ===")
	rangeDemo()

	fmt.Println("\n=== 5. Select Statement ===")
	selectDemo()

	fmt.Println("\n=== 6. Done Channel Pattern ===")
	doneChannelPattern()

	fmt.Println("\n=== 7. Fan-out / Fan-in ===")
	fanOutFanIn()

	fmt.Println("\n=== 8. Pipeline Pattern ===")
	pipelinePattern()

	fmt.Println("\n=== 9. Timeout Pattern ===")
	timeoutPattern()

	fmt.Println("\n=== 10. Nil Channel Trick ===")
	nilChannelTrick()
}

func unbufferedDemo() {
	// Unbuffered: synchronous communication
	// Send blocks until someone receives
	ch := make(chan string)

	go func() {
		time.Sleep(100 * time.Millisecond)
		ch <- "hello" // blocks until main receives
		fmt.Println("  Sender: message sent")
	}()

	msg := <-ch // blocks until goroutine sends
	fmt.Printf("  Receiver: got %q\n", msg)
}

func bufferedDemo() {
	// Buffered: async up to capacity
	ch := make(chan int, 3) // buffer size 3

	// Can send 3 without blocking
	ch <- 1
	ch <- 2
	ch <- 3
	// ch <- 4 // this would block!

	fmt.Printf("  Buffer: len=%d, cap=%d\n", len(ch), cap(ch))

	// Receive
	fmt.Printf("  Received: %d, %d, %d\n", <-ch, <-ch, <-ch)
}

func directionDemo() {
	ch := make(chan int, 1)

	// Send-only channel
	var sendOnly chan<- int = ch
	sendOnly <- 42

	// Receive-only channel
	var recvOnly <-chan int = ch
	val := <-recvOnly

	fmt.Printf("  Directional channels: sent and received %d\n", val)
}

// producer is send-only, consumer is receive-only
func produce(ch chan<- int) {
	for i := 0; i < 5; i++ {
		ch <- i
	}
	close(ch)
}

func consume(ch <-chan int) []int {
	var results []int
	for v := range ch {
		results = append(results, v)
	}
	return results
}

func rangeDemo() {
	ch := make(chan int, 5)

	go produce(ch)
	results := consume(ch)

	fmt.Printf("  Range over channel: %v\n", results)
	// Range exits when channel is closed
}

func selectDemo() {
	ch1 := make(chan string, 1)
	ch2 := make(chan string, 1)

	go func() {
		time.Sleep(50 * time.Millisecond)
		ch1 <- "from ch1"
	}()
	go func() {
		time.Sleep(100 * time.Millisecond)
		ch2 <- "from ch2"
	}()

	// Select first available
	for i := 0; i < 2; i++ {
		select {
		case msg := <-ch1:
			fmt.Printf("  Select: %s\n", msg)
		case msg := <-ch2:
			fmt.Printf("  Select: %s\n", msg)
		}
	}

	// Non-blocking select with default
	ch3 := make(chan int)
	select {
	case v := <-ch3:
		fmt.Printf("  Got: %d\n", v)
	default:
		fmt.Println("  Select: no message available (non-blocking)")
	}
}

func doneChannelPattern() {
	// Done channel for cancellation
	done := make(chan struct{}) // empty struct = zero memory
	results := make(chan int)

	// Worker that respects cancellation
	go func() {
		defer close(results)
		for i := 0; ; i++ {
			select {
			case <-done:
				fmt.Println("  Worker: cancelled, cleaning up")
				return
			case results <- i:
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	// Consume some results then cancel
	for i := 0; i < 5; i++ {
		fmt.Printf("  Got: %d\n", <-results)
	}
	close(done) // signal cancellation

	time.Sleep(50 * time.Millisecond) // let worker finish
}

func fanOutFanIn() {
	// Fan-out: one source, multiple workers
	// Fan-in: multiple sources, one destination

	source := make(chan int, 10)

	// Produce work
	go func() {
		for i := 0; i < 10; i++ {
			source <- i
		}
		close(source)
	}()

	// Fan-out to 3 workers
	numWorkers := 3
	workerResults := make([]<-chan int, numWorkers)
	for i := 0; i < numWorkers; i++ {
		workerResults[i] = worker(i, source)
	}

	// Fan-in: merge all worker results
	merged := fanIn(workerResults...)

	// Collect results
	var results []int
	for v := range merged {
		results = append(results, v)
	}
	fmt.Printf("  Fan-out/Fan-in results (10 items, 3 workers): %v\n", results)
}

func worker(id int, input <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for v := range input {
			// simulate work
			time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
			out <- v * v // square the input
		}
	}()
	return out
}

func fanIn(channels ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup

	for _, ch := range channels {
		wg.Add(1)
		go func(c <-chan int) {
			defer wg.Done()
			for v := range c {
				out <- v
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func pipelinePattern() {
	// Pipeline: chain of stages connected by channels
	// Each stage: receive → process → send

	// Stage 1: Generate numbers
	gen := func(nums ...int) <-chan int {
		out := make(chan int)
		go func() {
			defer close(out)
			for _, n := range nums {
				out <- n
			}
		}()
		return out
	}

	// Stage 2: Square numbers
	square := func(in <-chan int) <-chan int {
		out := make(chan int)
		go func() {
			defer close(out)
			for n := range in {
				out <- n * n
			}
		}()
		return out
	}

	// Stage 3: Filter (only even)
	filterEven := func(in <-chan int) <-chan int {
		out := make(chan int)
		go func() {
			defer close(out)
			for n := range in {
				if n%2 == 0 {
					out <- n
				}
			}
		}()
		return out
	}

	// Connect pipeline
	numbers := gen(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	squared := square(numbers)
	evens := filterEven(squared)

	// Consume
	fmt.Print("  Pipeline (gen → square → filterEven): ")
	for v := range evens {
		fmt.Printf("%d ", v)
	}
	fmt.Println()
}

func timeoutPattern() {
	ch := make(chan string)

	go func() {
		time.Sleep(200 * time.Millisecond) // slow response
		ch <- "result"
	}()

	select {
	case result := <-ch:
		fmt.Printf("  Got result: %s\n", result)
	case <-time.After(100 * time.Millisecond):
		fmt.Println("  Timeout! Operation took too long")
	}
}

func nilChannelTrick() {
	// Nil channel in select = disabled case
	// Useful for dynamically enabling/disabling channels

	ch1 := make(chan int, 2)
	ch2 := make(chan int, 2)

	ch1 <- 1
	ch1 <- 2
	ch2 <- 10
	ch2 <- 20

	close(ch1)
	close(ch2)

	var c1, c2 <-chan int = ch1, ch2
	var results []int

	for c1 != nil || c2 != nil {
		select {
		case v, ok := <-c1:
			if !ok {
				c1 = nil // disable this case
				continue
			}
			results = append(results, v)
		case v, ok := <-c2:
			if !ok {
				c2 = nil // disable this case
				continue
			}
			results = append(results, v)
		}
	}

	fmt.Printf("  Nil channel trick (merge until both closed): %v\n", results)
}
