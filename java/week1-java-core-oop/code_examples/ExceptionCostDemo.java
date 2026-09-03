// Chi phí exception: fillInStackTrace là phần đắt. So sánh exception thường
// vs exception tắt stack trace vs không dùng exception (control flow).
// Chạy: java ExceptionCostDemo.java
public class ExceptionCostDemo {

    // Exception thường: mỗi lần tạo chạy fillInStackTrace (duyệt cả stack)
    static class NormalEx extends RuntimeException {
        NormalEx() { super("normal"); }
    }

    // Exception tắt stack trace: super(msg, cause, enableSuppression, writableStackTrace=false)
    static class NoTraceEx extends RuntimeException {
        NoTraceEx() { super("fast", null, false, false); }
    }

    static final int N = 200_000;

    static long timeThrow(Runnable r) {
        long t0 = System.nanoTime();
        for (int i = 0; i < N; i++) {
            try { r.run(); } catch (RuntimeException ignored) {}
        }
        return (System.nanoTime() - t0) / 1_000_000;
    }

    public static void main(String[] args) {
        // warm-up cho JIT
        for (int i = 0; i < 50_000; i++) { try { throw new NormalEx(); } catch (Exception e) {} }

        long normal  = timeThrow(() -> { throw new NormalEx(); });
        long notrace = timeThrow(() -> { throw new NoTraceEx(); });

        // Control flow không exception (baseline)
        long t0 = System.nanoTime();
        long acc = 0;
        for (int i = 0; i < N; i++) acc += (i % 7 == 0) ? 1 : 0;
        long plain = (System.nanoTime() - t0) / 1_000_000;

        System.out.printf("Throw %,d lần:%n", N);
        System.out.printf("  Exception thường (fillInStackTrace): %d ms%n", normal);
        System.out.printf("  Exception tắt stack trace          : %d ms%n", notrace);
        System.out.printf("  Không exception (baseline)         : %d ms (acc=%d)%n", plain, acc);
        System.out.println("-> fillInStackTrace là phần đắt; đừng dùng exception cho control flow.");
    }
}
