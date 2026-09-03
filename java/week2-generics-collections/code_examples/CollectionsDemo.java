// Collections: fail-fast iterator, cách sửa list đúng, và khác biệt Map implementations.
// Chạy: java CollectionsDemo.java
import java.util.*;

public class CollectionsDemo {
    public static void main(String[] args) {
        // 1. Fail-fast: sửa cấu trúc khi for-each -> ConcurrentModificationException
        List<String> list = new ArrayList<>(List.of("", "a", "", "b"));
        try {
            for (String s : list) {
                if (s.isEmpty()) list.remove(s); // sửa trực tiếp -> lỗi
            }
        } catch (ConcurrentModificationException e) {
            System.out.println("Fail-fast: bắt được ConcurrentModificationException");
        }

        // Cách ĐÚNG: removeIf (hoặc iterator.remove)
        List<String> ok = new ArrayList<>(List.of("", "a", "", "b"));
        ok.removeIf(String::isEmpty);
        System.out.println("removeIf -> " + ok); // [a, b]

        // 2. Map implementations giữ thứ tự khác nhau
        Map<String, Integer> hash = new HashMap<>();
        Map<String, Integer> linked = new LinkedHashMap<>();
        Map<String, Integer> tree = new TreeMap<>();
        for (Map<String, Integer> m : List.of(hash, linked, tree)) {
            m.put("banana", 1); m.put("apple", 2); m.put("cherry", 3);
        }
        System.out.println("\nHashMap       (không đảm bảo thứ tự): " + hash);
        System.out.println("LinkedHashMap (thứ tự chèn)        : " + linked);
        System.out.println("TreeMap       (sắp xếp theo key)   : " + tree);

        // 3. LinkedHashMap access-order làm LRU cache đơn giản
        LinkedHashMap<Integer, String> lru = new LinkedHashMap<>(16, 0.75f, true) {
            @Override protected boolean removeEldestEntry(Map.Entry<Integer, String> e) {
                return size() > 3; // giữ tối đa 3 phần tử
            }
        };
        lru.put(1, "a"); lru.put(2, "b"); lru.put(3, "c");
        lru.get(1);          // truy cập 1 -> đẩy 1 thành mới nhất
        lru.put(4, "d");     // vượt 3 -> loại phần tử cũ nhất (2)
        System.out.println("\nLRU cache (cap=3) sau truy cập: " + lru.keySet()); // [3, 1, 4]
    }
}
