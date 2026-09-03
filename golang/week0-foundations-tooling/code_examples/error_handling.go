// Error handling hiện đại (Go 1.13+): wrapping, errors.Is, errors.As, custom type.
// Chạy: go run error_handling.go
package main

import (
	"errors"
	"fmt"
)

// --- Sentinel error: so khớp bằng errors.Is ---
var ErrNotFound = errors.New("record not found")

// --- Custom error type: trích field bằng errors.As ---
type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed on %q: %s", e.Field, e.Msg)
}

// repo layer trả sentinel error
func findUser(id string) error {
	if id == "" {
		return &ValidationError{Field: "id", Msg: "must not be empty"}
	}
	if id != "42" {
		return ErrNotFound
	}
	return nil
}

// service layer WRAP thêm context bằng %w (giữ nguyên chain)
func loadUser(id string) error {
	if err := findUser(id); err != nil {
		return fmt.Errorf("loadUser(%s): %w", id, err)
	}
	return nil
}

func main() {
	// 1. errors.Is — so khớp theo GIÁ TRỊ sentinel, kể cả khi đã bị wrap
	err := loadUser("999")
	fmt.Println("err:", err)
	fmt.Println("Is ErrNotFound?", errors.Is(err, ErrNotFound)) // true dù đã wrap

	// 2. errors.As — trích theo KIỂU để đọc field bên trong
	err = loadUser("")
	var ve *ValidationError
	if errors.As(err, &ve) {
		fmt.Printf("As ValidationError -> field=%s msg=%s\n", ve.Field, ve.Msg)
	}

	// 3. Unwrap thủ công (hiếm khi cần, thường dùng Is/As)
	fmt.Println("Unwrap:", errors.Unwrap(err))

	// 4. errors.Join (Go 1.20+): gộp nhiều lỗi
	multi := errors.Join(ErrNotFound, &ValidationError{Field: "email", Msg: "invalid"})
	fmt.Println("Joined Is ErrNotFound?", errors.Is(multi, ErrNotFound))
}
