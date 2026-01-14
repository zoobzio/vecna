package vecna

import (
	"errors"
	"testing"
)

func TestOp_String(t *testing.T) {
	tests := []struct {
		op   Op
		want string
	}{
		{Eq, "eq"},
		{Ne, "ne"},
		{Gt, "gt"},
		{Gte, "gte"},
		{Lt, "lt"},
		{Lte, "lte"},
		{In, "in"},
		{And, "and"},
		{Or, "or"},
		{Op(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.op.String(); got != tt.want {
				t.Errorf("Op.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFieldKind_String(t *testing.T) {
	tests := []struct {
		kind FieldKind
		want string
	}{
		{KindString, "string"},
		{KindInt, "int"},
		{KindFloat, "float"},
		{KindBool, "bool"},
		{KindSlice, "slice"},
		{KindUnknown, "unknown"},
		{FieldKind(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.kind.String(); got != tt.want {
				t.Errorf("FieldKind.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilter_Accessors(t *testing.T) {
	f := &Filter{
		op:    Eq,
		field: "category",
		value: "tech",
	}

	if f.Op() != Eq {
		t.Errorf("Filter.Op() = %v, want %v", f.Op(), Eq)
	}
	if f.Field() != "category" {
		t.Errorf("Filter.Field() = %v, want %v", f.Field(), "category")
	}
	if f.Value() != "tech" {
		t.Errorf("Filter.Value() = %v, want %v", f.Value(), "tech")
	}
	if f.Children() != nil {
		t.Errorf("Filter.Children() = %v, want nil", f.Children())
	}
	if f.Err() != nil {
		t.Errorf("Filter.Err() = %v, want nil", f.Err())
	}
}

func TestFilter_Children(t *testing.T) {
	child1 := &Filter{op: Eq, field: "a", value: 1}
	child2 := &Filter{op: Eq, field: "b", value: 2}

	f := &Filter{
		op:       And,
		children: []*Filter{child1, child2},
	}

	if f.Op() != And {
		t.Errorf("Filter.Op() = %v, want %v", f.Op(), And)
	}
	if len(f.Children()) != 2 {
		t.Errorf("len(Filter.Children()) = %v, want 2", len(f.Children()))
	}
}

func TestFilter_Err(t *testing.T) {
	t.Run("nil filter", func(t *testing.T) {
		var f *Filter
		if f.Err() != nil {
			t.Errorf("nil Filter.Err() = %v, want nil", f.Err())
		}
	})

	t.Run("filter with error", func(t *testing.T) {
		f := &Filter{err: ErrFieldNotFound}
		if !errors.Is(f.Err(), ErrFieldNotFound) {
			t.Errorf("Filter.Err() = %v, want %v", f.Err(), ErrFieldNotFound)
		}
	})

	t.Run("child with error", func(t *testing.T) {
		child := &Filter{err: ErrFieldNotFound}
		f := &Filter{
			op:       And,
			children: []*Filter{child},
		}
		if !errors.Is(f.Err(), ErrFieldNotFound) {
			t.Errorf("Filter.Err() = %v, want %v", f.Err(), ErrFieldNotFound)
		}
	})
}

func TestSpec_Field(t *testing.T) {
	spec := Spec{
		TypeName: "TestType",
		Fields: []FieldSpec{
			{Name: "category", GoName: "Category", Kind: KindString},
			{Name: "score", GoName: "Score", Kind: KindFloat},
		},
	}

	t.Run("found", func(t *testing.T) {
		f := spec.Field("category")
		if f == nil {
			t.Fatal("Spec.Field() returned nil")
		}
		if f.Name != "category" {
			t.Errorf("FieldSpec.Name = %v, want category", f.Name)
		}
	})

	t.Run("not found", func(t *testing.T) {
		f := spec.Field("nonexistent")
		if f != nil {
			t.Errorf("Spec.Field() = %v, want nil", f)
		}
	})
}
