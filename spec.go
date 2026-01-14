package vecna

import "fmt"

// FilterSpec represents a serializable filter specification.
// This enables programmatic filter construction from JSON or other external sources.
type FilterSpec struct {
	Op       string        `json:"op"`                 // Operator: "eq", "ne", "gt", "gte", "lt", "lte", "in", "and", "or"
	Field    string        `json:"field,omitempty"`    // Field name (for field conditions)
	Value    any           `json:"value,omitempty"`    // Comparison value (for field conditions)
	Children []*FilterSpec `json:"children,omitempty"` // Child filters (for and/or)
}

// FromSpec converts a FilterSpec to a validated Filter.
// The spec is validated against the schema defined by T.
// Any validation errors are accessible via Filter.Err().
func (b *Builder[T]) FromSpec(spec *FilterSpec) *Filter {
	if spec == nil {
		return &Filter{err: fmt.Errorf("%w: nil spec", ErrInvalidFilter)}
	}

	op, err := parseOp(spec.Op)
	if err != nil {
		return &Filter{err: err}
	}

	// Handle logical operators
	if op == And || op == Or {
		return b.fromLogicalSpec(op, spec.Children)
	}

	// Handle field operators
	return b.fromFieldSpec(op, spec.Field, spec.Value)
}

// fromLogicalSpec converts a logical operator spec (and/or) to a Filter.
func (b *Builder[T]) fromLogicalSpec(op Op, children []*FilterSpec) *Filter {
	if len(children) == 0 {
		return &Filter{
			op:  op,
			err: fmt.Errorf("%w: %s requires at least one child", ErrInvalidFilter, op),
		}
	}

	filters := make([]*Filter, len(children))
	for i, child := range children {
		filters[i] = b.FromSpec(child)
	}

	if op == And {
		return b.And(filters...)
	}
	return b.Or(filters...)
}

// fromFieldSpec converts a field operator spec to a Filter.
func (b *Builder[T]) fromFieldSpec(op Op, field string, value any) *Filter {
	fb := b.Where(field)

	switch op {
	case Eq:
		return fb.Eq(value)
	case Ne:
		return fb.Ne(value)
	case Gt:
		return fb.Gt(value)
	case Gte:
		return fb.Gte(value)
	case Lt:
		return fb.Lt(value)
	case Lte:
		return fb.Lte(value)
	case In:
		return b.fromInSpec(fb, value)
	default:
		return &Filter{
			op:    op,
			field: field,
			value: value,
			err:   fmt.Errorf("%w: unsupported field operator %s", ErrInvalidFilter, op),
		}
	}
}

// fromInSpec handles the In operator which expects a slice value.
func (*Builder[T]) fromInSpec(fb *FieldBuilder[T], value any) *Filter {
	// Value should be a slice when deserialized from JSON
	slice, ok := value.([]any)
	if !ok {
		// If it's already a typed slice, pass it through
		return fb.In(value)
	}
	return fb.In(slice...)
}

// parseOp converts a string operator to an Op constant.
func parseOp(s string) (Op, error) {
	switch s {
	case "eq":
		return Eq, nil
	case "ne":
		return Ne, nil
	case "gt":
		return Gt, nil
	case "gte":
		return Gte, nil
	case "lt":
		return Lt, nil
	case "lte":
		return Lte, nil
	case "in":
		return In, nil
	case "and":
		return And, nil
	case "or":
		return Or, nil
	default:
		return 0, fmt.Errorf("%w: unknown operator %q", ErrInvalidFilter, s)
	}
}
