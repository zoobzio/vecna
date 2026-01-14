package vecna

import (
	"errors"
	"testing"

	"github.com/zoobzio/sentinel"
)

// Test metadata struct.
type testMetadata struct {
	Category string   `json:"category"`
	Score    float64  `json:"score"`
	Count    int      `json:"count"`
	Active   bool     `json:"active"`
	Tags     []string `json:"tags"`
	Internal string   `json:"-"`
	NoTag    string
}

func TestNew(t *testing.T) {
	t.Run("valid struct", func(t *testing.T) {
		builder, err := New[testMetadata]()
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		if builder == nil {
			t.Fatal("New() returned nil")
		}
	})

	t.Run("non-struct type", func(t *testing.T) {
		_, err := New[string]()
		if !errors.Is(err, ErrNotStruct) {
			t.Errorf("New[string]() error = %v, want %v", err, ErrNotStruct)
		}
	})
}

func TestBuilder_Spec(t *testing.T) {
	builder, err := New[testMetadata]()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	spec := builder.Spec()

	if spec.TypeName != "testMetadata" {
		t.Errorf("Spec.TypeName = %v, want testMetadata", spec.TypeName)
	}

	// Check expected fields
	expectedFields := map[string]FieldKind{
		"category": KindString,
		"score":    KindFloat,
		"count":    KindInt,
		"active":   KindBool,
		"tags":     KindSlice,
		"NoTag":    KindString, // No json tag, uses Go name
	}

	for name, expectedKind := range expectedFields {
		field := spec.Field(name)
		if field == nil {
			t.Errorf("Field %q not found in spec", name)
			continue
		}
		if field.Kind != expectedKind {
			t.Errorf("Field %q Kind = %v, want %v", name, field.Kind, expectedKind)
		}
	}

	// Internal should be excluded (json:"-")
	if spec.Field("Internal") != nil {
		t.Error("Field 'Internal' should be excluded")
	}
	if spec.Field("-") != nil {
		t.Error("Field '-' should not exist")
	}
}

func TestBuilder_Where(t *testing.T) {
	builder, _ := New[testMetadata]()

	t.Run("valid field", func(t *testing.T) {
		fb := builder.Where("category")
		if fb.err != nil {
			t.Errorf("Where() error = %v", fb.err)
		}
		if fb.field != "category" {
			t.Errorf("FieldBuilder.field = %v, want category", fb.field)
		}
	})

	t.Run("invalid field", func(t *testing.T) {
		fb := builder.Where("nonexistent")
		if fb.err == nil {
			t.Error("Where() should have error for nonexistent field")
		}
		if !errors.Is(fb.err, ErrFieldNotFound) {
			t.Errorf("Where() error = %v, want %v", fb.err, ErrFieldNotFound)
		}
	})
}

func TestFieldBuilder_Operators(t *testing.T) {
	builder, _ := New[testMetadata]()

	tests := []struct {
		name    string
		filter  *Filter
		wantOp  Op
		wantVal any
	}{
		{"Eq", builder.Where("category").Eq("tech"), Eq, "tech"},
		{"Ne", builder.Where("category").Ne("other"), Ne, "other"},
		{"Gt", builder.Where("score").Gt(0.5), Gt, 0.5},
		{"Gte", builder.Where("score").Gte(0.5), Gte, 0.5},
		{"Lt", builder.Where("count").Lt(10), Lt, 10},
		{"Lte", builder.Where("count").Lte(10), Lte, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.filter.Op() != tt.wantOp {
				t.Errorf("Filter.Op() = %v, want %v", tt.filter.Op(), tt.wantOp)
			}
			if tt.filter.Field() != "category" && tt.filter.Field() != "score" && tt.filter.Field() != "count" {
				t.Errorf("Filter.Field() = %v, unexpected", tt.filter.Field())
			}
			if tt.filter.Err() != nil {
				t.Errorf("Filter.Err() = %v, want nil", tt.filter.Err())
			}
		})
	}
}

func TestFieldBuilder_In(t *testing.T) {
	builder, _ := New[testMetadata]()

	filter := builder.Where("category").In("tech", "science", "art")

	if filter.Op() != In {
		t.Errorf("Filter.Op() = %v, want %v", filter.Op(), In)
	}
	if filter.Field() != "category" {
		t.Errorf("Filter.Field() = %v, want category", filter.Field())
	}

	values, ok := filter.Value().([]any)
	if !ok {
		t.Fatalf("Filter.Value() type = %T, want []any", filter.Value())
	}
	if len(values) != 3 {
		t.Errorf("len(Filter.Value()) = %v, want 3", len(values))
	}
}

