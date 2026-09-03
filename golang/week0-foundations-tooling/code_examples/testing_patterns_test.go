// Testing chiều sâu: table-driven, subtest, mock qua interface, fuzzing.
// Chạy:
//
//	go test -v ./...
//	go test -cover ./...
//	go test -fuzz=FuzzNormalize -fuzztime=15s
package main

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// ---------- Code under test ----------

func Discount(price int, tier string) int {
	switch tier {
	case "gold":
		return price * 80 / 100
	case "silver":
		return price * 90 / 100
	default:
		return price
	}
}

// Normalize: dùng để demo fuzzing (property: chạy 2 lần = 1 lần / idempotent).
func Normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// Interface ở phía consumer -> dễ mock.
type UserRepo interface {
	FindByID(ctx context.Context, id string) (string, error)
}

type UserService struct{ repo UserRepo }

var ErrNotFound = errors.New("not found")

func (s UserService) Greeting(ctx context.Context, id string) (string, error) {
	name, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return "", err
	}
	return "Hello " + name, nil
}

// ---------- Table-driven + subtests ----------

func TestDiscount(t *testing.T) {
	tests := []struct {
		name  string
		price int
		tier  string
		want  int
	}{
		{"gold 20% off", 100, "gold", 80},
		{"silver 10% off", 100, "silver", 90},
		{"unknown tier no discount", 100, "bronze", 100},
		{"empty tier", 50, "", 50},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Discount(tt.price, tt.tier); got != tt.want {
				t.Errorf("Discount(%d,%q) = %d, want %d", tt.price, tt.tier, got, tt.want)
			}
		})
	}
}

// ---------- Mock qua interface ----------

type mockRepo struct {
	name string
	err  error
}

func (m mockRepo) FindByID(context.Context, string) (string, error) {
	return m.name, m.err
}

func TestGreeting(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		svc := UserService{repo: mockRepo{name: "Alice"}}
		got, err := svc.Greeting(context.Background(), "1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err) // Fatalf = dừng ngay như require
		}
		if got != "Hello Alice" {
			t.Errorf("got %q", got)
		}
	})

	t.Run("not found propagates error", func(t *testing.T) {
		svc := UserService{repo: mockRepo{err: ErrNotFound}}
		_, err := svc.Greeting(context.Background(), "x")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

// ---------- Fuzzing ----------

func FuzzNormalize(f *testing.F) {
	f.Add("  Hello ")
	f.Add("ALREADY-lower")
	f.Fuzz(func(t *testing.T, s string) {
		once := Normalize(s)
		twice := Normalize(once)
		if once != twice {
			t.Errorf("Normalize không idempotent: %q -> %q -> %q", s, once, twice)
		}
	})
}
