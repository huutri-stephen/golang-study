// Stream API: lazy evaluation, groupingBy, reduce, Collectors, Optional.
// Chạy: java StreamsDemo.java
import java.util.*;
import java.util.stream.*;

public class StreamsDemo {
    record User(String name, String city, int age) {}

    public static void main(String[] args) {
        List<User> users = List.of(
            new User("An", "Hanoi", 25),
            new User("Binh", "Hanoi", 17),
            new User("Cuong", "HCM", 30),
            new User("Dung", "HCM", 22),
            new User("Em", "Danang", 19)
        );

        // 1. Lazy evaluation + short-circuit: chỉ duyệt tới khi tìm thấy
        System.out.println("== Lazy demo ==");
        Optional<String> first = Stream.of("a", "bb", "ccc", "dddd")
            .filter(s -> { System.out.println("filter " + s); return s.length() > 1; })
            .map(s -> { System.out.println("map " + s); return s.toUpperCase(); })
            .findFirst();
        System.out.println("findFirst -> " + first.get()); // BB, không duyệt ccc/dddd

        // 2. filter + map + collect
        List<String> adults = users.stream()
            .filter(u -> u.age() >= 18)
            .map(User::name)
            .sorted()
            .toList();
        System.out.println("\nAdults: " + adults);

        // 3. groupingBy + downstream
        Map<String, List<String>> namesByCity = users.stream()
            .collect(Collectors.groupingBy(User::city,
                     Collectors.mapping(User::name, Collectors.toList())));
        System.out.println("Names by city: " + namesByCity);

        Map<String, Long> countByCity = users.stream()
            .collect(Collectors.groupingBy(User::city, Collectors.counting()));
        System.out.println("Count by city: " + countByCity);

        // 4. partitioningBy (chia theo boolean)
        Map<Boolean, List<String>> partition = users.stream()
            .collect(Collectors.partitioningBy(u -> u.age() >= 18,
                     Collectors.mapping(User::name, Collectors.toList())));
        System.out.println("Adult? -> " + partition);

        // 5. reduce + average
        int totalAge = users.stream().mapToInt(User::age).sum();
        double avgAge = users.stream().mapToInt(User::age).average().orElse(0);
        System.out.printf("Total age=%d, avg=%.1f%n", totalAge, avgAge);

        // 6. Optional dùng đúng cách
        String oldest = users.stream()
            .max(Comparator.comparingInt(User::age))
            .map(User::name)
            .orElse("none");
        System.out.println("Oldest: " + oldest);

        // 7. joining
        String csv = users.stream().map(User::name)
            .collect(Collectors.joining(", ", "[", "]"));
        System.out.println("CSV: " + csv);
    }
}
