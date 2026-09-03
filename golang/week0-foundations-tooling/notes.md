# Week 0 – Foundations & Tooling – Study Notes

> Phần "nền tảng ecosystem" mà roadmap.sh/golang nhấn mạnh nhưng plan interview thường bỏ qua:
> modules, testing chuyên sâu, tooling, error handling hiện đại, encoding/json, reflection.
> Đây là những thứ interviewer coi là "hiển nhiên phải biết" ở mức senior.

---

## 1. Go Modules & Dependency Management

### Khái niệm cốt lõi

```
module   = tập hợp package được version cùng nhau, khai báo trong go.mod
package  = thư mục chứa các file .go cùng khai báo `package x`
go.mod   = tên module + Go version + danh sách dependency trực tiếp/gián tiếp
go.sum   = checksum (hash) của từng dependency → đảm bảo tính toàn vẹn (supply-chain)
```

### go.mod ví dụ

```go
module github.com/acme/payment-service

go 1.22

require (
    github.com/jackc/pgx/v5 v5.5.1
    google.golang.org/grpc v1.62.0
)

require (
    github.com/stretchr/testify v1.9.0 // indirect
)

replace github.com/acme/shared => ../shared // local override khi dev
```

### Lệnh thường dùng

| Lệnh | Ý nghĩa |
|---|---|
| `go mod init <module>` | khởi tạo go.mod |
| `go get pkg@v1.2.3` | thêm/nâng dependency lên version cụ thể |
| `go get -u ./...` | nâng minor/patch cho toàn bộ |
| `go mod tidy` | thêm dep thiếu, xóa dep thừa, đồng bộ go.sum |
| `go mod download` | tải về module cache (`$GOPATH/pkg/mod`) |
| `go mod vendor` | copy dependency vào thư mục `vendor/` |
| `go mod why <pkg>` | vì sao dependency này có mặt |
| `go mod graph` | in đồ thị dependency |

### Semantic Versioning + MVS

- Version dạng `vMAJOR.MINOR.PATCH`. **Major ≥ 2 phải đổi import path**: `github.com/x/y/v2`.
- Go dùng **MVS (Minimal Version Selection)**: chọn version **thấp nhất thỏa mãn** tất cả yêu cầu — khác npm (chọn cao nhất). Điều này khiến build **reproducible** và ổn định.
- `go.sum` chứa hash → nếu ai đó sửa nội dung một version đã publish, build fail (chống tampering).

### Go Workspaces (Go 1.18+)

Khi làm nhiều module cùng lúc (monorepo) mà không muốn dùng `replace`:

```bash
go work init ./payment ./shared
go work use ./notification   # thêm module vào workspace
```

`go.work` chỉ dùng local, **không commit** (hoặc commit tùy team) — override go.mod khi build.

### Vendoring — khi nào dùng

- Build offline / CI không có internet.
- Audit dependency trực tiếp trong repo.
- Có `vendor/` thì `go build` tự dùng nó (bỏ qua module cache) trừ khi `-mod=mod`.

---

## 2. Testing (chiều sâu senior)

> Plan gốc chỉ có benchmark + `-race`. Interview senior kiểm tra **cách bạn tổ chức test**,
> không chỉ "biết viết test".

### Table-driven tests (idiom chuẩn Go)

```go
func TestDiscount(t *testing.T) {
    tests := []struct {
        name     string
        price    int
        tier     string
        want     int
    }{
        {"gold tier", 100, "gold", 80},
        {"silver tier", 100, "silver", 90},
        {"no tier", 100, "", 100},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) { // subtest → chạy/lọc riêng
            got := Discount(tt.price, tt.tier)
            if got != tt.want {
                t.Errorf("Discount(%d,%q) = %d, want %d", tt.price, tt.tier, got, tt.want)
            }
        })
    }
}
```

- `t.Run` tạo **subtest**: chạy riêng `go test -run TestDiscount/gold_tier`, báo cáo rõ case nào fail.
- `t.Parallel()` trong subtest → chạy song song (chú ý capture biến `tt` trước Go 1.22).

### Mocking qua interface

Go không có mock "magic" — mock bằng **interface + implementation giả**:

