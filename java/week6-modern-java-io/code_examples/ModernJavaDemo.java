// Modern Java: records, sealed, pattern matching switch, switch expression, text block.
// Chạy: java ModernJavaDemo.java   (JDK 17)
import java.util.List;

public class ModernJavaDemo {

    // Sealed interface + records => tổng kiểu đóng
    sealed interface Shape permits Circle, Square, Rectangle {}
    record Circle(double r) implements Shape {}
    record Square(double side) implements Shape {}
    record Rectangle(double w, double h) implements Shape {}

    // Record với compact constructor để validate
    record Range(int lo, int hi) {
        Range {
            if (lo > hi) throw new IllegalArgumentException("lo > hi: " + lo + ">" + hi);
        }
    }

    // instanceof pattern matching (GA từ Java 16) + sealed.
    // Ghi chú: switch pattern matching là preview ở JDK 17, GA ở JDK 21:
    //   return switch (s) {
    //       case Circle c    -> Math.PI * c.r() * c.r();
    //       case Square sq   -> sq.side() * sq.side();
    //       case Rectangle r -> r.w() * r.h();
    //   };  // JDK 21: exhaustive, không cần default
    static double area(Shape s) {
        if (s instanceof Circle c) return Math.PI * c.r() * c.r();
        if (s instanceof Square sq) return sq.side() * sq.side();
        if (s instanceof Rectangle r) return r.w() * r.h();
        throw new IllegalStateException("unknown shape: " + s);
    }

    public static void main(String[] args) {
        // 1. Record: equals/hashCode/toString tự sinh
        var c1 = new Circle(2);
        var c2 = new Circle(2);
        System.out.println("record toString : " + c1);
        System.out.println("record equals   : " + c1.equals(c2)); // true

        // 2. Sealed + switch pattern matching
        List<Shape> shapes = List.of(new Circle(1), new Square(2), new Rectangle(2, 3));
        for (Shape s : shapes) {
            System.out.printf("area(%s) = %.2f%n", s, area(s));
        }

        // 3. instanceof pattern matching
        Object obj = "hello world";
        if (obj instanceof String str && str.length() > 5) {
            System.out.println("\ninstanceof pattern -> upper: " + str.toUpperCase());
        }

        // 4. Switch expression với yield
        int month = 2;
        int days = switch (month) {
            case 1, 3, 5, 7, 8, 10, 12 -> 31;
            case 4, 6, 9, 11 -> 30;
            case 2 -> 28;
            default -> {
                yield -1;
            }
        };
        System.out.println("days in month " + month + " = " + days);

        // 5. Text block
        String json = """
            {
              "name": "an",
              "age": 25
            }""";
        System.out.println("\ntext block:\n" + json);

        // 6. Compact constructor validation
        try {
            new Range(5, 1);
        } catch (IllegalArgumentException e) {
            System.out.println("\nRange validation: " + e.getMessage());
        }

        // 7. Immutable factory
        var immutable = List.of("a", "b");
        try {
            immutable.add("c");
        } catch (UnsupportedOperationException e) {
            System.out.println("List.of() là immutable -> add ném UnsupportedOperationException");
        }
    }
}
