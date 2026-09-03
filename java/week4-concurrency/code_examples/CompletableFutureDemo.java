// CompletableFuture: async pipeline, combine song song, xử lý exception.
// Chạy: java CompletableFutureDemo.java
import java.util.concurrent.*;

public class CompletableFutureDemo {

    static String fetchUser(String id) {
        sleep(100);
        return "User(" + id + ")";
    }

    static int fetchOrderCount(String user) {
        sleep(100);
        return user.length();
    }

    public static void main(String[] args) throws Exception {
        long start = System.currentTimeMillis();

        // 1. Pipeline: supply -> transform -> compose
        CompletableFuture<Integer> pipeline = CompletableFuture
            .supplyAsync(() -> fetchUser("42"))
            .thenApply(u -> { System.out.println("got " + u); return u; })
            .thenCompose(u -> CompletableFuture.supplyAsync(() -> fetchOrderCount(u)));
        System.out.println("Order count = " + pipeline.get());

        // 2. Chạy 2 việc SONG SONG rồi combine (tổng ~100ms chứ không 200ms)
        long t0 = System.currentTimeMillis();
        CompletableFuture<String> a = CompletableFuture.supplyAsync(() -> fetchUser("A"));
        CompletableFuture<String> b = CompletableFuture.supplyAsync(() -> fetchUser("B"));
        String combined = a.thenCombine(b, (x, y) -> x + " + " + y).get();
        System.out.println("Combined = " + combined
                + "  (took " + (System.currentTimeMillis() - t0) + "ms, song song)");

        // 3. Exception handling: exceptionally cung cấp fallback
        String safe = CompletableFuture
            .<String>supplyAsync(() -> { throw new RuntimeException("boom"); })
            .thenApply(s -> s + "!")            // bị SKIP vì stage trước lỗi
            .exceptionally(ex -> "fallback (" + ex.getMessage() + ")")
            .get();
        System.out.println("With error -> " + safe);

        System.out.println("Total: " + (System.currentTimeMillis() - start) + "ms");
    }

    static void sleep(long ms) {
        try { Thread.sleep(ms); } catch (InterruptedException e) { Thread.currentThread().interrupt(); }
    }
}
