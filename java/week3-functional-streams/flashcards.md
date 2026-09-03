# Week 3 – Flashcards / Q&A (Deep Dive)

## Lambda internals

### Q: Lambda có phải đường tắt của anonymous inner class không?
**A:** Không. Anonymous class sinh class file riêng (Outer$1.class). Lambda dùng invokedynamic + bootstrap LambdaMetafactory.metafactory sinh lớp implement lúc runtime lần đầu rồi cache. Không sinh class con lúc compile, tránh nổ số class file.

---

### Q: Lambda có tạo object mới mỗi lần đánh giá không?
**A:** Lambda KHÔNG capture (stateless): JVM thường tái dùng cùng 1 instance. Lambda có capture: tạo instance mới giữ các biến bắt được (như closure). Vì thế stateless lambda gần như miễn phí.

---

### Q: Vì sao lambda capture phải effectively final?
**A:** Lambda capture biến local bằng cách copy giá trị (local nằm trên stack có thể biến mất trước khi lambda chạy). Cấm gán lại để tránh nhập nhằng giá trị giữa lambda và code ngoài.

---

### Q: Lambda trong instance method capture gì? Rủi ro?
**A:** Capture reference tới `this` (object bao ngoài) khi dùng field/method của nó. Nếu lambda sống lâu (đăng ký listener) có thể giữ object bao ngoài không cho GC → memory leak.

---

## Functional interfaces

### Q: Vì sao có biến thể IntFunction, ToIntFunction, IntStream?
**A:** Tránh autoboxing Integer trong hot path. Stream<Integer> box/unbox mỗi phần tử tốn RAM và GC; IntStream lưu int nguyên thuỷ, có sum/average sẵn. mapToInt để chuyển.

---

### Q: 4 loại method reference?
**A:** static (Integer::parseInt), instance của object cụ thể (out::println), instance của kiểu tham số (String::toUpperCase), constructor (ArrayList::new).

---

## Stream internals

### Q: Stream chạy thế nào bên dưới?
**A:** Source tạo Spliterator (tryAdvance duyệt, trySplit chia đôi). Intermediate op xâu chuỗi các Sink (lazy). Terminal op đẩy từng phần tử qua chuỗi Sink. Không tạo collection trung gian giữa các op.

---

### Q: filter().map().collect() có tạo list sau mỗi op không?
**A:** Không. Mỗi phần tử đi DỌC qua toàn bộ chuỗi Sink một lần (filter→map→collect), không phải filter tất cả rồi map tất cả. Nhờ đó tiết kiệm bộ nhớ và cho phép short-circuit.

---

### Q: Stateful vs stateless intermediate op?
**A:** Stateless (map/filter/peek) xử lý độc lập từng phần tử. Stateful (sorted/distinct/limit) cần buffer/nhìn nhiều phần tử, tạo rào cản trong pipeline, giảm hiệu quả (đặc biệt parallel). sorted phải gom hết mới sort.

---

### Q: Lazy evaluation nghĩa là gì trong Stream?
**A:** Intermediate op không chạy tới khi có terminal op. Kết hợp xử lý theo chiều dọc + short-circuit (findFirst/limit/anyMatch) cho phép dừng sớm, bỏ qua phần tử thừa.

---

## Collectors

### Q: Collector gồm những phần nào?
**A:** supplier (tạo container), accumulator (thêm phần tử), combiner (gộp 2 container khi parallel), finisher (biến đổi cuối). Combiner phải associative để parallel đúng.

---

### Q: toMap trùng key gây gì? Cách xử lý?
**A:** IllegalStateException (duplicate key). Cung cấp merge function: `toMap(k, v, (a,b)->a)` để chọn giá trị khi trùng, hoặc `(a,b)->b`, hoặc gộp.

---

### Q: groupingBy trả gì, lồng được không?
**A:** Map<K, List<T>> mặc định. Downstream collector: groupingBy(city, counting()) → Map<String,Long>. Lồng nhiều tầng: groupingBy(city, groupingBy(isAdult)) → Map<String, Map<Boolean, List>>.

---

## Optional

### Q: Optional dùng sai như thế nào?
**A:** get() không check (NoSuchElementException), Optional làm field/tham số (không Serializable, sai ý định), Optional<List> (nên trả list rỗng), if(isPresent())get() thay map/orElse. Optional chỉ nên là kiểu trả về "có thể không có".

---

## Parallel

### Q: parallelStream chạy trên pool nào? Rủi ro?
**A:** ForkJoinPool.commonPool() chung toàn JVM (size = cores-1). Trong web request, parallel stream chiếm common pool → chặn request khác → giảm throughput toàn app. Nếu cần, dùng ForkJoinPool riêng qua submit().get().

---

### Q: Khi nào KHÔNG nên parallel stream?
**A:** Data nhỏ / thao tác nhẹ (overhead chia+gộp > lợi ích), có shared mutable state/side effect, nguồn khó chia (LinkedList/IO), hoặc trong web request. Chỉ hiệu quả với data lớn + thao tác nặng độc lập + nguồn dễ chia (ArrayList/array).

---

### Q: forEach vs forEachOrdered trên parallel stream?
**A:** forEach không đảm bảo thứ tự (chạy song song). forEachOrdered giữ thứ tự gặp (encounter order) nhưng mất phần lớn lợi ích song song. Muốn kết quả có thứ tự thường dùng collect thay forEach.
