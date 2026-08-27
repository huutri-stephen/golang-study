// encoding/json thực chiến: struct tags, omitempty, custom marshal, streaming, số lớn.
// Chạy: go run json_patterns.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type User struct {
	ID       string    `json:"id"`
	Email    string    `json:"email"`
	Password string    `json:"-"`                  // KHÔNG bao giờ lộ ra JSON
	Nickname string    `json:"nickname,omitempty"` // bỏ khi rỗng
	Created  time.Time `json:"created_at"`
}

// Custom marshaling: format created_at thành date-only.
func (u User) MarshalJSON() ([]byte, error) {
	type alias User // alias để tránh gọi lại MarshalJSON -> đệ quy vô hạn
	return json.Marshal(struct {
		alias
		Created string `json:"created_at"`
	}{alias(u), u.Created.Format("2006-01-02")})
}

func main() {
	u := User{
		ID:       "u1",
		Email:    "a@b.com",
		Password: "secret", // sẽ bị bỏ
		Created:  time.Date(2026, 8, 25, 10, 0, 0, 0, time.UTC),
	}

	// Marshal: password bị loại, nickname rỗng bị omit, created_at date-only
	b, _ := json.MarshalIndent(u, "", "  ")
	fmt.Println("marshal:\n" + string(b))

	// Streaming decode + strict mode (reject field lạ)
	raw := `{"id":"u2","email":"c@d.com","role":"admin"}`
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.DisallowUnknownFields()
	var u2 User
	if err := dec.Decode(&u2); err != nil {
		fmt.Println("strict decode error (đúng như mong đợi):", err)
	}

	// Bẫy số lớn: unmarshal vào interface{} -> float64 mất chính xác
	var anyVal map[string]interface{}
	json.Unmarshal([]byte(`{"n": 9007199254740993}`), &anyVal)
	fmt.Printf("as interface{}: %v (float64 -> mất chính xác)\n", anyVal["n"])

	// Fix: dùng UseNumber
	dec2 := json.NewDecoder(bytes.NewBufferString(`{"n": 9007199254740993}`))
	dec2.UseNumber()
	var exact map[string]json.Number
	dec2.Decode(&exact)
	fmt.Println("with UseNumber:", exact["n"])
}
