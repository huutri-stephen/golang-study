# Week 1 – Flashcards / Q&A (Deep Dive)

## Bộ nhớ & Type system

### Q: Một object rỗng `new Object()` chiếm bao nhiêu byte? Vì sao?
**A:** 16 byte trên HotSpot 64-bit: 8 byte mark word + 4 byte klass pointer (compressed oops) = 12 byte header, đệm padding lên bội số 8 → 16 byte. Mọi object alignment 8 byte.

---

### Q: `Integer` chiếm bao nhiêu byte so với `int`?
**A:** `int` = 4 byte (primitive). `Integer` = 12 byte header + 4 byte value = 16 byte object trên heap + reference 4/8 byte để trỏ tới. Autoboxing tốn ~4x bộ nhớ và gây áp lực GC → tránh trong hot path.

---

### Q: Mark word lưu gì?
**A:** Identity hashCode (lazy, tính lần đầu khi cần), thông tin lock (biased/lightweight/heavyweight), và GC age. Không lưu địa chỉ bộ nhớ vì GC di chuyển object nhưng identity hashCode phải ổn định.

---

### Q: Vì sao int[] tốt hơn Integer[] trong hot path?
**A:** int[] lưu 4 byte/phần tử liền mạch (cache-friendly). Integer[] lưu reference trỏ tới các Integer object 16 byte rải rác trên heap → tốn RAM, nhiều cache miss, tăng áp lực GC.

---

## == vs equals (bytecode)

### Q: `==` biên dịch thành bytecode nào cho reference vs int?
**A:** Reference → `if_acmpeq/if_acmpne` (so sánh con trỏ). int → `if_icmpeq/...` (so sánh giá trị). `.equals()` → `invokevirtual` (dynamic dispatch, chạy code override).

---

### Q: Vì sao `==` trên String/Integer nguy hiểm dù đôi khi đúng?
**A:** Có thể đúng do trùng string pool hoặc Integer cache, tạo ảo giác code chạy được, rồi vỡ với dữ liệu ngoài pool/cache. Bug âm thầm. Luôn dùng `.equals()`.

---

## Autoboxing

### Q: `Integer x = 100` compiler chèn gì?
**A:** `Integer.valueOf(100)`. `valueOf` khác `new Integer` vì dùng IntegerCache (-128..127) trả object chung. `new Integer` luôn tạo object mới (đã deprecated).

---

### Q: Vì sao `Integer 127 == 127` true nhưng `128 == 128` false?
**A:** IntegerCache cache -128..127 nên 127 trả cùng object. 128 ngoài cache → `valueOf` tạo 2 object mới khác nhau → `==` (so địa chỉ) là false.

---

### Q: `int v = map.get(key)` khi key không tồn tại?
**A:** `get` trả `null`, autounbox gọi `null.intValue()` → NullPointerException. Bẫy phổ biến với Map<_, Integer>. Dùng `getOrDefault` hoặc kiểm tra null.

---

## String

### Q: String immutability được thực thi thế nào?
**A:** Field value (byte[] từ Java 9) là private final, không setter, không leak reference; class String là final (không kế thừa phá được). Nhờ đó thread-safe, cache hashCode, và pool tái dùng an toàn.

---

### Q: String pool nằm ở đâu? Literal được intern khi nào?
**A:** Bảng hash trong heap (từ Java 7; trước ở PermGen). Literal được intern tự động lúc class load qua bytecode `ldc` trỏ tới CONSTANT_String trong constant pool. `new String()` tạo object mới ngoài pool.

---

### Q: `a + b` trong vòng lặp vì sao O(n²)? Fix?
**A:** Mỗi `+=` biên dịch thành StringBuilder mới + toString mới (Java 8) hoặc invokedynamic (Java 9+), tạo object trung gian mỗi vòng. Dùng một StringBuilder ngoài loop → O(n).

---

### Q: Compact Strings (Java 9) là gì?
**A:** String dùng byte[] + coder (LATIN1 1 byte/char cho ASCII, UTF16 2 byte cho còn lại) thay char[] cố định 2 byte → tiết kiệm ~50% RAM cho text ASCII (đa số web app).

---

## OOP internals

### Q: Static dispatch vs dynamic dispatch?
**A:** Overriding → dynamic dispatch: `invokevirtual` tra vtable của kiểu runtime lúc chạy. Overloading → static dispatch: chọn method lúc compile theo kiểu tĩnh của tham số. invokeinterface tra itable.

---

### Q: Thứ tự khởi tạo class có kế thừa?
**A:** static cha → static con (1 lần khi load); rồi mỗi new: instance init + field cha → ctor cha → instance init + field con → ctor con.

---

### Q: Gọi method bị override từ constructor lớp cha thấy gì?
**A:** Dynamic dispatch gọi vào lớp con nhưng field lớp con CHƯA init → thấy giá trị mặc định (null/0). Anti-pattern nguy hiểm, tránh gọi overridable method trong constructor.

---

### Q: Diamond problem với default method giải quyết thế nào?
**A:** Nếu implement 2 interface có default method cùng tên, class BẮT BUỘC override, chọn tường minh bằng `A.super.method()`. Quy tắc: class thắng interface, interface con thắng cha.

---

## equals/hashCode

### Q: equals/hashCode contract và hệ quả vi phạm?
**A:** a.equals(b) ⇒ a.hashCode()==b.hashCode() (ngược lại không bắt buộc). Vi phạm: HashMap tính bucket bằng hashCode, 2 object equals mà hashCode khác rơi bucket khác → get/contains sai.

---

### Q: Vì sao dùng số 31 trong hashCode?
**A:** 31 lẻ + nguyên tố → phân phối hash tốt, giảm collision. `31*i == (i<<5)-i` nên JIT tối ưu bằng shift+subtract (nhanh). Là quy ước từ String.hashCode.

---

### Q: Bẫy đối xứng equals khi kế thừa?
**A:** Nếu lớp con thêm field vào equals, `parent.equals(child)` có thể true còn `child.equals(parent)` false → phá symmetric. Giải: composition thay kế thừa, hoặc record (final). getClass() thay instanceof phá Liskov.

---

## Exceptions

### Q: Phần nào của exception đắt nhất?
**A:** `fillInStackTrace()` chạy lúc tạo exception, duyệt toàn bộ stack. throw/catch bản thân rẻ. Vì thế không dùng exception cho control flow; hot exception có thể tắt stack trace qua constructor `super(msg, cause, false, false)`.

---

### Q: Suppressed exception trong try-with-resources?
**A:** Resource đóng theo thứ tự ngược khai báo. Nếu body ném exception VÀ close() cũng ném, exception của body là chính, exception của close gắn vào getSuppressed() (không mất).

---

### Q: Checked vs unchecked, khi nào dùng gì?
**A:** Checked (IOException) bị compiler ép handle/throws — lỗi phục hồi được, nhưng lan khắp signature. Unchecked (RuntimeException) — lỗi lập trình. Thực tế nhiều codebase (Spring) ưu tiên unchecked cho lỗi nghiệp vụ để signature sạch.
