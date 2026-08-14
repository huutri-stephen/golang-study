package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"
)

// --- Saga Orchestrator Pattern ---
// Coordinates distributed transaction across multiple services
// Each step has an action and a compensating action (rollback)

type SagaStep struct {
	Name       string
	Action     func(ctx context.Context, data *SagaData) error
	Compensate func(ctx context.Context, data *SagaData) error
}

type SagaData struct {
	OrderID   string
	UserID    string
	Amount    float64
	PaymentID string
	StockID   string
	ShipID    string
	Error     error
}

type SagaOrchestrator struct {
	steps []SagaStep
}

func NewSagaOrchestrator() *SagaOrchestrator {
	return &SagaOrchestrator{}
}

func (s *SagaOrchestrator) AddStep(step SagaStep) {
	s.steps = append(s.steps, step)
}

func (s *SagaOrchestrator) Execute(ctx context.Context, data *SagaData) error {
	completedSteps := make([]int, 0, len(s.steps))

	for i, step := range s.steps {
		log.Printf("[SAGA] Executing step %d: %s", i+1, step.Name)

		if err := step.Action(ctx, data); err != nil {
			log.Printf("[SAGA] Step %d failed: %v", i+1, err)
			data.Error = err

			// Compensate in reverse order
			s.compensate(ctx, data, completedSteps)
			return fmt.Errorf("saga failed at step '%s': %w", step.Name, err)
		}

		completedSteps = append(completedSteps, i)
		log.Printf("[SAGA] Step %d completed: %s", i+1, step.Name)
	}

	log.Println("[SAGA] All steps completed successfully")
	return nil
}

func (s *SagaOrchestrator) compensate(ctx context.Context, data *SagaData, completedSteps []int) {
	log.Println("[SAGA] Starting compensation...")

	// Compensate in reverse order
	for i := len(completedSteps) - 1; i >= 0; i-- {
		stepIdx := completedSteps[i]
		step := s.steps[stepIdx]

		if step.Compensate == nil {
			continue
		}

		log.Printf("[SAGA] Compensating step %d: %s", stepIdx+1, step.Name)

		if err := step.Compensate(ctx, data); err != nil {
			// Compensation failed — needs manual intervention
			log.Printf("[SAGA] CRITICAL: Compensation failed for step '%s': %v", step.Name, err)
			// In production: alert, store for manual retry
		}
	}

	log.Println("[SAGA] Compensation completed")
}

// --- Service Implementations (Simulated) ---

// Order Service
func createOrder(ctx context.Context, data *SagaData) error {
	// Simulate creating order
	data.OrderID = fmt.Sprintf("order_%d", time.Now().UnixNano())
	fmt.Printf("  → Created order: %s\n", data.OrderID)
	return nil
}

func cancelOrder(ctx context.Context, data *SagaData) error {
	fmt.Printf("  ← Cancelled order: %s\n", data.OrderID)
	return nil
}

// Payment Service
func chargePayment(ctx context.Context, data *SagaData) error {
	// Simulate payment
	if data.Amount > 10000 {
		return errors.New("payment declined: amount exceeds limit")
	}
	data.PaymentID = fmt.Sprintf("pay_%d", time.Now().UnixNano())
	fmt.Printf("  → Charged payment: %s ($%.2f)\n", data.PaymentID, data.Amount)
	return nil
}

func refundPayment(ctx context.Context, data *SagaData) error {
	fmt.Printf("  ← Refunded payment: %s ($%.2f)\n", data.PaymentID, data.Amount)
	return nil
}

// Inventory Service
func reserveStock(ctx context.Context, data *SagaData) error {
	// Simulate stock check — this one will fail in our demo
	if data.Amount > 5000 {
		return errors.New("insufficient stock")
	}
	data.StockID = fmt.Sprintf("stock_%d", time.Now().UnixNano())
	fmt.Printf("  → Reserved stock: %s\n", data.StockID)
	return nil
}

func releaseStock(ctx context.Context, data *SagaData) error {
	fmt.Printf("  ← Released stock: %s\n", data.StockID)
	return nil
}

// Shipping Service
func createShipment(ctx context.Context, data *SagaData) error {
	data.ShipID = fmt.Sprintf("ship_%d", time.Now().UnixNano())
	fmt.Printf("  → Created shipment: %s\n", data.ShipID)
	return nil
}

func cancelShipment(ctx context.Context, data *SagaData) error {
	fmt.Printf("  ← Cancelled shipment: %s\n", data.ShipID)
	return nil
}

// --- Demo ---

func main() {
	fmt.Println("=== Saga Orchestrator Pattern ===\n")

	// Define saga steps
	saga := NewSagaOrchestrator()
	saga.AddStep(SagaStep{
		Name:       "Create Order",
		Action:     createOrder,
		Compensate: cancelOrder,
	})
	saga.AddStep(SagaStep{
		Name:       "Charge Payment",
		Action:     chargePayment,
		Compensate: refundPayment,
	})
	saga.AddStep(SagaStep{
		Name:       "Reserve Stock",
		Action:     reserveStock,
		Compensate: releaseStock,
	})
	saga.AddStep(SagaStep{
		Name:       "Create Shipment",
		Action:     createShipment,
		Compensate: cancelShipment,
	})

	ctx := context.Background()

	// --- Scenario 1: Success ---
	fmt.Println("--- Scenario 1: Happy Path ($100 order) ---")
	data := &SagaData{UserID: "user-1", Amount: 100}
	err := saga.Execute(ctx, data)
	if err != nil {
		fmt.Printf("  FAILED: %v\n", err)
	} else {
		fmt.Println("  SUCCESS!")
	}

	fmt.Println("\n--- Scenario 2: Stock Failure ($7000 order) ---")
	// Payment succeeds but stock fails → compensate payment + order
	data2 := &SagaData{UserID: "user-2", Amount: 7000}
	err = saga.Execute(ctx, data2)
	if err != nil {
		fmt.Printf("  FAILED: %v\n", err)
	}

	fmt.Println("\n--- Scenario 3: Payment Failure ($15000 order) ---")
	// Payment fails → compensate order only
	data3 := &SagaData{UserID: "user-3", Amount: 15000}
	err = saga.Execute(ctx, data3)
	if err != nil {
		fmt.Printf("  FAILED: %v\n", err)
	}

	fmt.Println(`
Saga Pattern Key Points:

Orchestration vs Choreography:
┌─────────────────────────────────────────────────────────┐
│ Orchestration (shown above):                            │
│ • Central coordinator manages flow                      │
│ • Easy to understand and debug                          │
│ • Coordinator is single point of failure                │
│                                                         │
│ Choreography (event-driven):                            │
│ • Services emit/listen to events                        │
│ • No single point of failure                            │
│ • Harder to trace flow                                  │
└─────────────────────────────────────────────────────────┘

Rules:
• Each step MUST have a compensating action
• Compensations MUST be idempotent
• Compensations execute in reverse order
• If compensation fails → need manual intervention / alert
• Store saga state for recovery after crash
• Use saga ID for tracing across services
`)
}
