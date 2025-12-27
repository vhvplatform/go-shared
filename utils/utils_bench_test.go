package utils

import (
	"testing"
)

func BenchmarkToSnakeCase(b *testing.B) {
	input := "MyVariableName"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ToSnakeCase(input)
	}
}

func BenchmarkToSnakeCaseLong(b *testing.B) {
	input := "MyVeryLongVariableNameWithManyWords"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ToSnakeCase(input)
	}
}

func BenchmarkIsValidEmail(b *testing.B) {
	email := "test@example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsValidEmail(email)
	}
}

func BenchmarkIsValidEmailInvalid(b *testing.B) {
	email := "invalid-email"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsValidEmail(email)
	}
}

func BenchmarkContains(b *testing.B) {
	slice := []string{"apple", "banana", "cherry", "date", "elderberry"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Contains(slice, "cherry")
	}
}

func BenchmarkContainsMiss(b *testing.B) {
	slice := []string{"apple", "banana", "cherry", "date", "elderberry"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Contains(slice, "fig")
	}
}

func BenchmarkRemove(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		slice := []string{"apple", "banana", "cherry", "date", "elderberry"}
		b.StartTimer()
		_ = Remove(slice, "cherry")
	}
}

func BenchmarkHashPassword(b *testing.B) {
	password := "mySecurePassword123!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = HashPassword(password)
	}
}

func BenchmarkCheckPassword(b *testing.B) {
	password := "mySecurePassword123!"
	hash, _ := HashPassword(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CheckPassword(password, hash)
	}
}
