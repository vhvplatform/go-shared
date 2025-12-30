package clickhouse

import (
	"testing"
)

func TestQueryBuilder_BuildSelect(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(*QueryBuilder) *QueryBuilder
		expectedQuery string
		expectedArgs  int
	}{
		{
			name: "simple select all",
			setup: func(qb *QueryBuilder) *QueryBuilder {
				return qb.Table("users")
			},
			expectedQuery: "SELECT * FROM users",
			expectedArgs:  0,
		},
		{
			name: "select with columns",
			setup: func(qb *QueryBuilder) *QueryBuilder {
				return qb.Table("users").Select("id", "name", "email")
			},
			expectedQuery: "SELECT id, name, email FROM users",
			expectedArgs:  0,
		},
		{
			name: "select with where",
			setup: func(qb *QueryBuilder) *QueryBuilder {
				return qb.Table("users").Where("id", 1)
			},
			expectedQuery: "SELECT * FROM users WHERE id = ?",
			expectedArgs:  1,
		},
		{
			name: "select with multiple where",
			setup: func(qb *QueryBuilder) *QueryBuilder {
				return qb.Table("users").
					Where("status", "active").
					Where("role", "admin")
			},
			expectedQuery: "SELECT * FROM users WHERE status = ? AND role = ?",
			expectedArgs:  2,
		},
		{
			name: "select with order by",
			setup: func(qb *QueryBuilder) *QueryBuilder {
				return qb.Table("users").OrderBy("created_at", "DESC")
			},
			expectedQuery: "SELECT * FROM users ORDER BY created_at DESC",
			expectedArgs:  0,
		},
		{
			name: "select with limit and offset",
			setup: func(qb *QueryBuilder) *QueryBuilder {
				return qb.Table("users").Limit(10).Offset(20)
			},
			expectedQuery: "SELECT * FROM users LIMIT 10 OFFSET 20",
			expectedArgs:  0,
		},
		{
			name: "complex select",
			setup: func(qb *QueryBuilder) *QueryBuilder {
				return qb.Table("orders").
					Select("id", "user_id", "amount").
					Where("status", "completed").
					WhereGreaterThan("amount", 100).
					OrderBy("created_at", "DESC").
					Limit(50).
					Offset(0)
			},
			expectedQuery: "SELECT id, user_id, amount FROM orders WHERE status = ? AND amount > ? ORDER BY created_at DESC LIMIT 50",
			expectedArgs:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := NewQueryBuilder()
			qb = tt.setup(qb)
			query, args := qb.BuildSelect()

			if query != tt.expectedQuery {
				t.Errorf("Query mismatch.\nExpected: %s\nGot: %s", tt.expectedQuery, query)
			}

			if len(args) != tt.expectedArgs {
				t.Errorf("Args count mismatch. Expected: %d, Got: %d", tt.expectedArgs, len(args))
			}
		})
	}
}

func TestQueryBuilder_WhereIn(t *testing.T) {
	qb := NewQueryBuilder()
	query, args := qb.Table("users").
		WhereIn("id", []interface{}{1, 2, 3, 4, 5}).
		BuildSelect()

	expectedQuery := "SELECT * FROM users WHERE id IN (?,?,?,?,?)"
	if query != expectedQuery {
		t.Errorf("Query mismatch.\nExpected: %s\nGot: %s", expectedQuery, query)
	}

	if len(args) != 5 {
		t.Errorf("Expected 5 args, got %d", len(args))
	}
}

func TestQueryBuilder_WhereBetween(t *testing.T) {
	qb := NewQueryBuilder()
	query, args := qb.Table("orders").
		WhereBetween("amount", 100, 1000).
		BuildSelect()

	expectedQuery := "SELECT * FROM orders WHERE amount BETWEEN ? AND ?"
	if query != expectedQuery {
		t.Errorf("Query mismatch.\nExpected: %s\nGot: %s", expectedQuery, query)
	}

	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}
}

