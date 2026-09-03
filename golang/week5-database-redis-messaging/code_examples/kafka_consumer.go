package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"
)

// Kafka consumer patterns for interview discussion.
// In production: use github.com/segmentio/kafka-go or github.com/confluentinc/confluent-kafka-go

// --- Domain Types ---

type OrderEvent struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
	Timestamp int64   `json:"timestamp"`
}

type Message struct {
	Key       string
	Value     []byte
	Topic     string
	Partition int
	Offset    int64
}

// --- Consumer Interface ---

type Consumer interface {
	ReadMessage(ctx context.Context) (*Message, error)
	CommitMessage(ctx context.Context, msg *Message) error
	Close() error
}

type Producer interface {
	SendMessage(ctx context.Context, topic string, key string, value []byte) error
}

// --- Pattern 1: At-Least-Once Consumer ---

type OrderConsumer struct {
	consumer Consumer
	handler  OrderHandler
}

type OrderHandler interface {
	Process(ctx context.Context, event *OrderEvent) error
}

func (c *OrderConsumer) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Read message
		msg, err := c.consumer.ReadMessage(ctx)
		if err != nil {
			log.Printf("Read error: %v", err)
			continue
		}

		// Parse event
		var event OrderEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Parse error: %v (offset=%d)", err, msg.Offset)
			// Commit to skip bad message (or send to DLQ)
			c.consumer.CommitMessage(ctx, msg)
			continue
		}

		// Process
		if err := c.handler.Process(ctx, &event); err != nil {
			log.Printf("Process error: %v (event=%s)", err, event.ID)
			// DON'T commit → will re-read on restart (at-least-once)
			// Or: retry logic below
			continue
		}

		// Commit AFTER successful processing (at-least-once guarantee)
		if err := c.consumer.CommitMessage(ctx, msg); err != nil {
			log.Printf("Commit error: %v", err)
		}
	}
}

// --- Pattern 2: Retry with DLQ ---

type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

type RetryConsumer struct {
	consumer    Consumer
	producer    Producer
	handler     OrderHandler
	dlqTopic    string
	retryConfig RetryConfig
}

func (c *RetryConsumer) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		msg, err := c.consumer.ReadMessage(ctx)
		if err != nil {
			continue
		}

		var event OrderEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			c.sendToDLQ(ctx, msg, "parse_error", err)
			c.consumer.CommitMessage(ctx, msg)
			continue
		}

		// Retry with exponential backoff
		if err := c.processWithRetry(ctx, &event); err != nil {
			// All retries failed → send to DLQ
			c.sendToDLQ(ctx, msg, "process_failed", err)
		}

		c.consumer.CommitMessage(ctx, msg)
	}
}

