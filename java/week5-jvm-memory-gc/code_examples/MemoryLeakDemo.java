// Memory leak dù có GC: object không cần nữa nhưng vẫn reachable qua static collection.
// Chạy demo (heap nhỏ để thấy nhanh):
//   java -Xmx64m MemoryLeakDemo.java leak     -> OutOfMemoryError (leak)
//   java -Xmx64m MemoryLeakDemo.java noleak   -> chạy bình thường (dọn được)
// Xem GC hoạt động:
//   java -Xlog:gc -Xmx128m MemoryLeakDemo.java noleak
import java.util.*;

public class MemoryLeakDemo {

    // "Leak": static list giữ mọi object mãi mãi -> GC không thu được (vẫn reachable)
    static final List<byte[]> LEAK = new ArrayList<>();

    public static void main(String[] args) {
        String mode = args.length > 0 ? args[0] : "noleak";
        System.out.println("Mode: " + mode + " | maxHeap = "
                + Runtime.getRuntime().maxMemory() / (1024 * 1024) + " MB");

        try {
            for (int i = 0; i < 100_000; i++) {
                byte[] block = new byte[1024 * 1024]; // 1MB mỗi vòng

                if (mode.equals("leak")) {
                    LEAK.add(block);                  // GIỮ tham chiếu -> không thể GC
                } else {
                    // không giữ -> block thành rác ngay sau vòng lặp -> GC dọn được
                    if (block[0] != 0) System.out.println("touch");
                }

                if (i % 500 == 0) {
                    long used = (Runtime.getRuntime().totalMemory()
                            - Runtime.getRuntime().freeMemory()) / (1024 * 1024);
                    System.out.println("iter " + i + " | used ~ " + used + " MB"
                            + " | LEAK.size=" + LEAK.size());
                }
            }
            System.out.println("Hoàn thành (noleak): GC đã dọn rác liên tục, heap ổn định.");
        } catch (OutOfMemoryError e) {
            System.out.println(">>> OutOfMemoryError sau khi giữ " + LEAK.size()
                    + " MB. Object vẫn REACHABLE qua static list nên GC không thu được -> LEAK.");
        }
    }
}
