# Week 8 – Flashcards / Q&A (Deep Dive)

## IoC & Bean lifecycle

### Q: Bean lifecycle của Spring theo thứ tự nào?
**A:** Instantiate (constructor) → populate properties (DI) → Aware callbacks → BeanPostProcessor.before → @PostConstruct/afterPropertiesSet → BeanPostProcessor.after (AOP PROXY tạo ở đây) → sẵn sàng → @PreDestroy/destroy khi context đóng.

---

### Q: Proxy AOP được tạo ở bước nào? Hệ quả?
**A:** Ở postProcessAfterInitialization (BeanPostProcessor). Hệ quả: self-invocation (this.method()) bỏ qua proxy vì gọi thẳng object thật → @Transactional/@Async trong cùng bean không hoạt động.

---

### Q: BeanPostProcessor là gì?
**A:** Hook mở rộng mạnh nhất của Spring: chạy before/after mỗi bean init. AOP proxy, @Autowired resolution đều là BeanPostProcessor. Cho phép sửa/bọc bean trước khi dùng.

---

## Dependency Injection

### Q: Vì sao constructor injection tốt hơn field injection?
**A:** Cho phép field final (bất biến), dependency rõ ràng và bắt buộc (compiler ép), dễ test đơn vị (new + mock, không cần Spring/reflection), và phát hiện circular dependency sớm lúc khởi động.

---

### Q: Circular dependency: constructor vs field injection?
**A:** Constructor injection cả hai chiều → Spring không giải được → BeanCurrentlyInCreationException (fail fast, tốt). Field/setter singleton → giải được nhờ early reference (tiêm instance chưa init xong) → che giấu vấn đề thiết kế.

---

## Scope

### Q: Bean scope mặc định? Điều kiện?
**A:** Singleton — một instance trên toàn container, chia sẻ cho mọi nơi. Vì thế bean singleton PHẢI stateless/thread-safe. Scope khác: prototype (mỗi lần lấy tạo mới), request, session.

---

## AOP & Proxy

### Q: JDK dynamic proxy vs CGLIB?
**A:** JDK proxy implement interface (Proxy class). CGLIB sinh subclass của class (không cần interface, nhưng không proxy được final/private). Spring Boot mặc định CGLIB (proxyTargetClass=true).

---

### Q: Vì sao @Transactional self-invocation không hoạt động?
**A:** Spring AOP dùng proxy bọc ngoài; proxy chỉ chặn lời gọi TỪ NGOÀI. Gọi this.method() trong cùng bean đi thẳng object thật, bỏ qua proxy → advice (transaction) không chạy. Fix: tách bean khác hoặc self-inject.

---

### Q: @Transactional trên private method có tác dụng không?
**A:** Không. Proxy (JDK interface hoặc CGLIB subclass) chỉ chặn được public method gọi từ ngoài. private/final/non-public method không bị proxy chặn → annotation vô hiệu.

---

## @Transactional

### Q: @Transactional rollback khi nào?
**A:** Mặc định rollback với RuntimeException và Error; commit khi thành công. Checked exception KHÔNG rollback (cần rollbackFor). Bẫy: nuốt exception bên trong → transaction không thấy lỗi → commit.

---

### Q: REQUIRED vs REQUIRES_NEW?
**A:** REQUIRED (mặc định): join transaction hiện có hoặc tạo mới. REQUIRES_NEW: suspend transaction ngoài, tạo transaction mới độc lập → commit dù transaction ngoài rollback (dùng cho audit log).

---

## Auto-configuration

### Q: Spring Boot auto-configuration hoạt động thế nào?
**A:** @EnableAutoConfiguration nạp danh sách config class (từ AutoConfiguration.imports). Mỗi class dùng @Conditional: @ConditionalOnClass (có class trên classpath), @ConditionalOnMissingBean (chỉ tạo nếu user chưa định nghĩa), @ConditionalOnProperty. Classpath có gì thì kích hoạt cái đó.

---

### Q: Làm sao override một auto-configured bean?
**A:** Định nghĩa bean cùng kiểu trong config của bạn. Auto-config dùng @ConditionalOnMissingBean nên sẽ nhường (không tạo bean mặc định nữa).

---

### Q: @SpringBootApplication gồm gì?
**A:** @Configuration + @ComponentScan + @EnableAutoConfiguration. Quét component trong package, bật auto-config, và bản thân là config class.

---

## System Design

### Q: Circuit breaker giải quyết gì?
**A:** Khi downstream lỗi liên tục, breaker "mở" → fail nhanh không chờ timeout, thử lại sau (half-open). Tránh cascade failure và cạn tài nguyên (thread/connection chờ service chết). Dùng Resilience4j.

---

### Q: Đồng bộ (REST/gRPC) vs bất đồng bộ (queue) khi nào?
**A:** REST/gRPC khi cần phản hồi ngay, coupling chấp nhận được. Queue (Kafka/RabbitMQ) khi cần decoupling, chịu tải đỉnh (buffer), fan-out sự kiện, xử lý async — đổi lại phức tạp hơn và eventual consistency.

---

### Q: Làm service scale ngang thế nào?
**A:** Giữ stateless (đẩy state ra Redis/DB, không session cục bộ), đặt sau load balancer, DB read replica/sharding + pool đúng size, tác vụ nặng đẩy queue xử lý async, observability đầy đủ.

---

### Q: Idempotency key dùng để làm gì?
**A:** Client gửi key duy nhất cho mỗi thao tác (vd payment). Server lưu key + kết quả; nếu nhận lại key đã xử lý → trả kết quả cũ thay vì thực thi lại → an toàn khi client retry (tránh double charge).

---

### Q: Cấu trúc trả lời một bài system design?
**A:** Làm rõ yêu cầu & scale (QPS, data size) → định nghĩa API → data model → high-level architecture → đi sâu bottleneck (DB/cache/queue) → bàn trade-off. Không nhảy vào chi tiết trước khi chốt phạm vi.
