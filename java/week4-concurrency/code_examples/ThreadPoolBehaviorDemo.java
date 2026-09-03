// ThreadPoolExecutor: minh hoạ thứ tự core -> queue -> max -> reject,
// và vì sao bounded queue + CallerRunsPolicy tạo back-pressure.
// Chạy: java ThreadPoolBehaviorDemo.java
import java.util.concurrent.*;
import java.util.concurrent.atomic.AtomicInteger;

public class ThreadPoolBehaviorDemo {

    public static void main(String[] args) throws Exception {
        AtomicInteger running = new AtomicInteger();
        AtomicInteger rejected = new AtomicInteger();

        // core=2, max=4, bounded queue=2 -> tối đa xử lý 4 chạy + 2 chờ = 6 task đồng thời
        ThreadPoolExecutor pool = new ThreadPoolExecutor(
                2, 4, 30, TimeUnit.SECONDS,
                new ArrayBlockingQueue<>(2),                 // BOUNDED queue
                new ThreadPoolExecutor.AbortPolicy());       // đầy -> reject

        System.out.println("core=2, max=4, queue=2 -> submit 10 task");
        for (int i = 0; i < 10; i++) {
            final int id = i;
            try {
                pool.submit(() -> {
                    int n = running.incrementAndGet();
                    System.out.printf("task %d chạy (đang chạy=%d, poolSize=%d, queue=%d)%n",
                            id, n, pool.getPoolSize(), pool.getQueue().size());
                    sleep(200);
                    running.decrementAndGet();
                });
            } catch (RejectedExecutionException e) {
                rejected.incrementAndGet();
                System.out.println("task " + id + " BỊ REJECT (queue đầy + đạt max)");
            }
        }
        pool.shutdown();
        pool.awaitTermination(5, TimeUnit.SECONDS);
        System.out.println("=> Rejected " + rejected.get() + " task (10 - 4 chạy - 2 queue = 4 reject)");

        // Demo CallerRunsPolicy: không reject mà chạy trên thread gọi -> back-pressure
        System.out.println("\n== CallerRunsPolicy (back-pressure, không mất task) ==");
        ThreadPoolExecutor pool2 = new ThreadPoolExecutor(
                1, 1, 0, TimeUnit.SECONDS,
                new ArrayBlockingQueue<>(1),
                new ThreadPoolExecutor.CallerRunsPolicy());
        for (int i = 0; i < 5; i++) {
            final int id = i;
            pool2.submit(() -> {
                System.out.println("  task " + id + " chạy trên " + Thread.currentThread().getName());
                sleep(100);
            });
        }
        pool2.shutdown();
        pool2.awaitTermination(5, TimeUnit.SECONDS);
        System.out.println("-> vài task chạy trên thread 'main' (caller) -> làm chậm producer");
    }

    static void sleep(long ms) {
        try { Thread.sleep(ms); } catch (InterruptedException e) { Thread.currentThread().interrupt(); }
    }
}
