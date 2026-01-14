// Package vecna provides a schema-validated filter builder for vector database queries.
// It serves as the query language for vector metadata filtering, similar to how
// edamame provides SQL AST capabilities.
package vecna

import "errors"

// Errors returned by vecna.
var (
	// ErrNotStruct is returned when the type parameter is not a struct.
	ErrNotStruct = errors.New("vecna: type must be a struct")

	// ErrFieldNotFound is returned when a field name is not found in the schema.
	ErrFieldNotFound = errors.New("vecna: field not found")

	// ErrInvalidFilter is returned when a filter contains validation errors.
	ErrInvalidFilter = errors.New("vecna: invalid filter")
)

// Op represents a filter operator.
type Op uint8

// Filter operators.
const (
	Eq  Op = iota // Equal
	Ne            // Not equal
	Gt            // Greater than
	Gte           // Greater than or equal
	Lt            // Less than
	Lte           // Less than or equal
	In            // In set
	And           // Logical AND
	Or            // Logical OR
)

// String returns the string representation of the operator.
func (o Op) String() string {
	switch o {
	case Eq:
		return "eq"
	case Ne:
		return "ne"
	case Gt:
		return "gt"
	case Gte:
		return "gte"
	case Lt:
		return "lt"
	case Lte:
		return "lte"
	case In:
		return "in"
	case And:
		return "and"
	case Or:
		return "or"
	default:
		return "unknown"
	}
}

// Filter represents a filter condition or logical group.
// Construct filters using Builder[T].Where(), And(), or Or().
type Filter struct {
	op       Op
	field    string
	value    any
	children []*Filter
	err      error // Deferred error for invalid field
}

// Op returns the filter operator.
func (f *Filter) Op() Op {
	return f.op
}

// Field returns the field name for field conditions.
// Returns empty string for logical operators (And, Or).
func (f *Filter) Field() string {
	return f.field
}

// Value returns the comparison value for field conditions.
// Returns nil for logical operators (And, Or).
func (f *Filter) Value() any {
	return f.value
}

// Children returns the child filters for logical operators (And, Or).
// Returns nil for field conditions.
func (f *Filter) Children() []*Filter {
	return f.children
}

// Err returns any error that occurred during filter construction.
// This enables deferred error checking after building complex filters.
func (f *Filter) Err() error {
	if f == nil {
		return nil
	}
	if f.err != nil {
		return f.err
	}
	// Check children for errors
	for _, child := range f.children {
		if err := child.Err(); err != nil {
			return err
		}
	}
	return nil
}

// FieldKind categorizes field types for validation.
type FieldKind uint8

// Field kind constants.
const (
	KindString FieldKind = iota
	KindInt
	KindFloat
	KindBool
	KindSlice
	KindUnknown
)

// String returns the string representation of the field kind.
func (k FieldKind) String() string {
	switch k {
	case KindString:
		return "string"
	case KindInt:
		return "int"
	case KindFloat:
		return "float"
	case KindBool:
		return "bool"
	case KindSlice:
		return "slice"
	default:
		return "unknown"
	}
}

// FieldSpec describes a single filterable field.
type FieldSpec struct {
	Name   string    // JSON field name (from tag or Go name)
	GoName string    // Original Go field name
	Kind   FieldKind // Type category
}

// Spec describes the metadata schema extracted from T.
type Spec struct {
	TypeName string      // Go type name
	Fields   []FieldSpec // Filterable fields
}

// Field returns the FieldSpec for the given field name, or nil if not found.
func (s *Spec) Field(name string) *FieldSpec {
	for i := range s.Fields {
		if s.Fields[i].Name == name {
			return &s.Fields[i]
		}
	}
	return nil
}
