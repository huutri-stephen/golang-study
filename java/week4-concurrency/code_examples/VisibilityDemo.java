// JMM visibility: thiếu volatile -> thread reader có thể KHÔNG BAO GIỜ thấy flag đổi.
// Chạy: java VisibilityDemo.java
// (Có thể phải chạy vài lần; với JIT tối ưu, biến non-volatile thường bị cache trong thanh ghi.)
public class VisibilityDemo {

    static boolean plainFlag = false;        // KHÔNG volatile -> có thể không thấy thay đổi
    static volatile boolean volatileFlag = false;

    public static void main(String[] args) throws Exception {
        System.out.println("== Test volatile: reader dừng đúng lúc ==");
        Thread reader = new Thread(() -> {
            long spins = 0;
            while (!volatileFlag) spins++;    // volatile -> thấy ngay khi main set true
            System.out.println("volatile reader THOÁT sau " + spins + " vòng (thấy flag=true)");
        });
        reader.start();
        Thread.sleep(200);
        volatileFlag = true;                  // main ghi
        reader.join(2000);
        System.out.println("volatile reader còn sống? " + reader.isAlive() + " (false = tốt)");

        System.out.println("\n== Giải thích plain flag ==");
        System.out.println("Với biến KHÔNG volatile, JIT có thể nâng !plainFlag ra ngoài loop");
        System.out.println("(cache trong thanh ghi) -> reader có thể lặp VÔ HẠN dù main đã set true.");
        System.out.println("Đó là lý do flag chia sẻ giữa thread PHẢI volatile.");
        // Không demo trực tiếp plainFlag vì có thể treo vô hạn; chỉ giải thích an toàn.
    }
}
