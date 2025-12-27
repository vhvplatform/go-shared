package mongodb

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func BenchmarkNewQueryBuilder(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewQueryBuilder()
	}
}

func BenchmarkQueryBuilderWhere(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		qb := NewQueryBuilder()
		b.StartTimer()
		
		qb.Where("name", "John").
			Where("age", 30).
			Where("active", true)
	}
}

func BenchmarkQueryBuilderBuild(b *testing.B) {
	qb := NewQueryBuilder().
		Where("name", "John").
		Where("age", 30).
		WhereGreaterThan("score", 80)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = qb.Build()
	}
}

func BenchmarkQueryBuilderClone(b *testing.B) {
	qb := NewQueryBuilder().
		Where("name", "John").
		Where("age", 30).
		WhereGreaterThan("score", 80).
		WhereIn("tags", []string{"go", "mongodb"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = qb.Clone()
	}
}

func BenchmarkQueryBuilderReset(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		qb := NewQueryBuilder().
			Where("name", "John").
			Where("age", 30).
			WhereGreaterThan("score", 80)
		b.StartTimer()
		
		qb.Reset()
	}
}

func BenchmarkNewAggregationBuilder(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewAggregationBuilder()
	}
}

func BenchmarkAggregationBuilderMatch(b *testing.B) {
	filter := bson.M{"status": "active"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		ab := NewAggregationBuilder()
		b.StartTimer()
		
		ab.Match(filter)
	}
}

func BenchmarkAggregationBuilderBuild(b *testing.B) {
	ab := NewAggregationBuilder().
		Match(bson.M{"status": "active"}).
		Group(bson.M{"_id": "$category", "count": bson.M{"$sum": 1}}).
		Sort(bson.D{{Key: "count", Value: -1}}).
		Limit(10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ab.Build()
	}
}

func BenchmarkAggregationBuilderClone(b *testing.B) {
	ab := NewAggregationBuilder().
		Match(bson.M{"status": "active"}).
		Group(bson.M{"_id": "$category", "count": bson.M{"$sum": 1}}).
		Sort(bson.D{{Key: "count", Value: -1}}).
		Limit(10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ab.Clone()
	}
}

func BenchmarkPaginationParamsSkip(b *testing.B) {
	params := &PaginationParams{
		Page:     5,
		PageSize: 20,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = params.Skip()
	}
}

func BenchmarkNewPaginationParams(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewPaginationParams(1, 10)
	}
}
