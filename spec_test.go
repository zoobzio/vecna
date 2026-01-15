package vecna

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestBuilder_FromSpec_SimpleConditions(t *testing.T) {
	builder, _ := New[testMetadata]()

	tests := []struct {
		name      string
		spec      *FilterSpec
		wantOp    Op
		wantField string
		wantValue any
	}{
		{
			name:      "eq",
			spec:      &FilterSpec{Op: "eq", Field: "category", Value: "tech"},
			wantOp:    Eq,
			wantField: "category",
			wantValue: "tech",
		},
		{
			name:      "ne",
			spec:      &FilterSpec{Op: "ne", Field: "category", Value: "other"},
			wantOp:    Ne,
			wantField: "category",
			wantValue: "other",
		},
		{
			name:      "gt",
			spec:      &FilterSpec{Op: "gt", Field: "score", Value: 0.5},
			wantOp:    Gt,
			wantField: "score",
			wantValue: 0.5,
		},
		{
			name:      "gte",
			spec:      &FilterSpec{Op: "gte", Field: "score", Value: 0.5},
			wantOp:    Gte,
			wantField: "score",
			wantValue: 0.5,
		},
		{
			name:      "lt",
			spec:      &FilterSpec{Op: "lt", Field: "count", Value: 10},
			wantOp:    Lt,
			wantField: "count",
			wantValue: 10,
		},
		{
			name:      "lte",
			spec:      &FilterSpec{Op: "lte", Field: "count", Value: 10},
			wantOp:    Lte,
			wantField: "count",
			wantValue: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := builder.FromSpec(tt.spec)

			if filter.Op() != tt.wantOp {
				t.Errorf("Filter.Op() = %v, want %v", filter.Op(), tt.wantOp)
			}
			if filter.Field() != tt.wantField {
				t.Errorf("Filter.Field() = %v, want %v", filter.Field(), tt.wantField)
			}
			if filter.Err() != nil {
				t.Errorf("Filter.Err() = %v, want nil", filter.Err())
			}
		})
	}
}

func TestBuilder_FromSpec_In(t *testing.T) {
	builder, _ := New[testMetadata]()

	spec := &FilterSpec{
		Op:    "in",
		Field: "category",
		Value: []any{"tech", "science", "art"},
	}

	filter := builder.FromSpec(spec)

	if filter.Op() != In {
		t.Errorf("Filter.Op() = %v, want %v", filter.Op(), In)
	}
	if filter.Field() != "category" {
		t.Errorf("Filter.Field() = %v, want category", filter.Field())
	}
	if filter.Err() != nil {
		t.Errorf("Filter.Err() = %v, want nil", filter.Err())
	}

	values, ok := filter.Value().([]any)
	if !ok {
		t.Fatalf("Filter.Value() type = %T, want []any", filter.Value())
	}
	if len(values) != 3 {
		t.Errorf("len(Filter.Value()) = %v, want 3", len(values))
	}
}

func TestBuilder_FromSpec_In_TypedSlice(t *testing.T) {
	builder, _ := New[testMetadata]()

	// Typed slice (not []any) - exercises the else branch in fromInSpec
	spec := &FilterSpec{
		Op:    "in",
		Field: "category",
		Value: []string{"tech", "science"},
	}

	filter := builder.FromSpec(spec)

	if filter.Op() != In {
		t.Errorf("Filter.Op() = %v, want %v", filter.Op(), In)
	}
	if filter.Err() != nil {
		t.Errorf("Filter.Err() = %v, want nil", filter.Err())
	}
}

func TestBuilder_FromSpec_And(t *testing.T) {
	builder, _ := New[testMetadata]()

	spec := &FilterSpec{
		Op: "and",
		Children: []*FilterSpec{
			{Op: "eq", Field: "category", Value: "tech"},
			{Op: "gte", Field: "score", Value: 0.5},
		},
	}

	filter := builder.FromSpec(spec)

	if filter.Op() != And {
		t.Errorf("Filter.Op() = %v, want %v", filter.Op(), And)
	}
	if len(filter.Children()) != 2 {
		t.Errorf("len(Filter.Children()) = %v, want 2", len(filter.Children()))
	}
	if filter.Err() != nil {
		t.Errorf("Filter.Err() = %v, want nil", filter.Err())
	}
}

