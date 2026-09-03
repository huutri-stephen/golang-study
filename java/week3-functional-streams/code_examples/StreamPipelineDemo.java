// Stream pipeline internals: xử lý theo chiều dọc (không tạo collection trung gian),
// short-circuit, stateful op (sorted) là rào cản, và custom Collector 4 phần.
// Chạy: java StreamPipelineDemo.java
import java.util.*;
import java.util.stream.*;

public class StreamPipelineDemo {

    public static void main(String[] args) {
        // 1. Xử lý theo chiều dọc: log cho thấy filter->map từng phần tử, không theo lớp
        System.out.println("== Vertical processing (không tạo list trung gian) ==");
        List<Integer> out = Stream.of(1, 2, 3, 4)
            .filter(n -> { System.out.println("filter " + n); return n % 2 == 0; })
            .map(n -> { System.out.println("map " + n); return n * 10; })
            .collect(Collectors.toList());
        System.out.println("result = " + out);
        System.out.println("-> thứ tự log: filter 1, filter 2, map 2, filter 3, filter 4, map 4");

        // 2. Short-circuit: limit dừng sớm
        System.out.println("\n== Short-circuit (limit) ==");
        List<Integer> firstTwo = Stream.iterate(1, n -> n + 1)  // vô hạn
            .peek(n -> System.out.println("gen " + n))
            .filter(n -> n % 2 == 0)
            .limit(2)                                            // chỉ cần 2 -> dừng
            .collect(Collectors.toList());
        System.out.println("firstTwo = " + firstTwo);

        // 3. Stateful op: sorted phải gom hết trước khi phát ra
        System.out.println("\n== Stateful op: sorted là rào cản ==");
        Stream.of(3, 1, 2)
            .peek(n -> System.out.println("trước sorted: " + n))
            .sorted()
            .peek(n -> System.out.println("sau sorted: " + n))
            .forEach(n -> {});
        System.out.println("-> tất cả 'trước sorted' in xong mới tới 'sau sorted'");

        // 4. Custom Collector 4 phần: supplier, accumulator, combiner, finisher
        System.out.println("\n== Custom Collector (đếm + nối) ==");
        String joined = Stream.of("a", "b", "c").collect(Collector.of(
            StringBuilder::new,                       // supplier
            (sb, s) -> sb.append(s).append(","),      // accumulator
            (sb1, sb2) -> sb1.append(sb2),            // combiner (cho parallel)
            sb -> sb.length() > 0 ? sb.substring(0, sb.length() - 1) : ""  // finisher
        ));
        System.out.println("custom joined = " + joined);

        // 5. toMap trùng key -> cần merge function
        System.out.println("\n== toMap trùng key ==");
        try {
            Stream.of("apple", "avocado", "banana")
                .collect(Collectors.toMap(s -> s.charAt(0), s -> s)); // 2 key 'a' -> nổ
        } catch (IllegalStateException e) {
            System.out.println("Không merge function -> IllegalStateException (duplicate key)");
        }
        Map<Character,String> merged = Stream.of("apple", "avocado", "banana")
            .collect(Collectors.toMap(s -> s.charAt(0), s -> s, (x, y) -> x + "|" + y));
        System.out.println("có merge function -> " + merged);
    }
}
