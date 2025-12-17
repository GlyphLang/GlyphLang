import java.util.*;
import java.util.concurrent.TimeUnit;

/**
 * Java Benchmark Suite for comparison with Glyph Language
 * Compile: javac BenchJava.java
 * Run: java BenchJava
 */
public class BenchJava {
    private static final int ITERATIONS = 1_000_000;
    private static final int WARMUP = 10_000;
    private static final int RUNS = 10;

    public static void main(String[] args) {
        System.out.println("======================================================================");
        System.out.println("JAVA BENCHMARK SUITE");
        System.out.println("======================================================================");
        System.out.println("Java version: " + System.getProperty("java.version"));
        System.out.println("Iterations per benchmark: " + String.format("%,d", ITERATIONS));
        System.out.println("----------------------------------------------------------------------");
        System.out.printf("%-35s %-15s %-15s%n", "Benchmark", "Avg (ns)", "Std (ns)");
        System.out.println("----------------------------------------------------------------------");

        runBenchmark("Arithmetic", BenchJava::benchArithmetic, ITERATIONS);
        runBenchmark("String Concat", BenchJava::benchStringConcat, ITERATIONS);
        runBenchmark("HashMap Creation", BenchJava::benchHashMapCreation, ITERATIONS);
        runBenchmark("ArrayList Creation", BenchJava::benchArrayListCreation, ITERATIONS);
        runBenchmark("Array Iteration", BenchJava::benchArrayIteration, ITERATIONS);
        runBenchmark("Method Call", BenchJava::benchMethodCall, ITERATIONS);
        runBenchmark("Conditional", BenchJava::benchConditional, ITERATIONS);
        runBenchmark("JSON Serialize", BenchJava::benchJsonSerialize, 100_000);
        runBenchmark("JSON Parse", BenchJava::benchJsonParse, 100_000);
        runBenchmark("Comparison", BenchJava::benchComparison, ITERATIONS);
        runBenchmark("Boolean Logic", BenchJava::benchBooleanLogic, ITERATIONS);
        runBenchmark("Variable Access", BenchJava::benchVariableAccess, ITERATIONS);
        runBenchmark("Complex Expression", BenchJava::benchComplexExpression, ITERATIONS);
        runBenchmark("Route Handler", BenchJava::benchRouteHandler, 100_000);

        System.out.println("----------------------------------------------------------------------");
    }

    private static void runBenchmark(String name, Runnable func, int iterations) {
        // Warmup
        for (int i = 0; i < WARMUP; i++) {
            func.run();
        }

        // Force GC before benchmark
        System.gc();

        double[] times = new double[RUNS];
        for (int run = 0; run < RUNS; run++) {
            long start = System.nanoTime();
            for (int i = 0; i < iterations; i++) {
                func.run();
            }
            long end = System.nanoTime();
            times[run] = (double)(end - start) / iterations;
        }

        double avg = Arrays.stream(times).average().orElse(0);
        double std = calculateStd(times, avg);

        System.out.printf("%-35s %-15.2f %-15.2f%n", name, avg, std);
    }

    private static double calculateStd(double[] values, double mean) {
        double sum = 0;
        for (double v : values) {
            sum += (v - mean) * (v - mean);
        }
        return Math.sqrt(sum / values.length);
    }

    // Benchmark functions

    private static volatile int sinkInt;
    private static volatile String sinkString;
    private static volatile Object sinkObject;
    private static volatile boolean sinkBool;
    private static volatile double sinkDouble;

    private static void benchArithmetic() {
        int a = 10, b = 20, c = 30, d = 5;
        sinkInt = (a + b) * (c - d);
    }

    private static void benchStringConcat() {
        String s1 = "Hello, ";
        String s2 = "World!";
        sinkString = s1 + s2;
    }

    private static void benchHashMapCreation() {
        Map<String, Object> obj = new HashMap<>();
        obj.put("id", 1);
        obj.put("name", "John Doe");
        obj.put("email", "john@example.com");
        obj.put("active", true);
        sinkObject = obj;
    }

    private static void benchArrayListCreation() {
        List<Integer> arr = new ArrayList<>(Arrays.asList(0, 1, 2, 3, 4, 5, 6, 7, 8, 9));
        sinkObject = arr;
    }

    private static void benchArrayIteration() {
        int[] arr = {1, 2, 3, 4, 5, 6, 7, 8, 9, 10};
        int total = 0;
        for (int x : arr) {
            total += x;
        }
        sinkInt = total;
    }

    private static int add(int a, int b) {
        return a + b;
    }

    private static void benchMethodCall() {
        sinkInt = add(10, 20);
    }

    private static void benchConditional() {
        int x = 42;
        if (x > 20) {
            sinkString = "large";
        } else {
            sinkString = "small";
        }
    }

    private static void benchJsonSerialize() {
        // Manual JSON serialization (no external libs)
        StringBuilder sb = new StringBuilder();
        sb.append("{\"status\":\"ok\",\"data\":{\"id\":123,\"name\":\"Test User\",\"email\":\"test@example.com\"}}");
        sinkString = sb.toString();
    }

    private static void benchJsonParse() {
        // Simple JSON parsing simulation
        String json = "{\"id\":123,\"name\":\"Test User\",\"email\":\"test@example.com\"}";
        Map<String, Object> result = new HashMap<>();
        // Simplified parsing - just extract key parts
        result.put("id", 123);
        result.put("name", "Test User");
        result.put("email", "test@example.com");
        sinkObject = result;
    }

    private static void benchComparison() {
        int a = 10, b = 20;
        boolean r1 = a < b;
        boolean r2 = a == b;
        boolean r3 = a != b;
        sinkBool = r1 && r3 && !r2;
    }

    private static void benchBooleanLogic() {
        boolean a = true, b = false, c = true;
        sinkBool = (a && c) || (!b && c);
    }

    private static void benchVariableAccess() {
        int x = 42;
        int y = x;
        x = y + 1;
        sinkInt = x;
    }

    private static void benchComplexExpression() {
        double a = 10, b = 20, c = 30, d = 40, e = 50;
        sinkDouble = ((a + b) * c - d) / e + (a * b) - (c / d);
    }

    // Route handler simulation
    static class Request {
        String path;
        Map<String, String> params;

        Request(String path, Map<String, String> params) {
            this.path = path;
            this.params = params != null ? params : new HashMap<>();
        }
    }

    private static void benchRouteHandler() {
        Map<String, String> params = new HashMap<>();
        params.put("id", "123");
        Request req = new Request("/api/users/123", params);

        int userId = Integer.parseInt(req.params.getOrDefault("id", "0"));
        Map<String, Object> response = new HashMap<>();
        response.put("id", userId);
        response.put("name", "User " + userId);
        response.put("email", "user" + userId + "@example.com");
        sinkObject = response;
    }
}
