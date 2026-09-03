// PECS (Producer Extends, Consumer Super) + type erasure minh hoạ.
// Chạy: java PecsDemo.java
import java.util.*;

public class PecsDemo {

    // PRODUCER: chỉ ĐỌC ra Number -> ? extends
    static double total(List<? extends Number> src) {
        double s = 0;
        for (Number n : src) s += n.doubleValue();
        // src.add(1); // <- KHÔNG compile được: không biết kiểu con chính xác
        return s;
    }

    // CONSUMER: chỉ GHI Integer vào -> ? super
    static void fillWith123(List<? super Integer> dst) {
        dst.add(1);
        dst.add(2);
        dst.add(3);
        // Integer x = dst.get(0); // <- đọc ra chỉ được Object
    }

    public static void main(String[] args) {
        // Producer: truyền List con của Number đều được
        List<Integer> ints = List.of(1, 2, 3);
        List<Double> dbls = List.of(1.5, 2.5);
        System.out.println("total(ints) = " + total(ints)); // 6.0
        System.out.println("total(dbls) = " + total(dbls)); // 4.0

        // Consumer: truyền List cha của Integer đều được
        List<Number> numbers = new ArrayList<>();
        List<Object> objects = new ArrayList<>();
        fillWith123(numbers);
        fillWith123(objects);
        System.out.println("filled numbers = " + numbers);
        System.out.println("filled objects = " + objects);

        // Type erasure: List<String> và List<Integer> cùng class runtime
        List<String> ls = new ArrayList<>();
        List<Integer> li = new ArrayList<>();
        System.out.println("\nType erasure:");
        System.out.println("ls.getClass() == li.getClass() ? "
                + (ls.getClass() == li.getClass())); // true - cùng ArrayList.class
    }
}