func TestBuilder_FromSpec_Or(t *testing.T) {
	builder, _ := New[testMetadata]()

	spec := &FilterSpec{
		Op: "or",
		Children: []*FilterSpec{
			{Op: "eq", Field: "category", Value: "tech"},
			{Op: "eq", Field: "category", Value: "science"},
		},
	}

	filter := builder.FromSpec(spec)

	if filter.Op() != Or {
		t.Errorf("Filter.Op() = %v, want %v", filter.Op(), Or)
	}
	if len(filter.Children()) != 2 {
		t.Errorf("len(Filter.Children()) = %v, want 2", len(filter.Children()))
	}
	if filter.Err() != nil {
		t.Errorf("Filter.Err() = %v, want nil", filter.Err())
	}
}

func TestBuilder_FromSpec_Nested(t *testing.T) {
	builder, _ := New[testMetadata]()

	// category == "tech" AND (score >= 0.5 OR active == true)
	spec := &FilterSpec{
		Op: "and",
		Children: []*FilterSpec{
			{Op: "eq", Field: "category", Value: "tech"},
			{
				Op: "or",
				Children: []*FilterSpec{
					{Op: "gte", Field: "score", Value: 0.5},
					{Op: "eq", Field: "active", Value: true},
				},
			},
		},
	}

	filter := builder.FromSpec(spec)

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
		t.Errorf("Filter.Err() = %v, want nil", filter.Err())
	}
}

func TestBuilder_FromSpec_JSON(t *testing.T) {
	builder, _ := New[testMetadata]()

	jsonSpec := `{
		"op": "and",
		"children": [
			{"op": "eq", "field": "category", "value": "tech"},
			{"op": "gte", "field": "score", "value": 0.8}
		]
	}`

	var spec FilterSpec
	if err := json.Unmarshal([]byte(jsonSpec), &spec); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	filter := builder.FromSpec(&spec)

	if filter.Op() != And {
		t.Errorf("Filter.Op() = %v, want %v", filter.Op(), And)
	}
	if len(filter.Children()) != 2 {
		t.Errorf("len(Filter.Children()) = %v, want 2", len(filter.Children()))
	}
	if filter.Err() != nil {
		t.Errorf("Filter.Err() = %v, want nil", filter.Err())
	}

	// Verify first child
	first := filter.Children()[0]
	if first.Op() != Eq {
		t.Errorf("First child Op() = %v, want %v", first.Op(), Eq)
	}
	if first.Field() != "category" {
		t.Errorf("First child Field() = %v, want category", first.Field())
	}
}

func TestBuilder_FromSpec_InvalidField(t *testing.T) {
	builder, _ := New[testMetadata]()

	spec := &FilterSpec{Op: "eq", Field: "nonexistent", Value: "value"}

	filter := builder.FromSpec(spec)

	if filter.Err() == nil {
		t.Error("Filter.Err() should not be nil for invalid field")
	}
	if !errors.Is(filter.Err(), ErrFieldNotFound) {
		t.Errorf("Filter.Err() = %v, want %v", filter.Err(), ErrFieldNotFound)
	}
}

func TestBuilder_FromSpec_InvalidOperator(t *testing.T) {
	builder, _ := New[testMetadata]()

	spec := &FilterSpec{Op: "invalid", Field: "category", Value: "value"}

	filter := builder.FromSpec(spec)

	if filter.Err() == nil {
		t.Error("Filter.Err() should not be nil for invalid operator")
	}
	if !errors.Is(filter.Err(), ErrInvalidFilter) {
		t.Errorf("Filter.Err() = %v, want %v", filter.Err(), ErrInvalidFilter)
	}
}

func TestBuilder_FromSpec_NilSpec(t *testing.T) {
	builder, _ := New[testMetadata]()

	filter := builder.FromSpec(nil)

	if filter.Err() == nil {
		t.Error("Filter.Err() should not be nil for nil spec")
	}
	if !errors.Is(filter.Err(), ErrInvalidFilter) {
		t.Errorf("Filter.Err() = %v, want %v", filter.Err(), ErrInvalidFilter)
	}
}

