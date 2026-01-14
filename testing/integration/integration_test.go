package integration

import (
	"encoding/json"
	"testing"

	"github.com/zoobzio/vecna"
)

// Integration tests validate end-to-end filter construction workflows.

type ProductMetadata struct {
	Name     string   `json:"name"`
	Price    float64  `json:"price"`
	InStock  bool     `json:"in_stock"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
	Rating   int      `json:"rating"`
}

func TestIntegration_FullWorkflow(t *testing.T) {
	// Create builder
	builder, err := vecna.New[ProductMetadata]()
	if err != nil {
		t.Fatalf("failed to create builder: %v", err)
	}

	// Build complex filter
	filter := builder.And(
		builder.Where("category").Eq("electronics"),
		builder.Where("price").Lte(1000.0),
		builder.Or(
			builder.Where("in_stock").Eq(true),
			builder.Where("rating").Gte(4),
		),
	)

	// Validate no errors
	if err := filter.Err(); err != nil {
		t.Fatalf("unexpected filter error: %v", err)
	}

	// Validate structure
	if filter.Op() != vecna.And {
		t.Errorf("expected And operator, got %s", filter.Op())
	}
	if len(filter.Children()) != 3 {
		t.Errorf("expected 3 children, got %d", len(filter.Children()))
	}
}

func TestIntegration_SpecRoundTrip(t *testing.T) {
	builder, err := vecna.New[ProductMetadata]()
	if err != nil {
		t.Fatalf("failed to create builder: %v", err)
	}

	// Create spec from JSON (simulating external input)
	specJSON := `{
		"op": "and",
		"children": [
			{"op": "eq", "field": "category", "value": "books"},
			{"op": "lt", "field": "price", "value": 50}
		]
	}`

	var spec vecna.FilterSpec
	if err := json.Unmarshal([]byte(specJSON), &spec); err != nil {
		t.Fatalf("failed to parse spec JSON: %v", err)
	}

	// Convert to filter
	filter := builder.FromSpec(&spec)
	if err := filter.Err(); err != nil {
		t.Fatalf("unexpected filter error: %v", err)
	}

	// Validate
	if filter.Op() != vecna.And {
		t.Errorf("expected And, got %s", filter.Op())
	}
	children := filter.Children()
	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}
	if children[0].Field() != "category" {
		t.Errorf("expected field 'category', got %q", children[0].Field())
	}
}

func TestIntegration_ErrorPropagation(t *testing.T) {
	builder, err := vecna.New[ProductMetadata]()
	if err != nil {
		t.Fatalf("failed to create builder: %v", err)
	}

	// Build filter with invalid field buried in tree
	filter := builder.And(
		builder.Where("category").Eq("electronics"),
		builder.Or(
			builder.Where("invalid_field").Eq("value"), // Error here
			builder.Where("in_stock").Eq(true),
		),
	)

	// Error should propagate up
	if filter.Err() == nil {
		t.Error("expected error for invalid field, got nil")
	}
}

func TestIntegration_TypeValidation(t *testing.T) {
	builder, err := vecna.New[ProductMetadata]()
	if err != nil {
		t.Fatalf("failed to create builder: %v", err)
	}

	// Comparison operator on non-numeric field should error
	filter := builder.Where("category").Gt("electronics")
	if filter.Err() == nil {
		t.Error("expected error for comparison on string field, got nil")
	}

	// Comparison operator on numeric field should succeed
	filter = builder.Where("price").Gt(100.0)
	if err := filter.Err(); err != nil {
		t.Errorf("unexpected error for comparison on numeric field: %v", err)
	}
}

func TestIntegration_InOperator(t *testing.T) {
	builder, err := vecna.New[ProductMetadata]()
	if err != nil {
		t.Fatalf("failed to create builder: %v", err)
	}

	filter := builder.Where("category").In("electronics", "books", "clothing")
	if err := filter.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if filter.Op() != vecna.In {
		t.Errorf("expected In operator, got %s", filter.Op())
	}
}
