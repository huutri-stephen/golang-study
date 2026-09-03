// equals/hashCode contract + hệ quả với HashMap/HashSet.
// Chạy: java EqualsHashCodeDemo.java   (JDK 11+ single-file mode)
import java.util.*;

public class EqualsHashCodeDemo {

    // Class ĐÚNG: override cả equals lẫn hashCode dùng cùng field.
    record Good(int id, String email) {
        // record tự sinh equals/hashCode/toString đúng contract
    }

    // Class SAI: chỉ override equals, quên hashCode.
    static final class Bad {
        final int id;
        Bad(int id) { this.id = id; }
        @Override public boolean equals(Object o) {
            return o instanceof Bad b && b.id == id;
        }
        // KHÔNG override hashCode -> dùng Object.hashCode (theo địa chỉ)
    }

    public static void main(String[] args) {
        // 1. == vs equals
        String a = new String("hi");
        String b = new String("hi");
        System.out.println("a == b      : " + (a == b));       // false (khác object)
        System.out.println("a.equals(b) : " + a.equals(b));    // true  (cùng nội dung)
        System.out.println("intern ==   : " + (a.intern() == b.intern())); // true (pool)

        // 2. Integer cache bẫy
        Integer x1 = 127, x2 = 127, y1 = 128, y2 = 128;
        System.out.println("\n127 == 127 (Integer): " + (x1 == x2)); // true  (cache)
        System.out.println("128 == 128 (Integer): " + (y1 == y2)); // false (ngoài cache)

        // 3. Class ĐÚNG hoạt động tốt trong HashSet
        Set<Good> goodSet = new HashSet<>();
        goodSet.add(new Good(1, "a@x.com"));
        boolean found = goodSet.contains(new Good(1, "a@x.com"));
        System.out.println("\n[Good] contains equal object: " + found); // true

        // 4. Class SAI: "bằng nhau" nhưng HashSet không nhận ra
        Set<Bad> badSet = new HashSet<>();
        badSet.add(new Bad(1));
        Bad other = new Bad(1);
        System.out.println("[Bad] equals?            : " + badSet.iterator().next().equals(other)); // true
        System.out.println("[Bad] contains equal obj : " + badSet.contains(other)); // FALSE -> bug!
        System.out.println("  -> vì hashCode khác nhau nên rơi vào bucket khác");
    }
}