```go
type UserRepo interface {
    FindByID(ctx context.Context, id string) (*User, error)
}

type mockRepo struct {
    user *User
    err  error
}
func (m *mockRepo) FindByID(context.Context, string) (*User, error) {
    return m.user, m.err
}
```

- Thư viện: `stretchr/testify/mock`, `golang/mock` (gomock, sinh code từ interface), `matryer/moq`.
- **Nguyên tắc:** định nghĩa interface **ở nơi tiêu thụ** (consumer), không ở nơi implement → dễ mock, tránh coupling.

### Coverage

```bash
go test -cover ./...                       # % coverage nhanh
go test -coverprofile=cover.out ./...      # xuất profile
go tool cover -html=cover.out              # xem heatmap trên browser
go tool cover -func=cover.out              # coverage theo từng func
```

- Đừng thần thánh hóa 100% coverage — coverage cao ≠ test tốt. Ưu tiên **branch quan trọng + edge case**.

### Fuzzing (Go 1.18+)

Tự sinh input ngẫu nhiên tìm case gây crash/sai:

```go
func FuzzParse(f *testing.F) {
    f.Add("valid-input")          // seed corpus
    f.Fuzz(func(t *testing.T, s string) {
        got, err := Parse(s)
        if err == nil {
            // property: parse rồi serialize lại phải bằng input
            if Serialize(got) != s { t.Errorf("round-trip mismatch") }
        }
    })
}
```

```bash
go test -fuzz=FuzzParse -fuzztime=30s
```

### testify & helpers

```go
assert.Equal(t, want, got)   // ghi lỗi nhưng tiếp tục
require.NoError(t, err)       // ghi lỗi và DỪNG test ngay (dùng cho precondition)
```

- `t.Helper()` trong helper function → stacktrace trỏ về dòng gọi, không phải trong helper.
- `t.Cleanup(fn)` → dọn dẹp sau test (thay defer, chạy cả với subtest).

### Integration test

- Build tag để tách: `//go:build integration` → `go test -tags=integration`.
- `testcontainers-go`: spin Postgres/Redis/Kafka thật trong Docker cho test → sát production.

---

## 3. Tooling & Code Quality

| Tool | Vai trò |
|---|---|
| `gofmt` / `goimports` | format chuẩn + tự sắp xếp import (goimports thêm/bỏ import) |
| `go vet` | phát hiện bug tĩnh: printf sai format, lock copy, unreachable code |
| `staticcheck` | linter mạnh (SA checks): dead code, sai dùng API, simplify |
| `golangci-lint` | **meta-linter** chạy nhiều linter song song (govet, staticcheck, errcheck, ineffassign, gosec...) — chuẩn de-facto trong CI |
| `delve` (`dlv`) | debugger chính thức của Go: breakpoint, step, inspect goroutine |
| `go build -race` | phát hiện data race lúc runtime |
| `pprof` | profiling CPU/heap/goroutine (xem Week 3) |

### golangci-lint config mẫu (`.golangci.yml`)

```yaml
linters:
  enable:
    - errcheck      # lỗi bị bỏ (không check err)
    - gosec         # lỗ hổng bảo mật
    - govet
    - staticcheck
    - ineffassign   # gán biến không dùng
    - revive        # style
```

### delve nhanh

```bash
dlv debug ./cmd/server    # build + debug
# trong dlv:
break main.go:42          # đặt breakpoint
continue                  # chạy tới breakpoint
goroutines                # liệt kê goroutine (rất hữu ích debug deadlock)
print myVar               # in giá trị
```

---

## 4. Error Handling hiện đại (Go 1.13+)

> Plan gốc dừng ở panic/recover. Phần dưới là thứ **hay bị hỏi** và **hay dùng thực tế**.

### Wrapping với `%w`

```go
if err != nil {
    return fmt.Errorf("load user %s: %w", id, err) // %w giữ nguyên err gốc trong chain
}
```

- `%w` → error được **bọc** (unwrappable). `%v` → chỉ nối chuỗi (mất chain).

### errors.Is vs errors.As

