package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
)

// This file demonstrates database transaction patterns in Go.
// Requires: go get github.com/lib/pq (PostgreSQL driver)
// For demonstration purposes, using interface to show patterns.

// --- Repository Pattern with Transactions ---

type DB interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type Order struct {
	ID     int64
	UserID int64
	Amount float64
	Status string
}

type OrderRepository struct {
	db DB
}

// --- Pattern 1: Simple Transaction ---

func (r *OrderRepository) CreateOrder(ctx context.Context, order *Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	// CRITICAL: always handle rollback
	defer tx.Rollback() // no-op if committed

	// Insert order
	err = tx.QueryRowContext(ctx,
		`INSERT INTO orders (user_id, amount, status) VALUES ($1, $2, $3) RETURNING id`,
		order.UserID, order.Amount, "pending",
	).Scan(&order.ID)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	// Deduct balance
	result, err := tx.ExecContext(ctx,
		`UPDATE wallets SET balance = balance - $1 WHERE user_id = $2 AND balance >= $1`,
		order.Amount, order.UserID,
	)
	if err != nil {
		return fmt.Errorf("deduct balance: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("insufficient balance")
	}

	// Commit
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

// --- Pattern 2: Transaction with Function ---

// WithTransaction executes fn within a transaction.
// Rolls back on error, commits on success.
func WithTransaction(ctx context.Context, db *sql.DB, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // re-throw
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback failed: %v (original: %w)", rbErr, err)
		}
		return err
	}

	return tx.Commit()
}

// Usage:
// err := WithTransaction(ctx, db, func(tx *sql.Tx) error {
//     _, err := tx.ExecContext(ctx, "INSERT INTO ...", args...)
//     if err != nil { return err }
//     _, err = tx.ExecContext(ctx, "UPDATE ...", args...)
//     return err
// })

// --- Pattern 3: SELECT ... FOR UPDATE (Pessimistic Locking) ---

func (r *OrderRepository) TransferMoney(ctx context.Context, fromID, toID int64, amount float64) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead, // higher isolation for transfers
	})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Lock rows in consistent order (prevent deadlock)
	// Always lock lower ID first
	var firstID, secondID int64
	if fromID < toID {
		firstID, secondID = fromID, toID
	} else {
		firstID, secondID = toID, fromID
	}

	// Lock both accounts
	var balance1, balance2 float64
	err = tx.QueryRowContext(ctx,
		"SELECT balance FROM accounts WHERE id = $1 FOR UPDATE",
		firstID,
	).Scan(&balance1)
	if err != nil {
		return fmt.Errorf("lock account %d: %w", firstID, err)
	}

	err = tx.QueryRowContext(ctx,
		"SELECT balance FROM accounts WHERE id = $1 FOR UPDATE",
		secondID,
	).Scan(&balance2)
	if err != nil {
		return fmt.Errorf("lock account %d: %w", secondID, err)
	}

	// Determine which is sender
	var senderBalance float64
	if fromID == firstID {
		senderBalance = balance1
	} else {
		senderBalance = balance2
	}

	if senderBalance < amount {
		return errors.New("insufficient funds")
	}

	// Execute transfer
	_, err = tx.ExecContext(ctx,
		"UPDATE accounts SET balance = balance - $1 WHERE id = $2",
		amount, fromID,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		"UPDATE accounts SET balance = balance + $1 WHERE id = $2",
		amount, toID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// --- Pattern 4: Optimistic Locking with Version ---

type Product struct {
	ID      int64
	Name    string
	Stock   int
	Version int // optimistic lock version
}

var ErrConflict = errors.New("optimistic lock conflict")

func (r *OrderRepository) DeductStock(ctx context.Context, productID int64, quantity int) error {
	// Read current state
	var product Product
	err := r.db.QueryRowContext(ctx,
		"SELECT id, stock, version FROM products WHERE id = $1",
		productID,
	).Scan(&product.ID, &product.Stock, &product.Version)
	if err != nil {
		return err
	}

	if product.Stock < quantity {
		return errors.New("insufficient stock")
	}

	// Update with version check (optimistic lock)
	tx, _ := r.db.BeginTx(ctx, nil)
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx,
		`UPDATE products 
		 SET stock = stock - $1, version = version + 1 
		 WHERE id = $2 AND version = $3`,
		quantity, productID, product.Version,
	)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		// Another transaction modified it → retry
		return ErrConflict
	}

	return tx.Commit()
}

// Retry with optimistic locking
func DeductStockWithRetry(ctx context.Context, repo *OrderRepository, productID int64, qty int) error {
	const maxRetries = 3
	for i := 0; i < maxRetries; i++ {
		err := repo.DeductStock(ctx, productID, qty)
		if err == nil {
			return nil
		}
		if !errors.Is(err, ErrConflict) {
			return err // non-retryable error
		}
		// Exponential backoff
		time.Sleep(time.Duration(i*10) * time.Millisecond)
	}
	return errors.New("max retries exceeded")
}

// --- Pattern 5: Connection Pool Configuration ---

func configurePool(db *sql.DB) {
	// Max open connections (active + idle)
	db.SetMaxOpenConns(25)

	// Max idle connections (kept ready for reuse)
	db.SetMaxIdleConns(10)

	// Max lifetime (recycle connections)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Max idle time (close stale idle connections)
	db.SetConnMaxIdleTime(1 * time.Minute)
}

func main() {
	log.Println("Database transaction patterns demo")
	fmt.Println(`
Patterns demonstrated:
1. Simple transaction (begin → operations → commit/rollback)
2. WithTransaction helper (function-based)
3. SELECT FOR UPDATE (pessimistic locking)
4. Optimistic locking (version column)
5. Connection pool configuration

Key Rules:
• Always defer tx.Rollback() (no-op after commit)
• Use context for timeout/cancellation
• Lock in consistent order (prevent deadlock)
• Keep transactions short
• Handle ErrNoRows explicitly
• Use appropriate isolation level
• Monitor connection pool stats
`)
}
