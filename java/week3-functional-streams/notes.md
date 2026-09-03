# Week 3 – Functional Programming & Streams – Deep Dive Notes

> Không dừng ở "biết dùng lambda". Ở mức senior phải giải thích **lambda KHÔNG phải anonymous class**
> (nó là `invokedynamic` + `LambdaMetafactory`), Stream pipeline chạy thế nào (Spliterator + Sink),
> và ForkJoinPool chung ảnh hưởng ra sao khi parallel.

## Mục lục
1. [Lambda: invokedynamic, không phải anonymous class](#1-lambda-invokedynamic)
2. [Capture & effectively final](#2-capture--effectively-final)
3. [Functional interfaces đầy đủ](#3-functional-interfaces-đầy-đủ)
4. [Method reference (4 loại)](#4-method-reference)
5. [Stream pipeline internals (Spliterator + Sink)](#5-stream-pipeline-internals)
6. [Lazy evaluation & short-circuit](#6-lazy-evaluation--short-circuit)
7. [Collectors sâu](#7-collectors-sâu)
8. [Optional đúng cách](#8-optional-đúng-cách)
9. [Parallel stream & ForkJoinPool](#9-parallel-stream--forkjoinpool)
10. [Follow-up questions](#10-follow-up-questions)

---

## 1. Lambda: invokedynamic

> Sai lầm phổ biến: "lambda là đường tắt của anonymous inner class". **Không đúng.**

### Anonymous class vs lambda ở mức bytecode

```java
Runnable a = new Runnable() { public void run() { System.out.println("x"); } };
Runnable b = () -> System.out.println("x");
```

- Anonymous class → sinh **class file riêng** (`Outer$1.class`), `new Outer$1()` lúc runtime.
- Lambda → **không** sinh class con lúc compile. Compiler tạo một **`invokedynamic`** call site + một method `private static lambda$main$0` chứa thân lambda.
- Lần đầu chạy, JVM gọi **bootstrap method `LambdaMetafactory.metafactory`** → sinh (spin) một lớp implement `Runnable` trỏ tới `lambda$main$0`, rồi cache lại. Các lần sau tái dùng.

### Vì sao thiết kế vậy

- Không "nổ" số class file (mỗi lambda 1 class như anonymous class → tăng thời gian load, metaspace).
- JVM tự do chọn chiến lược tạo instance (có thể reuse với lambda không capture → **stateless lambda là singleton**).

```java
// Lambda KHÔNG capture -> JVM có thể tái dùng cùng 1 instance mỗi lần
Supplier<String> s1 = () -> "hi";
Supplier<String> s2 = () -> "hi";   // có thể là cùng object (không đảm bảo, nhưng hay xảy ra)
```

> **Follow-up:** "Lambda có tạo object mới mỗi lần không?" → lambda không capture: thường reuse 1 instance. Lambda có capture: tạo instance mới giữ biến bắt được (giống closure).

---

## 2. Capture & effectively final

```java
int base = 10;                 // effectively final (không gán lại)
Function<Integer,Integer> f = x -> x + base;   // capture base bằng giá trị (copy)
// base = 20; -> nếu bỏ comment: LỖI COMPILE "must be final or effectively final"
```

- Lambda **capture biến local bằng giá trị** (copy vào field ẩn của instance lambda). Vì local nằm trên stack, có thể biến mất trước lambda → phải copy.
- Cấm gán lại để tránh nhập nhằng "giá trị nào" giữa lambda và code ngoài.
- Capture `this`/field của object → capture reference (thấy thay đổi của field). Đây là lý do lambda trong instance method giữ reference tới object bao ngoài (có thể gây leak nếu lambda sống lâu).

---

## 3. Functional interfaces đầy đủ

`@FunctionalInterface` = đúng 1 abstract method (SAM); có thể thêm default/static.

| Interface | Method | Ghi chú |
|---|---|---|
| `Function<T,R>` | `R apply(T)` | có `andThen`, `compose` |
| `BiFunction<T,U,R>` | `R apply(T,U)` | |
| `Supplier<T>` | `T get()` | lazy, factory |
| `Consumer<T>` | `void accept(T)` | `andThen` |
| `BiConsumer<T,U>` | `void accept(T,U)` | |
| `Predicate<T>` | `boolean test(T)` | `and/or/negate` |
| `UnaryOperator<T>` | `T apply(T)` | Function<T,T> |
| `BinaryOperator<T>` | `T apply(T,T)` | dùng trong reduce |

### Biến thể primitive (tránh boxing)

`IntFunction`, `ToIntFunction`, `IntPredicate`, `IntUnaryOperator`, `IntBinaryOperator`... → tránh autobox `Integer` trong hot path. Stream primitive: `IntStream`, `LongStream`, `DoubleStream`.

```java
// mapToInt -> IntStream: không box, sum() sẵn
int total = users.stream().mapToInt(User::age).sum();
```

---

## 4. Method reference

| Loại | Cú pháp | Tương đương lambda |
|---|---|---|
| static | `Integer::parseInt` | `s -> Integer.parseInt(s)` |
| instance của object cụ thể | `out::println` | `x -> out.println(x)` |
| instance của kiểu tham số | `String::toUpperCase` | `s -> s.toUpperCase()` |
| constructor | `ArrayList::new` | `() -> new ArrayList<>()` |

---

## 5. Stream pipeline internals

> "Stream chạy thế nào bên dưới?" — điểm senior.

### 3 thành phần

- **Source** → tạo **Spliterator** (splittable iterator: `tryAdvance` duyệt từng phần tử, `trySplit` chia đôi cho parallel).
- **Intermediate ops** → xâu chuỗi các **Sink** (mỗi op là một stage biến đổi). Chưa chạy gì (lazy).
- **Terminal op** → khởi động: đẩy từng phần tử qua chuỗi Sink.

### Không có collection trung gian

Điểm mấu chốt: `filter().map().collect()` **KHÔNG** tạo list sau filter rồi list sau map. Mỗi phần tử đi **dọc** qua toàn bộ chuỗi Sink một lần (pull/push fusion). Nhờ đó tiết kiệm bộ nhớ và cho phép short-circuit.

```
phần tử 1 → filter → map → collect
phần tử 2 → filter → map → collect   (KHÔNG phải: filter tất cả rồi map tất cả)
```

### Stateful vs stateless intermediate

- **Stateless**: `map`, `filter`, `peek` — xử lý độc lập từng phần tử.
- **Stateful**: `sorted`, `distinct`, `limit` — cần "nhìn" nhiều phần tử/buffer, tạo rào cản (barrier) trong pipeline, làm giảm hiệu quả (đặc biệt parallel). `sorted` phải gom hết mới sort được.

---

## 6. Lazy evaluation & short-circuit

```java
Stream.of("a", "bb", "ccc", "dddd")
    .filter(s -> { System.out.println("filter " + s); return s.length() > 1; })
    .map(s -> { System.out.println("map " + s); return s.toUpperCase(); })
    .findFirst();
// filter a, filter bb, map bb  -> DỪNG (không đụng ccc, dddd)
```

- Không có terminal op → không chạy gì (`peek` không log nếu thiếu terminal).
- Short-circuit ops: `findFirst`, `findAny`, `anyMatch`, `allMatch`, `noneMatch`, `limit` → dừng sớm nhờ lazy.

---

## 7. Collectors sâu

```java
// groupingBy + downstream
Map<String, Long> countByCity =
    users.stream().collect(groupingBy(User::city, counting()));

// groupingBy lồng nhiều tầng
Map<String, Map<Boolean, List<User>>> byCityThenAdult =
    users.stream().collect(groupingBy(User::city, groupingBy(u -> u.age() >= 18)));

// toMap + merge function (tránh IllegalStateException khi trùng key)
Map<String, Integer> byName =
    users.stream().collect(toMap(User::name, User::age, (a, b) -> a));  // giữ giá trị đầu

// teeing (Java 12): gộp 2 collector
record MinMax(int min, int max) {}
MinMax mm = nums.stream().collect(teeing(
    minBy(naturalOrder()).map(o -> o.orElse(0)),  // ... (minh hoạ ý tưởng)
    maxBy(naturalOrder()).map(o -> o.orElse(0)),
    MinMax::new));
```

### Collector là gì bên dưới

Một `Collector` gồm 4 phần: **supplier** (tạo container), **accumulator** (thêm phần tử), **combiner** (gộp 2 container khi parallel), **finisher** (biến đổi cuối). Hiểu điều này giúp viết custom collector và biết vì sao parallel cần combiner **kết hợp được (associative)**.

### Bẫy toMap trùng key

`toMap` không có merge function → **`IllegalStateException`** khi 2 phần tử cùng key. Luôn cung cấp merge function nếu key có thể trùng.

---

## 8. Optional đúng cách

```java
// TỐT
String name = repo.findById(id).map(User::name).orElse("unknown");
repo.findById(id).ifPresentOrElse(this::use, this::handleMissing);   // Java 9+
User u = repo.findById(id).orElseThrow(() -> new NotFoundException(id));
```

| Dùng SAI | Vì sao |
|---|---|
| `opt.get()` không check | ném NoSuchElementException (như NPE) |
| `if (opt.isPresent()) opt.get()` | dài, nên map/orElse/ifPresent |
| Optional làm field/tham số | Optional không Serializable, tăng overhead, sai ý định thiết kế |
| `Optional<List<T>>` | trả list rỗng thay vì Optional rỗng |

Optional là **kiểu trả về** biểu diễn "có thể không có", không phải công cụ chống null khắp nơi.

---

## 9. Parallel stream & ForkJoinPool

### Chạy trên đâu

- `parallelStream()` dùng **`ForkJoinPool.commonPool()`** — **chung toàn JVM**, size mặc định = số core - 1.
- Chia việc bằng `trySplit` (chia đôi Spliterator đệ quy), gộp bằng combiner. Mô hình fork/join + work-stealing.

### Khi nào hiệu quả (checklist)

- Dữ liệu **lớn** (thường ≥ hàng chục nghìn) và thao tác **nặng, độc lập**.
- Nguồn **dễ chia đều**: array/ArrayList (SIZED, SUBSIZED) tốt; LinkedList/IO stream kém.
- **Không** shared mutable state, không side effect, combiner associative.

### Khi nào KHÔNG nên (bẫy production)

- Trong **web request**: parallel stream chiếm common pool → chặn các request khác dùng chung pool → giảm throughput toàn app. Nếu bắt buộc parallel, dùng **ForkJoinPool riêng**:

```java
ForkJoinPool pool = new ForkJoinPool(4);
pool.submit(() -> list.parallelStream().map(...).collect(toList())).get();
```

- Thao tác nhẹ / data nhỏ: overhead chia + gộp + đồng bộ > lợi ích → chậm hơn tuần tự.
- `forEach` trên parallel không đảm bảo thứ tự; muốn giữ thứ tự dùng `forEachOrdered` (mất lợi ích song song).

---

## 10. Follow-up questions

1. Lambda có phải anonymous class không? → Không; invokedynamic + LambdaMetafactory, không sinh class con lúc compile.
2. Lambda tạo object mới mỗi lần? → không capture: thường reuse; có capture: tạo mới giữ biến.
3. Vì sao capture phải effectively final? → capture bằng copy giá trị (local trên stack có thể biến mất); cấm gán lại tránh nhập nhằng.
4. Stream có tạo collection trung gian giữa các op không? → không; mỗi phần tử đi dọc qua chuỗi Sink một lần.
5. Stateful op nào gây rào cản? → sorted, distinct, limit — cần buffer/nhìn nhiều phần tử.
6. 4 thành phần của Collector? → supplier, accumulator, combiner, finisher.
7. toMap trùng key nổ gì? → IllegalStateException; cung cấp merge function.
8. parallelStream chạy trên pool nào? → ForkJoinPool.commonPool chung JVM; tránh trong web request.
9. Vì sao IntStream tốt hơn Stream<Integer>? → không boxing/unboxing, ít GC, có sum/average sẵn.