func TestBuilder_FromSpec_EmptyAndChildren(t *testing.T) {
	builder, _ := New[testMetadata]()

	spec := &FilterSpec{Op: "and", Children: []*FilterSpec{}}

	filter := builder.FromSpec(spec)

	if filter.Err() == nil {
		t.Error("Filter.Err() should not be nil for empty children")
	}
	if !errors.Is(filter.Err(), ErrInvalidFilter) {
		t.Errorf("Filter.Err() = %v, want %v", filter.Err(), ErrInvalidFilter)
	}
}

func TestBuilder_FromSpec_NestedError(t *testing.T) {
	builder, _ := New[testMetadata]()

	spec := &FilterSpec{
		Op: "and",
		Children: []*FilterSpec{
			{Op: "eq", Field: "category", Value: "tech"},
			{Op: "eq", Field: "invalid", Value: "value"},
		},
	}

	filter := builder.FromSpec(spec)

	if filter.Err() == nil {
		t.Error("Filter.Err() should not be nil for nested invalid field")
	}
	if !errors.Is(filter.Err(), ErrFieldNotFound) {
		t.Errorf("Filter.Err() = %v, want %v", filter.Err(), ErrFieldNotFound)
	}
}

func TestBuilder_FromSpec_ComparisonOnString(t *testing.T) {
	builder, _ := New[testMetadata]()

	spec := &FilterSpec{Op: "gt", Field: "category", Value: "value"}

	filter := builder.FromSpec(spec)

	if filter.Err() == nil {
		t.Error("Filter.Err() should not be nil for comparison on string")
	}
	if !errors.Is(filter.Err(), ErrInvalidFilter) {
		t.Errorf("Filter.Err() = %v, want %v", filter.Err(), ErrInvalidFilter)
	}
}

