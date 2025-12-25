# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of shared library extracted from monorepo
- `auth` package for authentication and authorization helpers
- `config` package for configuration management with Viper
- `context` package for request context management
- `errors` package for custom error types
- `httpclient` package for HTTP client utilities
- `jwt` package for JWT token management
- `logger` package for structured logging with Zap
- `middleware` package for Gin middleware collection
- `mongodb` package for MongoDB utilities with tenant isolation
- `rabbitmq` package for RabbitMQ integration
- `redis` package for Redis client with multi-tenant support
- `response` package for standard API responses
- `tenant` package for multi-strategy tenant resolution
- `utils` package for common utility functions
- `validation` package for request validation
- Comprehensive documentation and examples
- GitHub Actions workflows for CI/CD
- MIT License

### Changed
- Module path changed from `github.com/longvhv/saas-framework-go/pkg` to `github.com/vhvcorp/go-shared`

### Deprecated
- None

### Removed
- None

### Fixed
- None

### Security
- None

## [1.0.0] - TBD

Initial stable release.

---

## Version History Guidelines

### Types of Changes
- **Added** for new features
- **Changed** for changes in existing functionality
- **Deprecated** for soon-to-be removed features
- **Removed** for now removed features
- **Fixed** for any bug fixes
- **Security** for vulnerability fixes

### Version Numbers
- **MAJOR** version (X.0.0) - Incompatible API changes
- **MINOR** version (0.X.0) - Add functionality in a backwards compatible manner
- **PATCH** version (0.0.X) - Backwards compatible bug fixes
