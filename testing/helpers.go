// Package testing provides test helpers for vecna.
package testing

import (
	"testing"

	"github.com/zoobzio/vecna"
)

// DocumentMetadata is a standard test schema for vecna tests.
type DocumentMetadata struct {
	Category string   `json:"category"`
	Score    float64  `json:"score"`
	Active   bool     `json:"active"`
	Tags     []string `json:"tags"`
	Count    int      `json:"count"`
}

// NewTestBuilder creates a Builder[DocumentMetadata] for testing.
// Fails the test if builder creation fails.
func NewTestBuilder(t *testing.T) *vecna.Builder[DocumentMetadata] {
	t.Helper()
	builder, err := vecna.New[DocumentMetadata]()
	if err != nil {
		t.Fatalf("failed to create test builder: %v", err)
	}
	return builder
}

// AssertNoError fails the test if the filter has an error.
func AssertNoError(t *testing.T, f *vecna.Filter) {
	t.Helper()
	if err := f.Err(); err != nil {
		t.Errorf("unexpected filter error: %v", err)
	}
}

// AssertError fails the test if the filter does not have an error.
func AssertError(t *testing.T, f *vecna.Filter) {
	t.Helper()
	if f.Err() == nil {
		t.Error("expected filter error, got nil")
	}
}

// AssertOp fails the test if the filter operator doesn't match expected.
func AssertOp(t *testing.T, f *vecna.Filter, expected vecna.Op) {
	t.Helper()
	if f.Op() != expected {
		t.Errorf("expected op %s, got %s", expected, f.Op())
	}
}

// AssertField fails the test if the filter field doesn't match expected.
func AssertField(t *testing.T, f *vecna.Filter, expected string) {
	t.Helper()
	if f.Field() != expected {
		t.Errorf("expected field %q, got %q", expected, f.Field())
	}
}

// AssertChildCount fails the test if the filter doesn't have the expected number of children.
func AssertChildCount(t *testing.T, f *vecna.Filter, expected int) {
	t.Helper()
	if len(f.Children()) != expected {
		t.Errorf("expected %d children, got %d", expected, len(f.Children()))
	}
}
