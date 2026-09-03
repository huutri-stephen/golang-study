# Week 6 – Flashcards / Q&A (Deep Dive)

## Records

### Q: Record được compiler sinh ra những gì?
**A:** Field private final cho mỗi component, canonical constructor, accessor tên component (x() không phải getX()), equals/hashCode/toString (qua invokedynamic + ObjectMethods.bootstrap), class final kế thừa java.lang.Record. Không extends class khác.

---

### Q: Record có immutable sâu không?
**A:** Không hẳn. Field là final nhưng nếu component là kiểu mutable (List, mảng) thì nội dung vẫn thay đổi được qua reference. Muốn bất biến sâu phải defensive copy trong compact constructor và accessor.

---

### Q: Compact constructor chạy khi nào?
**A:** Chạy trước phần gán field ẩn. Không khai báo tham số/không gán this.field; gán vào tham số để normalize thì compiler tự copy vào field. Dùng để validate hoặc chuẩn hoá.

---

## Sealed

### Q: Sealed class enforce ở đâu?
**A:** Compile-time (chỉ class trong permits được implement) VÀ runtime (ghi vào class file attribute PermittedSubclasses). Subtype phải là final, sealed, hoặc non-sealed.

---

### Q: Sealed + record cho phép mô hình gì?
**A:** Algebraic data type: sum type (Shape = Circle | Square | Rectangle qua sealed) + product type (record). Cho phép switch pattern matching exhaustive không cần default, thay visitor pattern.

---

## Pattern matching & switch

### Q: instanceof pattern matching lợi gì?
**A:** `if (o instanceof String s)` vừa check kiểu vừa bind biến s đã cast (flow scoping: s chỉ dùng được khi chắc đúng kiểu). Bỏ cast thủ công dễ sai. GA từ Java 16.

---

### Q: Switch expression khác switch statement?
**A:** Expression trả giá trị; arrow -> không fall-through (không cần break); yield trả giá trị từ block; exhaustive với enum/sealed (thiếu nhánh = lỗi compile). An toàn hơn switch cũ.

---

### Q: switch trên String desugar thế nào?
**A:** 2 tầng: dùng hashCode() để nhóm các case (tableswitch/lookupswitch trên hash), rồi equals() xác nhận đúng chuỗi (tránh hash collision). Switch trên int/enum dùng tableswitch/lookupswitch trực tiếp.

---

## Text block & var

### Q: Text block xử lý thụt lề thế nào?
**A:** Compiler loại "incidental whitespace" (thụt lề chung, xác định bởi vị trí """ đóng). \ cuối dòng nối dòng, \s giữ space cuối. Chỉ là syntactic sugar, kết quả là String thường.

---

### Q: var có phải dynamic typing không?
**A:** Không. var là local type inference: compiler suy kiểu từ vế phải, kiểu vẫn tĩnh và cố định, không overhead runtime. Chỉ cho biến local có khởi tạo, không field/tham số/return/null.

---

## IO / NIO

### Q: java.io dùng design pattern nào?
**A:** Decorator: bọc stream để thêm khả năng, vd new DataInputStream(new BufferedInputStream(new FileInputStream(f))). InputStreamReader là cầu byte→char kèm charset (nên chỉ định UTF-8).

---

### Q: Vì sao buffered stream nhanh hơn nhiều?
**A:** Đọc từng byte trực tiếp = 1 syscall/byte, rất chậm. BufferedInputStream đọc một khối vào bộ nhớ rồi phục vụ từng byte từ buffer → giảm số syscall hàng nghìn lần.

---

### Q: 3 khái niệm cốt lõi của NIO?
**A:** Buffer (vùng nhớ có position/limit/capacity, flip() chuyển ghi↔đọc, có direct buffer off-heap), Channel (hai chiều, làm việc với buffer, có thể non-blocking), Selector (1 thread select() giám sát nhiều channel).

---

### Q: NIO khác IO blocking thế nào về mô hình thread?
**A:** IO cũ: 1 thread/1 connection (blocking) → nhiều kết nối tốn nhiều thread. NIO với selector: 1 thread giám sát N channel, chỉ xử lý channel sẵn sàng → 1 thread phục vụ hàng nghìn kết nối (nền của Netty).

---

### Q: Direct buffer là gì, đánh đổi?
**A:** ByteBuffer off-heap (allocateDirect): giảm copy giữa JVM heap và OS khi I/O → nhanh cho I/O lớn. Đánh đổi: cấp phát/thu hồi đắt hơn, không do GC thường quản (dễ leak nếu giữ lâu).

---

## Serialization

### Q: Vì sao tránh Java Serialization?
**A:** Deserialize dữ liệu không tin cậy có thể RCE (gadget chain, readObject chạy code); gắn chặt cấu trúc class (serialVersionUID, khó đổi version); không cross-language; nhị phân khó debug. Thay bằng JSON hoặc Protobuf/Avro.

---

### Q: Nếu buộc dùng Java serialization thì phòng thủ thế nào?
**A:** Dùng ObjectInputFilter (Java 9+) để whitelist/blacklist class được deserialize, giới hạn depth/số object, chặn các gadget class nguy hiểm. Không bao giờ deserialize dữ liệu từ nguồn không tin cậy mà không lọc.
