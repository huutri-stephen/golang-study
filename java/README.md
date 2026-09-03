# Senior Java Interview Preparation

Bộ tài liệu ôn tập phỏng vấn **Senior Backend Engineer – Java**, xây dựng theo
[roadmap.sh/java](https://roadmap.sh/java) và tham khảo cấu trúc của bộ tài liệu Golang trong repo.

> Nội dung roadmap được diễn giải lại cho phù hợp license.

## Mục tiêu

Ôn tập có hệ thống trong 8 tuần: Java core, generics/collections, functional/streams, concurrency,
JVM/memory/GC, modern Java, build/testing/ecosystem, và Spring/system design/mock. Trọng tâm mức
senior: hiểu **cơ chế bên dưới** (JVM, memory model, collection internals) và **trade-off**, không chỉ cú pháp.

## Cấu trúc thư mục

```
java/
├── README.md                              # file này
├── senior_java_interview_plan.md          # master plan (roadmap, question bank, final checklist)
├── checklist/
│   └── weekly_checklist.md                # tracking tiến độ theo tuần
├── week1-java-core-oop/                    # ==/equals, String pool, OOP, equals/hashCode, exceptions
│   ├── notes.md
│   ├── flashcards.md
│   └── code_examples/
│       ├── StringImmutabilityDemo.java
│       └── EqualsHashCodeDemo.java
├── week2-generics-collections/             # generics/PECS/type erasure, HashMap internals
│   ├── notes.md, flashcards.md
│   └── code_examples/{CollectionsDemo, PecsDemo}.java
├── week3-functional-streams/               # lambda, Stream API, Optional, lazy evaluation
│   ├── notes.md, flashcards.md
│   └── code_examples/StreamsDemo.java
├── week4-concurrency/                      # executor, CompletableFuture, JMM, locks, virtual threads
│   ├── notes.md, flashcards.md
│   └── code_examples/{RaceConditionDemo, CompletableFutureDemo}.java
├── week5-jvm-memory-gc/                     # class loading, heap, GC, memory leak, troubleshooting
│   ├── notes.md, flashcards.md
│   └── code_examples/MemoryLeakDemo.java
├── week6-modern-java-io/                    # records, sealed, pattern matching, switch expr, IO/NIO
│   ├── notes.md, flashcards.md
│   └── code_examples/ModernJavaDemo.java
├── week7-build-testing-ecosystem/          # Maven/Gradle, JUnit5, Mockito, JPA/N+1, logging
│   ├── notes.md, flashcards.md
│   └── code_examples/TestingConceptsDemo.java
├── week8-spring-system-design-mock/        # Spring IoC/DI/Boot/REST, system design, mock interview
│   ├── notes.md, flashcards.md
│   └── mock_scenarios.md
└── html/                                   # study hub web (deep-dive, dark theme)
    ├── index.html
    ├── week1.html ... week8.html
    └── assets/{style.css, app.js}
```

## Cách dùng

1. Đọc `senior_java_interview_plan.md` để nắm roadmap tổng quan và tự chấm question bank.
2. Mỗi tuần: đọc `notes.md` → chạy `code_examples/` → tự kiểm bằng `flashcards.md`.
3. Đánh dấu tiến độ trong `checklist/weekly_checklist.md`.
4. Mở `html/index.html` trên trình duyệt để học dạng web deep-dive (flashcard bấm để mở).
5. Tuần 8: luyện `mock_scenarios.md` theo 4 vòng.

## Chạy code examples

Môi trường: **JDK 17** (LTS). Không cần Maven/Gradle — dùng single-file source mode:

```bash
cd week1-java-core-oop/code_examples
java StringImmutabilityDemo.java

# Ví dụ cần JVM flag (Week 5):
cd ../../week5-jvm-memory-gc/code_examples
java -Xmx64m MemoryLeakDemo.java leak      # minh hoạ memory leak -> OOM
java -Xmx64m MemoryLeakDemo.java noleak    # heap ổn định, GC dọn được
```

Ghi chú: switch **pattern matching** là preview ở JDK 17 (GA từ JDK 21); code example dùng
`instanceof` pattern (đã GA ở JDK 16) để chạy được không cần `--enable-preview`.
Các file test/ecosystem (JUnit/Mockito, Spring) được minh hoạ khái niệm bằng stdlib thuần
vì môi trường không có build tool; thực tế dùng Maven/Gradle + các thư viện tương ứng.
