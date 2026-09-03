# Week 2 – Generics & Collections – Deep Dive Notes

> Generics + Collections là câu hỏi internals dày đặc nhất. Ở mức senior phải giải thích được
> **bridge method** do erasure sinh ra, **hàm spread hash** của HashMap, ngưỡng **treeify**,
> cơ chế **resize** và vì sao ConcurrentHashMap **không khoá cả map**.

## Mục lục
1. [Generics & type erasure sâu](#1-generics--type-erasure-sâu)
2. [Bridge method](#2-bridge-method)
3. [PECS & variance](#3-pecs--variance)
4. [Reifiable type & mảng generic](#4-reifiable-type--mảng-generic)
5. [ArrayList internals (grow, modCount)](#5-arraylist-internals)
6. [HashMap internals: hash → bucket → treeify → resize](#6-hashmap-internals)
7. [ConcurrentHashMap: CAS + synchronized bin](#7-concurrenthashmap)
8. [Iterator: fail-fast vs fail-safe](#8-iterator-fail-fast-vs-fail-safe)
9. [Follow-up questions](#9-follow-up-questions)

---

## 1. Generics & type erasure sâu

### Erasure làm gì

Generics chỉ tồn tại lúc compile để **type-check**. Sau đó compiler **xoá** thông tin type:

- `List<String>` → `List` (raw).
- `T` không bounded → `Object`; `T extends Number` → `Number` (erasure về bound đầu tiên).
- Compiler chèn **cast** tự động ở nơi lấy phần tử ra.

```java
List<String> l = new ArrayList<>();
l.add("x");
String s = l.get(0);   // bytecode: invokeinterface List.get -> checkcast String
```

`checkcast` được compiler chèn ⇒ nếu heap pollution (dùng raw type nhét sai kiểu) thì `ClassCastException` nổ tại điểm lấy ra, không phải điểm nhét vào.

### Hệ quả của erasure

| Không làm được | Vì sao |
|---|---|
| `new T()` | runtime không biết T là gì |
| `new T[n]` | không tạo được mảng của type bị xoá |
| `instanceof List<String>` | runtime chỉ thấy `List` |
| overload `f(List<String>)` và `f(List<Integer>)` | sau erasure cùng chữ ký `f(List)` |
| `T.class` | không có class literal của type param |
| catch `T extends Exception` generic | erasure phá type safety của catch |

> **Follow-up:** "Làm sao vượt qua giới hạn `new T()`?" → truyền `Class<T> clazz` rồi `clazz.getDeclaredConstructor().newInstance()`, hoặc truyền `Supplier<T>`.

---

## 2. Bridge method

Erasure gây vấn đề khi override generic method. Compiler sinh **bridge method** để giữ đúng đa hình.

```java
class Node<T> {
    T value;
    void set(T value) { this.value = value; }
}
class StringNode extends Node<String> {
    @Override void set(String value) { ... }   // muốn override set(T)
}
```

- Sau erasure, `Node.set` có chữ ký `set(Object)`, còn `StringNode.set` là `set(String)` → **không** phải override (khác tham số).
- Compiler sinh **bridge**: `StringNode.set(Object o) { set((String) o); }` → nối `invokevirtual set(Object)` về `set(String)`. Bridge method là synthetic, có thể thấy qua `javap`.

> Bridge method giải thích vì sao đôi khi stack trace/reflection thấy method "lạ" với tham số Object.

---

## 3. PECS & variance

Java generics **invariant**: `List<String>` KHÔNG phải `List<Object>` (khác Java array covariant). Wildcards cho variance có kiểm soát:

```java
// PRODUCER extends: chỉ đọc ra
double sum(List<? extends Number> src) {
    double s = 0; for (Number n : src) s += n.doubleValue(); return s;
    // src.add(x) -> KHÔNG compile: không biết kiểu con chính xác
}
// CONSUMER super: chỉ ghi vào
void addInts(List<? super Integer> dst) {
    dst.add(1);                    // OK
    // Integer i = dst.get(0) -> chỉ lấy được Object
}
```

| Muốn | Dùng | Đọc | Ghi |
|---|---|---|---|
| Producer | `? extends T` | ra `T` | chỉ `null` |
| Consumer | `? super T` | ra `Object` | `T` và con |
| Cả hai | `T` cụ thể | `T` | `T` |

Ví dụ chuẩn thư viện: `Collections.copy(List<? super T> dest, List<? extends T> src)`.

---

## 4. Reifiable type & mảng generic

- **Reifiable** = type mà runtime giữ đầy đủ thông tin (primitive, non-generic class, `List<?>`, `Object`...). Generic parameterized type **không** reifiable.
- Mảng là **covariant + reifiable** → xung đột với generics (invariant + non-reifiable):

```java
Object[] arr = new String[3];
arr[0] = 42;   // biên dịch OK, RUNTIME ArrayStoreException (mảng nhớ type)

// List<String>[] a = new List<String>[3];  // LỖI COMPILE: generic array creation
List<String>[] a = (List<String>[]) new List[3];  // workaround unchecked, dễ heap pollution
```

→ Ưu tiên `List<List<String>>` thay vì mảng của generic.

---

## 5. ArrayList internals

### Grow strategy

- Backing array `Object[] elementData`. Default capacity **10** (lazy, cấp phát ở lần add đầu).
- Khi đầy: `newCapacity = oldCapacity + (oldCapacity >> 1)` → **tăng 1.5x**. Copy sang mảng mới bằng `Arrays.copyOf` (O(n) nhưng amortized O(1) mỗi add).

```
10 → 15 → 22 → 33 → 49 → ...
```

- `add` cuối: amortized O(1). `add(index, e)` / `remove(index)`: O(n) do `System.arraycopy` dịch phần tử.
- **Prealloc** khi biết size: `new ArrayList<>(expectedSize)` tránh nhiều lần resize.

### modCount

- Biến đếm số lần **sửa cấu trúc** (add/remove làm đổi size). Iterator lưu `expectedModCount` lúc tạo, mỗi `next()` so sánh → khác nhau ném `ConcurrentModificationException` (fail-fast).

---

## 6. HashMap internals

> Câu hỏi kinh điển nhất. Cần vẽ được luồng: **hash → chọn bucket → xử lý collision → treeify → resize**.

### Cấu trúc

- `Node<K,V>[] table` — mảng bucket, độ dài luôn **lũy thừa 2**.
- Mỗi Node giữ `hash, key, value, next` (linked list). Khi treeify → `TreeNode` (cây đỏ-đen).

### Hàm spread hash (vì sao XOR dịch 16 bit)

```java
static final int hash(Object key) {
    int h;
    return (key == null) ? 0 : (h = key.hashCode()) ^ (h >>> 16);
}
```

- Index bucket = `(n - 1) & hash`. Vì `n` là lũy thừa 2, `(n-1)` chỉ giữ **bit thấp** → nếu hashCode chỉ khác nhau ở bit cao thì cùng bucket (collision cao).
- `h ^ (h >>> 16)` **trộn bit cao xuống thấp** → phân phối đều hơn với chi phí 1 XOR. Đây là đánh đổi chất lượng/tốc độ.

### Vì sao capacity lũy thừa 2

- Dùng `(n-1) & hash` thay `% n` — nhanh hơn nhiều (bit AND vs chia). Chỉ đúng khi n lũy thừa 2 (mask liên tục các bit thấp).

### Collision → treeify

- Cùng bucket → nối linked list. Tra cứu O(k) với k là độ dài list.
- Khi list trong 1 bucket ≥ **TREEIFY_THRESHOLD = 8** VÀ `table.length ≥ MIN_TREEIFY_CAPACITY = 64` → chuyển thành **cây đỏ-đen** → tra cứu O(log k). Nếu table < 64 thì **resize** trước thay vì treeify.
- Khi cây co lại ≤ **UNTREEIFY_THRESHOLD = 6** → chuyển ngược về list.
- Vì sao ngưỡng 8: theo phân phối Poisson với hash tốt, xác suất một bucket có 8 phần tử là ~0.00000006 → treeify là "van an toàn" chống hash xấu/tấn công, không phải trường hợp thường.

### Resize

- Load factor mặc định **0.75**. Khi `size > capacity * 0.75` → threshold vượt → **capacity gấp đôi** + rehash.
- Java 8 tối ưu rehash: với capacity gấp đôi, mỗi phần tử **hoặc ở nguyên index cũ, hoặc index cũ + oldCap** (dựa 1 bit `hash & oldCap`) → không cần tính lại toàn bộ, tách list thành 2 (lo/hi).
- Load factor 0.75 là đánh đổi: thấp hơn → ít collision nhưng tốn RAM & resize sớm; cao hơn → tiết kiệm RAM nhưng nhiều collision.

### Liên hệ equals/hashCode

`get(key)`: tính hash → chọn bucket → duyệt tìm node có `hash bằng && (key == k || key.equals(k))`. Sai contract equals/hashCode → tìm sai bucket hoặc so sai → mất entry (xem Week 1).

---

## 7. ConcurrentHashMap

> "Làm sao thread-safe mà vẫn nhanh?" — điểm phân biệt.

### Java 7 vs Java 8

- **Java 7**: chia thành 16 **Segment** (mỗi segment là một HashTable con có lock riêng) → concurrency level = số segment.
- **Java 8**: bỏ segment. Khoá **mịn hơn tới từng bucket (bin)**:
  - Bucket rỗng → thêm bằng **CAS** (`tabAt`/`casTabAt`), lock-free.
  - Bucket đã có node → `synchronized` trên **node đầu bucket** đó (chỉ khoá 1 bucket, các bucket khác song song).
  - `size` dùng `LongAdder`-style (mảng `CounterCell`) để tránh contention khi đếm.

### Vì sao không cho null

- Với map thường, `get(k)==null` mơ hồ: "không có key" hay "value là null"? Trong môi trường đa luồng, không thể dùng `containsKey` để phân biệt an toàn (race). Nên CHM cấm cả null key lẫn null value.

### So sánh

| | HashMap | ConcurrentHashMap | Hashtable | synchronizedMap |
|---|---|---|---|---|
| Thread-safe | không | có | có | có |
| Khoá | — | CAS + sync per-bin | sync cả map | sync cả map |
| null | cho | không | không | cho |
| Đa luồng | — | cao | thấp | thấp |

---

## 8. Iterator: fail-fast vs fail-safe

### Fail-fast (ArrayList, HashMap)

- Dựa `modCount`. Sửa cấu trúc khi đang duyệt (không qua iterator) → `ConcurrentModificationException` **ở lần next() kế tiếp** (best-effort, không đảm bảo).

```java
// SAI
for (String s : list) if (s.isEmpty()) list.remove(s);   // CME
// ĐÚNG
Iterator<String> it = list.iterator();
while (it.hasNext()) if (it.next().isEmpty()) it.remove();  // iterator.remove cập nhật modCount
// hoặc gọn hơn:
list.removeIf(String::isEmpty);
```

### Fail-safe (CopyOnWriteArrayList, ConcurrentHashMap)

- **CopyOnWriteArrayList**: mỗi lần ghi copy cả mảng; iterator duyệt trên snapshot cũ → không CME nhưng không thấy thay đổi mới, và ghi O(n) → chỉ hợp **đọc nhiều ghi rất ít** (listener list).
- **ConcurrentHashMap iterator**: weakly consistent — phản ánh trạng thái tại/sau khi tạo iterator, không ném CME.

---

## 9. Follow-up questions

1. Bridge method là gì, khi nào sinh ra? → override generic method; compiler nối chữ ký sau erasure.
2. Vì sao không tạo được `new List<String>[n]`? → generic không reifiable + mảng reifiable → không đảm bảo type safety.
3. HashMap treeify khi nào, vì sao ngưỡng 8? → bucket ≥8 & table ≥64; Poisson với hash tốt khiến 8 cực hiếm, chỉ là van an toàn.
4. Resize Java 8 tối ưu rehash thế nào? → tách lo/hi bằng bit `hash & oldCap`, không rehash toàn bộ.
5. Vì sao `h ^ (h >>> 16)`? → trộn bit cao xuống thấp vì index chỉ dùng bit thấp `(n-1) & hash`.
6. ConcurrentHashMap Java 8 khoá gì? → CAS cho bucket rỗng, synchronized trên node đầu bucket khi có collision.
7. Vì sao CHM cấm null? → tránh mơ hồ "không có key" vs "value null" trong môi trường đa luồng.
8. CopyOnWriteArrayList phù hợp khi nào? → đọc nhiều, ghi cực ít (ghi O(n) copy toàn mảng).
9. ArrayList grow bao nhiêu? → 1.5x (oldCap + oldCap>>1), default cap 10 lazy.
