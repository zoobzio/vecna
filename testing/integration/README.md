# Integration Tests

End-to-end tests for vecna.

## Running

```bash
make test-integration
```

## Test Coverage

| Test | Description |
|------|-------------|
| `TestIntegration_FullWorkflow` | Complete filter construction and validation |
| `TestIntegration_SpecRoundTrip` | JSON spec parsing and filter conversion |
| `TestIntegration_ErrorPropagation` | Error bubbling through nested filters |
| `TestIntegration_TypeValidation` | Operator/type compatibility checks |
| `TestIntegration_InOperator` | Set membership filter construction |

## Writing Integration Tests

Integration tests should:

1. Test complete workflows, not isolated functions
2. Use realistic schema types
3. Validate both success and error paths
4. Test edge cases in filter tree construction

## Naming Convention

All integration tests use the `TestIntegration_` prefix to enable targeted execution:

```bash
go test -run Integration ./...
```
