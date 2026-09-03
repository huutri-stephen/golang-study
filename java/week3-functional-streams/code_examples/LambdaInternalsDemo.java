// Lambda internals: lambda KHÔNG phải anonymous class; stateless lambda được tái dùng;
// capturing lambda tạo instance mới; anonymous class sinh class riêng.
// Xem bằng chứng: java LambdaInternalsDemo.java  (chú ý getClass() và số class file)
import java.util.function.*;

public class LambdaInternalsDemo {

    public static void main(String[] args) {
        // 1. Stateless lambda (không capture): MỖI call site (dòng lambda) được cache riêng.
        //    Cùng một call site đánh giá nhiều lần -> reuse cùng instance.
        //    Hai call site khác nhau (s1, s2 ở 2 dòng) -> 2 instance khác nhau.
        Supplier<String> s1 = () -> "hi";
        Supplier<String> s2 = () -> "hi";
        System.out.println("== Stateless lambda ==");
        System.out.println("s1 class: " + s1.getClass().getSimpleName()); // tên $$Lambda...
        System.out.println("s1 == s2 ? " + (s1 == s2)
                + "  (false - 2 call site khác nhau; nhưng cùng 1 call site lặp lại thì reuse)");

        // Chứng minh reuse tại CÙNG call site: gọi cùng factory 2 lần
        Supplier<String> r1 = sameCallSite();
        Supplier<String> r2 = sameCallSite();
        System.out.println("r1 == r2 ? " + (r1 == r2)
                + "  (true - cùng call site, stateless -> JVM reuse instance)");

        // 2. Capturing lambda -> tạo instance mới giữ biến bắt được
        System.out.println("\n== Capturing lambda ==");
        Function<Integer,Integer> a = makeAdder(10);
        Function<Integer,Integer> b = makeAdder(20);
        System.out.println("adder(10).apply(5) = " + a.apply(5)); // 15
        System.out.println("adder(20).apply(5) = " + b.apply(5)); // 25
        System.out.println("a == b ? " + (a == b) + "  (false - mỗi cái giữ base riêng)");

        // 3. So sánh với anonymous class: sinh class file riêng (Outer$1)
        System.out.println("\n== Anonymous class ==");
        Runnable anon = new Runnable() { public void run() {} };
        System.out.println("anon class: " + anon.getClass().getName()
                + "  (có $1 -> class file riêng sinh lúc compile)");
        Runnable lam = () -> {};
        System.out.println("lambda class: " + lam.getClass().getName()
                + "  ($$Lambda -> sinh lúc runtime bởi LambdaMetafactory)");

        // 4. Effectively final: nếu gán lại biến bị capture sẽ lỗi compile
        int base = 100;
        Supplier<Integer> get = () -> base;   // capture base = 100
        // base = 200; // <- bỏ comment sẽ LỖI COMPILE: variable must be effectively final
        System.out.println("\ncaptured base = " + get.get());
    }

    static Function<Integer,Integer> makeAdder(int base) {
        return x -> x + base;   // capture base -> closure
    }

    // Cùng một call site (dòng lambda này) -> stateless -> JVM cache & reuse instance
    static Supplier<String> sameCallSite() {
        return () -> "hi";
    }
}
