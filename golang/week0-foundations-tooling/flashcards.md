# Week 0 – Flashcards / Q&A

## Modules & Dependency Management

### Q: go.mod và go.sum khác nhau gì?
**A:** `go.mod` khai báo module, Go version và danh sách dependency (trực tiếp + indirect). `go.sum` chứa checksum (hash) của từng version dependency để đảm bảo tính toàn vẹn — chống ai đó sửa nội dung một version đã publish (supply-chain security).

---

### Q: `go mod tidy` làm gì?
**A:** Thêm dependency bị thiếu trong go.mod, xóa dependency không còn dùng, và đồng bộ go.sum. Nên chạy trước khi commit.

---

### Q: Go dùng thuật toán nào để chọn version dependency? Khác npm thế nào?
**A:** MVS (Minimal Version Selection) — chọn version **thấp nhất** thỏa mãn tất cả yêu cầu. npm chọn cao nhất. MVS làm build reproducible, ổn định, tránh "đột nhiên vỡ vì minor version mới".

---

### Q: Vì sao major version ≥ 2 phải đổi import path?
**A:** Semantic Import Versioning: `github.com/x/y/v2`. Major version = breaking change, đổi path cho phép v1 và v2 cùng tồn tại trong một build (import compatibility rule).

---

### Q: Khi nào dùng go workspace (`go.work`) thay vì `replace`?
**A:** Khi dev nhiều module cùng lúc (monorepo). `go work use ./m1 ./m2` override go.mod local mà không phải sửa `replace` trong từng go.mod rồi lỡ commit nhầm. go.work thường không commit.

---

## Testing

### Q: Table-driven test là gì và vì sao là idiom chuẩn?
**A:** Định nghĩa slice các case `{name, input, want}` rồi loop chạy. Thêm case = thêm 1 dòng. Kết hợp `t.Run(tt.name, ...)` tạo subtest → chạy/lọc/báo cáo từng case riêng biệt.

---

### Q: Mock trong Go làm thế nào? (không có mock magic như Java)
**A:** Định nghĩa **interface** ở nơi tiêu thụ (consumer), rồi viết implementation giả cho test. Có thể dùng gomock (sinh code), testify/mock, hoặc moq. Nguyên tắc: interface nhỏ, đặt ở consumer để dễ mock và giảm coupling.

---

### Q: `assert` vs `require` trong testify?
**A:** `assert` ghi lỗi nhưng test tiếp tục. `require` ghi lỗi và **dừng test ngay** — dùng cho precondition (vd `require.NoError` sau khi setup, nếu fail thì các assert sau vô nghĩa).

---

### Q: Fuzzing giải quyết vấn đề gì?
**A:** Tự sinh input ngẫu nhiên (từ seed corpus) để tìm case gây panic/sai mà con người khó nghĩ ra — đặc biệt cho parser, encoder, xử lý string. Chạy `go test -fuzz=Fn`. Kiểm tra property (vd round-trip parse→serialize).

---

### Q: `t.Helper()` và `t.Cleanup()` dùng để làm gì?
**A:** `t.Helper()` đánh dấu func là helper → stacktrace lỗi trỏ về dòng gọi thật, không phải trong helper. `t.Cleanup(fn)` đăng ký dọn dẹp sau test (hoạt động đúng cả với subtest và parallel), sạch hơn defer.

---

### Q: Coverage 100% có nghĩa test tốt không?
**A:** Không. Coverage đo dòng được chạy, không đo assertion có ý nghĩa. Code chạy qua nhưng không kiểm tra kết quả vẫn tính coverage. Ưu tiên phủ branch quan trọng + edge case hơn là con số.

---

## Tooling

### Q: `go vet` khác `golangci-lint` thế nào?
**A:** `go vet` là công cụ stdlib phát hiện bug tĩnh cơ bản (printf sai, copy lock). `golangci-lint` là meta-linter chạy hàng chục linter (govet, staticcheck, errcheck, gosec...) song song — chuẩn de-facto trong CI.