func TestFieldBuilder_InvalidField(t *testing.T) {
	builder, _ := New[testMetadata]()

	filter := builder.Where("nonexistent").Eq("value")

	if filter.Err() == nil {
		t.Error("Filter.Err() should not be nil for invalid field")
	}
	if !errors.Is(filter.Err(), ErrFieldNotFound) {
		t.Errorf("Filter.Err() = %v, want %v", filter.Err(), ErrFieldNotFound)
	}
}

func TestFieldBuilder_ComparisonOnString(t *testing.T) {
	builder, _ := New[testMetadata]()

	// Comparison operators should error on string fields
	filter := builder.Where("category").Gt("value")

	if filter.Err() == nil {
		t.Error("Gt on string field should have error")
	}
	if !errors.Is(filter.Err(), ErrInvalidFilter) {
		t.Errorf("Filter.Err() = %v, want %v", filter.Err(), ErrInvalidFilter)
	}
}

func TestBuilder_And(t *testing.T) {
	builder, _ := New[testMetadata]()

	f1 := builder.Where("category").Eq("tech")
	f2 := builder.Where("score").Gte(0.5)

	combined := builder.And(f1, f2)

	if combined.Op() != And {
		t.Errorf("And Filter.Op() = %v, want %v", combined.Op(), And)
	}
	if len(combined.Children()) != 2 {
		t.Errorf("len(And Filter.Children()) = %v, want 2", len(combined.Children()))
	}
	if combined.Err() != nil {
		t.Errorf("And Filter.Err() = %v, want nil", combined.Err())
	}
}

func TestBuilder_Or(t *testing.T) {
	builder, _ := New[testMetadata]()

	f1 := builder.Where("category").Eq("tech")
	f2 := builder.Where("category").Eq("science")

	combined := builder.Or(f1, f2)

	if combined.Op() != Or {
		t.Errorf("Or Filter.Op() = %v, want %v", combined.Op(), Or)
	}
	if len(combined.Children()) != 2 {
		t.Errorf("len(Or Filter.Children()) = %v, want 2", len(combined.Children()))
	}
}

func TestBuilder_NestedFilters(t *testing.T) {
	builder, _ := New[testMetadata]()

	// category == "tech" AND (score >= 0.5 OR active == true)
	filter := builder.And(
		builder.Where("category").Eq("tech"),
		builder.Or(
			builder.Where("score").Gte(0.5),
			builder.Where("active").Eq(true),
		),
	)

	if filter.Op() != And {
		t.Errorf("Filter.Op() = %v, want %v", filter.Op(), And)
	}
	if len(filter.Children()) != 2 {
		t.Errorf("len(Filter.Children()) = %v, want 2", len(filter.Children()))
	}

	orChild := filter.Children()[1]
	if orChild.Op() != Or {
		t.Errorf("Child Filter.Op() = %v, want %v", orChild.Op(), Or)
	}
	if filter.Err() != nil {
		t.Errorf("Nested Filter.Err() = %v, want nil", filter.Err())
	}
}

func TestBuilder_NestedFilterWithError(t *testing.T) {
	builder, _ := New[testMetadata]()

	// Nested filter with invalid field
	filter := builder.And(
		builder.Where("category").Eq("tech"),
		builder.Where("invalid").Eq("value"),
	)

	if filter.Err() == nil {
		t.Error("Nested filter with invalid field should have error")
	}
	if !errors.Is(filter.Err(), ErrFieldNotFound) {
		t.Errorf("Filter.Err() = %v, want %v", filter.Err(), ErrFieldNotFound)
	}
}

func TestResolveFieldKind(t *testing.T) {
	tests := []struct {
		typeName string
		want     FieldKind
	}{
		{"int", KindInt},
		{"int64", KindInt},
		{"uint32", KindInt},
		{"float64", KindFloat},
		{"float32", KindFloat},
		{"bool", KindBool},
		{"string", KindString},
		{"complex128", KindUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			got := resolveFieldKind(sentinel.KindScalar, tt.typeName)
			if got != tt.want {
				t.Errorf("resolveFieldKind(%q) = %v, want %v", tt.typeName, got, tt.want)
			}
		})
	}
}

func TestResolveFieldKind_Slice(t *testing.T) {
	got := resolveFieldKind(sentinel.KindSlice, "[]string")
	if got != KindSlice {
		t.Errorf("resolveFieldKind(slice) = %v, want %v", got, KindSlice)
	}
}

func TestResolveFieldKind_UnknownKind(t *testing.T) {
	// Test outer default branch for unsupported kinds (struct, map, etc.)
	tests := []sentinel.FieldKind{
		sentinel.KindStruct,
		sentinel.KindMap,
		sentinel.KindPointer,
		sentinel.KindInterface,
	}

	for _, kind := range tests {
		t.Run(string(kind), func(t *testing.T) {
			got := resolveFieldKind(kind, "SomeType")
			if got != KindUnknown {
				t.Errorf("resolveFieldKind(%s) = %v, want %v", kind, got, KindUnknown)
			}
		})
	}
}
