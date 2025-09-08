# Contributing to git-assist

Thank you for your interest in contributing to git-assist! This document provides guidelines and information for contributors.

## Code of Conduct

This project adheres to a code of conduct that we expect all contributors to follow. Please be respectful and constructive in all interactions.

## Getting Started

### Prerequisites

- Go 1.19 or later
- Git 2.20 or later
- Make (for build automation)

### Development Setup

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/your-username/git-assist.git
   cd git-assist
   ```

3. Set up the development environment:
   ```bash
   make dev-setup
   ```

4. Verify the setup:
   ```bash
   make test
   make build
   ./git-assist --version
   ```

### Project Structure

```
git-assist/
├── cmd/                    # Main application entry point
├── internal/
│   ├── ai/                # AI provider integrations
│   ├── cli/               # Command-line interface
│   ├── config/            # Configuration management
│   ├── git/               # Git operations
│   └── repository/        # Repository analysis
├── docs/                  # Documentation
├── scripts/               # Build and deployment scripts
└── tests/                 # Integration tests
```

## Development Workflow

### 1. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
```

Use descriptive branch names:
- `feature/add-github-integration`
- `fix/config-file-parsing`
- `docs/update-installation-guide`

### 2. Make Changes

- Write clean, readable code
- Follow Go conventions and best practices
- Add tests for new functionality
- Update documentation as needed

### 3. Test Your Changes

```bash
# Run unit tests
make test

# Run integration tests
make test-integration

# Test the built binary
make build
./git-assist --help
```

### 4. Commit Your Changes

Use conventional commit format:
```bash
git commit -m "feat(analyze): add dependency analysis for Python projects"
git commit -m "fix(config): handle missing config file gracefully"
git commit -m "docs(readme): update installation instructions"
```

### 5. Push and Create Pull Request

```bash
git push origin feature/your-feature-name
```

Then create a pull request on GitHub with:
- Clear description of changes
- Reference to any related issues
- Screenshots for UI changes
- Test results and verification steps

## Coding Standards

### Go Style Guide

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for formatting
- Use `golint` and `go vet` for static analysis
- Write meaningful variable and function names

### Code Organization

- Keep functions small and focused
- Use interfaces for testability
- Handle errors explicitly
- Add comments for public APIs

### Example Code Style

```go
// AnalyzeRepository performs comprehensive repository analysis
func AnalyzeRepository(path string, options AnalysisOptions) (*RepositoryAnalysis, error) {
    if path == "" {
        return nil, fmt.Errorf("repository path cannot be empty")
    }
    
    repo, err := git.OpenRepository(path)
    if err != nil {
        return nil, fmt.Errorf("failed to open repository: %w", err)
    }
    
    analysis := &RepositoryAnalysis{
        Path:      path,
        Timestamp: time.Now(),
    }
    
    if options.IncludeStructure {
        if err := analyzeStructure(repo, analysis); err != nil {
            return nil, fmt.Errorf("structure analysis failed: %w", err)
        }
    }
    
    return analysis, nil
}
```

## Testing Guidelines

### Unit Tests

- Write tests for all public functions
- Use table-driven tests for multiple scenarios
- Mock external dependencies
- Aim for high test coverage

```go
func TestAnalyzeRepository(t *testing.T) {
    tests := []struct {
        name    string
        path    string
        options AnalysisOptions
        want    *RepositoryAnalysis
        wantErr bool
    }{
        {
            name: "valid repository",
            path: "/path/to/repo",
            options: AnalysisOptions{IncludeStructure: true},
            want: &RepositoryAnalysis{Path: "/path/to/repo"},
            wantErr: false,
        },
        {
            name: "empty path",
            path: "",
            options: AnalysisOptions{},
            want: nil,
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := AnalyzeRepository(tt.path, tt.options)
            if (err != nil) != tt.wantErr {
                t.Errorf("AnalyzeRepository() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("AnalyzeRepository() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Integration Tests

- Test complete workflows
- Use temporary repositories
- Test error conditions
- Verify command-line interface

## Documentation

### Code Documentation

- Document all public functions and types
- Use Go doc conventions
- Include examples for complex APIs
- Keep documentation up to date

### User Documentation

- Update README.md for new features
- Add examples and use cases
- Update command help text
- Create or update wiki pages

## Submitting Changes

### Pull Request Guidelines

1. **Title**: Use descriptive titles that explain the change
2. **Description**: Include:
   - What the change does
   - Why it's needed
   - How to test it
   - Any breaking changes

3. **Size**: Keep PRs focused and reasonably sized
4. **Tests**: Include tests for new functionality
5. **Documentation**: Update docs for user-facing changes

### Review Process

1. Automated checks must pass (tests, linting, build)
2. At least one maintainer review required
3. Address review feedback promptly
4. Squash commits before merging (if requested)

### Merge Requirements

- All tests passing
- Code review approved
- Documentation updated
- No merge conflicts
- Conventional commit format

## Issue Reporting

### Bug Reports

Include:
- git-assist version (`git-assist --version`)
- Operating system and version
- Go version (if building from source)
- Steps to reproduce
- Expected vs actual behavior
- Error messages or logs

### Feature Requests

Include:
- Clear description of the feature
- Use cases and motivation
- Proposed implementation (if any)
- Alternatives considered

### Questions and Support

- Check existing documentation first
- Search existing issues
- Use GitHub Discussions for questions
- Provide context and examples

## Release Process

### Version Numbering

We follow [Semantic Versioning](https://semver.org/):
- MAJOR: Breaking changes
- MINOR: New features (backward compatible)
- PATCH: Bug fixes (backward compatible)

### Release Checklist

1. Update version in code
2. Update CHANGELOG.md
3. Create release notes
4. Tag the release
5. Build and publish binaries
6. Update documentation

## Community

### Communication Channels

- GitHub Issues: Bug reports and feature requests
- GitHub Discussions: Questions and community support
- Pull Requests: Code contributions and reviews

### Getting Help

- Read the documentation thoroughly
- Search existing issues and discussions
- Ask specific questions with context
- Be patient and respectful

## Recognition

Contributors are recognized in:
- CHANGELOG.md for significant contributions
- GitHub contributors page
- Release notes for major features

Thank you for contributing to git-assist and helping make Git more accessible and intelligent for developers everywhere!
