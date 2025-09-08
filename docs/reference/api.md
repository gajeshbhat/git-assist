# API Reference

This page contains auto-generated documentation from the Go source code.

## Package: internal/git

### Repository Operations

```go

// Repository represents a Git repository
type Repository struct {
    path string
}

// OpenRepository opens a Git repository at the specified path
func OpenRepository(path string) (*Repository, error)

// Analyze performs comprehensive analysis of the repository
func (r *Repository) Analyze() (*Analysis, error)
```


*This documentation is auto-generated from Go source code.*
