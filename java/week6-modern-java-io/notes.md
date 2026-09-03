# Week 6 – Modern Java & IO/NIO – Deep Dive Notes

> Không dừng ở "record gọn hơn class". Phải giải thích **record biên dịch thành gì**,
> **sealed enforce ở compile + runtime thế nào**, **switch expression desugar ra sao**,
> và **NIO buffer/channel/selector** khác IO blocking chỗ nào.

## Mục lục
1. [Records internals](#1-records-internals)
2. [Sealed classes](#2-sealed-classes)
3. [Pattern matching (instanceof, switch)](#3-pattern-matching)
4. [Switch expression desugaring](#4-switch-expression)
5. [Text blocks](#5-text-blocks)
6. [var & local type inference](#6-var)
7. [IO: stream, decorator, buffered](#7-io-stream-decorator)
8. [NIO: buffer, channel, selector](#8-nio)
9. [Serialization & vì sao tránh](#9-serialization)
10. [Follow-up questions](#10-follow-up-questions)

---

## 1. Records internals

### Record biên dịch thành gì

```java
record Point(int x, int y) {}
```

Compiler sinh (xem qua `javap`):

- `private final int x; private final int y;` — field final.
- Canonical constructor `Point(int, int)`.
- Accessor `x()`, `y()` (không phải `getX()`).
- `equals`, `hashCode`, `toString` **dựa trên `invokedynamic` + `ObjectMethods.bootstrap`** — không sinh code tay mà uỷ quyền cho bootstrap method sinh runtime, đảm bảo nhất quán trên tất cả component.
- Class là `final`, kế thừa `java.lang.Record` (không extends class khác được).

### Compact constructor

```java
record Range(int lo, int hi) {
    Range {                                   // không khai báo tham số, không gán this.x
        if (lo > hi) throw new IllegalArgumentException("lo > hi");
        // có thể normalize: lo = Math.max(lo, 0);  -> gán vào tham số, tự copy vào field
    }
}
```

Compact constructor chạy **trước** phần gán field ẩn. Gán vào tham số (không phải `this.field`) → compiler tự copy vào field cuối.

### Giới hạn & khi dùng

- Không thêm instance field ngoài component; không extends; implicitly final.
- Có thể implement interface, thêm static field/method, override accessor.
- Dùng cho: DTO, value object, key của Map/Set (equals/hashCode chuẩn sẵn), kết quả trung gian.

> **Follow-up:** "Record có thực sự immutable sâu không?" → field final nhưng nếu component là kiểu mutable (List, mảng) thì nội dung vẫn đổi được. Muốn bất biến sâu phải defensive copy trong compact constructor + accessor.

---

## 2. Sealed classes

```java
sealed interface Shape permits Circle, Square, Rectangle {}
record Circle(double r) implements Shape {}
record Square(double s) implements Shape {}
record Rectangle(double w, double h) implements Shape {}
```

- Giới hạn **tập subtype được phép** qua `permits` (có thể bỏ `permits` nếu subtype ở cùng file).
- Subtype phải là `final`, `sealed`, hoặc `non-sealed`.
- Enforce ở **compile-time** (không class ngoài permits được implement) VÀ ghi vào **class file** (attribute `PermittedSubclasses`) → runtime cũng kiểm tra.
- Cho phép compiler biết **tập đóng** → switch pattern matching **exhaustive** (không cần default).

### Sealed + record = algebraic data type

Mô hình "sum type" (Shape = Circle | Square | Rectangle) + "product type" (record). Rất mạnh cho domain modeling, thay thế visitor pattern cồng kềnh.

---

## 3. Pattern matching

### instanceof pattern (GA Java 16)

```java
if (obj instanceof String s && s.length() > 3) {   // check + bind + guard, bỏ cast
    use(s);
}
```

- Bind biến `s` chỉ trong scope mà điều kiện chắc chắn đúng (flow scoping). Bỏ cast thủ công dễ sai.

### switch pattern matching (preview J17, GA J21)

```java
// GA từ Java 21 (J17 cần --enable-preview):
double area = switch (shape) {
    case Circle c    -> Math.PI * c.r() * c.r();
    case Square s    -> s.s() * s.s();
    case Rectangle r -> r.w() * r.h();
    // sealed -> exhaustive, KHÔNG cần default
};
```

- **Record deconstruction pattern** (J21): `case Circle(double r) -> ...` bóc trực tiếp component.
- **Guard**: `case Integer i when i > 0 -> ...`.

> Trong tài liệu này code example dùng **if-instanceof** (GA ở J16) để chạy được trên JDK 17 không cần `--enable-preview`.

---

## 4. Switch expression

### Cú pháp mới (J14)

```java
int days = switch (month) {
    case JAN, MAR, MAY -> 31;      // arrow: không fall-through
    case FEB -> 28;
    default -> { int d = compute(month); yield d; }   // block dùng yield trả giá trị
};
```

- **Expression** trả giá trị (khác statement). Arrow `->` không fall-through (không cần `break`).
- `yield` trả giá trị từ block.
- **Exhaustiveness**: với enum/sealed, compiler yêu cầu đủ nhánh (hoặc default) → an toàn hơn switch cũ (quên case = lỗi compile thay vì bug runtime).

### Desugaring

Switch trên `int`/enum → bytecode `tableswitch` (dày, index trực tiếp) hoặc `lookupswitch` (thưa, tìm nhị phân). Switch trên `String` → dùng `hashCode()` để nhóm rồi `equals()` xác nhận (2 tầng). Switch pattern matching → dùng `invokedynamic` (SwitchBootstraps).

---

## 5. Text blocks

```java
String json = """
    {
      "name": "an"
    }""";                // Java 15+
```

- Bỏ escape `\n`, `\"`. Compiler xử lý **incidental whitespace** (thụt lề chung bị loại dựa vị trí dấu `"""` đóng).
- `\` cuối dòng = nối dòng (không xuống dòng); `\s` = giữ space cuối dòng.
- Chỉ là **syntactic sugar** — kết quả vẫn là `String` thường (không kiểu mới).

---

## 6. var

- `var` = **local type inference**: compiler suy kiểu từ vế phải. **Kiểu vẫn tĩnh & cố định** — không phải dynamic typing, không có overhead runtime.
- Chỉ dùng cho **biến local có khởi tạo**; không field, tham số, return type, hay khởi tạo `null`.
- Nên dùng khi kiểu hiển nhiên (`var list = new ArrayList<String>()`); tránh khi làm mất rõ ràng (`var x = getResult()`).

---

## 7. IO: stream, decorator

### Byte vs char stream

| | Byte stream | Char stream |
|---|---|---|
| Lớp gốc | `InputStream`/`OutputStream` | `Reader`/`Writer` |
| Đơn vị | byte (nhị phân) | char (text + encoding) |
| Ví dụ | `FileInputStream` | `FileReader`, `InputStreamReader` |

### Decorator pattern (điểm senior)

`java.io` là ví dụ kinh điển của **Decorator pattern** — bọc stream để thêm khả năng:

```java
// bọc lớp: file -> buffer -> data
DataInputStream in = new DataInputStream(
    new BufferedInputStream(
        new FileInputStream("f.bin")));
```

- `InputStreamReader` là cầu byte→char (kèm charset). Luôn chỉ định **UTF-8** tránh phụ thuộc default charset của OS (nguồn bug mojibake).

### Vì sao buffered

Đọc từng byte trực tiếp = 1 syscall/byte → rất chậm. `BufferedInputStream` đọc một khối vào bộ nhớ, phục vụ từng byte từ buffer → giảm số syscall hàng nghìn lần.

---

## 8. NIO

> "NIO khác IO thế nào?" — cần nói được buffer/channel/selector + non-blocking.

### 3 khái niệm cốt lõi

- **Buffer** (`ByteBuffer`...) — vùng nhớ có `position`, `limit`, `capacity`. `flip()` chuyển từ ghi sang đọc. Có thể là heap buffer hoặc **direct buffer** (off-heap, tránh copy khi I/O).
- **Channel** (`FileChannel`, `SocketChannel`) — hai chiều (đọc & ghi), làm việc với buffer, có thể **non-blocking**.
- **Selector** — một thread `select()` giám sát **nhiều channel** (nhiều kết nối), chỉ xử lý channel sẵn sàng → **1 thread phục vụ hàng nghìn kết nối** (nền của Netty, server hiệu năng cao).

```
IO cũ:  1 thread / 1 connection (blocking)  -> nhiều kết nối = nhiều thread = tốn
NIO:    1 thread / N connection (selector)  -> scale kết nối tốt
```

### java.nio.file (API file hiện đại)

```java
Path p = Path.of("/tmp/data.txt");
Files.writeString(p, "hi", StandardOpenOption.CREATE);
List<String> lines = Files.readAllLines(p);
try (var s = Files.lines(p)) { s.filter(l -> !l.isBlank()).forEach(System.out::println); }  // stream, lazy, phải close
```

`Path`/`Files` tiện & an toàn hơn `java.io.File` cũ (xử lý lỗi rõ ràng, thao tác atomic, symbolic link).

---

## 9. Serialization

### Vì sao nên tránh Java Serialization

- **Bảo mật**: deserialize dữ liệu không tin cậy có thể dẫn tới **RCE** (gadget chain) — nhóm lỗ hổng nổi tiếng. `readObject` chạy code trong quá trình dựng lại object.
- **Gắn chặt**: `serialVersionUID` + cấu trúc class → đổi class dễ vỡ tương thích.
- **Không cross-language**, định dạng nhị phân khó debug.

### Thay bằng

- **JSON** (Jackson/Gson) cho API — human-readable, cross-language.
- **Protobuf/Avro** cho hiệu năng + schema evolution.
- Nếu buộc dùng Java serialization: đặt **serialization filter** (`ObjectInputFilter`, Java 9+) để chặn class nguy hiểm.

---

## 10. Follow-up questions

1. Record sinh equals/hashCode thế nào? → qua invokedynamic + ObjectMethods.bootstrap, nhất quán mọi component.
2. Record có immutable sâu không? → không nếu component mutable; cần defensive copy.
3. Sealed enforce ở đâu? → compile-time + class file attribute PermittedSubclasses (runtime).
4. instanceof pattern lợi gì? → check + bind + flow scoping, bỏ cast thủ công.
5. Switch expression khác statement? → trả giá trị, arrow không fall-through, yield, exhaustive với enum/sealed.
6. switch trên String desugar thế nào? → hashCode nhóm + equals xác nhận (2 tầng).
7. var có làm chậm runtime không? → không, chỉ suy kiểu compile-time, kiểu vẫn tĩnh.
8. Selector cho phép làm gì? → 1 thread giám sát nhiều channel → phục vụ nhiều kết nối (Netty).
9. Direct buffer là gì? → buffer off-heap, giảm copy khi I/O, nhưng cấp phát/thu hồi đắt hơn.
10. Vì sao tránh Java serialization? → RCE khi deserialize dữ liệu không tin cậy; gắn chặt; không cross-language.
