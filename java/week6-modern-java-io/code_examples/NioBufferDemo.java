// NIO ByteBuffer: position/limit/capacity, flip() chuyển ghi->đọc; và Files/Path hiện đại.
// Chạy: java NioBufferDemo.java
import java.io.IOException;
import java.nio.ByteBuffer;
import java.nio.charset.StandardCharsets;
import java.nio.file.*;

public class NioBufferDemo {

    static String state(ByteBuffer b) {
        return "pos=" + b.position() + " limit=" + b.limit() + " cap=" + b.capacity();
    }

    public static void main(String[] args) throws IOException {
        // 1. Vòng đời buffer: allocate -> ghi -> flip -> đọc
        System.out.println("== ByteBuffer lifecycle ==");
        ByteBuffer buf = ByteBuffer.allocate(16);
        System.out.println("mới allocate : " + state(buf));   // pos=0 limit=16 cap=16

        buf.put("hello".getBytes(StandardCharsets.UTF_8));     // ghi 5 byte
        System.out.println("sau khi ghi  : " + state(buf));    // pos=5 limit=16

        buf.flip();                                            // chuyển sang chế độ đọc
        System.out.println("sau flip()   : " + state(buf));    // pos=0 limit=5

        byte[] read = new byte[buf.remaining()];
        buf.get(read);
        System.out.println("đọc ra       : \"" + new String(read, StandardCharsets.UTF_8) + "\"");
        System.out.println("sau khi đọc  : " + state(buf));    // pos=5 limit=5

        // 2. Heap buffer vs direct buffer
        ByteBuffer direct = ByteBuffer.allocateDirect(16);
        System.out.println("\nheap buffer isDirect  = " + buf.isDirect());
        System.out.println("direct buffer isDirect = " + direct.isDirect()
                + "  (off-heap, giảm copy khi I/O)");

        // 3. java.nio.file: Path + Files (API hiện đại)
        System.out.println("\n== Files / Path ==");
        Path tmp = Files.createTempFile("nio-demo", ".txt");
        Files.writeString(tmp, "line1\nline2\n\nline4\n");
        try (var lines = Files.lines(tmp)) {                   // stream lazy, phải close
            long nonBlank = lines.filter(l -> !l.isBlank()).count();
            System.out.println("dòng không rỗng = " + nonBlank);
        }
        System.out.println("size = " + Files.size(tmp) + " byte");
        Files.delete(tmp);
        System.out.println("đã xoá file tạm");
    }
}
