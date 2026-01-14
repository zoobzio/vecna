package benchmarks

import (
	"testing"

	"github.com/zoobzio/vecna"
)

type BenchMetadata struct {
	Category string   `json:"category"`
	Score    float64  `json:"score"`
	Active   bool     `json:"active"`
	Tags     []string `json:"tags"`
}

func BenchmarkNew(b *testing.B) {
	for b.Loop() {
		_, _ = vecna.New[BenchMetadata]()
	}
}

func BenchmarkWhereEq(b *testing.B) {
	builder, _ := vecna.New[BenchMetadata]()
	b.ResetTimer()
	for b.Loop() {
		_ = builder.Where("category").Eq("tech")
	}
}

func BenchmarkWhereComparison(b *testing.B) {
	builder, _ := vecna.New[BenchMetadata]()
	b.ResetTimer()
	for b.Loop() {
		_ = builder.Where("score").Gte(0.8)
	}
}

func BenchmarkAnd(b *testing.B) {
	builder, _ := vecna.New[BenchMetadata]()
	f1 := builder.Where("category").Eq("tech")
	f2 := builder.Where("active").Eq(true)
	b.ResetTimer()
	for b.Loop() {
		_ = builder.And(f1, f2)
	}
}

func BenchmarkComplexFilter(b *testing.B) {
	builder, _ := vecna.New[BenchMetadata]()
	b.ResetTimer()
	for b.Loop() {
		_ = builder.And(
			builder.Where("category").Eq("tech"),
			builder.Or(
				builder.Where("score").Gte(0.8),
				builder.Where("active").Eq(true),
			),
		)
	}
}

func BenchmarkFromSpec(b *testing.B) {
	builder, _ := vecna.New[BenchMetadata]()
	spec := &vecna.FilterSpec{
		Op: "and",
		Children: []*vecna.FilterSpec{
			{Op: "eq", Field: "category", Value: "tech"},
			{Op: "gte", Field: "score", Value: 0.8},
		},
	}
	b.ResetTimer()
	for b.Loop() {
		_ = builder.FromSpec(spec)
	}
}

func BenchmarkFilterErr(b *testing.B) {
	builder, _ := vecna.New[BenchMetadata]()
	filter := builder.And(
		builder.Where("category").Eq("tech"),
		builder.Or(
			builder.Where("score").Gte(0.8),
			builder.Where("active").Eq(true),
		),
	)
	b.ResetTimer()
	for b.Loop() {
		_ = filter.Err()
	}
}
