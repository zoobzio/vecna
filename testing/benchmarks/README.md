# Benchmarks

Performance tests for vecna.

## Running

```bash
make test-bench
```

Or with more detail:

```bash
go test -bench=. -benchmem -benchtime=5s ./testing/benchmarks/...
```

## Benchmarks

| Benchmark | Description |
|-----------|-------------|
| `BenchmarkNew` | Builder creation with schema extraction |
| `BenchmarkWhereEq` | Simple equality filter construction |
| `BenchmarkWhereComparison` | Numeric comparison filter construction |
| `BenchmarkAnd` | Logical AND combination |
| `BenchmarkComplexFilter` | Nested filter tree construction |
| `BenchmarkFromSpec` | Filter construction from FilterSpec |
| `BenchmarkFilterErr` | Error checking on filter tree |

## Performance Notes

- Builder creation (`New[T]()`) involves reflection and should be done once per schema
- Filter construction is lightweight after builder creation
- `Err()` traverses the entire filter tree; cache results if checking multiple times
