// HashMap internals: hàm spread hash, phân phối bucket, và bad hashCode gây treeify/chậm.
// Chạy: java HashMapInternalsDemo.java
import java.util.*;

public class HashMapInternalsDemo {

    // Tái hiện hàm spread hash của HashMap (h ^ (h >>> 16)) và index bucket.
    static int spread(int h) { return h ^ (h >>> 16); }
    static int bucketIndex(int hash, int capacity) { return (capacity - 1) & spread(hash); }

    // Key có hashCode XẤU: mọi key rơi vào cùng 1 giá trị hash -> cùng bucket.
    record BadKey(int id) {
        @Override public int hashCode() { return 42; }   // hằng số -> collision tối đa
    }
    record GoodKey(int id) {
        // record tự sinh hashCode dựa id -> phân phối tốt
    }

    public static void main(String[] args) {
        // 1. Minh hoạ spread + index bucket
        System.out.println("== spread hash & bucket index (capacity=16) ==");
        for (int raw : new int[]{ 0x0000_0001, 0x0001_0001, 0x1234_0000, 0x1234_5678 }) {
            System.out.printf("hashCode=0x%08X  spread=0x%08X  bucket=%2d%n",
                    raw, spread(raw), bucketIndex(raw, 16));
        }
        System.out.println("-> 0x00010001 và 0x00000001 khác bit cao; nếu KHÔNG spread thì");
        System.out.println("   (15 & raw) đều = 1 (cùng bucket). Spread trộn bit cao xuống.");

        // 2. Phân phối bucket: good vs bad hashCode
        int n = 50_000;
        int[] goodDist = distribution(new GoodKeyFactory(), n, 1 << 16);
        int[] badDist  = distribution(new BadKeyFactory(),  n, 1 << 16);
        System.out.printf("%n== Phân phối %d key vào 65536 bucket ==%n", n);
        System.out.println("GoodKey: max bucket = " + max(goodDist) + " phần tử (phân phối đều)");
        System.out.println("BadKey : max bucket = " + max(badDist)  + " phần tử (dồn 1 bucket -> treeify)");

        // 3. Hiệu năng: 100k get trên map key tốt vs key xấu
        Map<GoodKey, Integer> good = new HashMap<>();
        Map<BadKey, Integer> bad = new HashMap<>();
        for (int i = 0; i < 20_000; i++) { good.put(new GoodKey(i), i); bad.put(new BadKey(i), i); }
        long tg = timeGet(() -> { for (int i = 0; i < 20_000; i++) good.get(new GoodKey(i)); });
        long tb = timeGet(() -> { for (int i = 0; i < 20_000; i++) bad.get(new BadKey(i)); });
        System.out.printf("%n20k get GoodKey: %d ms | BadKey: %d ms%n", tg, tb);
        System.out.println("-> BadKey chậm hơn nhiều dù đã treeify (O(log n) thay O(1)).");
    }

    interface KeyFactory { int hashOf(int i); }
    static class GoodKeyFactory implements KeyFactory { public int hashOf(int i){ return new GoodKey(i).hashCode(); } }
    static class BadKeyFactory  implements KeyFactory { public int hashOf(int i){ return new BadKey(i).hashCode(); } }

    static int[] distribution(KeyFactory f, int n, int buckets) {
        int[] dist = new int[buckets];
        for (int i = 0; i < n; i++) dist[bucketIndex(f.hashOf(i), buckets)]++;
        return dist;
    }
    static int max(int[] a) { int m = 0; for (int x : a) m = Math.max(m, x); return m; }
    static long timeGet(Runnable r) { long t0 = System.nanoTime(); r.run(); return (System.nanoTime()-t0)/1_000_000; }
}
