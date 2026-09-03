// Sealed + record = algebraic data type; exhaustive matching.
// Dùng if-instanceof (GA Java 16) để chạy trên JDK 17 không cần --enable-preview.
// Ghi chú: Java 21 cho phép switch pattern + record deconstruction gọn hơn (xem comment).
// Chạy: java SealedShapeDemo.java
import java.util.List;

public class SealedShapeDemo {

    sealed interface Shape permits Circle, Square, Rectangle {}
    record Circle(double r) implements Shape {}
    record Square(double s) implements Shape {}
    record Rectangle(double w, double h) implements Shape {}

    // Java 17 (GA): if-instanceof pattern
    static double area(Shape shape) {
        if (shape instanceof Circle c)    return Math.PI * c.r() * c.r();
        if (shape instanceof Square s)    return s.s() * s.s();
        if (shape instanceof Rectangle r) return r.w() * r.h();
        throw new IllegalStateException("unknown: " + shape);
        // Java 21 GA tương đương (exhaustive, không cần default/throw):
        //   return switch (shape) {
        //       case Circle(double r)       -> Math.PI * r * r;   // record deconstruction
        //       case Square(double s)       -> s * s;
        //       case Rectangle(double w, double h) -> w * h;
        //   };
    }

    // Guard + instanceof
    static String classify(Shape shape) {
        if (shape instanceof Circle c && c.r() > 10) return "big circle";
        if (shape instanceof Circle) return "small circle";
        return "polygon";
    }

    public static void main(String[] args) {
        List<Shape> shapes = List.of(new Circle(2), new Square(3), new Rectangle(2, 5), new Circle(20));
        System.out.println("== area (exhaustive over sealed) ==");
        for (Shape s : shapes) System.out.printf("%-30s area=%.2f%n", s, area(s));

        System.out.println("\n== classify (guard) ==");
        for (Shape s : shapes) System.out.printf("%-30s -> %s%n", s, classify(s));

        // Record: equals/hashCode/toString tự sinh
        System.out.println("\n== record identity ==");
        System.out.println("Circle(2).equals(Circle(2)) = " + new Circle(2).equals(new Circle(2)));
        System.out.println("toString: " + new Rectangle(2, 5));
    }
}
