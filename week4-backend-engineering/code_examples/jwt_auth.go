// JWT (HS256) tự implement bằng stdlib để hiểu bản chất + HTTP auth middleware.
// Thực tế nên dùng github.com/golang-jwt/jwt, nhưng đây là để thấy rõ cơ chế ký/verify.
// Chạy: go run jwt_auth.go
package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

var secret = []byte("super-secret-key-from-env")

type Claims struct {
	Sub   string   `json:"sub"`
	Roles []string `json:"roles"`
	Exp   int64    `json:"exp"`
}

func b64(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

// SignJWT tạo token dạng header.payload.signature
func SignJWT(c Claims) string {
	header, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	payload, _ := json.Marshal(c)
	signingInput := b64(header) + "." + b64(payload)

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(signingInput))
	sig := mac.Sum(nil)
	return signingInput + "." + b64(sig)
}

// VerifyJWT kiểm tra chữ ký (constant-time) và hạn dùng.
func VerifyJWT(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("malformed token")
	}
	signingInput := parts[0] + "." + parts[1]

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(signingInput))
	expectedSig := mac.Sum(nil)

	gotSig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, errors.New("bad signature encoding")
	}
	// So sánh constant-time chống timing attack
	if !hmac.Equal(expectedSig, gotSig) {
		return nil, errors.New("signature mismatch")
	}

	payload, _ := base64.RawURLEncoding.DecodeString(parts[1])
	var c Claims
	if err := json.Unmarshal(payload, &c); err != nil {
		return nil, err
	}
	if time.Now().Unix() > c.Exp {
		return nil, errors.New("token expired")
	}
	return &c, nil
}

// --- HTTP auth middleware ---

type ctxKey string

const claimsKey ctxKey = "claims"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "missing bearer token", http.StatusUnauthorized) // 401
			return
		}
		claims, err := VerifyJWT(strings.TrimPrefix(auth, "Bearer "))
		if err != nil {
			http.Error(w, "invalid token: "+err.Error(), http.StatusUnauthorized) // 401
			return
		}
		ctx := context.WithValue(r.Context(), claimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ClaimsFromContext lấy claims đã xác thực ra khỏi context trong handler.
func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(claimsKey).(*Claims)
	return c, ok
}

func main() {
	// 1. Ký token hợp lệ (còn hạn)
	good := SignJWT(Claims{Sub: "u42", Roles: []string{"admin"}, Exp: time.Now().Add(15 * time.Minute).Unix()})
	fmt.Println("token:", good[:40], "...")

	if c, err := VerifyJWT(good); err == nil {
		fmt.Printf("verify OK -> sub=%s roles=%v\n", c.Sub, c.Roles)
	}

	// 2. Token bị sửa payload -> chữ ký sai
	tampered := good[:len(good)-4] + "AAAA"
	if _, err := VerifyJWT(tampered); err != nil {
		fmt.Println("tampered ->", err)
	}

	// 3. Token hết hạn
	expired := SignJWT(Claims{Sub: "u1", Exp: time.Now().Add(-time.Minute).Unix()})
	if _, err := VerifyJWT(expired); err != nil {
		fmt.Println("expired ->", err)
	}

	// 4. Thử middleware qua httptest
	protected := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("secret data"))
	}))
	// request không token -> 401
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	protected.ServeHTTP(rec, req)
	fmt.Println("no-token status:", rec.Code)
	// request có token -> 200
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("Authorization", "Bearer "+good)
	rec2 := httptest.NewRecorder()
	protected.ServeHTTP(rec2, req2)
	fmt.Println("with-token status:", rec2.Code, "body:", rec2.Body.String())
}
