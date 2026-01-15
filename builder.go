package vecna

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/zoobzio/sentinel"
)

// Builder provides schema-validated filter construction for type T.
// Create a Builder using New[T]().
type Builder[T any] struct {
	spec   Spec
	fields map[string]*FieldSpec // field name -> spec for O(1) lookup
}

// New creates a schema-validated Builder for metadata type T.
// Uses sentinel to extract field metadata from T.
// Field names are resolved from: json tag > Go field name.
// Fields with json:"-" are excluded.
func New[T any]() (*Builder[T], error) {
	// Register json tag for extraction before inspection
	sentinel.Tag("json")

	metadata, err := sentinel.TryInspect[T]()
	if err != nil {
		return nil, ErrNotStruct
	}

	spec := Spec{
		TypeName: metadata.TypeName,
		Fields:   make([]FieldSpec, 0, len(metadata.Fields)),
	}

	fields := make(map[string]*FieldSpec)

	for _, field := range metadata.Fields {
		// Get field name from json tag or use Go name
		name := resolveFieldName(field)
		if name == "-" || name == "" {
			continue // Skip excluded fields
		}

		kind := resolveFieldKind(field.Kind, field.Type)

		fieldSpec := FieldSpec{
			Name:   name,
			GoName: field.Name,
			Kind:   kind,
		}
		spec.Fields = append(spec.Fields, fieldSpec)
		fields[name] = &spec.Fields[len(spec.Fields)-1]
	}

	return &Builder[T]{
		spec:   spec,
		fields: fields,
	}, nil
}

// resolveFieldName extracts the field name from json tag or falls back to Go name.
func resolveFieldName(field sentinel.FieldMetadata) string {
	if jsonTag, ok := field.Tags["json"]; ok {
		// Parse json tag (format: "name,omitempty")
		parts := strings.Split(jsonTag, ",")
		if len(parts) > 0 && parts[0] != "" {
			return parts[0]
		}
	}
	return field.Name
}

// resolveFieldKind maps sentinel's FieldKind to vecna's FieldKind.
func resolveFieldKind(kind sentinel.FieldKind, typeName string) FieldKind {
	switch kind {
	case sentinel.KindScalar:
		// Further classify scalars by type name
		switch {
		case strings.HasPrefix(typeName, "int"), strings.HasPrefix(typeName, "uint"):
			return KindInt
		case strings.HasPrefix(typeName, "float"):
			return KindFloat
		case typeName == "bool":
			return KindBool
		case typeName == "string":
			return KindString
		default:
			return KindUnknown
		}
	case sentinel.KindSlice:
		return KindSlice
	default:
		return KindUnknown
	}
}

// Spec returns the schema for documentation/export.
func (b *Builder[T]) Spec() Spec {
	return b.spec
}

// Where begins a filter condition on a field.
// If the field doesn't exist in T, the returned FieldBuilder will
// produce a Filter with an error accessible via Filter.Err().
func (b *Builder[T]) Where(field string) *FieldBuilder[T] {
	spec, ok := b.fields[field]
	if !ok {
		return &FieldBuilder[T]{
			builder: b,
			field:   field,
			spec:    nil,
			err:     fmt.Errorf("%w: %s", ErrFieldNotFound, field),
		}
	}
	return &FieldBuilder[T]{
		builder: b,
		field:   field,
		spec:    spec,
		err:     nil,
	}
}

// And combines filters with logical AND.
// Returns a Filter that matches when all child filters match.
func (*Builder[T]) And(filters ...*Filter) *Filter {
	return &Filter{
		op:       And,
		children: filters,
	}
}

// Or combines filters with logical OR.
// Returns a Filter that matches when any child filter matches.
func (*Builder[T]) Or(filters ...*Filter) *Filter {
	return &Filter{
		op:       Or,
		children: filters,
	}
}

// Not negates a filter.
// Returns a Filter that matches when the child filter does not match.
func (*Builder[T]) Not(filter *Filter) *Filter {
	return &Filter{
		op:       Not,
		children: []*Filter{filter},
	}
}

// FieldBuilder constructs conditions for a specific field.
type FieldBuilder[T any] struct {
	builder *Builder[T]
	field   string
	spec    *FieldSpec
	err     error
}

