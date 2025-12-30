package clickhouse

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// BatchInserter provides efficient batch insert operations
type BatchInserter struct {
	client *Client
	batch  driver.Batch
}

// NewBatchInserter creates a new batch inserter
func (c *Client) NewBatchInserter(ctx context.Context, query string) (*BatchInserter, error) {
	batch, err := c.PrepareBatch(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare batch: %w", err)
	}

	return &BatchInserter{
		client: c,
		batch:  batch,
	}, nil
}

// Append adds a row to the batch
func (b *BatchInserter) Append(args ...interface{}) error {
	return b.batch.Append(args...)
}

// Send commits the batch to ClickHouse
func (b *BatchInserter) Send() error {
	return b.batch.Send()
}

// Abort aborts the batch operation
func (b *BatchInserter) Abort() error {
	return b.batch.Abort()
}

// BatchInsert is a helper function to insert multiple rows efficiently
func (c *Client) BatchInsert(ctx context.Context, table string, columns []string, rows [][]interface{}) error {
	if len(rows) == 0 {
		return nil
	}

	// Build the INSERT query
	query := fmt.Sprintf("INSERT INTO %s", table)

	// Prepare batch
	batch, err := c.NewBatchInserter(ctx, query)
	if err != nil {
		return err
	}

	// Add all rows to batch
	for _, row := range rows {
		if err := batch.Append(row...); err != nil {
			batch.Abort()
			return fmt.Errorf("failed to append row to batch: %w", err)
		}
	}

	// Send the batch
	if err := batch.Send(); err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}

	return nil
}
