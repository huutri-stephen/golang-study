// Patterns truy cập DB bằng database/sql (stdlib): pool config, scan, rows,
// transaction an toàn với defer Rollback, chống SQL injection.
// Thực tế cần driver: _ "github.com/jackc/pgx/v5/stdlib" (hoặc lib/pq, mysql).
// Chạy: go run db_access_patterns.go  (không cần DB thật — minh hoạ cấu trúc)
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type User struct {
	ID    int64
	Email string
}

var ErrNotFound = errors.New("user not found")

// configurePool: sql.DB là POOL, không phải 1 connection.
func configurePool(db *sql.DB) {
	db.SetMaxOpenConns(25)                 // tổng connection đồng thời tối đa
	db.SetMaxIdleConns(10)                 // giữ sẵn để tái dùng
	db.SetConnMaxLifetime(5 * time.Minute) // tái tạo connection cũ (tránh bị LB/DB đóng ngầm)
	db.SetConnMaxIdleTime(1 * time.Minute)
}

// getUser: QueryRow + Scan, map sql.ErrNoRows -> lỗi domain bằng errors.Is.
func getUser(ctx context.Context, db *sql.DB, id int64) (*User, error) {
	var u User
	// Placeholder $1 -> driver escape -> chống SQL injection. KHÔNG nối chuỗi.
	err := db.QueryRowContext(ctx, `SELECT id, email FROM users WHERE id = $1`, id).
		Scan(&u.ID, &u.Email)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, ErrNotFound
	case err != nil:
		return nil, fmt.Errorf("getUser(%d): %w", id, err)
	}
	return &u, nil
}

// listActive: query nhiều dòng — phải defer rows.Close() và kiểm tra rows.Err().
func listActive(ctx context.Context, db *sql.DB) ([]User, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, email FROM users WHERE active = $1`, true)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // BẮT BUỘC: trả connection về pool

	var out []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err() // lỗi phát sinh giữa chừng (vd mất kết nối)
}

// transfer: transaction an toàn — defer Rollback ngay sau BeginTx.
func transfer(ctx context.Context, db *sql.DB, from, to int64, amount int64) (err error) {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	// Commit OK -> Rollback là no-op. Return sớm vì lỗi -> tự rollback. Không bao giờ leak tx.
	defer func() { _ = tx.Rollback() }()

	if _, err = tx.ExecContext(ctx,
		`UPDATE accounts SET balance = balance - $1 WHERE id = $2`, amount, from); err != nil {
		return fmt.Errorf("debit: %w", err)
	}
	if _, err = tx.ExecContext(ctx,
		`UPDATE accounts SET balance = balance + $1 WHERE id = $2`, amount, to); err != nil {
		return fmt.Errorf("credit: %w", err)
	}
	return tx.Commit()
}

func main() {
	// sql.Open KHÔNG kết nối ngay — chỉ tạo pool (lazy). Cần Ping để kiểm tra thật.
	db, err := sql.Open("pgx", "postgres://user:pass@localhost:5432/app")
	if err != nil {
		// Không có driver "pgx" đăng ký trong ví dụ standalone -> minh hoạ thông báo.
		fmt.Println("sql.Open error (thiếu driver — đúng như mong đợi trong demo):", err)
		fmt.Println("Các pattern quan trọng đã minh hoạ trong file:")
		fmt.Println("  - configurePool: sql.DB là POOL, set MaxOpen/MaxIdle/ConnMaxLifetime")
		fmt.Println("  - getUser: QueryRow+Scan, map sql.ErrNoRows bằng errors.Is")
		fmt.Println("  - listActive: defer rows.Close() + rows.Err()")
		fmt.Println("  - transfer: defer tx.Rollback() ngay sau BeginTx -> không leak tx")
		fmt.Println("  - placeholder $1 chống SQL injection, KHÔNG nối chuỗi")
		return
	}
	defer db.Close()
	configurePool(db)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		fmt.Println("ping error:", err)
	}
}