---

### Q: Công cụ debug goroutine deadlock?
**A:** `delve` (dlv) — lệnh `goroutines` liệt kê toàn bộ goroutine và trạng thái, giúp thấy goroutine nào đang block ở đâu. Ngoài ra gửi SIGQUIT hoặc `GODEBUG` để dump stack, và pprof goroutine profile.

---

## Error Handling

### Q: `%w` khác `%v` trong fmt.Errorf?
**A:** `%w` bọc error gốc, giữ nguyên chain → dùng được `errors.Is/As` để bóc ra sau. `%v` chỉ nối chuỗi, mất chain. Chỉ dùng một `%w` mỗi Errorf.

---

### Q: errors.Is vs errors.As?
**A:** `errors.Is(err, target)` so khớp theo **giá trị** (sentinel như `sql.ErrNoRows`) trong chain. `errors.As(err, &target)` trích theo **kiểu** (custom struct) để đọc field bên trong. Is → "có phải lỗi X?"; As → "lấy lỗi kiểu T ra".

---

### Q: Nên log hay return error?
**A:** Chọn một, không cả hai ở cùng chỗ (tránh log trùng). Thường: layer dưới **wrap + return** kèm context; layer ngoài cùng (handler/main) **log một lần**. 

---

### Q: Khi nào dùng panic thay vì return error?
**A:** Chỉ cho lỗi lập trình không hồi phục (invariant vỡ, index out of range logic), hoặc init fail lúc khởi động. KHÔNG dùng cho lỗi nghiệp vụ/expected (record not found, validation) — những cái đó return error.

---

## encoding/json

### Q: Vì sao field struct không được serialize ra JSON?
**A:** Field không exported (chữ thường đầu) → json package bỏ qua vì dùng reflection chỉ thấy exported field. Phải viết hoa chữ đầu.

---

### Q: `json:"-"` và `json:",omitempty"` khác gì?
**A:** `-` = không bao giờ serialize field (vd password). `omitempty` = bỏ field khi giá trị là zero value (0, "", nil, empty). Lưu ý omitempty không áp dụng cho struct lồng non-pointer.

---

### Q: Unmarshal số JSON vào interface{} bị gì?
**A:** Luôn thành `float64` → mất chính xác với int64 lớn. Dùng `json.Number` (qua `dec.UseNumber()`) hoặc unmarshal vào kiểu cụ thể.

---

### Q: Vì sao nên dùng json.Decoder thay Unmarshal cho HTTP body?
**A:** Decoder stream trực tiếp từ io.Reader, không cần đọc hết vào []byte trước → tiết kiệm RAM với body lớn. Thêm `DisallowUnknownFields()` để reject field lạ (strict API).

---

## Reflection

### Q: json.Marshal hoạt động thế nào bên dưới?
**A:** Dùng reflection (`reflect`) duyệt các field exported của struct, đọc struct tag `json:"..."` để lấy tên key và các option (omitempty, -), rồi encode đệ quy. Vì reflection nên nó chậm hơn code sinh sẵn và cần field exported.

---

### Q: Nhược điểm của reflection?
**A:** Chậm (nhiều lần so với code tĩnh), mất type-safety (lỗi runtime thay vì compile-time), khó đọc. Chỉ dùng cho serialization/framework/validator, tránh trong hot path.

---

## Config & Logging

### Q: Thứ tự ưu tiên config phổ biến?
**A:** flag > env > config file > default. Theo 12-factor, config (đặc biệt secret) nên đến từ env var, không hardcode, không commit.

---

### Q: Vì sao dùng structured logging (slog/zap)?
**A:** Log dạng key-value/JSON dễ query, filter, aggregate trên ELK/Loki/Grafana (vd lọc theo trace_id, user_id). Plain text khó parse. `log/slog` là chuẩn stdlib từ Go 1.21.
