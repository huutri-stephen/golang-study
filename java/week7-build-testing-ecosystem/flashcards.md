# Week 7 – Flashcards / Q&A (Deep Dive)

## Maven / Gradle

### Q: Maven resolve version conflict thế nào?
**A:** Nearest-wins: chọn version ở độ sâu nhỏ nhất trong cây dependency; cùng độ sâu thì khai báo trước thắng. Ép version bằng dependencyManagement hoặc import BOM. Khác Gradle (highest-version wins).

---

### Q: Dependency scope của Maven?
**A:** compile (mọi nơi, mặc định), provided (compile+test, không đóng gói - servlet API), runtime (không cần compile - JDBC driver), test (chỉ test - JUnit/Mockito). Scope quyết định classpath ở từng giai đoạn.

---

### Q: Gradle nhanh hơn Maven nhờ gì?
**A:** Task-based DAG chỉ chạy task cần thiết; incremental build (input không đổi → UP-TO-DATE, bỏ qua); build cache tái dùng output giữa máy/CI; daemon. Đổi lại script phức tạp hơn XML khai báo của Maven.

---

## JUnit 5

### Q: JUnit 5 gồm mấy phần?
**A:** Platform (nền chạy test, launcher), Jupiter (API viết test mới + extension), Vintage (chạy test JUnit 3/4 cũ).

---

### Q: JUnit 5 tạo bao nhiêu instance test?
**A:** Mặc định PER_METHOD: mỗi test method một instance mới → cô lập trạng thái. @TestInstance(PER_CLASS) dùng một instance cho cả class (cho phép @BeforeAll non-static).

---

### Q: Extension model của JUnit 5 thay gì của JUnit 4?
**A:** Thay Runner + Rule bằng Extension interface (BeforeEachCallback, ParameterResolver...). @ExtendWith(MockitoExtension.class) cắm Mockito, @ExtendWith(SpringExtension.class) cho Spring Test. assertAll gom nhiều assert báo tất cả lỗi.

---

## Mockito

### Q: Mockito tạo mock bằng cách nào?
**A:** Dùng ByteBuddy (trước CGLIB) sinh subclass runtime của class được mock, override method để trả stub/ghi tương tác. Vì thế không mock được final class/method (không override được) trừ khi dùng mockito-inline.

---

### Q: @Mock vs @InjectMocks vs @Spy?
**A:** @Mock tạo mock rỗng. @InjectMocks tạo instance thật của class test và tự inject các @Mock (constructor→setter→field). @Spy bọc object thật, chỉ stub method chỉ định, còn lại chạy thật.

---

### Q: Vì sao không mock được final method?
**A:** Mockito tạo subclass và override; final method không override được. Cần mockito-inline (bytecode instrumentation) hoặc thiết kế test qua interface. Static/private cũng cần mockito-inline (mockStatic).

---

## JDBC / HikariCP

### Q: Vì sao luôn dùng connection pool?
**A:** Tạo connection DB tốn kém (TCP + TLS + auth + session DB, hàng ms). Pool giữ sẵn connection mở tái sử dụng → giảm latency. HikariCP nhanh nhờ ConcurrentBag lock-free.

---

### Q: Pool lớn có luôn tốt hơn không?
**A:** Không. Nhiều connection hơn số DB xử lý được → context switch + lock contention ở DB → chậm hơn. Pool nhỏ (10-20) thường tốt hơn. Tham khảo: connections ≈ core*2 + spindle. Đo thực nghiệm.

---

### Q: maxLifetime của pool để làm gì?
**A:** Tái tạo connection sau một khoảng thời gian, tránh dùng connection cũ đã bị DB/load balancer đóng ngầm (gây "connection reset"). Nên đặt nhỏ hơn timeout của DB/LB.

---

## Hibernate

### Q: Persistence context là gì?
**A:** Bộ nhớ đệm entity managed trong một transaction (L1 cache, luôn bật). Query cùng entity 2 lần trong 1 transaction → lần 2 lấy từ context, không query DB lại. EntityManager/Session quản lý nó.

---

### Q: Dirty checking hoạt động thế nào?
**A:** Hibernate chụp snapshot entity managed lúc load. Lúc flush/commit, so sánh state hiện tại với snapshot, tự sinh UPDATE cho field đã đổi. Nên chỉ cần setEntity.setField() trong transaction, KHÔNG cần gọi save().

---

### Q: Vì sao set field entity ngoài transaction không update DB?
**A:** Ngoài transaction entity ở trạng thái detached, không được persistence context theo dõi → không dirty checking → không sinh UPDATE. Phải merge lại vào một transaction.

---

### Q: flush vs commit?
**A:** flush đồng bộ persistence context xuống DB (chạy SQL) nhưng chưa commit transaction; xảy ra tự động trước query và lúc commit. commit = flush + kết thúc transaction (DB commit thật).

---

### Q: L1 vs L2 cache?
**A:** L1 = persistence context, per-transaction, luôn bật. L2 = second-level cache (@Cacheable + Ehcache/Caffeine), chia sẻ giữa các session/transaction, phải cấu hình, cẩn thận invalidation.

---

## N+1 & fetch

### Q: N+1 problem và cách fix?
**A:** 1 query lấy N entity + N query lazy lấy association từng cái = 1+N query. Fix: JOIN FETCH (JPQL), @EntityGraph (khai báo đồ thị fetch), @BatchSize (gom thành query IN). Không dùng EAGER bừa.

---

### Q: Fetch type mặc định của các association?
**A:** @ManyToOne và @OneToOne mặc định EAGER. @OneToMany và @ManyToMany (collection) mặc định LAZY. LAZY trả proxy; truy cập sau khi session đóng → LazyInitializationException.

---

### Q: LazyInitializationException xảy ra khi nào?
**A:** Truy cập một lazy association sau khi persistence context (session/transaction) đã đóng → proxy không thể load. Fix: fetch trong transaction (JOIN FETCH/@EntityGraph) hoặc map sang DTO trong transaction.

---

## Transaction & JSON

### Q: Bẫy @Transactional self-invocation?
**A:** Gọi method @Transactional từ method khác TRONG CÙNG class → không qua proxy Spring → annotation không có tác dụng (không mở transaction). Lỗi phổ biến nhất với Spring AOP. Tách sang bean khác hoặc self-inject.

---

### Q: Propagation REQUIRED vs REQUIRES_NEW?
**A:** REQUIRED (mặc định): join transaction hiện có hoặc tạo mới. REQUIRES_NEW: luôn tạo transaction mới, suspend cái đang chạy (dùng cho audit log cần commit độc lập dù transaction ngoài rollback).

---

### Q: Vì sao không serialize entity Hibernate ra JSON trực tiếp?
**A:** Lazy association là proxy → serialize gây LazyInitializationException hoặc tải dư; circular reference (2 entity trỏ nhau) gây StackOverflow. Nên map sang DTO. Dùng @JsonIgnore/@JsonManagedReference nếu buộc.

---

### Q: Vì sao log parameterized `log.info("x={}", x)` tốt hơn nối chuỗi?
**A:** Chỉ format/nối chuỗi khi log level được bật. `log.debug("..." + heavy())` luôn tính heavy() dù debug tắt. Parameterized tránh chi phí thừa và an toàn hơn.
