# CRUSH.md - Codebase Guidelines for AI Agents

## Build/Lint/Test Commands

```bash
# Install dependencies
go mod tidy

# Build application
go build -o dist/hardcover-rss cmd/api/main.go
task build

# Run application
go run cmd/api/main.go
task run

# Development with live reload
task dev

# Run all tests
go test ./... -v
task test

# Run a single test
go test -v ./path/to/package -run TestName

# Generate GraphQL code
task generate

# Download GraphQL schema
task download-schema

# Clean build artifacts
task clean
```

## Code Style Guidelines

### Imports
- Standard library imports first
- External packages second
- Internal packages last
- No blank lines within import groups

### Naming Conventions
- Variables: camelCase
- Functions: PascalCase (exported), camelCase (unexported)
- Types: PascalCase (exported), camelCase (unexported)
- Struct fields: PascalCase (exported), camelCase (unexported)

### Formatting
- Line length: ~80-100 characters
- Indentation: Tabs
- Blank lines between logical sections
- Comments for exported functions and complex logic

### Error Handling
- Explicit error checking with if statements
- Errors passed up the call stack (no panics in library code)
- Context-aware error messages with fmt.Sprintf
- Log fatal only in main functions for unrecoverable errors

### Types and Structs
- Clear, descriptive field names
- Interface definitions for dependencies
- Dependency injection through struct fields
- Embedded structs for composition

### Additional Patterns
- Context usage for timeouts/cancellation
- Table-driven tests
- Template rendering with embedded filesystem
- HTTP handlers following chi router patterns
- Middleware composition for cross-cutting concerns