func (c *RetryConsumer) processWithRetry(ctx context.Context, event *OrderEvent) error {
	var lastErr error
	delay := c.retryConfig.BaseDelay

	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("Retry %d/%d for event %s", attempt, c.retryConfig.MaxRetries, event.ID)
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
			// Exponential backoff
			delay *= 2
			if delay > c.retryConfig.MaxDelay {
				delay = c.retryConfig.MaxDelay
			}
		}

		lastErr = c.handler.Process(ctx, event)
		if lastErr == nil {
			return nil
		}

		// Check if error is retryable
		if !isRetryable(lastErr) {
			return lastErr
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (c *RetryConsumer) sendToDLQ(ctx context.Context, msg *Message, reason string, err error) {
	dlqMessage := map[string]interface{}{
		"original_topic":     msg.Topic,
		"original_partition": msg.Partition,
		"original_offset":    msg.Offset,
		"original_key":       msg.Key,
		"original_value":     string(msg.Value),
		"error_reason":       reason,
		"error_message":      err.Error(),
		"timestamp":          time.Now().Unix(),
	}

	data, _ := json.Marshal(dlqMessage)
	if err := c.producer.SendMessage(ctx, c.dlqTopic, msg.Key, data); err != nil {
		log.Printf("Failed to send to DLQ: %v", err)
	}
}

func isRetryable(err error) bool {
	// Non-retryable errors
	var validationErr *ValidationError
	if errors.As(err, &validationErr) {
		return false
	}
	// Network errors, timeouts → retryable
	return true
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation: %s - %s", e.Field, e.Message)
}

// --- Pattern 3: Idempotent Consumer ---

type IdempotentHandler struct {
	store   ProcessedStore
	handler OrderHandler
}

type ProcessedStore interface {
	IsProcessed(ctx context.Context, messageID string) (bool, error)
	MarkProcessed(ctx context.Context, messageID string) error
}

func (h *IdempotentHandler) Process(ctx context.Context, event *OrderEvent) error {
	// Check if already processed
	processed, err := h.store.IsProcessed(ctx, event.ID)
	if err != nil {
		return fmt.Errorf("check processed: %w", err)
	}
	if processed {
		log.Printf("Event %s already processed, skipping", event.ID)
		return nil // idempotent — skip duplicate
	}

	// Process
	if err := h.handler.Process(ctx, event); err != nil {
		return err
	}

	// Mark as processed
	if err := h.store.MarkProcessed(ctx, event.ID); err != nil {
		// Non-critical: next time will process again (at-least-once)
		log.Printf("Failed to mark processed: %v", err)
	}

	return nil
}

// --- Pattern 4: Batch Consumer ---

type BatchConsumer struct {
	consumer  Consumer
	handler   BatchHandler
	batchSize int
	timeout   time.Duration
}

type BatchHandler interface {
	ProcessBatch(ctx context.Context, events []*OrderEvent) error
}

func (c *BatchConsumer) Start(ctx context.Context) error {
	batch := make([]*Message, 0, c.batchSize)
	timer := time.NewTimer(c.timeout)

	for {
		select {
		case <-ctx.Done():
			// Process remaining batch before exit
			if len(batch) > 0 {
				c.processBatch(ctx, batch)
			}
			return ctx.Err()

		case <-timer.C:
			// Timeout: process whatever we have
			if len(batch) > 0 {
				c.processBatch(ctx, batch)
				batch = batch[:0]
			}
			timer.Reset(c.timeout)

		default:
			msg, err := c.consumer.ReadMessage(ctx)
			if err != nil {
				continue
			}
			batch = append(batch, msg)

			// Batch full
			if len(batch) >= c.batchSize {
				c.processBatch(ctx, batch)
				batch = batch[:0]
				timer.Reset(c.timeout)
			}
		}
	}
}

func (c *BatchConsumer) processBatch(ctx context.Context, msgs []*Message) {
	events := make([]*OrderEvent, 0, len(msgs))
	for _, msg := range msgs {
		var event OrderEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			continue
		}
		events = append(events, &event)
	}

	if err := c.handler.ProcessBatch(ctx, events); err != nil {
		log.Printf("Batch processing failed: %v", err)
		return
	}

	// Commit last message in batch
	if len(msgs) > 0 {
		c.consumer.CommitMessage(ctx, msgs[len(msgs)-1])
	}
}

func main() {
	log.Println("Kafka consumer patterns demo")
	fmt.Println(`
Patterns demonstrated:
1. At-Least-Once Consumer
   - Process first, commit after
   - May reprocess on failure (need idempotent handler)

2. Retry with DLQ
   - Exponential backoff retry
   - Non-retryable → immediate DLQ
   - Max retries exceeded → DLQ
   - DLQ includes original message + error info

3. Idempotent Consumer
   - Check processed store before processing
   - Skip duplicates gracefully
   - Works with at-least-once delivery

4. Batch Consumer
   - Accumulate messages until batch size or timeout
   - Better throughput for I/O-bound processing
   - Commit after batch success

Key Interview Points:
• Ordering: guaranteed only within same partition
• Partition key: determines partition → ordering scope
• Consumer group: parallel consumers, each owns partitions
• Rebalancing: triggered by consumer join/leave
• Offset management: auto vs manual commit
• Backpressure: buffered channel + worker pool
`)
}