// Eq creates an equality filter (field == value).
func (fb *FieldBuilder[T]) Eq(value any) *Filter {
	return fb.makeFilter(Eq, value)
}

// Ne creates a not-equal filter (field != value).
func (fb *FieldBuilder[T]) Ne(value any) *Filter {
	return fb.makeFilter(Ne, value)
}

// Gt creates a greater-than filter (field > value).
func (fb *FieldBuilder[T]) Gt(value any) *Filter {
	return fb.makeFilter(Gt, value)
}

// Gte creates a greater-than-or-equal filter (field >= value).
func (fb *FieldBuilder[T]) Gte(value any) *Filter {
	return fb.makeFilter(Gte, value)
}

// Lt creates a less-than filter (field < value).
func (fb *FieldBuilder[T]) Lt(value any) *Filter {
	return fb.makeFilter(Lt, value)
}

// Lte creates a less-than-or-equal filter (field <= value).
func (fb *FieldBuilder[T]) Lte(value any) *Filter {
	return fb.makeFilter(Lte, value)
}

// In creates a set membership filter (field IN values).
func (fb *FieldBuilder[T]) In(values ...any) *Filter {
	// Convert variadic args to a slice for consistent handling
	return fb.makeFilter(In, values)
}

// Nin creates a set exclusion filter (field NOT IN values).
func (fb *FieldBuilder[T]) Nin(values ...any) *Filter {
	// Convert variadic args to a slice for consistent handling
	return fb.makeFilter(Nin, values)
}

// Like creates a pattern matching filter (field LIKE pattern).
// Pattern syntax is provider-dependent but typically supports % and _ wildcards.
func (fb *FieldBuilder[T]) Like(pattern string) *Filter {
	return fb.makeFilter(Like, pattern)
}

// Contains creates an array membership filter (array field contains value).
func (fb *FieldBuilder[T]) Contains(value any) *Filter {
	return fb.makeFilter(Contains, value)
}

// makeFilter creates a Filter with the given operator and value.
func (fb *FieldBuilder[T]) makeFilter(op Op, value any) *Filter {
	if fb.err != nil {
		return &Filter{
			op:    op,
			field: fb.field,
			value: value,
			err:   fb.err,
		}
	}

	// Validate value type against field kind
	if err := fb.validateValue(op, value); err != nil {
		return &Filter{
			op:    op,
			field: fb.field,
			value: value,
			err:   err,
		}
	}

	return &Filter{
		op:    op,
		field: fb.field,
		value: value,
	}
}

// validateValue checks if the value type is compatible with the field kind and operator.
func (fb *FieldBuilder[T]) validateValue(op Op, value any) error {
	if fb.spec == nil {
		return nil // Already has an error
	}

	// For In/Nin operators, validate the slice elements
	if op == In || op == Nin {
		return validateInValue(value)
	}

	// For Like operator, require string field
	if op == Like && fb.spec.Kind != KindString {
		return fmt.Errorf("%w: operator %s not valid for %s field %s",
			ErrInvalidFilter, op, fb.spec.Kind, fb.field)
	}

	// For Contains operator, require slice field
	if op == Contains && fb.spec.Kind != KindSlice {
		return fmt.Errorf("%w: operator %s not valid for %s field %s",
			ErrInvalidFilter, op, fb.spec.Kind, fb.field)
	}

	// For comparison operators on non-numeric fields
	if isComparisonOp(op) && !isNumericKind(fb.spec.Kind) {
		return fmt.Errorf("%w: operator %s not valid for %s field %s",
			ErrInvalidFilter, op, fb.spec.Kind, fb.field)
	}

	return nil
}

// validateInValue validates values for the In operator.
func validateInValue(value any) error {
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("%w: In operator requires slice of values", ErrInvalidFilter)
	}
	return nil
}

// isComparisonOp returns true if the operator is a comparison (not equality).
func isComparisonOp(op Op) bool {
	return op == Gt || op == Gte || op == Lt || op == Lte
}

// isNumericKind returns true if the field kind is numeric.
func isNumericKind(kind FieldKind) bool {
	return kind == KindInt || kind == KindFloat
}
