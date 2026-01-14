# Testing

Testing strategy for vecna.

## Structure

```
testing/
├── helpers.go          # Domain-specific test helpers
├── helpers_test.go     # Tests for helpers themselves
├── benchmarks/         # Performance tests
└── integration/        # End-to-end tests
```

## Running Tests

```bash
# All tests
make test

# Unit tests only (fast)
make test-unit

# Integration tests
make test-integration

# Benchmarks
make test-bench
```

## Test Helpers

The `testing` package provides domain-specific helpers:

- `NewTestBuilder(t)` - Creates a standard test builder
- `AssertNoError(t, f)` - Validates filter has no errors
- `AssertError(t, f)` - Validates filter has an error
- `AssertOp(t, f, op)` - Validates filter operator
- `AssertField(t, f, name)` - Validates filter field name
- `AssertChildCount(t, f, n)` - Validates child filter count

## Conventions

1. All helpers call `t.Helper()` for clean stack traces
2. All helpers accept `*testing.T` as first parameter
3. Helpers are domain-specific to vecna, not generic utilities
