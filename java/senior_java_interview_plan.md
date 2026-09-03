# Senior Backend Engineer – Java Interview Preparation Plan

> Bám theo roadmap.sh/java (diễn giải lại theo license), tham khảo cấu trúc plan Golang.
> Nội dung nhắm mức **senior backend**: hiểu sâu JVM/memory/concurrency, không chỉ cú pháp.
> Nguồn tham khảo roadmap: [roadmap.sh/java](https://roadmap.sh/java). Nội dung được diễn giải lại cho phù hợp license.

## 1. Mục tiêu

Chuẩn bị cho phỏng vấn Senior Backend Engineer – Java, tập trung vào:

- Java core & OOP idiomatic (equals/hashCode, immutability, exceptions)
- Generics & Collections internals (HashMap, ConcurrentHashMap, type erasure)
- Functional programming & Stream API
- Concurrency & multithreading (executors, CompletableFuture, JMM, virtual threads)
- JVM internals, memory model, garbage collection & tuning
- Modern Java (records, sealed, pattern matching, switch expressions) + IO/NIO
- Build tools, testing, ORM/persistence, ecosystem
- Spring/Spring Boot, microservices, system design
- Coding interview + production troubleshooting + behavioral

**Môi trường:** JDK 17 (LTS). Code examples chạy standalone bằng single-file source mode:
`java TênFile.java` (không cần Maven/Gradle). Nơi cần framework/deps sẽ ghi rõ.

---

# 2. Roadmap tổng quan – 8 tuần

| Tuần | Chủ đề | Priority |
|---|---|---|
| 1 | Java Core & OOP | ⭐⭐⭐⭐⭐ |
| 2 | Generics & Collections | ⭐⭐⭐⭐⭐ |
| 3 | Functional & Streams API | ⭐⭐⭐⭐ |
| 4 | Concurrency & Multithreading | ⭐⭐⭐⭐⭐ |
| 5 | JVM, Memory & Garbage Collection | ⭐⭐⭐⭐⭐ |
| 6 | Modern Java & IO/NIO | ⭐⭐⭐⭐ |
| 7 | Build, Testing & Ecosystem | ⭐⭐⭐⭐ |
| 8 | Spring, System Design & Mock | ⭐⭐⭐⭐⭐ |

## Tỷ lệ thời gian đề xuất

- 30% Theory (JVM, memory, concurrency internals)
- 30% Coding
- 25% System Design / Architecture
- 15% Mock Interview / Review

---

# 3. Week 1 – Java Core & OOP

## Types & Memory basics

- [ ] Primitive vs reference types, autoboxing/unboxing
- [ ] Stack vs heap (biến local vs object)
- [ ] `==` vs `.equals()`
- [ ] String pool, immutability của String, `String` vs `StringBuilder` vs `StringBuffer`
- [ ] Integer cache (-128..127)

## OOP

- [ ] Encapsulation, Inheritance, Polymorphism, Abstraction
- [ ] `interface` vs `abstract class`; default methods
- [ ] Method overloading vs overriding; covariant return
- [ ] `final`, `static`, khối init, thứ tự khởi tạo
- [ ] Composition over inheritance

## Contracts quan trọng

- [ ] `equals()` / `hashCode()` contract (5 quy tắc)
- [ ] `Comparable` vs `Comparator`
- [ ] `Object` methods (`toString`, `clone`, `getClass`)

## Exceptions

- [ ] Checked vs unchecked; hierarchy (`Throwable`/`Error`/`Exception`)
- [ ] try-with-resources, `AutoCloseable`
- [ ] `finally`, exception chaining, custom exception
- [ ] Best practice: fail-fast, không nuốt exception

---

# 4. Week 2 – Generics & Collections

## Generics

- [ ] Type parameter, bounded types (`<T extends Number>`)
- [ ] Wildcards: `? extends` / `? super`, quy tắc **PECS**
- [ ] Type erasure & hệ quả (không `new T()`, không array of generic)
- [ ] Generic methods, generic class

## Collections Framework

- [ ] Hierarchy: `Collection`, `List`, `Set`, `Queue`, `Map`
- [ ] `ArrayList` vs `LinkedList` (khi nào dùng gì)
- [ ] `HashMap` internals: bucket, hash, treeify (Java 8+), resize
- [ ] `HashSet` / `LinkedHashMap` / `TreeMap` (sorted)
- [ ] `ConcurrentHashMap` vs `Collections.synchronizedMap` vs `Hashtable`
- [ ] Fail-fast vs fail-safe iterator, `ConcurrentModificationException`
- [ ] `equals`/`hashCode` ảnh hưởng thế nào tới HashMap

---

# 5. Week 3 – Functional & Streams API

## Functional

- [ ] Lambda expression, capture biến (effectively final)
- [ ] Functional interfaces: `Function`, `Supplier`, `Consumer`, `Predicate`, `BiFunction`
- [ ] Method reference (4 loại)
- [ ] `@FunctionalInterface`

## Stream API

- [ ] Intermediate vs terminal operations; lazy evaluation
- [ ] `map`/`filter`/`reduce`/`collect`
- [ ] `Collectors` (toList, groupingBy, joining, partitioningBy)
- [ ] `Optional` (dùng đúng cách, tránh `.get()`)
- [ ] Parallel stream (khi nào nên/không nên)
- [ ] Stream vs loop (readability vs performance)

---

# 6. Week 4 – Concurrency & Multithreading

## Cơ bản

- [ ] Thread lifecycle; `Runnable` vs `Callable` vs `Future`
- [ ] `Thread` vs `ExecutorService`; các loại thread pool
- [ ] `CompletableFuture` (async pipeline, combine, exception handling)

## Đồng bộ hoá

- [ ] `synchronized` (method/block), intrinsic lock/monitor
- [ ] `volatile` (visibility, không atomic)
- [ ] `Lock`/`ReentrantLock`, `ReadWriteLock`, `Condition`
- [ ] `Semaphore`, `CountDownLatch`, `CyclicBarrier`
- [ ] Atomic classes (`AtomicInteger`, CAS)

## Nâng cao

- [ ] Java Memory Model (JMM), happens-before
- [ ] Deadlock / livelock / starvation; cách phát hiện & tránh
- [ ] `ThreadLocal`
- [ ] Virtual threads (Java 21, Project Loom) — biết khái niệm
- [ ] Concurrent collections

---

# 7. Week 5 – JVM, Memory & Garbage Collection

## JVM architecture

- [ ] Class loading: loading → linking → initialization
- [ ] ClassLoader hierarchy (bootstrap/platform/application), delegation
- [ ] Runtime data areas: heap, stack, metaspace, PC register
- [ ] JIT compilation (C1/C2), interpretation vs compilation

## Memory & GC

- [ ] Heap layout: young (eden/survivor) + old gen
- [ ] Object lifecycle, minor vs major/full GC
- [ ] GC algorithms: Serial, Parallel, **G1**, **ZGC**, Shenandoah
- [ ] GC roots, reachability, reference types (strong/soft/weak/phantom)
- [ ] Memory leak trong Java (dù có GC): static ref, listener, ThreadLocal, ClassLoader leak
- [ ] Tuning flags cơ bản (`-Xms`/`-Xmx`, GC selection)

## Troubleshooting

- [ ] `OutOfMemoryError` các loại (heap, metaspace, GC overhead)
- [ ] Heap dump, thread dump, tools (jstat, jmap, jstack, VisualVM, async-profiler)

---

# 8. Week 6 – Modern Java & IO/NIO

## Modern Java (11 → 21)

- [ ] `var` (local type inference)
- [ ] **Records** (immutable data carrier)
- [ ] **Sealed classes/interfaces** (permits)
- [ ] **Pattern matching** (instanceof, switch)
- [ ] **Switch expressions** (yield, arrow)
- [ ] Text blocks
- [ ] `Optional`, `List.of()`/`Map.of()` (immutable factories)

## IO / NIO

- [ ] Byte stream vs char stream; buffered streams
- [ ] `try-with-resources` cho IO
- [ ] NIO: `Path`, `Files`, channel/buffer
- [ ] Serialization (và vì sao nên tránh / thay bằng JSON)

---

# 9. Week 7 – Build, Testing & Ecosystem

## Build tools

- [ ] Maven: lifecycle, `pom.xml`, dependency scope, transitive deps
- [ ] Gradle: task-based, Groovy/Kotlin DSL, so với Maven
- [ ] Dependency management, version conflict resolution

## Testing

- [ ] JUnit 5 (Jupiter): `@Test`, lifecycle, assertions, parameterized
- [ ] Mockito: mock/stub/verify, `@Mock`/`@InjectMocks`
- [ ] Test structure (AAA), coverage, integration test

## Persistence & ecosystem

- [ ] JDBC cơ bản, connection pool (HikariCP)
- [ ] JPA/Hibernate: entity, lazy vs eager, N+1, transaction
- [ ] Logging: SLF4J + Logback/Log4j2
- [ ] JSON: Jackson (serialize/deserialize)

---

# 10. Week 8 – Spring, System Design & Mock

## Spring / Spring Boot

- [ ] IoC container, Dependency Injection (constructor injection)
- [ ] Bean lifecycle, scope, `@Component`/`@Service`/`@Repository`
- [ ] Spring Boot auto-configuration, starter
- [ ] REST controller, `@RestController`, exception handling (`@ControllerAdvice`)
- [ ] Transaction (`@Transactional`), propagation, isolation
- [ ] Spring AOP cơ bản

## System Design

- [ ] REST API design, versioning, idempotency
- [ ] Microservices: service discovery, config, resilience (circuit breaker)
- [ ] Caching, message queue (Kafka/RabbitMQ)
- [ ] Scalability, load balancing, observability

## Mock

- [ ] Coding round (data structures, concurrency task)
- [ ] JVM/GC deep dive Q&A
- [ ] System design round
- [ ] Behavioral / senior discussion

---

# 11. Question Bank

## Core & OOP

- [ ] `==` vs `equals()` khác nhau thế nào?
- [ ] equals/hashCode contract? Vi phạm thì sao?
- [ ] Vì sao String immutable? Lợi ích?
- [ ] String pool hoạt động thế nào? `new String("x")` khác literal?
- [ ] interface vs abstract class — khi nào dùng?
- [ ] Checked vs unchecked exception?
- [ ] `final` vs `finally` vs `finalize`?

## Generics & Collections

- [ ] Type erasure là gì? Hệ quả?
- [ ] PECS — giải thích?
- [ ] HashMap hoạt động bên trong thế nào? Treeify khi nào?
- [ ] HashMap vs ConcurrentHashMap vs Hashtable?
- [ ] Fail-fast vs fail-safe iterator?
- [ ] ArrayList vs LinkedList?

## Functional & Streams

- [ ] Lambda capture biến — vì sao phải effectively final?
- [ ] Intermediate vs terminal operation? Lazy nghĩa là gì?
- [ ] Khi nào KHÔNG nên dùng parallel stream?
- [ ] Optional dùng sai như thế nào?

## Concurrency

- [ ] `volatile` giải quyết gì? Có làm atomic không?
- [ ] `synchronized` vs `ReentrantLock`?
- [ ] happens-before là gì?
- [ ] `Runnable` vs `Callable`?
- [ ] CompletableFuture xử lý exception thế nào?
- [ ] Virtual thread vs platform thread?
- [ ] ConcurrentHashMap tránh khoá toàn bộ map thế nào?

## JVM & GC

- [ ] Class loading các giai đoạn? Delegation model?
- [ ] Stack vs heap lưu gì?
- [ ] GC roots là gì? Object nào bị thu hồi?
- [ ] G1 vs ZGC?
- [ ] Java có GC sao vẫn memory leak? Ví dụ?
- [ ] Các loại OutOfMemoryError?
- [ ] Strong/soft/weak/phantom reference?

## Modern Java & Ecosystem

- [ ] Record khác class thường thế nào?
- [ ] Sealed class dùng để làm gì?
- [ ] Maven dependency scope?
- [ ] Hibernate N+1 problem và cách fix?
- [ ] `@Transactional` propagation?
- [ ] Constructor injection vs field injection — vì sao nên constructor?

---

# 12. Final Preparation Checklist

- [ ] equals/hashCode contract + immutability
- [ ] Generics wildcard + PECS + type erasure
- [ ] HashMap / ConcurrentHashMap internals
- [ ] Stream API + Optional idiomatic
- [ ] Concurrency: executor, CompletableFuture, JMM, locks
- [ ] JVM class loading + memory areas
- [ ] GC algorithms + tuning + leak detection
- [ ] Modern Java: records/sealed/pattern matching
- [ ] Build (Maven/Gradle) + JUnit5/Mockito
- [ ] JPA/Hibernate + N+1
- [ ] Spring IoC/DI/Boot/REST/@Transactional
- [ ] System design + microservices
- [ ] Mock interview đủ 4 vòng

---

# 13. Recommended Interview Mindset

- Trả lời có **cấu trúc**: khái niệm → cơ chế bên dưới → trade-off → ví dụ thực tế.
- Với câu JVM/GC/concurrency: luôn nói được "vì sao" chứ không chỉ "là gì".
- Chủ động nêu edge case và cách bạn kiểm chứng (profiler, benchmark).
- Thành thật khi không chắc; nêu cách bạn sẽ tìm hiểu.
