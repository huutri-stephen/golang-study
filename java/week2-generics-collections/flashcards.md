# Week 2 – Flashcards / Q&A (Deep Dive)

## Generics & Erasure

### Q: Type erasure làm gì với `T` và `T extends Number`?
**A:** T không bound → erase về Object; T extends Number → erase về Number (bound đầu tiên). Compiler chèn checkcast ở nơi lấy phần tử ra. Runtime không còn thông tin generic.

---

### Q: Vì sao ClassCastException do heap pollution nổ ở điểm lấy ra, không phải nhét vào?
**A:** Vì compiler chèn `checkcast` tại điểm lấy phần tử (get) chứ không phải lúc add. Dùng raw type nhét sai kiểu qua được compile, tới khi get + cast mới nổ.

---

### Q: Bridge method là gì?
**A:** Method synthetic compiler sinh khi override generic method. Sau erasure Node.set(T)→set(Object) nhưng StringNode.set(String) khác chữ ký; compiler sinh bridge set(Object){set((String)o);} để giữ đa hình.

---

### Q: Làm sao vượt giới hạn `new T()` do erasure?
**A:** Truyền Class<T> rồi clazz.getDeclaredConstructor().newInstance(), hoặc truyền Supplier<T> vào. Vì runtime không biết T là gì nên cần thông tin type từ bên ngoài.

---

### Q: Generics của Java invariant nghĩa là gì?
**A:** List<String> KHÔNG phải List<Object> (không gán được), khác mảng covariant. Muốn variance có kiểm soát dùng wildcard: ? extends (covariant, đọc), ? super (contravariant, ghi).

---

### Q: Vì sao mảng generic bị cấm?
**A:** Mảng reifiable + covariant (nhớ type runtime, ném ArrayStoreException). Generic non-reifiable (erasure). Kết hợp sẽ phá type safety mà không phát hiện được runtime → compiler cấm `new List<String>[n]`.

---

## PECS

### Q: PECS đầy đủ?
**A:** Producer Extends Consumer Super. `? extends T`: đọc ra T, chỉ ghi null. `? super T`: ghi T, đọc ra Object. Vừa đọc vừa ghi: dùng T cụ thể. VD: Collections.copy(List<? super T> dest, List<? extends T> src).

---

## ArrayList

### Q: ArrayList grow bao nhiêu mỗi lần?
**A:** newCap = oldCap + (oldCap >> 1) = 1.5x. Default cap 10 (lazy, cấp ở add đầu). Copy bằng Arrays.copyOf O(n) nhưng amortized O(1)/add. Biết size trước thì new ArrayList<>(size) để tránh resize.

---

### Q: add(index) và remove(index) của ArrayList phức tạp bao nhiêu?
**A:** O(n) vì System.arraycopy dịch các phần tử sau vị trí đó. get/set là O(1). add cuối amortized O(1). Đây là lý do LinkedList đôi khi được nhắc, nhưng thực tế ArrayList vẫn thắng nhờ cache locality.

---

## HashMap

### Q: Index bucket được tính thế nào?
**A:** `(n-1) & hash(key)` với n là capacity (lũy thừa 2). Dùng bit AND thay % vì nhanh hơn; chỉ đúng khi n lũy thừa 2.

---

### Q: Vì sao HashMap dùng `h ^ (h >>> 16)`?
**A:** Index chỉ dùng bit thấp `(n-1) & hash`. Nếu hashCode khác nhau chỉ ở bit cao thì cùng bucket → collision. XOR dịch 16 trộn bit cao xuống thấp để phân phối đều, chi phí chỉ 1 phép XOR.

---

### Q: Treeify khi nào? Vì sao ngưỡng 8?
**A:** Bucket ≥8 phần tử VÀ table.length ≥64 → chuyển linked list thành cây đỏ-đen (O(log k)). Nếu table<64 thì resize trước. Ngưỡng 8 vì với hash tốt (Poisson) xác suất bucket có 8 phần tử ~6e-8 → treeify chỉ là van chống hash xấu/tấn công.

---

### Q: Resize Java 8 tối ưu rehash thế nào?
**A:** Capacity gấp đôi. Mỗi phần tử hoặc ở index cũ, hoặc index cũ + oldCap, quyết định bằng 1 bit `hash & oldCap`. Tách list thành 2 (lo/hi) mà không tính lại hash toàn bộ.

---

### Q: Load factor 0.75 đánh đổi gì?
**A:** size > cap*0.75 thì resize. Thấp hơn → ít collision nhưng tốn RAM, resize sớm. Cao hơn → tiết kiệm RAM nhưng nhiều collision, tra cứu chậm. 0.75 là cân bằng thời gian/không gian.

---

## ConcurrentHashMap

### Q: ConcurrentHashMap Java 8 khoá gì (khác Java 7)?
**A:** Java 7 chia 16 Segment mỗi cái có lock. Java 8 bỏ segment: bucket rỗng thêm bằng CAS (lock-free); bucket có node thì synchronized trên node đầu bucket đó → chỉ khoá 1 bucket, các bucket khác song song.

---

### Q: Vì sao ConcurrentHashMap cấm null?
**A:** get(k)==null mơ hồ giữa "không có key" và "value null". Đa luồng không thể dùng containsKey phân biệt an toàn (race). Nên cấm cả null key và null value.

---

### Q: ConcurrentHashMap đếm size thế nào để tránh contention?
**A:** Dùng mảng CounterCell (kiểu LongAdder): mỗi thread cộng vào cell riêng, size = baseCount + tổng các cell. Tránh mọi thread tranh một biến counter chung.

---

## Iterator

### Q: Fail-fast dựa vào gì? Có đảm bảo không?
**A:** Dựa modCount (đếm số lần sửa cấu trúc). Iterator lưu expectedModCount, mỗi next() so sánh, khác thì ném ConcurrentModificationException. Best-effort, KHÔNG đảm bảo phát hiện mọi trường hợp.

---

### Q: Xóa phần tử khi duyệt sao cho đúng?
**A:** Dùng Iterator.remove() (cập nhật cả modCount và expectedModCount), hoặc list.removeIf(predicate), hoặc duyệt bản copy. Không dùng for-each rồi collection.remove() (gây CME).

---

### Q: CopyOnWriteArrayList phù hợp khi nào?
**A:** Đọc nhiều, ghi cực ít (listener list, config). Ghi copy toàn mảng O(n); đọc lock-free trên snapshot. Iterator không ném CME nhưng không thấy thay đổi sau khi tạo.