func TestParseOp(t *testing.T) {
	tests := []struct {
		input   string
		want    Op
		wantErr bool
	}{
		{"eq", Eq, false},
		{"ne", Ne, false},
		{"gt", Gt, false},
		{"gte", Gte, false},
		{"lt", Lt, false},
		{"lte", Lte, false},
		{"in", In, false},
		{"nin", Nin, false},
		{"like", Like, false},
		{"contains", Contains, false},
		{"and", And, false},
		{"or", Or, false},
		{"not", Not, false},
		{"invalid", 0, true},
		{"", 0, true},
		{"EQ", 0, true}, // case-sensitive
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseOp(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("parseOp() should return error")
				}
				return
			}

			if err != nil {
				t.Errorf("parseOp() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("parseOp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuilder_FromSpec_Nin(t *testing.T) {
	builder, _ := New[testMetadata]()

	spec := &FilterSpec{
		Op:    "nin",
		Field: "category",
		Value: []any{"spam", "junk", "deleted"},
	}

	filter := builder.FromSpec(spec)

	if filter.Op() != Nin {
		t.Errorf("Filter.Op() = %v, want %v", filter.Op(), Nin)
	}
	if filter.Field() != "category" {
		t.Errorf("Filter.Field() = %v, want category", filter.Field())
	}
	if filter.Err() != nil {
		t.Errorf("Filter.Err() = %v, want nil", filter.Err())
	}

	values, ok := filter.Value().([]any)
	if !ok {
		t.Fatalf("Filter.Value() type = %T, want []any", filter.Value())
	}
	if len(values) != 3 {
		t.Errorf("len(Filter.Value()) = %v, want 3", len(values))
	}
}

func TestBuilder_FromSpec_Nin_TypedSlice(t *testing.T) {
	builder, _ := New[testMetadata]()

	// Typed slice (not []any) - exercises the else branch in fromNinSpec
	spec := &FilterSpec{
		Op:    "nin",
		Field: "category",
		Value: []string{"spam", "junk"},
	}

	filter := builder.FromSpec(spec)

	if filter.Op() != Nin {
		t.Errorf("Filter.Op() = %v, want %v", filter.Op(), Nin)
	}
	if filter.Err() != nil {
		t.Errorf("Filter.Err() = %v, want nil", filter.Err())
	}
}

func TestBuilder_FromSpec_Like(t *testing.T) {
	builder, _ := New[testMetadata]()

	spec := &FilterSpec{
		Op:    "like",
		Field: "category",
		Value: "%tech%",
	}

	filter := builder.FromSpec(spec)

	if filter.Op() != Like {
		t.Errorf("Filter.Op() = %v, want %v", filter.Op(), Like)
	}
	if filter.Field() != "category" {
		t.Errorf("Filter.Field() = %v, want category", filter.Field())
	}
	if filter.Value() != "%tech%" {
		t.Errorf("Filter.Value() = %v, want %%tech%%", filter.Value())
	}
	if filter.Err() != nil {
		t.Errorf("Filter.Err() = %v, want nil", filter.Err())
	}
}

func TestBuilder_FromSpec_Like_InvalidValue(t *testing.T) {
	builder, _ := New[testMetadata]()

	// Like requires string value
	spec := &FilterSpec{
		Op:    "like",
		Field: "category",
		Value: 123, // not a string
	}

	filter := builder.FromSpec(spec)

	if filter.Err() == nil {
		t.Error("Filter.Err() should not be nil for non-string like value")
	}
	if !errors.Is(filter.Err(), ErrInvalidFilter) {
		t.Errorf("Filter.Err() = %v, want %v", filter.Err(), ErrInvalidFilter)
	}
}

func TestBuilder_FromSpec_Contains(t *testing.T) {
	builder, _ := New[testMetadata]()

	spec := &FilterSpec{
		Op:    "contains",
		Field: "tags",
		Value: "featured",
	}

	filter := builder.FromSpec(spec)

	if filter.Op() != Contains {
		t.Errorf("Filter.Op() = %v, want %v", filter.Op(), Contains)
	}
	if filter.Field() != "tags" {
		t.Errorf("Filter.Field() = %v, want tags", filter.Field())
	}
	if filter.Value() != "featured" {
		t.Errorf("Filter.Value() = %v, want featured", filter.Value())
	}
	if filter.Err() != nil {
		t.Errorf("Filter.Err() = %v, want nil", filter.Err())
	}
}

func TestBuilder_FromSpec_Not(t *testing.T) {
	builder, _ := New[testMetadata]()

	spec := &FilterSpec{
		Op: "not",
		Children: []*FilterSpec{
			{Op: "eq", Field: "category", Value: "spam"},
		},
	}

	filter := builder.FromSpec(spec)

	if filter.Op() != Not {
		t.Errorf("Filter.Op() = %v, want %v", filter.Op(), Not)
	}
	if len(filter.Children()) != 1 {
		t.Errorf("len(Filter.Children()) = %v, want 1", len(filter.Children()))
	}
	if filter.Children()[0].Op() != Eq {
		t.Errorf("Child Filter.Op() = %v, want %v", filter.Children()[0].Op(), Eq)
	}
	if filter.Err() != nil {
		t.Errorf("Filter.Err() = %v, want nil", filter.Err())
	}
}

func TestBuilder_FromSpec_Not_MultipleChildren(t *testing.T) {
	builder, _ := New[testMetadata]()

	// Not requires exactly one child
	spec := &FilterSpec{
		Op: "not",
		Children: []*FilterSpec{
			{Op: "eq", Field: "category", Value: "spam"},
			{Op: "eq", Field: "category", Value: "junk"},
		},
	}

	filter := builder.FromSpec(spec)

	if filter.Err() == nil {
		t.Error("Filter.Err() should not be nil for not with multiple children")
	}
	if !errors.Is(filter.Err(), ErrInvalidFilter) {
		t.Errorf("Filter.Err() = %v, want %v", filter.Err(), ErrInvalidFilter)
	}
}

func TestBuilder_FromSpec_Not_Nested(t *testing.T) {
	builder, _ := New[testMetadata]()

	// NOT (category == "spam" OR category == "junk")
	spec := &FilterSpec{
		Op: "not",
		Children: []*FilterSpec{
			{
				Op: "or",
				Children: []*FilterSpec{
					{Op: "eq", Field: "category", Value: "spam"},
					{Op: "eq", Field: "category", Value: "junk"},
				},
			},
		},
	}

	filter := builder.FromSpec(spec)

	if filter.Op() != Not {
		t.Errorf("Filter.Op() = %v, want %v", filter.Op(), Not)
	}
	if len(filter.Children()) != 1 {
		t.Errorf("len(Filter.Children()) = %v, want 1", len(filter.Children()))
	}

	orChild := filter.Children()[0]
	if orChild.Op() != Or {
		t.Errorf("Child Filter.Op() = %v, want %v", orChild.Op(), Or)
	}
	if filter.Err() != nil {
		t.Errorf("Filter.Err() = %v, want nil", filter.Err())
	}
}
