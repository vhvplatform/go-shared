package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TransactionFunc is a function that executes operations within a transaction
type TransactionFunc func(sessCtx mongo.SessionContext) error

// Transaction executes a function within a MongoDB transaction
func (c *Client) Transaction(ctx context.Context, fn TransactionFunc) error {
	session, err := c.Client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		return nil, fn(sessCtx)
	})

	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}

	return nil
}

// TransactionWithOptions executes a function within a MongoDB transaction with custom options
func (c *Client) TransactionWithOptions(ctx context.Context, fn TransactionFunc, opts *options.TransactionOptions) error {
	session, err := c.Client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		return nil, fn(sessCtx)
	}, opts)

	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}

	return nil
}
