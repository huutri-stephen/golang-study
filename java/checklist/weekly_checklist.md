# ✅ Weekly Progress Checklist – Senior Java

## Week 1 – Java Core & OOP

### Types & Memory
- [ ] Primitive vs reference, autoboxing
- [ ] Stack vs heap
- [ ] `==` vs `.equals()`
- [ ] String pool, immutability
- [ ] String vs StringBuilder vs StringBuffer
- [ ] Integer cache (-128..127)

### OOP
- [ ] 4 trụ cột OOP
- [ ] interface vs abstract class
- [ ] overloading vs overriding
- [ ] composition over inheritance
- [ ] thứ tự khởi tạo

### Contracts & Exceptions
- [ ] equals/hashCode contract (5 quy tắc)
- [ ] Comparable vs Comparator
- [ ] checked vs unchecked
- [ ] try-with-resources

### Code Examples
- [ ] StringImmutabilityDemo.java
- [ ] EqualsHashCodeDemo.java

---

## Week 2 – Generics & Collections

- [ ] Bounded types, wildcards
- [ ] PECS (Producer Extends, Consumer Super)
- [ ] Type erasure + hệ quả
- [ ] ArrayList vs LinkedList
- [ ] HashMap internals (bucket, treeify, resize)
- [ ] HashMap vs ConcurrentHashMap vs Hashtable
- [ ] Fail-fast vs fail-safe iterator
- [ ] TreeMap / LinkedHashMap

### Code Examples
- [ ] CollectionsDemo.java
- [ ] PecsDemo.java

---

## Week 3 – Functional & Streams

- [ ] Lambda + effectively final
- [ ] Functional interfaces (Function/Supplier/Consumer/Predicate)
- [ ] Method reference (4 loại)
- [ ] Intermediate vs terminal + lazy evaluation
- [ ] Collectors (groupingBy, partitioningBy, joining)
- [ ] Optional đúng cách
- [ ] Parallel stream (khi nào không dùng)

### Code Examples
- [ ] StreamsDemo.java

---

## Week 4 – Concurrency & Multithreading

- [ ] Runnable vs Callable vs Future
- [ ] ExecutorService + thread pools
- [ ] CompletableFuture (pipeline, combine, exception)
- [ ] volatile (visibility) vs atomicity
- [ ] synchronized vs ReentrantLock
- [ ] happens-before (JMM)
- [ ] Atomic + CAS
- [ ] Semaphore/CountDownLatch/CyclicBarrier
- [ ] Deadlock + cách tránh
- [ ] ThreadLocal + bẫy trong pool
- [ ] Virtual threads (Java 21)

### Code Examples
- [ ] RaceConditionDemo.java
- [ ] CompletableFutureDemo.java

---

## Week 5 – JVM, Memory & GC

- [ ] Class loading 3 giai đoạn
- [ ] Parent delegation model
- [ ] JIT (C1/C2)
- [ ] Stack/heap/metaspace
- [ ] Heap layout (young/old, minor/major GC)
- [ ] GC roots + reachability
- [ ] GC collectors (G1, ZGC)
- [ ] Reference types (strong/soft/weak/phantom)
- [ ] Memory leak dù có GC
- [ ] Các loại OutOfMemoryError
- [ ] Công cụ (jstat/jmap/jstack/MAT)

### Code Examples
- [ ] MemoryLeakDemo.java (chạy `-Xmx64m ... leak` và `... noleak`)

---

## Week 6 – Modern Java & IO

- [ ] var (local type inference)
- [ ] Records + compact constructor
- [ ] Sealed classes
- [ ] Pattern matching (instanceof, switch)
- [ ] Switch expression + yield
- [ ] Text blocks
- [ ] Immutable factories (List.of...)
- [ ] Byte vs char stream, buffered
- [ ] NIO (Path/Files)
- [ ] Vì sao tránh Java Serialization

### Code Examples
- [ ] ModernJavaDemo.java

---

## Week 7 – Build, Testing & Ecosystem

- [ ] Maven lifecycle + dependency scope
- [ ] Maven vs Gradle
- [ ] JUnit 5 (lifecycle, parameterized, assertThrows)
- [ ] AAA pattern
- [ ] Mockito (mock/stub/verify, @Mock/@InjectMocks)
- [ ] Connection pool (HikariCP)
- [ ] JPA/Hibernate lazy vs eager
- [ ] N+1 problem + fix
- [ ] @Transactional rollback
- [ ] SLF4J + Logback, Jackson

### Code Examples
- [ ] TestingConceptsDemo.java

---

## Week 8 – Spring, System Design & Mock

- [ ] IoC & DI (constructor injection)
- [ ] Bean scope & lifecycle
- [ ] Spring Boot auto-config + starter
- [ ] REST controller + @RestControllerAdvice
- [ ] @Transactional (propagation, self-invocation bẫy)
- [ ] Spring AOP
- [ ] REST API design + microservices + resilience
- [ ] Caching, message queue, observability
- [ ] Mock vòng 1: Coding
- [ ] Mock vòng 2: JVM deep dive
- [ ] Mock vòng 3: System design
- [ ] Mock vòng 4: Behavioral

### Materials
- [ ] mock_scenarios.md (4 vòng)
