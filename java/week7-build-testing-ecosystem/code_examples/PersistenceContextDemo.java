// Mô phỏng persistence context (L1 cache) + dirty checking + N+1 vs batch fetch của Hibernate.
// Thuần stdlib (không có Hibernate/Maven) để thấy CƠ CHẾ. Chạy: java PersistenceContextDemo.java
import java.util.*;

public class PersistenceContextDemo {

    static class User {
        final long id; String email;
        User(long id, String email) { this.id = id; this.email = email; }
        User copy() { return new User(id, email); }   // để chụp snapshot
    }

    // "DB" giả + đếm số query để minh hoạ L1 cache và N+1
    static class FakeDb {
        final Map<Long, User> store = new HashMap<>();
        int selectCount = 0, updateCount = 0;
        User select(long id) { selectCount++; User u = store.get(id); return u == null ? null : u.copy(); }
        List<User> selectIn(Collection<Long> ids) {   // batch: 1 query
            selectCount++;
            var out = new ArrayList<User>();
            for (Long id : ids) if (store.containsKey(id)) out.add(store.get(id).copy());
            return out;
        }
        void update(User u) { updateCount++; store.put(u.id, u.copy()); }
    }

    // Persistence context: cache entity managed + snapshot để dirty checking
    static class PersistenceContext implements AutoCloseable {
        final FakeDb db;
        final Map<Long, User> managed = new HashMap<>();
        final Map<Long, User> snapshot = new HashMap<>();
        PersistenceContext(FakeDb db) { this.db = db; }

        User find(long id) {
            if (managed.containsKey(id)) return managed.get(id);   // L1 cache hit -> KHÔNG query
            User u = db.select(id);
            if (u != null) { managed.put(id, u); snapshot.put(id, u.copy()); }
            return u;
        }
        // flush: dirty checking - so managed với snapshot, chỉ UPDATE cái đổi
        void flush() {
            for (User cur : managed.values()) {
                User snap = snapshot.get(cur.id);
                if (snap != null && !Objects.equals(cur.email, snap.email)) {
                    db.update(cur);
                    snapshot.put(cur.id, cur.copy());
                }
            }
        }
        public void close() { flush(); }   // commit -> flush
    }

    public static void main(String[] args) {
        FakeDb db = new FakeDb();
        for (long i = 1; i <= 5; i++) db.store.put(i, new User(i, "u" + i + "@x.com"));

        // 1. L1 cache: find cùng id 2 lần -> chỉ 1 query
        System.out.println("== L1 cache (persistence context) ==");
        try (var ctx = new PersistenceContext(db)) {
            ctx.find(1); ctx.find(1); ctx.find(1);
            System.out.println("find(1) x3 -> selectCount = " + db.selectCount + " (chỉ 1 query, còn lại từ cache)");
        }

        // 2. Dirty checking: chỉ setField, không gọi update -> flush tự UPDATE
        System.out.println("\n== Dirty checking ==");
        db.updateCount = 0;
        try (var ctx = new PersistenceContext(db)) {
            User u = ctx.find(2);
            u.email = "changed@x.com";    // chỉ set field
            // KHÔNG gọi db.update -> close()/flush() phát hiện thay đổi và UPDATE
        }
        System.out.println("chỉ set email -> updateCount = " + db.updateCount + " (dirty checking tự UPDATE)");
        System.out.println("email trong db bây giờ = " + db.store.get(2L).email);

        // 3. N+1 vs batch
        System.out.println("\n== N+1 vs batch fetch ==");
        db.selectCount = 0;
        List<Long> ids = List.of(1L, 2L, 3L, 4L, 5L);
        for (Long id : ids) db.select(id);            // N+1: mỗi id 1 query
        System.out.println("N+1  : selectCount = " + db.selectCount + " (1 + N)");
        db.selectCount = 0;
        db.selectIn(ids);                              // batch: 1 query IN(...)
        System.out.println("batch: selectCount = " + db.selectCount + " (JOIN FETCH/@BatchSize)");
    }
}
