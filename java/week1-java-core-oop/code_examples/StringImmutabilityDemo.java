// String immutability, pool, và hiệu năng nối chuỗi.
// Chạy: java StringImmutabilityDemo.java
public class StringImmutabilityDemo {
    public static void main(String[] args) {
        // 1. Immutability: "thay đổi" tạo object mới, bản gốc không đổi
        String s = "hello";
        String upper = s.toUpperCase();
        System.out.println("original: " + s);      // hello (không đổi)
        System.out.println("upper   : " + upper);   // HELLO (object mới)

        // 2. String pool
        String p1 = "java";
        String p2 = "java";
        String p3 = new String("java");
        System.out.println("\np1 == p2 (literal)  : " + (p1 == p2)); // true - cùng pool
        System.out.println("p1 == p3 (new)      : " + (p1 == p3)); // false - heap mới
        System.out.println("p1 == p3.intern()   : " + (p1 == p3.intern())); // true

        // 3. Nối chuỗi: + trong loop tạo nhiều object; StringBuilder hiệu quả
        int n = 50_000;
        long t0 = System.nanoTime();
        String concat = "";
        for (int i = 0; i < n; i++) concat += "x";   // O(n^2): mỗi vòng tạo String mới
        long t1 = System.nanoTime();

        StringBuilder sb = new StringBuilder();
        for (int i = 0; i < n; i++) sb.append("x");  // O(n): buffer mutable
        String built = sb.toString();
        long t2 = System.nanoTime();

        System.out.printf("%n+= concat  : %.1f ms (len=%d)%n", (t1 - t0) / 1e6, concat.length());
        System.out.printf("StringBuilder: %.1f ms (len=%d)%n", (t2 - t1) / 1e6, built.length());
        System.out.println("-> StringBuilder nhanh hơn nhiều khi nối nhiều lần");
    }
}
