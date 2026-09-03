# Week 7 – Build, Testing & Ecosystem – Deep Dive Notes

> Kỹ năng "làm việc thực tế" nhưng đi sâu: **Maven dependency resolution (nearest-wins, BOM)**,
> **JUnit 5 extension model**, **Mockito dùng proxy/bytecode gì**, và trọng tâm phỏng vấn backend:
> **persistence context, dirty checking, N+1, fetch strategy** của Hibernate.

## Mục lục
1. [Maven: lifecycle, scope, dependency resolution](#1-maven)
2. [Gradle: task graph, incremental, cache](#2-gradle)
3. [JUnit 5 architecture & extension model](#3-junit-5)
4. [Mockito internals](#4-mockito-internals)
5. [JDBC & HikariCP](#5-jdbc--hikaricp)
6. [Hibernate: persistence context & dirty checking](#6-hibernate-persistence-context)
7. [N+1 & fetch strategies](#7-n1--fetch-strategies)
8. [Transaction & isolation trong JPA](#8-transaction--isolation)
9. [Logging & JSON](#9-logging--json)
10. [Follow-up questions](#10-follow-up-questions)

---

## 1. Maven

### Lifecycle & phase

Ba lifecycle: `clean`, `default`, `site`. Default lifecycle (rút gọn):
`validate → compile → test → package → verify → install → deploy`. Chạy phase nào thì mọi phase **trước** nó chạy theo. Mỗi phase gắn các **goal** của plugin.

### Dependency scope

| Scope | compile | test | runtime | đóng gói | Ví dụ |
|---|---|---|---|---|---|
| `compile` (default) | ✓ | ✓ | ✓ | ✓ | thư viện chính |
| `provided` | ✓ | ✓ | ✗ | ✗ | servlet API (container cấp) |
| `runtime` | ✗ | ✓ | ✓ | ✓ | JDBC driver |
| `test` | ✗ | ✓ | ✗ | ✗ | JUnit, Mockito |

### Dependency resolution: nearest-wins

- Dependency **transitive** (dep của dep) tự kéo theo. Khi hai đường dẫn cùng thư viện khác version → Maven chọn **"nearest wins"**: version ở **độ sâu nhỏ nhất** trong cây; cùng độ sâu thì **khai báo trước thắng**.
- Khác Gradle (mặc định chọn **version cao nhất**). Đây là nguồn "diamond dependency" khó chịu.
- `mvn dependency:tree` xem cây; `<dependencyManagement>` để **ép version** thống nhất (không tự thêm dependency, chỉ khai báo version). **BOM** (Bill of Materials, import scope) chuẩn hoá version một hệ (vd `spring-boot-dependencies`).

> **Follow-up:** "Có version conflict, làm sao ép?" → dependencyManagement hoặc exclude transitive + khai báo version mong muốn tường minh.

---

## 2. Gradle

- **Task-based**: build là một **DAG các task**; Gradle chỉ chạy task cần thiết.
- **Incremental build**: task có input/output; nếu input không đổi → **UP-TO-DATE** (bỏ qua). **Build cache** tái dùng output giữa các máy/CI.
- Resolution mặc định: **highest version wins** (khác Maven nearest-wins).
- Nhanh hơn Maven cho project lớn nhờ incremental + cache + daemon; đổi lại script phức tạp hơn.

---

## 3. JUnit 5

### Kiến trúc 3 phần

- **Platform** — nền chạy test (launcher, tích hợp IDE/build tool).
- **Jupiter** — API viết test mới (`@Test`, `@ParameterizedTest`, extension).
- **Vintage** — chạy test JUnit 3/4 cũ.

### Lifecycle

```java
@BeforeAll static void once() {}       // 1 lần trước tất cả (static, trừ khi PER_CLASS)
@BeforeEach void before() {}           // trước mỗi test
@Test void t() {}
@AfterEach void after() {}
@AfterAll static void done() {}
```

Mặc định JUnit 5 tạo **instance mới cho mỗi test method** (`PER_METHOD`) → cô lập trạng thái. `@TestInstance(PER_CLASS)` cho phép `@BeforeAll` non-static.

### Extension model (thay Runner/Rule của JUnit 4)

- Điểm mở rộng qua interface: `BeforeEachCallback`, `AfterEachCallback`, `ParameterResolver`, `TestExecutionExceptionHandler`...
- `@ExtendWith(MockitoExtension.class)` cắm Mockito vào lifecycle. `@ExtendWith(SpringExtension.class)` cho Spring Test.
- Assertions: `assertEquals`, `assertThrows`, `assertAll` (nhóm — báo tất cả lỗi thay vì dừng ở lỗi đầu), `assertTimeout`.

```java
@ParameterizedTest
@CsvSource({ "100,gold,80", "100,silver,90" })
void discount(int price, String tier, int expected) {
    assertEquals(expected, Discount.apply(price, tier));
}
```

---

## 4. Mockito internals

### Cách Mockito tạo mock

- Dùng **ByteBuddy** (trước là CGLIB) sinh một **subclass** của class được mock lúc runtime, override method để trả giá trị stub / ghi lại tương tác. Vì thế:
  - **Không mock được `final` class/method** (không override được) trừ khi bật `mockito-inline` (dùng instrumentation).
  - Không mock được `static`/`private` mặc định (cần `mockito-inline` + `mockStatic`).
- `when(x).thenReturn(y)` — stub. `verify(mock).method()` — kiểm tra tương tác (số lần, thứ tự, tham số qua ArgumentMatcher).

### @Mock, @InjectMocks, @Spy

- `@Mock` tạo mock rỗng. `@InjectMocks` tạo instance thật của class test và **tự inject** các `@Mock` (ưu tiên constructor → setter → field).
- `@Spy` bọc object thật, chỉ stub method được chỉ định (còn lại chạy thật).

```java
@ExtendWith(MockitoExtension.class)
class UserServiceTest {
    @Mock UserRepo repo;
    @InjectMocks UserService service;
    @Test void t() {
        when(repo.findById(1L)).thenReturn(new User(1L, "An"));
        assertEquals("An", service.getName(1L));
        verify(repo, times(1)).findById(1L);
    }
}
```

> **Follow-up:** "Vì sao không mock được final?" → Mockito subclass + override; final không override được. Dùng mockito-inline (bytecode instrumentation) hoặc thiết kế để test qua interface.

### Test pyramid & coverage

- Nhiều **unit** (nhanh, cô lập bằng mock) > ít **integration** (thật hơn, chậm) > rất ít **e2e**.
- Coverage đo **dòng chạy qua**, không đo assertion có ý nghĩa → không thần thánh hoá con số; ưu tiên branch/edge case.

---

## 5. JDBC & HikariCP

### JDBC cơ bản

`Connection → PreparedStatement → ResultSet`. Luôn `PreparedStatement` với `?` (chống SQL injection + reuse execution plan). Đóng resource bằng try-with-resources.

### Vì sao connection pool

- Tạo connection DB **tốn kém**: TCP handshake + TLS + xác thực + cấp session ở DB (hàng ms–chục ms). Pool giữ sẵn connection mở, tái sử dụng → giảm latency đột biến.
- **HikariCP** (mặc định Spring Boot): nhanh nhất nhờ code tối giản, dùng `ConcurrentBag` (cấu trúc lock-free thread-local), tránh khoá.

### Sizing pool (điểm senior)

- Pool **quá lớn** phản tác dụng: nhiều connection hơn số core DB xử lý được → context switch + tranh chấp lock ở DB → chậm hơn.
- Công thức tham khảo (PostgreSQL): `connections ≈ (core_count * 2) + effective_spindle_count`. Thường pool nhỏ (10–20) phục vụ tốt hơn nhiều so với trực giác. Đo bằng thực nghiệm.
- Tham số quan trọng: `maximumPoolSize`, `connectionTimeout` (chờ lấy connection), `maxLifetime` (tái tạo connection cũ tránh bị DB/LB đóng ngầm), `idleTimeout`.

---

## 6. Hibernate: persistence context

> Phần quan trọng nhất của Week 7 cho backend. Interviewer đào rất sâu.

### Persistence context = first-level cache

- `EntityManager` (JPA) / `Session` (Hibernate) quản lý một **persistence context**: bộ nhớ đệm các entity đang "managed" trong một transaction.
- Truy vấn cùng entity 2 lần trong 1 transaction → lần 2 lấy từ persistence context (**không** query DB lại). Đây là **L1 cache**, luôn bật, phạm vi transaction.

### Entity lifecycle states

```
new/transient  --persist-->  managed  --detach/close-->  detached
                                │  --remove-->  removed
managed: được persistence context theo dõi (dirty checking)
```

### Dirty checking (automatic)

- Với entity **managed**, Hibernate chụp snapshot lúc load. Khi transaction **flush/commit**, nó **so sánh** state hiện tại với snapshot → tự sinh `UPDATE` cho field đã đổi. **Không cần gọi `save()`**.

```java
@Transactional
void changeEmail(long id, String email) {
    User u = em.find(User.class, id);   // managed
    u.setEmail(email);                   // chỉ set field
    // KHÔNG cần em.merge/save -> flush lúc commit tự sinh UPDATE (dirty checking)
}
```

> **Follow-up:** "Vì sao set field mà không save vẫn update DB?" → dirty checking trên entity managed lúc flush. "Vì sao ngoài transaction set không update?" → entity detached, không còn được theo dõi.

### flush vs commit

- **flush**: đồng bộ persistence context xuống DB (chạy SQL) nhưng **chưa commit**. Xảy ra tự động trước query (để thấy thay đổi) và lúc commit.
- **commit**: flush + kết thúc transaction (DB commit).

### L1 vs L2 cache

- **L1** (persistence context): per-transaction, luôn bật.
- **L2** (second-level, `@Cacheable` + provider như Ehcache/Caffeine): chia sẻ giữa các session/transaction, phải cấu hình, cẩn thận invalidation.

---

## 7. N+1 & fetch strategies

### N+1 problem

```java
List<Order> orders = em.createQuery("select o from Order o", Order.class).getResultList(); // 1 query
for (Order o : orders) o.getUser().getName();   // MỖI order 1 query lazy -> N query
// Tổng: 1 + N query
```

### Fetch types

- `@ManyToOne` / `@OneToOne`: mặc định **EAGER**.
- `@OneToMany` / `@ManyToMany`: mặc định **LAZY** (collection).
- **LAZY** trả proxy; truy cập ngoài transaction → **`LazyInitializationException`** (session đã đóng).

### Cách fix N+1

| Cách | Ghi chú |
|---|---|
| `JOIN FETCH` (JPQL) | `select o from Order o join fetch o.user` → 1 query. Cẩn thận nhân bản với nhiều collection. |
| `@EntityGraph` | khai báo đồ thị fetch, dùng lại được, sạch hơn |
| `@BatchSize(size=n)` | gom N query lazy thành `⌈N/n⌉` query `IN (...)` |
| `hibernate.default_batch_fetch_size` | batch toàn cục |

> Không dùng EAGER bừa để "fix" N+1 → gây tải dư và Cartesian product khi nhiều association.

---

## 8. Transaction & isolation

- `@Transactional` (Spring) bọc method trong transaction; rollback mặc định với **RuntimeException**, commit khi thành công (checked exception cần `rollbackFor`).
- **Propagation**: `REQUIRED` (mặc định, join hoặc tạo mới), `REQUIRES_NEW` (luôn transaction mới, suspend cái cũ), `NESTED` (savepoint)...
- **Isolation**: `READ_COMMITTED` (mặc định đa số DB), `REPEATABLE_READ`, `SERIALIZABLE` → đánh đổi consistency vs concurrency (dirty read, non-repeatable read, phantom).
- **Bẫy self-invocation**: gọi method `@Transactional` từ method khác **trong cùng class** → không qua proxy → annotation **không có tác dụng**. Đây là lỗi phổ biến nhất với Spring AOP.

---

## 9. Logging & JSON

### Logging

- **SLF4J** là **facade** (API); binding tới impl runtime: **Logback** (mặc định Spring Boot), **Log4j2**.
- Log **parameterized**: `log.info("user {} did {}", userId, action)` — chỉ nối chuỗi khi level được bật (tránh chi phí thừa). Không `log.info("... " + heavyToString())`.
- Level: TRACE < DEBUG < INFO < WARN < ERROR. Production thường INFO.

### JSON (Jackson)

- `ObjectMapper` (thread-safe, tái dùng singleton). Annotation: `@JsonProperty`, `@JsonIgnore`, `@JsonInclude(NON_NULL)`, `@JsonFormat`.
- Bẫy: circular reference (2 entity trỏ nhau) → StackOverflow; dùng `@JsonManagedReference`/`@JsonBackReference` hoặc DTO. **Không serialize entity Hibernate trực tiếp** (lazy proxy → lỗi/tải dư) → map sang **DTO**.

---

## 10. Follow-up questions

1. Maven resolve version conflict thế nào? → nearest-wins (độ sâu nhỏ nhất, rồi khai báo trước); ép bằng dependencyManagement/BOM.
2. Maven vs Gradle resolution? → Maven nearest-wins; Gradle highest-version.
3. JUnit 5 tạo mấy instance test? → mặc định PER_METHOD (mỗi test 1 instance, cô lập).
4. Mockito tạo mock bằng gì? Vì sao không mock final? → ByteBuddy subclass + override; final không override được (cần mockito-inline).
5. Vì sao pool nhỏ đôi khi tốt hơn pool lớn? → nhiều connection hơn DB xử lý được gây context switch/lock contention.
6. Persistence context là gì? → L1 cache per-transaction, dirty checking tự sinh UPDATE.
7. Vì sao set field entity mà không save vẫn update? → dirty checking trên entity managed lúc flush.
8. LazyInitializationException khi nào? → truy cập lazy association sau khi session/transaction đóng.
9. N+1 và cách fix? → 1 query chính + N lazy; fix JOIN FETCH/@EntityGraph/@BatchSize.
10. Bẫy @Transactional self-invocation? → gọi trong cùng class bỏ qua proxy → không có transaction.
11. Vì sao không serialize entity Hibernate ra JSON? → lazy proxy + circular reference; dùng DTO.
