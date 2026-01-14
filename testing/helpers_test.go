package testing

import (
	"testing"

	"github.com/zoobzio/vecna"
)

func TestNewTestBuilder(t *testing.T) {
	builder := NewTestBuilder(t)
	if builder == nil {
		t.Fatal("NewTestBuilder returned nil")
	}

	spec := builder.Spec()
	if spec.TypeName != "DocumentMetadata" {
		t.Errorf("expected type name DocumentMetadata, got %s", spec.TypeName)
	}
}

func TestAssertNoError(t *testing.T) {
	builder := NewTestBuilder(t)
	filter := builder.Where("category").Eq("test")
	AssertNoError(t, filter)
}

func TestAssertError(t *testing.T) {
	builder := NewTestBuilder(t)
	filter := builder.Where("nonexistent").Eq("test")
	AssertError(t, filter)
}

func TestAssertOp(t *testing.T) {
	builder := NewTestBuilder(t)
	filter := builder.Where("category").Eq("test")
	AssertOp(t, filter, vecna.Eq)
}

func TestAssertField(t *testing.T) {
	builder := NewTestBuilder(t)
	filter := builder.Where("category").Eq("test")
	AssertField(t, filter, "category")
}

func TestAssertChildCount(t *testing.T) {
	builder := NewTestBuilder(t)
	filter := builder.And(
		builder.Where("category").Eq("tech"),
		builder.Where("active").Eq(true),
	)
	AssertChildCount(t, filter, 2)
}