func TestQueryBuilder_WhereLike(t *testing.T) {
	qb := NewQueryBuilder()
	query, args := qb.Table("users").
		WhereLike("email", "%@example.com").
		BuildSelect()

	expectedQuery := "SELECT * FROM users WHERE email LIKE ?"
	if query != expectedQuery {
		t.Errorf("Query mismatch.\nExpected: %s\nGot: %s", expectedQuery, query)
	}

	if len(args) != 1 {
		t.Errorf("Expected 1 arg, got %d", len(args))
	}
}

func TestQueryBuilder_BuildCount(t *testing.T) {
	qb := NewQueryBuilder()
	query, args := qb.Table("users").
		Where("status", "active").
		BuildCount()

	expectedQuery := "SELECT COUNT(*) FROM users WHERE status = ?"
	if query != expectedQuery {
		t.Errorf("Query mismatch.\nExpected: %s\nGot: %s", expectedQuery, query)
	}

	if len(args) != 1 {
		t.Errorf("Expected 1 arg, got %d", len(args))
	}
}

func TestQueryBuilder_BuildDelete(t *testing.T) {
	qb := NewQueryBuilder()
	query, args := qb.Table("users").
		Where("id", 123).
		BuildDelete()

	expectedQuery := "ALTER TABLE users DELETE WHERE id = ?"
	if query != expectedQuery {
		t.Errorf("Query mismatch.\nExpected: %s\nGot: %s", expectedQuery, query)
	}

	if len(args) != 1 {
		t.Errorf("Expected 1 arg, got %d", len(args))
	}
}

func TestQueryBuilder_Clone(t *testing.T) {
	original := NewQueryBuilder().
		Table("users").
		Where("status", "active")

	clone := original.Clone()
	clone.Where("role", "admin")

	// Original should not be modified
	originalQuery, originalArgs := original.BuildSelect()
	expectedOriginal := "SELECT * FROM users WHERE status = ?"
	if originalQuery != expectedOriginal {
		t.Errorf("Original query was modified. Expected: %s, Got: %s", expectedOriginal, originalQuery)
	}
	if len(originalArgs) != 1 {
		t.Errorf("Original args count mismatch. Expected: 1, Got: %d", len(originalArgs))
	}

	// Clone should have both conditions
	cloneQuery, cloneArgs := clone.BuildSelect()
	expectedClone := "SELECT * FROM users WHERE status = ? AND role = ?"
	if cloneQuery != expectedClone {
		t.Errorf("Clone query mismatch. Expected: %s, Got: %s", expectedClone, cloneQuery)
	}
	if len(cloneArgs) != 2 {
		t.Errorf("Clone args count mismatch. Expected: 2, Got: %d", len(cloneArgs))
	}
}

func TestQueryBuilder_Reset(t *testing.T) {
	qb := NewQueryBuilder().
		Table("users").
		Where("status", "active").
		Limit(10)

	qb.Reset()

	// After reset, should be empty
	qb.Table("orders")
	query, args := qb.BuildSelect()

	expectedQuery := "SELECT * FROM orders"
	if query != expectedQuery {
		t.Errorf("Query after reset mismatch. Expected: %s, Got: %s", expectedQuery, query)
	}
	if len(args) != 0 {
		t.Errorf("Args should be empty after reset. Got: %d", len(args))
	}
}

func BenchmarkQueryBuilder_BuildSelect(b *testing.B) {
	qb := NewQueryBuilder().
		Table("users").
		Select("id", "name", "email").
		Where("status", "active").
		WhereGreaterThan("created_at", "2024-01-01").
		OrderBy("created_at", "DESC").
		Limit(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qb.BuildSelect()
	}
}

func BenchmarkQueryBuilder_Clone(b *testing.B) {
	qb := NewQueryBuilder().
		Table("users").
		Where("status", "active").
		OrderBy("created_at", "DESC")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = qb.Clone()
	}
}