```go
var ErrNotFound = errors.New("not found") // sentinel error

// Is: so khớp với một GIÁ TRỊ error cụ thể trong chain
if errors.Is(err, ErrNotFound) { ... }
if errors.Is(err, sql.ErrNoRows) { ... }

// As: trích ra một KIỂU error cụ thể trong chain
var pgErr *pgconn.PgError
if errors.As(err, &pgErr) {
    if pgErr.Code == "23505" { // unique violation
        return ErrDuplicate
    }
}
```

| | So khớp theo | Dùng khi |
|---|---|---|
| `errors.Is` | giá trị (sentinel) | "có phải lỗi X không?" |
| `errors.As` | kiểu (custom struct) | cần đọc field trong error |

### Custom error type

```go
type ValidationError struct {
    Field string
    Msg   string
}
func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Msg)
}
```

### Nguyên tắc senior

- **Xử lý lỗi một lần**: hoặc log, hoặc return — đừng vừa log vừa return (log trùng).
- Wrap kèm **context** (id, thao tác) khi trả lên, để log cuối cùng đủ thông tin trace.
- Sentinel error cho control-flow ổn định giữa các layer; custom type khi cần metadata.
- `panic` chỉ cho lỗi lập trình không thể hồi phục (nil pointer logic), KHÔNG cho lỗi nghiệp vụ.

---

## 5. encoding/json (thực chiến)

> JSON gần như chắc chắn xuất hiện trong coding round backend.

### Struct tags

```go
type User struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    Password  string    `json:"-"`                 // KHÔNG bao giờ serialize
    Nickname  string    `json:"nickname,omitempty"` // bỏ nếu zero value
    CreatedAt time.Time `json:"created_at"`
    Internal  int       `json:"-"`
}
```

### Marshal / Unmarshal

```go
b, err := json.Marshal(user)          // struct → []byte
err = json.Unmarshal(b, &user)        // []byte → struct

// Streaming (file/HTTP body lớn) — không load hết vào RAM:
dec := json.NewDecoder(r.Body)
dec.DisallowUnknownFields()           // reject field lạ (strict API)
err = dec.Decode(&user)
```

### Bẫy thường gặp

- Field **không exported (chữ thường)** → json bỏ qua. Muốn serialize phải viết hoa.
- `omitempty` **không** áp dụng cho struct lồng non-pointer (struct rỗng vẫn xuất). Dùng con trỏ hoặc `*T`.
- Số trong JSON unmarshal vào `interface{}` → luôn thành `float64` (mất chính xác với int64 lớn). Dùng `json.Number` hoặc `dec.UseNumber()`.
- `time.Time` mặc định theo RFC3339.

### Custom marshaling

```go
func (u User) MarshalJSON() ([]byte, error) {
    type alias User // tránh đệ quy vô hạn
    return json.Marshal(struct {
        alias
        CreatedAt string `json:"created_at"`
    }{alias(u), u.CreatedAt.Format("2006-01-02")})
}
```

---

## 6. Reflection (reflect) – biết đủ để giải thích

- `reflect.TypeOf` / `reflect.ValueOf` → soi kiểu/giá trị lúc runtime.
- Nền tảng của `encoding/json`, ORM (GORM), validator (đọc struct tag).
- **Đắt** (chậm hơn code tĩnh nhiều lần) và **mất type-safety** → chỉ dùng khi thật cần (serialization, framework), tránh trong hot path.
- Câu hay hỏi: *"json.Marshal hoạt động thế nào?"* → dùng reflection duyệt field + đọc struct tag để quyết định tên key và có serialize hay không.

---

## 7. Bổ sung nhẹ: Config & Logging (ecosystem)

### Config

- Thứ tự ưu tiên phổ biến: **flag > env > file > default**.
- `flag` (stdlib), `os.Getenv`, `viper` (file+env+watch), `envconfig` (map env vào struct).
- 12-factor: config qua **env var**, không hardcode, không commit secret.

### Logging

- `log/slog` (stdlib từ Go 1.21): **structured logging** chính thức — nên dùng cho project mới.
- `zap` (Uber, nhanh nhất), `zerolog` (zero-alloc), `logrus` (cũ, phổ biến, chậm hơn).
- Structured (key-value/JSON) > plain text → dễ query trên ELK/Loki.

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
logger.Info("payment processed", "user_id", uid, "amount", amt, "trace_id", tid)
```
