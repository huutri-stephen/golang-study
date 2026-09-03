// ArrayList grow 1.5x (mô phỏng công thức thực), tác động prealloc, và fail-fast modCount.
// Ghi chú: JDK 17 module system chặn reflection vào java.util nên ta MÔ PHỎNG công thức
// grow thay vì đọc field private (nếu muốn đọc thật: java --add-opens java.base/java.util=ALL-UNNAMED).
// Chạy: java ArrayListGrowDemo.java
import java.util.*;

public class ArrayListGrowDemo {

    // Công thức grow thực của ArrayList (Java 8+): newCap = oldCap + (oldCap >> 1)
    static int nextCapacity(int oldCap) { return oldCap + (oldCap >> 1); }

    public static void main(String[] args) {
        // 1. Chuỗi capacity growth: 10 -> 15 -> 22 -> 33 -> 49 ...
        System.out.println("== ArrayList capacity growth (newCap = oldCap + oldCap>>1) ==");
        int cap = 10;   // default capacity
        System.out.print("10");
        for (int i = 0; i < 8; i++) { cap = nextCapacity(cap); System.out.print(" -> " + cap); }
        System.out.println("\n-> tăng 1.5x mỗi lần đầy; copy sang mảng mới (amortized O(1)/add)");

        // 2. Prealloc vs không prealloc
        int n = 5_000_000;
        long t0 = System.nanoTime();
        List<Integer> noPre = new ArrayList<>();       // grow nhiều lần
        for (int i = 0; i < n; i++) noPre.add(i);
        long tNo = (System.nanoTime() - t0) / 1_000_000;

        long t1 = System.nanoTime();
        List<Integer> pre = new ArrayList<>(n);        // cấp sẵn -> không resize
        for (int i = 0; i < n; i++) pre.add(i);
        long tPre = (System.nanoTime() - t1) / 1_000_000;

        System.out.printf("%nThêm %,d phần tử: không prealloc %d ms | prealloc %d ms%n", n, tNo, tPre);

        // Đếm số lần resize lý thuyết khi không prealloc (từ cap 10)
        int resizes = 0, c = 10;
        while (c < n) { c = nextCapacity(c); resizes++; }
        System.out.println("Số lần resize (mỗi lần copy toàn mảng): ~" + resizes);

        // 3. Fail-fast: sửa cấu trúc khi đang duyệt
        System.out.println("\n== Fail-fast (modCount) ==");
        List<Integer> l = new ArrayList<>(List.of(1, 2, 3, 4));
        try {
            for (Integer x : l) if (x == 2) l.remove(x);   // sửa trực tiếp -> CME
        } catch (ConcurrentModificationException e) {
            System.out.println("Bắt ConcurrentModificationException khi remove trong for-each");
        }
        // Cách đúng: Iterator.remove hoặc removeIf
        Iterator<Integer> it = l.iterator();
        while (it.hasNext()) if (it.next() == 3) it.remove();
        System.out.println("sau iterator.remove(3) -> " + l);
        l.removeIf(x -> x == 2);
        System.out.println("sau removeIf(==2)      -> " + l);
    }
}
