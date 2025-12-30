package clickhouse

import (
	"context"
	"fmt"
)

// Transaction represents a ClickHouse transaction-like operation
// Note: ClickHouse has limited transaction support compared to traditional RDBMS.
// This provides a consistent API for batch operations that should be executed together.
type Transaction struct {
	client     *Client
	ctx        context.Context
	statements []statement
}

type statement struct {
	query string
	args  []interface{}
}

// BeginTx creates a new transaction-like context
// Note: This doesn't create a real database transaction in ClickHouse
// but provides a way to group operations together
func (c *Client) BeginTx(ctx context.Context) *Transaction {
	return &Transaction{
		client:     c,
		ctx:        ctx,
		statements: make([]statement, 0),
	}
}

// Exec adds a query to the transaction
func (t *Transaction) Exec(query string, args ...interface{}) error {
	t.statements = append(t.statements, statement{
		query: query,
		args:  args,
	})
	return nil
}

// Commit executes all queued statements
func (t *Transaction) Commit() error {
	for i, stmt := range t.statements {
		if err := t.client.Exec(t.ctx, stmt.query, stmt.args...); err != nil {
			return fmt.Errorf("failed to execute statement %d: %w", i+1, err)
		}
	}
	return nil
}

// Rollback clears all queued statements
// Note: Since ClickHouse doesn't support traditional rollback,
// this only clears the queued statements before they're executed
func (t *Transaction) Rollback() error {
	t.statements = make([]statement, 0)
	return nil
}

// WithTransaction executes a function within a transaction-like context
func (c *Client) WithTransaction(ctx context.Context, fn func(*Transaction) error) error {
	tx := c.BeginTx(ctx)

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
