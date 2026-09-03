// Minh hoạ KHÁI NIỆM testing (mock qua interface, AAA) + N+1 problem — thuần stdlib.
// Thực tế dùng JUnit5 + Mockito (cần Maven/Gradle), nhưng đây để thấy rõ bản chất.
// Chạy: java TestingConceptsDemo.java
import java.util.*;

public class TestingConceptsDemo {

    // ---- Code under test ----
    record User(long id, String name) {}

    interface UserRepo {
        User findById(long id);
        List<User> findByOrderIds(List<Long> orderIds); // batch: fix N+1
    }

    static class UserService {
        private final UserRepo repo;
        UserService(UserRepo repo) { this.repo = repo; }
        String getName(long id) {
            User u = repo.findById(id);
            if (u == null) throw new NoSuchElementException("user " + id);
            return u.name();
        }
    }

    // ---- Mock repo thủ công (giống điều Mockito làm tự động) ----
    static class MockUserRepo implements UserRepo {
        Map<Long, User> data = new HashMap<>();
        int findByIdCalls = 0;   // đếm để "verify"
        public User findById(long id) { findByIdCalls++; return data.get(id); }
        public List<User> findByOrderIds(List<Long> ids) {
            var out = new ArrayList<User>();
            for (Long i : ids) if (data.containsKey(i)) out.add(data.get(i));
            return out;
        }
    }

    // ---- Mini test framework (thay JUnit cho demo) ----
    static int passed = 0, failed = 0;
    static void test(String name, Runnable body) {
        try { body.run(); passed++; System.out.println("  [PASS] " + name); }
        catch (Throwable t) { failed++; System.out.println("  [FAIL] " + name + " -> " + t.getMessage()); }
    }
    static void assertEq(Object expected, Object actual) {
        if (!Objects.equals(expected, actual))
            throw new AssertionError("expected=" + expected + " actual=" + actual);
    }
    static void assertThrows(Class<? extends Throwable> type, Runnable r) {
        try { r.run(); } catch (Throwable t) {
            if (type.isInstance(t)) return;
            throw new AssertionError("wrong exception: " + t.getClass().getSimpleName());
        }
        throw new AssertionError("expected " + type.getSimpleName() + " but none thrown");
    }

    public static void main(String[] args) {
        System.out.println("== Unit tests với mock (AAA pattern) ==");

        test("getName returns name when found", () -> {
            // Arrange
            var repo = new MockUserRepo();
            repo.data.put(1L, new User(1L, "An"));
            var service = new UserService(repo);
            // Act
            String name = service.getName(1L);
            // Assert
            assertEq("An", name);
            assertEq(1, repo.findByIdCalls);          // "verify" findById gọi 1 lần
        });

        test("getName throws when not found", () -> {
            var service = new UserService(new MockUserRepo());
            assertThrows(NoSuchElementException.class, () -> service.getName(99L));
        });

        System.out.printf("Result: %d passed, %d failed%n%n", passed, failed);

        // == Minh hoạ N+1 vs batch fetch ==
        System.out.println("== N+1 problem vs batch fetch ==");
        var repo = new MockUserRepo();
        for (long i = 1; i <= 5; i++) repo.data.put(i, new User(i, "U" + i));
        List<Long> orderUserIds = List.of(1L, 2L, 3L, 4L, 5L);

        // N+1: mỗi id một query
        repo.findByIdCalls = 0;
        for (Long id : orderUserIds) repo.findById(id);
        System.out.println("N+1  : " + repo.findByIdCalls + " queries (1 + N)");

        // Batch: gom một query
        List<User> batched = repo.findByOrderIds(orderUserIds);
        System.out.println("Batch: 1 query lấy " + batched.size() + " user (JOIN FETCH/@BatchSize)");
    }
}
