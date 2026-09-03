# Week 1 – Java Core & OOP – Deep Dive Notes

> Nền tảng bắt buộc, nhưng đi tới tận **bytecode và bố cục bộ nhớ object**. Ở mức senior,
> interviewer không hỏi "equals là gì" mà hỏi "điều gì xảy ra ở mức JVM khi bạn gọi equals",
> "một Integer chiếm bao nhiêu byte", "String pool nằm ở đâu trong bộ nhớ".

## Mục lục
1. [Type system & bộ nhớ object](#1-type-system--bộ-nhớ-object)
2. [`==` vs `equals` ở mức bytecode](#2--vs-equals-ở-mức-bytecode)
3. [Autoboxing & Integer cache](#3-autoboxing--integer-cache)
4. [String: pool, immutability, nối chuỗi](#4-string-pool-immutability-nối-chuỗi)
5. [OOP internals: dispatch, vtable, init order](#5-oop-internals-dispatch-vtable-init-order)
6. [interface vs abstract class (+ default method)](#6-interface-vs-abstract-class)
7. [equals() & hashCode() – contract sâu](#7-equals--hashcode--contract-sâu)
8. [Exceptions: hierarchy, chi phí, best practice](#8-exceptions)
9. [Follow-up questions (mock)](#9-follow-up-questions)

---

## 1. Type system & bộ nhớ object

### 8 primitive & kích thước

| Type | Bits | Range / ghi chú |
|---|---|---|
| `boolean` | JVM không định nghĩa (thường 1 byte, trong mảng có thể bit-packed) | true/false |
| `byte` | 8 | -128..127 |
| `short` | 16 | |
| `char` | 16 | **unsigned** 0..65535 (UTF-16 code unit) |
| `int` | 32 | |
| `long` | 64 | |
| `float` | 32 | IEEE-754 |
| `double` | 64 | IEEE-754 |

- Primitive nằm **trực tiếp** trên stack (biến local) hoặc inline trong object (field).
- Reference là một "con trỏ được quản lý" (thường 4 byte khi bật **compressed oops** cho heap < 32GB, 8 byte nếu không) trỏ tới object trên heap.

### Bố cục một object trên heap (HotSpot 64-bit)

```
┌─────────────────────────── Object ───────────────────────────┐
│ Object Header                                                 │
│   ├── Mark Word      (8 byte): hashCode, GC age, lock state   │
│   └── Klass Pointer  (4 byte nếu compressed / 8 byte)         │
│ [Array length]       (4 byte, chỉ với mảng)                   │
│ Instance fields      (int 4, long 8, reference 4/8, ...)      │
│ Padding              (đệm cho bội số 8 byte)                  │
└───────────────────────────────────────────────────────────────┘
```

- **Mark word** giữ identity hashCode (lazy — chỉ tính khi gọi `System.identityHashCode`/`Object.hashCode` lần đầu), thông tin lock (biased/lightweight/heavyweight), và GC age.
- **Klass pointer** trỏ tới metadata class trong Metaspace (không phải object `Class`).
- Mọi object **alignment 8 byte** → luôn đệm padding.
- Ví dụ: `new Object()` = 12 byte header → pad lên **16 byte**. `Integer` = 12 header + 4 (int value) = **16 byte** (so với `int` chỉ 4 byte → autoboxing đắt gấp 4 lần + gây áp lực GC).

> **Follow-up hay gặp:** "Tại sao dùng int[] thay Integer[] trong hot path?" → int[] lưu 4 byte/phần tử liền mạch (cache-friendly); Integer[] lưu reference (4/8 byte) trỏ tới các Integer object rải rác 16 byte mỗi cái → tốn RAM, nhiều cache miss, áp lực GC.

---

## 2. `==` vs `equals` ở mức bytecode

### Nguồn Java

```java
String a = new String("hi");
String b = new String("hi");
boolean r1 = (a == b);        // false
boolean r2 = a.equals(b);     // true
```

### Bytecode tương ứng (rút gọn từ `javap -c`)

```
// a == b  -> so sánh 2 reference trên stack
aload_1
aload_2
if_acmpne  ...        // "if reference compare not equal" -> so sánh ĐỊA CHỈ

// a.equals(b) -> gọi virtual method
aload_1
aload_2
invokevirtual java/lang/String.equals:(Ljava/lang/Object;)Z
```

- `==` biên dịch thành **`if_acmpeq/if_acmpne`** với reference → so sánh con trỏ, không đụng tới nội dung.
- `.equals()` biên dịch thành **`invokevirtual`** → dynamic dispatch, chạy code override của `String`.
- Với primitive, `==` là `if_icmpeq` (int), `lcmp` (long)... → so sánh giá trị bit.

### Vì sao nhầm lẫn nguy hiểm

`==` trên wrapper hoặc String **đôi khi đúng do trùng cache/pool**, tạo ảo giác code chạy được, rồi vỡ ở dữ liệu khác. Đây là bug âm thầm.

---

## 3. Autoboxing & Integer cache

### Cơ chế autoboxing

```java
Integer x = 100;         // compiler chèn: Integer.valueOf(100)
int y = x;               // compiler chèn: x.intValue()
```

`javac` chèn `Integer.valueOf` (box) và `intValue` (unbox) tự động. `valueOf` khác `new Integer` — nó **dùng cache**.

### IntegerCache

```java
// Trích logic Integer.valueOf (diễn giải lại):
public static Integer valueOf(int i) {
    if (i >= -128 && i <= IntegerCache.high)  // mặc định high = 127
        return IntegerCache.cache[i + 128];   // trả object CHUNG trong cache
    return new Integer(i);                     // ngoài cache -> object mới
}
```

- Cache `-128..127` (giới hạn trên có thể chỉnh qua `-XX:AutoBoxCacheMax`).
- Hệ quả kinh điển:

```java
Integer a = 127, b = 127;   a == b;  // true  (cùng object cache)
Integer c = 128, d = 128;   c == d;  // false (2 object khác nhau)
```

- `Long`, `Short`, `Byte`, `Character` cũng có cache tương tự (-128..127). `Boolean` cache cả 2 giá trị.

### Bẫy NPE khi unbox null

```java
Map<String, Integer> m = new HashMap<>();
int v = m.get("missing");   // get trả null -> unbox null.intValue() -> NullPointerException
```

> **Follow-up:** "Vì sao `a == b` với Integer 1000 có thể true trong một JVM?" → nếu chạy với `-XX:AutoBoxCacheMax=1000` thì 1000 nằm trong cache. Quy tắc an toàn: **luôn `.equals()` cho wrapper**.

---

## 4. String: pool, immutability, nối chuỗi

### Immutability được thực thi thế nào

- Field `value` (mảng byte từ Java 9, trước đó là `char[]`) là `private final`, không có setter, không leak reference ra ngoài.
- Class `String` là `final` → không thể kế thừa để phá bất biến.
- Từ Java 9: **Compact Strings** — String dùng `byte[]` + `coder` (LATIN1 1 byte/char, hoặc UTF16 2 byte/char) → tiết kiệm ~50% RAM cho text ASCII.

### String pool nằm ở đâu

- Là một bảng hash trong heap (từ Java 7 trở đi; trước đó ở PermGen). Literal String được **intern tự động** lúc class load qua bytecode `ldc` trỏ tới `CONSTANT_String_info` trong constant pool.

```java
String s1 = "hi";              // ldc -> intern tự động, lấy từ pool
String s2 = "hi";              // cùng object trong pool
String s3 = new String("hi");  // new StringBuilder object mới trên heap
s1 == s2;          // true
s1 == s3;          // false
s1 == s3.intern(); // true  (intern trả về object trong pool)
```

### Nối chuỗi: `+` được biên dịch thành gì

- Java 8: `a + b` → `new StringBuilder().append(a).append(b).toString()`.
- Java 9+: dùng **`invokedynamic` + `StringConcatFactory`** (indy) → JIT sinh chiến lược nối tối ưu lúc runtime.
- **Trong vòng lặp**, mỗi `+=` tạo một StringBuilder + String mới → **O(n²)**. Phải dùng một `StringBuilder` bên ngoài loop → O(n).

```java
// XẤU: O(n^2)
String s = "";
for (int i = 0; i < n; i++) s += i;

// TỐT: O(n)
StringBuilder sb = new StringBuilder();
for (int i = 0; i < n; i++) sb.append(i);
```

### String vs StringBuilder vs StringBuffer

| | Mutable | Đồng bộ | Ghi chú |
|---|---|---|---|
| `String` | không | — (immutable → thread-safe) | mọi thao tác tạo object mới |
| `StringBuilder` | có | không | mặc định cho nối chuỗi 1 thread |
| `StringBuffer` | có | có (`synchronized`) | chậm hơn, hầu như không cần ngày nay |

> **Follow-up:** "Vì sao String làm key HashMap tốt?" → immutable ⇒ hashCode không đổi sau khi put (không lạc bucket), và được cache trong field `hash` (tính 1 lần).

---

## 5. OOP internals: dispatch, vtable, init order

### 4 trụ cột (nhanh, nhưng nhìn ở góc cơ chế)

- **Encapsulation** — `private` field + method có kiểm soát; JVM không ép, chỉ compiler + reflection guard.
- **Inheritance** — `extends`; mỗi class có một **vtable** (method table) trong metadata.
- **Polymorphism** — overriding qua dynamic dispatch; overloading qua static dispatch.
- **Abstraction** — interface/abstract.

### Static dispatch vs dynamic dispatch

```java
class Animal { String sound() { return "?"; } }
class Dog extends Animal { @Override String sound() { return "woof"; } }

Animal a = new Dog();
a.sound();   // "woof"
```

- Overriding → `invokevirtual`: JVM tra **vtable** của kiểu runtime (Dog) → gọi `Dog.sound`. Đây là **dynamic dispatch** (quyết định lúc chạy).
- Overloading → quyết định lúc **compile** dựa kiểu tĩnh của tham số → `invokevirtual`/`invokestatic` tới đúng phương thức đã chọn sẵn.
- `invokeinterface` cho gọi qua interface (tra itable), `invokestatic` cho static, `invokespecial` cho constructor/private/super.

### Covariant return

```java
class A { Object f() {...} }
class B extends A { @Override String f() {...} }  // return hẹp hơn (String <: Object) - hợp lệ
```

### Thứ tự khởi tạo (rất hay hỏi)

```
1. static của lớp cha (1 lần khi class load)
2. static của lớp con
3. --- mỗi lần new: ---
4. instance init block + field của cha
5. constructor của cha
6. instance init block + field của con
7. constructor của con
```

```java
class Base {
    static { System.out.println("1 base static"); }
    { System.out.println("3 base instance"); }
    Base() { System.out.println("4 base ctor"); }
}
class Child extends Base {
    static { System.out.println("2 child static"); }
    { System.out.println("5 child instance"); }
    Child() { System.out.println("6 child ctor"); }
}
// new Child() -> 1,2 (lần đầu),3,4,5,6
```

> **Follow-up:** "Gọi method bị override từ constructor của lớp cha thì sao?" → dynamic dispatch gọi vào lớp con **khi field lớp con chưa init** → thấy giá trị mặc định (null/0). Đây là anti-pattern nguy hiểm.

### Composition over inheritance

Kế thừa tạo coupling chặt (fragile base class: đổi cha vỡ con), lộ toàn bộ API cha, và Java chỉ đơn kế thừa. Ưu tiên **composition** (has-a) + interface trừ khi thực sự "is-a" và ổn định.

---

## 6. interface vs abstract class

| | interface | abstract class |
|---|---|---|
| Đa kế thừa | có (implement nhiều) | không |
| Field | chỉ `public static final` | field bất kỳ (có state) |
| Constructor | không | có |
| Method có thân | `default`, `static`, `private` (Java 9+) | có |
| Khi dùng | định nghĩa capability/contract | chia sẻ state + code chung |

### Default method & vấn đề "diamond"

```java
interface A { default String hi() { return "A"; } }
interface B { default String hi() { return "B"; } }
class C implements A, B {
    // BẮT BUỘC override, nếu không -> lỗi compile "inherits unrelated defaults"
    @Override public String hi() { return A.super.hi(); }  // chọn tường minh
}
```

- Default method ra đời (Java 8) để **thêm method vào interface mà không phá code cũ** (vd `Collection.stream()`).
- Quy tắc phân giải: **class thắng interface**; interface con thắng interface cha; còn lại phải override tường minh.

---

## 7. equals() & hashCode() – contract sâu

### 5 quy tắc của equals

1. Reflexive: `x.equals(x)` = true.
2. Symmetric: `x.equals(y)` ⇔ `y.equals(x)`.
3. Transitive: `x.equals(y) && y.equals(z)` ⇒ `x.equals(z)`.
4. Consistent: gọi nhiều lần cùng kết quả (khi object không đổi).
5. `x.equals(null)` = false.

### Contract với hashCode

- `a.equals(b)` ⇒ **bắt buộc** `a.hashCode() == b.hashCode()`.
- Ngược lại KHÔNG bắt buộc (collision là hợp lệ).
- Vi phạm ⇒ HashMap tính bucket bằng hashCode, nếu 2 object equals mà hashCode khác → rơi **bucket khác** → `get`/`contains` trả sai.

### Bẫy đối xứng khi kế thừa

```java
// Point vs ColorPoint: nếu ColorPoint.equals so cả color còn Point.equals chỉ so toạ độ
// point.equals(colorPoint) = true nhưng colorPoint.equals(point) = false -> VỠ đối xứng.
```

→ Giải pháp: **composition thay kế thừa**, hoặc dùng `getClass()` thay `instanceof` (đánh đổi: phá Liskov với proxy/subclass). `record` giải quyết gọn vì final.

### Cách viết đúng (Java 16+ với pattern matching)

```java
@Override public boolean equals(Object o) {
    if (this == o) return true;                    // shortcut + xử lý reflexive
    if (!(o instanceof User u)) return false;      // null-safe + type check + bind
    return id == u.id && Objects.equals(email, u.email);
}
@Override public int hashCode() {
    return Objects.hash(id, email);                // ĐÚNG các field như equals
}
```

- `Objects.hash(...)` tiện nhưng tạo mảng varargs (chút overhead) — hot path có thể tự viết `31 * result + field`.
- Vì sao số **31**: lẻ + nguyên tố, `31 * i == (i << 5) - i` (JIT tối ưu bằng shift), phân phối hash tốt.

> **Follow-up:** "hashCode của Object mặc định là gì?" → identity hashCode, lazy lưu trong mark word (không phải địa chỉ bộ nhớ — vì GC di chuyển object nhưng hashCode phải ổn định).

---

## 8. Exceptions

### Hierarchy

```
Throwable
 ├── Error              (OutOfMemoryError, StackOverflowError) — KHÔNG catch
 └── Exception
      ├── RuntimeException   (unchecked: NPE, IllegalArgument, IndexOutOfBounds)
      └── (checked)          (IOException, SQLException)
```

### Chi phí của exception (điểm senior)

- Đắt nhất là **`fillInStackTrace()`** — chạy khi tạo exception, duyệt toàn bộ stack. `throw`/`catch` bản thân rẻ.
- Vì thế: **không dùng exception cho control flow** (vd thoát vòng lặp). Với exception "nóng" có thể override `fillInStackTrace()` để bỏ stack trace (dùng hạn chế).

```java
// Exception không stack trace (cho control-flow performance-sensitive, dùng cẩn thận)
static final class FastEx extends RuntimeException {
    FastEx() { super(null, null, false, false); } // writableStackTrace=false
}
```

### try-with-resources & suppressed exceptions

```java
try (var a = open("A"); var b = open("B")) {   // đóng theo thứ tự NGƯỢC: b rồi a
    use(a, b);
}
// Nếu body ném exception VÀ close() cũng ném -> exception của body là chính,
// exception của close được gắn vào getSuppressed()
```

### Best practice

- Xử lý một lần: log **hoặc** rethrow, không cả hai.
- Wrap giữ cause: `throw new ServiceException("load " + id, e)`.
- Catch cụ thể trước tổng quát; không catch `Throwable`/`Error`.
- Lỗi nghiệp vụ → unchecked (tránh checked lan khắp signature).

---

## 9. Follow-up questions

Bộ câu hỏi interviewer hay hỏi tiếp (tự trả lời to):

1. `new Object()` chiếm bao nhiêu byte và vì sao? → 16 byte (12 header + 4 padding).
2. Identity hashCode lưu ở đâu, có phải địa chỉ bộ nhớ không? → mark word, không phải địa chỉ (GC di chuyển object).
3. `a + b + c` với String tạo mấy object ở Java 8 vs Java 9? → J8: qua StringBuilder; J9+: invokedynamic một lần.
4. Vì sao `Integer` cache -128..127? → tối ưu cho các giá trị nhỏ hay dùng (vòng lặp, flag).
5. Gọi overridden method trong constructor cha → thấy gì? → giá trị mặc định của field con (chưa init).
6. Vì sao `record` không phá đối xứng equals? → final, generated equals so tất cả component nhất quán.
7. `==` biên dịch thành bytecode nào cho reference vs int? → `if_acmp*` vs `if_icmp*`.
8. Compact Strings tiết kiệm gì? → dùng byte[] LATIN1 cho ASCII, ~50% RAM.
