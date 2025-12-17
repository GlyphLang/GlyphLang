#!/usr/bin/env python3
"""
Python Benchmark Suite for comparison with Glyph Language
Run: python bench_python.py
"""

import time
import json
import statistics
from typing import Dict, List, Any

# Number of iterations for each benchmark
ITERATIONS = 1_000_000
WARMUP = 10_000

def benchmark(name: str, func, iterations: int = ITERATIONS):
    """Run a benchmark and return results"""
    # Warmup
    for _ in range(WARMUP):
        func()

    # Actual benchmark
    times = []
    for _ in range(10):  # 10 runs
        start = time.perf_counter_ns()
        for _ in range(iterations):
            func()
        end = time.perf_counter_ns()
        times.append((end - start) / iterations)

    avg = statistics.mean(times)
    std = statistics.stdev(times) if len(times) > 1 else 0

    return {
        'name': name,
        'avg_ns': avg,
        'std_ns': std,
        'iterations': iterations
    }


# Benchmark functions

def bench_arithmetic():
    """Simple arithmetic: (10 + 20) * (30 - 5)"""
    a, b, c, d = 10, 20, 30, 5
    result = (a + b) * (c - d)
    return result

def bench_string_concat():
    """String concatenation"""
    s1 = "Hello, "
    s2 = "World!"
    result = s1 + s2
    return result

def bench_dict_creation():
    """Create a dictionary (object)"""
    obj = {
        'id': 1,
        'name': 'John Doe',
        'email': 'john@example.com',
        'active': True
    }
    return obj

def bench_list_creation():
    """Create a list with 10 elements"""
    arr = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]
    return arr

def bench_list_iteration():
    """Iterate over a list and sum"""
    arr = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
    total = 0
    for x in arr:
        total += x
    return total

def bench_function_call():
    """Function call overhead"""
    def add(a, b):
        return a + b
    return add(10, 20)

def bench_conditional():
    """Conditional branching"""
    x = 42
    if x > 20:
        result = "large"
    else:
        result = "small"
    return result

def bench_json_serialize():
    """JSON serialization (API response simulation)"""
    obj = {
        'status': 'ok',
        'data': {
            'id': 123,
            'name': 'Test User',
            'email': 'test@example.com'
        }
    }
    return json.dumps(obj)

def bench_json_parse():
    """JSON parsing (API request simulation)"""
    json_str = '{"id": 123, "name": "Test User", "email": "test@example.com"}'
    return json.loads(json_str)

def bench_comparison():
    """Comparison operations"""
    a, b = 10, 20
    r1 = a < b
    r2 = a == b
    r3 = a != b
    return r1 and r3 and not r2

def bench_boolean_logic():
    """Boolean logic operations"""
    a, b, c = True, False, True
    result = (a and c) or (not b and c)
    return result

def bench_variable_access():
    """Variable read/write"""
    x = 42
    y = x
    x = y + 1
    return x

def bench_complex_expression():
    """Complex nested expression"""
    a, b, c, d, e = 10, 20, 30, 40, 50
    result = ((a + b) * c - d) / e + (a * b) - (c / d)
    return result

# HTTP route simulation
class Request:
    def __init__(self, path: str, params: Dict[str, Any] = None):
        self.path = path
        self.params = params or {}

def bench_route_handler():
    """Simulate a simple route handler"""
    req = Request('/api/users/123', {'id': '123'})
    user_id = int(req.params.get('id', 0))
    response = {
        'id': user_id,
        'name': 'User ' + str(user_id),
        'email': f'user{user_id}@example.com'
    }
    return response


def run_all_benchmarks():
    """Run all benchmarks and print results"""
    print("=" * 70)
    print("PYTHON BENCHMARK SUITE")
    print("=" * 70)
    print(f"Python version: {__import__('sys').version.split()[0]}")
    print(f"Iterations per benchmark: {ITERATIONS:,}")
    print("-" * 70)
    print(f"{'Benchmark':<35} {'Avg (ns)':<15} {'Std (ns)':<15}")
    print("-" * 70)

    benchmarks = [
        ('Arithmetic', bench_arithmetic),
        ('String Concat', bench_string_concat),
        ('Dict Creation', bench_dict_creation),
        ('List Creation', bench_list_creation),
        ('List Iteration', bench_list_iteration),
        ('Function Call', bench_function_call),
        ('Conditional', bench_conditional),
        ('JSON Serialize', bench_json_serialize, 100_000),
        ('JSON Parse', bench_json_parse, 100_000),
        ('Comparison', bench_comparison),
        ('Boolean Logic', bench_boolean_logic),
        ('Variable Access', bench_variable_access),
        ('Complex Expression', bench_complex_expression),
        ('Route Handler', bench_route_handler, 100_000),
    ]

    results = []
    for item in benchmarks:
        name = item[0]
        func = item[1]
        iters = item[2] if len(item) > 2 else ITERATIONS

        result = benchmark(name, func, iters)
        results.append(result)
        print(f"{result['name']:<35} {result['avg_ns']:<15.2f} {result['std_ns']:<15.2f}")

    print("-" * 70)
    print("\nJSON Output for comparison:")
    print(json.dumps(results, indent=2))

    return results


if __name__ == '__main__':
    run_all_benchmarks()
