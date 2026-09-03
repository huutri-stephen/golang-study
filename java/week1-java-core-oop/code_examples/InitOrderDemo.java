// Thứ tự khởi tạo class có kế thừa + bẫy gọi overridable method trong constructor.
// Chạy: java InitOrderDemo.java
public class InitOrderDemo {

    static class Base {
        static { System.out.println("1. Base static block"); }
        { System.out.println("3. Base instance block"); }
        Base() {
            System.out.println("4. Base constructor");
            init();   // BẪY: gọi method overridable -> dynamic dispatch xuống Child
        }
        void init() { System.out.println("   Base.init()"); }
    }

    static class Child extends Base {
        static { System.out.println("2. Child static block"); }
        private String name = "child-field";   // gán khi tới lượt init field của Child
        { System.out.println("5. Child instance block"); }
        Child() { System.out.println("6. Child constructor"); }

        @Override void init() {
            // Chạy TRONG constructor Base -> field name CHƯA được gán -> null
            System.out.println("   Child.init() thấy name = " + name);
        }
    }

    public static void main(String[] args) {
        System.out.println("=== Lần new đầu tiên (kích hoạt static) ===");
        new Child();
        System.out.println("\n=== Lần new thứ hai (static KHÔNG chạy lại) ===");
        new Child();
    }
}
