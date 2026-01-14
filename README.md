# vecna

[![CI](https://github.com/zoobzio/vecna/actions/workflows/ci.yml/badge.svg)](https://github.com/zoobzio/vecna/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/zoobzio/vecna/branch/main/graph/badge.svg)](https://codecov.io/gh/zoobzio/vecna)
[![Go Report Card](https://goreportcard.com/badge/github.com/zoobzio/vecna)](https://goreportcard.com/report/github.com/zoobzio/vecna)
[![CodeQL](https://github.com/zoobzio/vecna/actions/workflows/codeql.yml/badge.svg)](https://github.com/zoobzio/vecna/actions/workflows/codeql.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/zoobzio/vecna.svg)](https://pkg.go.dev/github.com/zoobzio/vecna)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod-go-version/zoobzio/vecna)](https://github.com/zoobzio/vecna/blob/main/go.mod)
[![Release](https://img.shields.io/github/v/release/zoobzio/vecna)](https://github.com/zoobzio/vecna/releases)

Schema-validated filter builder for vector databases. Build type-safe metadata filters with compile-time field validation.

## Filters That Know Your Schema

```go
type DocumentMetadata struct {
    Category string   `json:"category"`
    Score    float64  `json:"score"`
    Tags     []string `json:"tags"`
}

builder, _ := vecna.New[DocumentMetadata]()

// Valid - field exists
filter := builder.Where("category").Eq("tech")

// Error - typo caught immediately
filter := builder.Where("categroy").Eq("tech")
// filter.Err() returns: vecna: field not found: categroy
```

## Install

```bash
go get github.com/zoobzio/vecna
```

Requires Go 1.24 or higher.

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/zoobzio/vecna"
)

type Metadata struct {
    Category string   `json:"category"`
    Score    float64  `json:"score"`
    Active   bool     `json:"active"`
    Tags     []string `json:"tags"`
}

func main() {
    // Create a schema-validated builder
    builder, err := vecna.New[Metadata]()
    if err != nil {
        panic(err)
    }

    // Build filters with validation
    filter := builder.And(
        builder.Where("category").Eq("tech"),
        builder.Or(
            builder.Where("score").Gte(0.8),
            builder.Where("active").Eq(true),
        ),
    )

    // Check for construction errors
    if err := filter.Err(); err != nil {
        panic(err)
    }

    // Use filter with your vector database provider
    fmt.Printf("Filter: %s %s %v\n", filter.Op(), filter.Field(), filter.Value())
}
```

## Capabilities

| Feature | Description | Docs |
|---------|-------------|------|
| Schema Validation | Field names validated against your struct at build time | [Concepts](docs/1.learn/3.concepts.md) |
| Type Checking | Comparison operators blocked on non-numeric fields | [Concepts](docs/1.learn/3.concepts.md) |
| Deferred Errors | Build complex filters, check errors once at the end | [Quickstart](docs/1.learn/2.quickstart.md) |
| Nested Logic | Combine filters with And/Or for complex queries | [API](docs/4.reference/1.api.md) |
| JSON Tags | Uses existing `json` tags - no new annotations needed | [Concepts](docs/1.learn/3.concepts.md) |
| Serializable Specs | Build filters from JSON for dynamic/LLM-generated queries | [Specs](docs/2.guides/1.specs.md) |

## Why vecna?

- **Catch typos early** — Field name errors surface when building filters, not at query time
- **Type safety** — Comparison operators only work on numeric fields
- **Zero new tags** — Uses your existing `json` struct tags
- **Provider agnostic** — Filter AST translates to any vector database

## Filters as Data

Filters can be built programmatically from serializable specs — enabling dynamic queries, stored filter configurations, and LLM-generated filters.

```go
// Filter spec from JSON (API request, config file, LLM output)
specJSON := `{
    "op": "and",
    "children": [
        {"op": "eq", "field": "category", "value": "tech"},
        {"op": "gte", "field": "score", "value": 0.8}
    ]
}`

var spec vecna.FilterSpec
json.Unmarshal([]byte(specJSON), &spec)

// Convert to validated filter
builder, _ := vecna.New[Metadata]()
filter := builder.FromSpec(&spec)

if err := filter.Err(); err != nil {
    // Schema validation still applies - invalid fields caught here
}
```

The spec is validated against your schema. Field typos and type mismatches are caught at conversion time, not when the query hits your database.

## Ecosystem

vecna powers the filter system in [grub](https://github.com/zoobzio/grub), providing schema-validated queries across vector database providers.

```go
// In grub - vecna filters translate to provider-specific syntax
store, _ := grub.NewPinecone[Document](index, grub.WithNamespace("docs"))

filter := store.Filter().And(
    store.Filter().Where("category").Eq("tech"),
    store.Filter().Where("score").Gte(0.8),
)

results, _ := store.Query(ctx, embedding, grub.WithFilter(filter))
```

Supported providers: Pinecone, Qdrant, Weaviate, Milvus, pgvector.

## Documentation

**Learn**
- [Overview](docs/1.learn/1.overview.md) — What vecna is and why
- [Quickstart](docs/1.learn/2.quickstart.md) — Get started in minutes
- [Concepts](docs/1.learn/3.concepts.md) — Core concepts and design
- [Architecture](docs/1.learn/4.architecture.md) — System design

**Guides**
- [Specs](docs/2.guides/1.specs.md) — Building filters from JSON
- [Testing](docs/2.guides/2.testing.md) — Testing with vecna
- [Troubleshooting](docs/2.guides/3.troubleshooting.md) — Common issues

**Reference**
- [API Reference](docs/4.reference/1.api.md) — Function documentation
- [Types Reference](docs/4.reference/2.types.md) — Type definitions
- [Operators](docs/4.reference/3.operators.md) — All filter operators
- [pkg.go.dev](https://pkg.go.dev/github.com/zoobzio/vecna) — Generated docs

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development workflow.

## License

MIT - see [LICENSE](LICENSE) for details.
