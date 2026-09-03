# Week 8 – Spring, System Design & Mock – Deep Dive Notes

> Tuần tổng hợp đi sâu: **Spring container khởi tạo bean thế nào**, **proxy (JDK dynamic vs CGLIB)**
> đứng sau `@Transactional`/AOP, **auto-configuration hoạt động ra sao**, và khung **system design**.

## Mục lục
1. [IoC container & bean lifecycle](#1-ioc-container--bean-lifecycle)
2. [Dependency injection: cơ chế & chọn kiểu](#2-dependency-injection)
3. [Bean scope & circular dependency](#3-bean-scope--circular-dependency)
4. [Spring AOP & proxy (JDK vs CGLIB)](#4-spring-aop--proxy)
5. [@Transactional internals](#5-transactional-internals)
6. [Spring Boot auto-configuration](#6-spring-boot-auto-configuration)
7. [REST & exception handling](#7-rest--exception-handling)
8. [System design cho Java backend](#8-system-design)
9. [Follow-up questions](#9-follow-up-questions)

---

## 1. IoC container & bean lifecycle

### Container làm gì

Spring **IoC container** (`ApplicationContext`) đọc định nghĩa bean (annotation/config), tạo instance, tiêm dependency, quản lý vòng đời. **Inversion of Control**: object không tự tạo dependency, container tiêm vào.

### Bean lifecycle (thứ tự chi tiết)

```
1. Instantiate (constructor)
2. Populate properties (dependency injection)
3. Aware callbacks (BeanNameAware, ApplicationContextAware...)
4. BeanPostProcessor.postProcessBeforeInitialization   <-- @PostConstruct chạy ở đây
5. InitializingBean.afterPropertiesSet / @PostConstruct / init-method
6. BeanPostProcessor.postProcessAfterInitialization    <-- AOP PROXY được tạo ở đây!
7. (bean sẵn sàng dùng)
8. DisposableBean.destroy / @PreDestroy / destroy-method (khi context đóng)
```

- **Điểm mấu chốt**: **proxy AOP được tạo ở bước 6** (`postProcessAfterInitialization`). Đây là lý do `@Transactional`/`@Async` không hoạt động khi tự gọi trong cùng bean (self-invocation) — vì gọi trực tiếp `this`, không qua proxy.
- `BeanPostProcessor` là hook mở rộng mạnh nhất (AOP, `@Autowired` resolution... đều là BeanPostProcessor).

---

## 2. Dependency injection

### 3 kiểu & vì sao constructor

```java
@Service
class OrderService {
    private final PaymentGateway gateway;    // final -> bất biến
    OrderService(PaymentGateway gateway) {   // constructor injection (KHUYẾN NGHỊ)
        this.gateway = gateway;
    }
}
```

| Kiểu | final | Test không Spring | Bắt buộc dependency | Circular dep |
|---|---|---|---|---|
| **Constructor** | ✓ | dễ (new + mock) | ✓ (compile ép) | phát hiện sớm (fail fast) |
| Field (`@Autowired`) | ✗ | khó (cần reflection) | ✗ | ẩn, dễ vô tình |
| Setter | ✗ | trung bình | ✗ (optional) | cho phép |

- Từ Spring 4.3, class có **1 constructor** thì không cần `@Autowired`.
- Constructor injection là **best practice**: bất biến, rõ ràng, dễ test đơn vị (không cần container), và phát hiện circular dependency lúc khởi động.

---

## 3. Bean scope & circular dependency

### Scope

- `singleton` (mặc định): **một instance** trên toàn container → phải **stateless/thread-safe**.
- `prototype`: mỗi lần lấy tạo instance mới (container không quản lý destroy).
- Web: `request`, `session`, `application`.

### Circular dependency

- A cần B, B cần A. Với **constructor injection** cả hai → Spring **không giải được** (cần A xong mới tạo B mà B lại cần A) → `BeanCurrentlyInCreationException` lúc khởi động (fail fast — tốt).
- Với **field/setter injection singleton** → Spring giải được nhờ **early reference** (tạo instance chưa init xong, tiêm vào rồi hoàn thiện) → che giấu vấn đề thiết kế.
- Cách xử lý đúng: **thiết kế lại** (tách trách nhiệm), hoặc `@Lazy`, hoặc dùng `ObjectProvider`. Circular dependency thường là mùi thiết kế xấu.

---

## 4. Spring AOP & proxy

> "@Transactional hoạt động thế nào?" luôn dẫn tới proxy. Phải phân biệt JDK dynamic proxy vs CGLIB.

### Proxy-based AOP

Spring AOP dùng **runtime proxy** (không weaving bytecode như AspectJ). Khi bean có aspect (`@Transactional`, `@Async`, `@Cacheable`), container thay bean thật bằng **proxy** bọc ngoài, chèn logic (advice) trước/sau method.

### JDK dynamic proxy vs CGLIB

| | JDK dynamic proxy | CGLIB |
|---|---|---|
| Cơ chế | `Proxy` implement **interface** | sinh **subclass** của class |
| Điều kiện | bean có interface | không cần interface |
| Giới hạn | chỉ proxy method của interface | không proxy được `final` class/method, `private` |
| Spring Boot | — | **mặc định** (`proxyTargetClass=true`) |

- Hệ quả CGLIB: `@Transactional` trên **`final` method/class → không proxy được → không có transaction** (âm thầm).

### Vì sao self-invocation không hoạt động

```java
@Service
class UserService {
    public void outer() { inner(); }           // gọi this.inner() -> KHÔNG qua proxy
    @Transactional public void inner() { ... }  // annotation bị bỏ qua!
}
```

Proxy chỉ chặn được lời gọi **từ ngoài vào**. `this.inner()` gọi thẳng object thật, bỏ qua proxy → không mở transaction. Fix: tách `inner` sang bean khác, hoặc self-inject proxy, hoặc `AopContext.currentProxy()`.

---

## 5. @Transactional internals

- Advice `TransactionInterceptor` (qua proxy) bọc method: **mở** transaction (hoặc join) trước, **commit** nếu thành công, **rollback** nếu ném exception.
- **Rollback rule**: mặc định rollback với `RuntimeException` + `Error`; **KHÔNG** rollback với checked exception (phải `@Transactional(rollbackFor = Exception.class)`).
- **Propagation**:
  - `REQUIRED` (mặc định): join transaction hiện có, hoặc tạo mới.
  - `REQUIRES_NEW`: suspend transaction ngoài, tạo transaction mới độc lập (audit log commit dù ngoài rollback).
  - `NESTED`: savepoint trong transaction hiện có.
  - `SUPPORTS`, `MANDATORY`, `NEVER`...
- **Bẫy thường gặp**:
  1. Self-invocation (mục 4).
  2. `@Transactional` trên method `private` / non-public → proxy không chặn được → vô hiệu.
  3. Checked exception mà quên `rollbackFor` → **commit** dù có lỗi.
  4. Bắt exception bên trong (nuốt) → transaction không thấy lỗi → không rollback.

---

## 6. Spring Boot auto-configuration

### Cơ chế

- `@SpringBootApplication` = `@Configuration` + `@ComponentScan` + `@EnableAutoConfiguration`.
- `@EnableAutoConfiguration` nạp danh sách auto-config class từ `META-INF/spring/org.springframework.boot.autoconfigure.AutoConfiguration.imports` (Spring Boot 2.7+; trước ở `spring.factories`).
- Mỗi auto-config class dùng **`@Conditional`** để chỉ kích hoạt khi điều kiện đúng:
  - `@ConditionalOnClass` (có class X trên classpath — vd có `DataSource` thì cấu hình JDBC).
  - `@ConditionalOnMissingBean` (chỉ tạo nếu người dùng chưa định nghĩa → cho phép override).
  - `@ConditionalOnProperty` (bật theo config).
- **Starter** (`spring-boot-starter-web`) chỉ là gói dependency kéo sẵn thư viện tương thích → classpath có gì thì auto-config kích hoạt cái đó.

> **Follow-up:** "Làm sao override một auto-configured bean?" → định nghĩa bean cùng kiểu; `@ConditionalOnMissingBean` khiến auto-config nhường.

---

## 7. REST & exception handling

```java
@RestController
@RequestMapping("/api/users")
class UserController {
    private final UserService service;
    UserController(UserService s) { this.service = s; }

    @GetMapping("/{id}")
    UserDto get(@PathVariable long id) { return service.get(id); }

    @PostMapping
    @ResponseStatus(HttpStatus.CREATED)
    UserDto create(@RequestBody @Valid CreateUserRequest req) { return service.create(req); }
}

@RestControllerAdvice
class ApiExceptionHandler {
    @ExceptionHandler(NotFoundException.class)
    @ResponseStatus(HttpStatus.NOT_FOUND)
    ErrorResponse handle(NotFoundException e) { return new ErrorResponse(e.getMessage()); }
}
```

- `@RestController` = `@Controller` + `@ResponseBody` (trả body JSON qua Jackson).
- `@RestControllerAdvice` gom xử lý exception tập trung → response lỗi nhất quán.
- `@Valid` + Bean Validation (`@NotNull`, `@Size`...) validate input.
- **Không trả entity Hibernate trực tiếp** → dùng DTO (tránh lazy proxy + lộ field).

---

## 8. System design

### REST API design

- Resource-oriented, method + status code đúng chuẩn; versioning (`/v1`), pagination, **idempotency key** cho POST (tránh double charge khi retry).

### Microservices resilience

- **Circuit breaker** (Resilience4j): sau ngưỡng lỗi, "mở" → fail nhanh, thử lại sau (half-open) → tránh cascade failure.
- Retry + timeout + bulkhead (cô lập tài nguyên). API gateway (routing/auth/rate limit).
- Giao tiếp: REST/gRPC (đồng bộ) vs Kafka/RabbitMQ (bất đồng bộ, decoupling, chịu tải đỉnh — đổi lại eventual consistency).

### Caching & data

- Redis cho read-heavy (cache-aside); cẩn thận invalidation, stampede (dùng lock/TTL jitter).
- DB: read replica, sharding, connection pool đúng size (Week 7).

### Observability

- Logs structured (JSON), metrics (Micrometer + Prometheus), tracing (OpenTelemetry), health check (`/actuator/health`) cho K8s readiness/liveness.

### Scalability

- Service **stateless** → scale ngang sau load balancer; state đẩy ra Redis/DB. Tác vụ nặng → queue xử lý async.

---

## 9. Follow-up questions

1. Proxy AOP được tạo ở bước nào của bean lifecycle? → postProcessAfterInitialization (BeanPostProcessor).
2. Vì sao self-invocation @Transactional không hoạt động? → gọi this bỏ qua proxy; proxy chỉ chặn lời gọi từ ngoài.
3. JDK dynamic proxy vs CGLIB? → interface-based vs subclass-based; CGLIB không proxy được final; Spring Boot mặc định CGLIB.
4. @Transactional rollback với checked exception không? → không, cần rollbackFor.
5. Vì sao constructor injection tốt hơn field? → final/bất biến, dễ test không cần Spring, phát hiện circular dep sớm.
6. Circular dependency: constructor vs field injection? → constructor fail fast (tốt); field giải được qua early reference (che giấu vấn đề).
7. Auto-configuration hoạt động thế nào? → @EnableAutoConfiguration nạp config class + @Conditional (OnClass/OnMissingBean) kích hoạt theo classpath.
8. Làm sao override auto-configured bean? → định nghĩa bean cùng kiểu; @ConditionalOnMissingBean nhường.
9. Bean singleton cần điều kiện gì? → stateless/thread-safe vì chia sẻ toàn container.
10. Circuit breaker giải quyết gì? → fail nhanh khi downstream lỗi, tránh cascade + cạn tài nguyên.
11. Vì sao không trả entity Hibernate ra REST? → lazy proxy + circular reference; dùng DTO.
