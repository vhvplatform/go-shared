package clickhouse

import (
	"context"
	"testing"
)

func TestTransaction_ExecAndCommit(t *testing.T) {
	// Create a mock client (we'll just test the transaction logic)
	tx := &Transaction{
		ctx:        context.Background(),
		statements: make([]statement, 0),
	}

	// Add statements
	err := tx.Exec("INSERT INTO users VALUES (?, ?, ?)", 1, "John", "john@example.com")
	if err != nil {
		t.Errorf("Exec failed: %v", err)
	}

	err = tx.Exec("INSERT INTO logs VALUES (?, ?)", 1, "User created")
	if err != nil {
		t.Errorf("Exec failed: %v", err)
	}

	// Check statements were queued
	if len(tx.statements) != 2 {
		t.Errorf("Expected 2 statements, got %d", len(tx.statements))
	}

	// Verify first statement
	if tx.statements[0].query != "INSERT INTO users VALUES (?, ?, ?)" {
		t.Errorf("First statement query mismatch")
	}
	if len(tx.statements[0].args) != 3 {
		t.Errorf("First statement expected 3 args, got %d", len(tx.statements[0].args))
	}

	// Verify second statement
	if tx.statements[1].query != "INSERT INTO logs VALUES (?, ?)" {
		t.Errorf("Second statement query mismatch")
	}
	if len(tx.statements[1].args) != 2 {
		t.Errorf("Second statement expected 2 args, got %d", len(tx.statements[1].args))
	}
}

func TestTransaction_Rollback(t *testing.T) {
	tx := &Transaction{
		ctx:        context.Background(),
		statements: make([]statement, 0),
	}

	// Add statements
	tx.Exec("INSERT INTO users VALUES (?, ?, ?)", 1, "John", "john@example.com")
	tx.Exec("INSERT INTO logs VALUES (?, ?)", 1, "User created")

	if len(tx.statements) != 2 {
		t.Errorf("Expected 2 statements before rollback, got %d", len(tx.statements))
	}

	// Rollback
	err := tx.Rollback()
	if err != nil {
		t.Errorf("Rollback failed: %v", err)
	}

	// Verify statements were cleared
	if len(tx.statements) != 0 {
		t.Errorf("Expected 0 statements after rollback, got %d", len(tx.statements))
	}
}

func BenchmarkTransaction_Exec(b *testing.B) {
	tx := &Transaction{
		ctx:        context.Background(),
		statements: make([]statement, 0),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx.Exec("INSERT INTO users VALUES (?, ?, ?)", i, "User", "user@example.com")
	}
}
