# Contributing to SaaS Shared Go Library

Thank you for your interest in contributing to the SaaS Shared Go Library! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Coding Standards](#coding-standards)
- [Package Guidelines](#package-guidelines)

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for all contributors.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/go-shared.git
   cd go-shared
   ```
3. **Add the upstream remote**:
   ```bash
   git remote add upstream https://github.com/vhvcorp/go-shared.git
   ```

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Git
- golangci-lint for linting

### Install Dependencies

```bash
go mod download
```

### Install Development Tools

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## Making Changes

### 1. Create a Branch

Create a branch for your changes:

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

Branch naming conventions:
- `feature/` - for new features
- `fix/` - for bug fixes
- `docs/` - for documentation changes
- `refactor/` - for code refactoring

### 2. Make Your Changes

- Write clear, concise code
- Follow the existing code style
- Add tests for new functionality
- Update documentation as needed

### 3. Test Your Changes

Run the full test suite:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out -covermode=atomic ./...

# View coverage report
go tool cover -html=coverage.out
```

### 4. Lint Your Code

```bash
golangci-lint run ./...
```

## Testing

### Writing Tests

- All new code must include unit tests
- Tests should be in `*_test.go` files alongside the code
- Use table-driven tests when appropriate
- Mock external dependencies

Example test structure:

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:    "valid input",
            input:   "test",
            want:    "expected",
            wantErr: false,
        },
        {
            name:    "invalid input",
            input:   "",
            want:    "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := MyFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("MyFunction() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("MyFunction() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Coverage Requirements

- New code should maintain or improve overall coverage
- Aim for at least 80% coverage for new packages
- Critical paths should have 100% coverage

## Submitting Changes

### 1. Commit Your Changes

Write clear commit messages:

```bash
git commit -m "feat: add new authentication method

- Implement OAuth2 authentication
- Add comprehensive tests
- Update documentation"
```

Commit message format:
- `feat:` - new feature
- `fix:` - bug fix
- `docs:` - documentation changes
- `test:` - test updates
- `refactor:` - code refactoring
- `chore:` - maintenance tasks

### 2. Push to Your Fork

```bash
git push origin feature/your-feature-name
```

### 3. Create a Pull Request

1. Go to the original repository on GitHub
2. Click "New Pull Request"
3. Select your fork and branch
4. Fill in the PR template with:
   - Clear description of changes
   - Related issue numbers
   - Testing performed
   - Breaking changes (if any)

### 4. Code Review

- Address reviewer feedback promptly
- Make requested changes in new commits
- Keep the PR focused and reasonably sized

## Coding Standards

### Go Code Style

Follow standard Go conventions:

- Use `gofmt` to format code
- Use meaningful variable and function names
- Keep functions small and focused
- Document exported functions and types
- Handle errors explicitly

### Documentation

- Add godoc comments for all exported types and functions
- Include usage examples in documentation
- Update README.md for new packages or significant changes

Example documentation:

```go
// UserService provides user management functionality.
// It handles user creation, retrieval, updates, and deletion
// with full multi-tenant support.
type UserService struct {
    db *mongodb.Database
}

// NewUserService creates a new UserService instance.
// It requires a MongoDB database connection and will panic if db is nil.
//
// Example:
//   db := mongodb.Connect(ctx, "mongodb://localhost:27017")
//   userService := NewUserService(db)
func NewUserService(db *mongodb.Database) *UserService {
    if db == nil {
        panic("database connection is required")
    }
    return &UserService{db: db}
}
```

### Error Handling

- Return errors, don't panic (except in initialization)
- Wrap errors with context using `fmt.Errorf`
- Use custom error types for domain-specific errors

```go
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}
```

## Package Guidelines

### Creating a New Package

When adding a new package:

1. **Create the package directory** with a clear, descriptive name
2. **Add a package-level godoc comment** explaining the package purpose
3. **Include examples** in the documentation
4. **Write comprehensive tests**
5. **Update the main README.md** with package description and examples
6. **Ensure no circular dependencies**

### Package Structure

```
newpackage/
├── doc.go              # Package documentation
├── newpackage.go       # Main implementation
├── newpackage_test.go  # Tests
├── types.go            # Type definitions (if needed)
└── examples/           # Usage examples (optional)
```

### Backward Compatibility

- Maintain backward compatibility within major versions
- Deprecate features before removing them
- Document breaking changes clearly
- Use semantic versioning

### Dependencies

- Minimize external dependencies
- Use well-maintained, popular libraries
- Avoid dependencies with restrictive licenses
- Keep dependencies up to date

## Questions?

If you have questions:

1. Check existing issues and discussions
2. Create a new issue with the `question` label
3. Be specific and provide context

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Thank You!

Your contributions help make this library better for everyone. We appreciate your time and effort!
