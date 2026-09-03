// Race condition: volatile KHÔNG đủ, cần atomic/synchronized.
// Chạy: java RaceConditionDemo.java
import java.util.concurrent.*;
import java.util.concurrent.atomic.AtomicInteger;

public class RaceConditionDemo {
    static volatile int volatileCounter = 0;   // volatile: visibility, KHÔNG atomic
    static int syncCounter = 0;
    static final AtomicInteger atomicCounter = new AtomicInteger(0);
    static final Object lock = new Object();

    static final int THREADS = 8;
    static final int PER_THREAD = 100_000;
    static final int EXPECTED = THREADS * PER_THREAD;

    public static void main(String[] args) throws Exception {
        ExecutorService pool = Executors.newFixedThreadPool(THREADS);
        CountDownLatch latch = new CountDownLatch(THREADS);

        for (int i = 0; i < THREADS; i++) {
            pool.submit(() -> {
                for (int j = 0; j < PER_THREAD; j++) {
                    volatileCounter++;                     // RACE: read-modify-write
                    synchronized (lock) { syncCounter++; } // an toàn
                    atomicCounter.incrementAndGet();       // an toàn (CAS)
                }
                latch.countDown();
            });
        }
        latch.await();
        pool.shutdown();

        System.out.println("Expected           : " + EXPECTED);
        System.out.println("volatile (SAI)     : " + volatileCounter
                + (volatileCounter == EXPECTED ? "" : "  <- mất update do race!"));
        System.out.println("synchronized (đúng): " + syncCounter);
        System.out.println("atomic (đúng)      : " + atomicCounter.get());
    }
}
