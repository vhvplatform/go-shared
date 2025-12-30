package clickhouse

import (
	"fmt"
	"strings"
)

// QueryBuilder provides a fluent API for building SQL queries
type QueryBuilder struct {
	table      string
	columns    []string
	conditions []string
	orderBy    []string
	limit      int
	offset     int
	args       []interface{}
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		columns:    make([]string, 0),
		conditions: make([]string, 0),
		orderBy:    make([]string, 0),
		args:       make([]interface{}, 0),
	}
}

// Table sets the table name
func (qb *QueryBuilder) Table(table string) *QueryBuilder {
	qb.table = table
	return qb
}

// Select sets the columns to select
func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	qb.columns = columns
	return qb
}

// Where adds a WHERE condition
func (qb *QueryBuilder) Where(column string, value interface{}) *QueryBuilder {
	qb.conditions = append(qb.conditions, fmt.Sprintf("%s = ?", column))
	qb.args = append(qb.args, value)
	return qb
}

// WhereIn adds a WHERE IN condition
func (qb *QueryBuilder) WhereIn(column string, values []interface{}) *QueryBuilder {
	placeholders := make([]string, len(values))
	for i := range values {
		placeholders[i] = "?"
		qb.args = append(qb.args, values[i])
	}
	qb.conditions = append(qb.conditions, fmt.Sprintf("%s IN (%s)", column, strings.Join(placeholders, ",")))
	return qb
}

// WhereGreaterThan adds a WHERE > condition
func (qb *QueryBuilder) WhereGreaterThan(column string, value interface{}) *QueryBuilder {
	qb.conditions = append(qb.conditions, fmt.Sprintf("%s > ?", column))
	qb.args = append(qb.args, value)
	return qb
}

// WhereLessThan adds a WHERE < condition
func (qb *QueryBuilder) WhereLessThan(column string, value interface{}) *QueryBuilder {
	qb.conditions = append(qb.conditions, fmt.Sprintf("%s < ?", column))
	qb.args = append(qb.args, value)
	return qb
}

// WhereBetween adds a WHERE BETWEEN condition
func (qb *QueryBuilder) WhereBetween(column string, from, to interface{}) *QueryBuilder {
	qb.conditions = append(qb.conditions, fmt.Sprintf("%s BETWEEN ? AND ?", column))
	qb.args = append(qb.args, from, to)
	return qb
}

// WhereLike adds a WHERE LIKE condition
func (qb *QueryBuilder) WhereLike(column string, pattern string) *QueryBuilder {
	qb.conditions = append(qb.conditions, fmt.Sprintf("%s LIKE ?", column))
	qb.args = append(qb.args, pattern)
	return qb
}

// OrderBy adds an ORDER BY clause
func (qb *QueryBuilder) OrderBy(column string, direction string) *QueryBuilder {
	qb.orderBy = append(qb.orderBy, fmt.Sprintf("%s %s", column, direction))
	return qb
}

// Limit sets the LIMIT clause
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

// Offset sets the OFFSET clause
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offset = offset
	return qb
}

// BuildSelect builds a SELECT query
func (qb *QueryBuilder) BuildSelect() (string, []interface{}) {
	query := strings.Builder{}

	// SELECT clause
	if len(qb.columns) == 0 {
		query.WriteString("SELECT *")
	} else {
		query.WriteString("SELECT ")
		query.WriteString(strings.Join(qb.columns, ", "))
	}

	// FROM clause
	query.WriteString(fmt.Sprintf(" FROM %s", qb.table))

	// WHERE clause
	if len(qb.conditions) > 0 {
		query.WriteString(" WHERE ")
		query.WriteString(strings.Join(qb.conditions, " AND "))
	}

	// ORDER BY clause
	if len(qb.orderBy) > 0 {
		query.WriteString(" ORDER BY ")
		query.WriteString(strings.Join(qb.orderBy, ", "))
	}

	// LIMIT clause
	if qb.limit > 0 {
		query.WriteString(fmt.Sprintf(" LIMIT %d", qb.limit))
	}

	// OFFSET clause (only if greater than 0)
	if qb.offset > 0 {
		query.WriteString(fmt.Sprintf(" OFFSET %d", qb.offset))
	}

	return query.String(), qb.args
}

// BuildCount builds a COUNT query
func (qb *QueryBuilder) BuildCount() (string, []interface{}) {
	query := strings.Builder{}
	query.WriteString(fmt.Sprintf("SELECT COUNT(*) FROM %s", qb.table))

	// WHERE clause
	if len(qb.conditions) > 0 {
		query.WriteString(" WHERE ")
		query.WriteString(strings.Join(qb.conditions, " AND "))
	}

	return query.String(), qb.args
}

// BuildDelete builds a DELETE query
func (qb *QueryBuilder) BuildDelete() (string, []interface{}) {
	query := strings.Builder{}
	query.WriteString(fmt.Sprintf("ALTER TABLE %s DELETE", qb.table))

	// WHERE clause
	if len(qb.conditions) > 0 {
		query.WriteString(" WHERE ")
		query.WriteString(strings.Join(qb.conditions, " AND "))
	}

	return query.String(), qb.args
}

// Reset resets the query builder
func (qb *QueryBuilder) Reset() *QueryBuilder {
	qb.table = ""
	qb.columns = make([]string, 0)
	qb.conditions = make([]string, 0)
	qb.orderBy = make([]string, 0)
	qb.limit = 0
	qb.offset = 0
	qb.args = make([]interface{}, 0)
	return qb
}

// Clone creates a copy of the query builder
func (qb *QueryBuilder) Clone() *QueryBuilder {
	return &QueryBuilder{
		table:      qb.table,
		columns:    append([]string{}, qb.columns...),
		conditions: append([]string{}, qb.conditions...),
		orderBy:    append([]string{}, qb.orderBy...),
		limit:      qb.limit,
		offset:     qb.offset,
		args:       append([]interface{}{}, qb.args...),
	}
}
