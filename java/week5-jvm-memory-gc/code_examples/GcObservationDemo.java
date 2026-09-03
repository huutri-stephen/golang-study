// Quan sát GC: object chết trẻ (young GC dọn được) vs object sống lâu (promote lên old).
// Chạy có GC log để thấy minor GC:
//   java -Xlog:gc -Xmx128m GcObservationDemo.java
// Hoặc chạy thường để xem thống kê alloc:
//   java GcObservationDemo.java
public class GcObservationDemo {

    // Giữ vài object "sống lâu" -> bị promote lên old gen qua nhiều lần minor GC
    static final java.util.List<byte[]> longLived = new java.util.ArrayList<>();

    public static void main(String[] args) {
        Runtime rt = Runtime.getRuntime();
        long gcBefore = totalGcCount();

        System.out.println("== Cấp phát: đa số chết trẻ, một ít sống lâu ==");
        long allocated = 0;
        for (int i = 0; i < 200_000; i++) {
            byte[] garbage = new byte[1024];   // 1KB: chết trẻ -> young GC dọn
            allocated += garbage.length;
            if (garbage[0] != 0) System.out.print("");   // dùng để tránh bị tối ưu bỏ

            if (i % 20_000 == 0) {
                longLived.add(new byte[64 * 1024]);       // 64KB giữ lại -> sống lâu
            }
        }

        long gcAfter = totalGcCount();
        System.out.printf("Đã cấp phát ~%,d MB rác tạm%n", allocated / (1024 * 1024));
        System.out.printf("Số object sống lâu giữ lại: %d (~%d MB)%n",
                longLived.size(), longLived.size() * 64 / 1024);
        System.out.printf("Số lần GC xảy ra trong vòng lặp: %d%n", gcAfter - gcBefore);
        System.out.printf("Heap used hiện tại: ~%d MB / max %d MB%n",
                (rt.totalMemory() - rt.freeMemory()) / (1024 * 1024),
                rt.maxMemory() / (1024 * 1024));
        System.out.println("-> Rác 1KB chết trẻ được minor GC dọn liên tục (heap không phình).");
        System.out.println("-> Chạy với -Xlog:gc để thấy các dòng 'Pause Young' và promotion.");
    }

    // Đếm tổng số lần GC qua GarbageCollectorMXBean (an toàn, không cần reflection nội bộ)
    static long totalGcCount() {
        long sum = 0;
        for (var bean : java.lang.management.ManagementFactory.getGarbageCollectorMXBeans()) {
            long c = bean.getCollectionCount();
            if (c > 0) sum += c;
        }
        return sum;
    }
}
