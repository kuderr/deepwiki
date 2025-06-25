# Contributing to DeepWiki CLI

Thank you for your interest in contributing to DeepWiki CLI! This document provides guidelines and information for contributors.

## 🤝 Ways to Contribute

### 🐛 Bug Reports

- Report bugs through [GitHub Issues](https://github.com/your-org/deepwiki-cli/issues)
- Include detailed reproduction steps
- Provide environment information
- Attach relevant logs and configuration

### 💡 Feature Requests

- Suggest new features via [GitHub Issues](https://github.com/your-org/deepwiki-cli/issues)
- Explain the use case and benefits
- Consider implementation complexity
- Discuss with maintainers before starting work

### 📖 Documentation Improvements

- Fix typos and improve clarity
- Add examples and tutorials
- Update configuration documentation
- Improve troubleshooting guides

### 🔧 Code Contributions

- Fix bugs and implement features
- Improve performance and reliability
- Add comprehensive tests
- Follow coding standards

## 🚀 Getting Started

### Prerequisites

- **Go 1.24+** installed
- **Git** for version control
- **OpenAI API Key** for testing (optional for some contributions)
- **Make** for build automation (optional)

### Development Setup

1. **Fork and Clone**

```bash
# Fork the repository on GitHub
# Then clone your fork
git clone https://github.com/YOUR-USERNAME/deepwiki-cli.git
cd deepwiki-cli

# Add upstream remote
git remote add upstream https://github.com/your-org/deepwiki-cli.git
```

2. **Install Dependencies**

```bash
# Download Go modules
go mod download

# Verify installation
go mod verify
```

3. **Build and Test**

```bash
# Build the project
go build -o deepwiki-cli

# Run tests
go test ./...

# Run specific tests
go test ./pkg/scanner/ -v
go test ./internal/config/ -v
```

4. **Verify Setup**

```bash
# Test basic functionality (requires OPENAI_API_KEY)
export OPENAI_API_KEY="your-test-key"
./deepwiki-cli version
./deepwiki-cli generate --dry-run --verbose
```

## 📝 Development Workflow

### 1. Create Feature Branch

```bash
# Update your fork
git checkout main
git pull upstream main

# Create feature branch
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b fix/issue-number-description
```

### 2. Make Changes

- Follow the coding standards (see below)
- Write tests for new functionality
- Update documentation as needed
- Keep commits focused and atomic

### 3. Test Your Changes

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run integration tests
go test -tags=integration ./...

# Run benchmarks
go test -bench=. ./pkg/output/

# Test with real projects
export OPENAI_API_KEY="your-key"
./deepwiki-cli generate --dry-run /path/to/test/project
```

### 4. Commit Changes

```bash
# Stage changes
git add .

# Commit with descriptive message
git commit -m "feat: add support for custom templates

- Add template loading functionality
- Support for custom Markdown templates
- Add tests for template processing
- Update documentation

Fixes #123"
```

### 5. Submit Pull Request

```bash
# Push to your fork
git push origin feature/your-feature-name

# Create pull request on GitHub
# Fill out the pull request template
# Link to related issues
```

## 🏗️ Project Structure

Understanding the codebase organization:

```
deepwiki-cli/
├── cmd/                    # CLI commands (Cobra)
│   ├── root.go            # Root command and global flags
│   ├── generate.go        # Generate command implementation
│   └── config.go          # Config commands (init, validate)
├── internal/              # Internal packages (not for external use)
│   ├── config/            # Configuration management
│   ├── logging/           # Structured logging with slog
│   └── prompts/           # AI prompt templates and management
├── pkg/                   # Public packages (reusable components)
│   ├── scanner/           # File discovery and filtering
│   ├── processor/         # Text processing and chunking
│   ├── embeddings/        # Vector embeddings and search
│   ├── rag/              # Retrieval-Augmented Generation
│   ├── generator/         # Wiki structure and content generation
│   ├── openai/           # OpenAI API client integration
│   └── output/           # File generation and formatting
├── examples/              # Example configurations and tutorials
├── docs/                 # Project documentation
├── integration_test.go   # Integration tests
├── go.mod               # Go module definition
├── go.sum              # Go module checksums
├── main.go             # Application entry point
├── README.md           # Project overview
├── CONTRIBUTING.md     # This file
└── TODO.md            # Development progress tracking
```

### Package Responsibilities

- **cmd/**: CLI interface, argument parsing, user interaction
- **internal/config/**: YAML parsing, environment variables, validation
- **internal/logging/**: Structured logging, multiple outputs, component isolation
- **internal/prompts/**: AI prompt templates, localization, prompt engineering
- **pkg/scanner/**: File system traversal, language detection, content analysis
- **pkg/processor/**: Text chunking, tokenization, preprocessing
- **pkg/embeddings/**: Vector generation, similarity search, caching
- **pkg/rag/**: Document retrieval, context ranking, relevance scoring
- **pkg/generator/**: Wiki structure creation, content generation, progress tracking
- **pkg/openai/**: API client, rate limiting, error handling, cost tracking
- **pkg/output/**: File generation, formatting, organization, concurrent processing

## 🎯 Coding Standards

### Go Style Guidelines

Follow standard Go conventions:

1. **Formatting**

```bash
# Use gofmt for consistent formatting
gofmt -w .

# Use goimports for import management
goimports -w .
```

2. **Naming Conventions**

```go
// Use CamelCase for exported functions
func GenerateWikiStructure() {}

// Use camelCase for unexported functions
func parseXMLResponse() {}

// Use descriptive names
func (s *Scanner) ScanDirectory(path string) (*ScanResult, error) {}

// Avoid abbreviations unless widely understood
func ProcessConfiguration() {} // Good
func ProcConf() {}             // Avoid
```

3. **Error Handling**

```go
// Always handle errors explicitly
result, err := someFunction()
if err != nil {
    return fmt.Errorf("failed to process data: %w", err)
}

// Use wrapped errors for context
if err := validateInput(data); err != nil {
    return fmt.Errorf("input validation failed: %w", err)
}
```

4. **Package Documentation**

```go
// Package scanner provides file discovery and filtering capabilities.
//
// The scanner package implements intelligent file system traversal with
// support for multiple programming languages, content analysis, and
// configurable filtering rules.
package scanner
```

5. **Function Documentation**

```go
// ScanDirectory recursively scans a directory and returns filtered file information.
//
// The function traverses the directory tree starting from the given path,
// applies the configured filters, and returns detailed information about
// each relevant file found.
//
// Parameters:
//   - path: The root directory path to scan
//   - options: Filtering and processing options
//
// Returns:
//   - *ScanResult: Contains discovered files and statistics
//   - error: Any error encountered during scanning
func (s *Scanner) ScanDirectory(path string, options *ScanOptions) (*ScanResult, error) {
    // Implementation...
}
```

### Testing Standards

1. **Test Coverage**

- Aim for 80%+ test coverage
- Test both success and error cases
- Include edge cases and boundary conditions

2. **Test Organization**

```go
func TestScanDirectory(t *testing.T) {
    tests := []struct {
        name     string
        path     string
        options  *ScanOptions
        want     int  // expected file count
        wantErr  bool
    }{
        {
            name: "valid directory",
            path: "testdata/sample-project",
            options: &ScanOptions{
                IncludeExtensions: []string{".go", ".md"},
            },
            want:    5,
            wantErr: false,
        },
        {
            name:    "non-existent directory",
            path:    "testdata/does-not-exist",
            options: &ScanOptions{},
            want:    0,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            scanner := NewScanner(tt.options)
            result, err := scanner.ScanDirectory(tt.path)

            if (err != nil) != tt.wantErr {
                t.Errorf("ScanDirectory() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if !tt.wantErr && len(result.Files) != tt.want {
                t.Errorf("ScanDirectory() files = %d, want %d", len(result.Files), tt.want)
            }
        })
    }
}
```

3. **Test Data Management**

```bash
# Create test data directories
mkdir -p testdata/sample-project
echo "package main" > testdata/sample-project/main.go
echo "# Test Project" > testdata/sample-project/README.md

# Use t.TempDir() for temporary files
func TestFileCreation(t *testing.T) {
    tempDir := t.TempDir() // Automatically cleaned up
    filePath := filepath.Join(tempDir, "test.txt")
    // ... test implementation
}
```

4. **Benchmark Tests**

```go
func BenchmarkScanDirectory(b *testing.B) {
    scanner := NewScanner(&ScanOptions{})

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := scanner.ScanDirectory("testdata/large-project")
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### Documentation Standards

1. **Code Comments**

```go
// Good: Explains why, not just what
// Use exponential backoff to handle temporary API failures
// that commonly occur during high traffic periods
time.Sleep(time.Duration(attempt) * time.Second)

// Avoid: Obvious comments
// Increment the counter
counter++
```

2. **README Updates**

- Update README.md for user-facing changes
- Include examples for new features
- Update configuration documentation

3. **API Documentation**

- Document all exported functions and types
- Include usage examples
- Explain parameter constraints

## 🧪 Testing Guidelines

### Unit Tests

1. **Test Structure**

```go
func TestFunctionName(t *testing.T) {
    // Arrange: Set up test data
    input := "test input"
    expected := "expected output"

    // Act: Execute the function
    result, err := FunctionName(input)

    // Assert: Verify results
    if err != nil {
        t.Fatalf("FunctionName() error = %v", err)
    }
    if result != expected {
        t.Errorf("FunctionName() = %v, want %v", result, expected)
    }
}
```

2. **Test Coverage**

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# View coverage by package
go test -cover ./pkg/scanner/
go test -cover ./internal/config/
```

3. **Mock Dependencies**

```go
// Use interfaces for testability
type OpenAIClient interface {
    GenerateCompletion(prompt string) (string, error)
}

// Create mock implementations
type MockOpenAIClient struct {
    Response string
    Error    error
}

func (m *MockOpenAIClient) GenerateCompletion(prompt string) (string, error) {
    return m.Response, m.Error
}
```

### Integration Tests

1. **Test Real Workflows**

```go
// +build integration

func TestFullDocumentationGeneration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // Create test project
    projectDir := createTestProject(t)
    defer os.RemoveAll(projectDir)

    // Run full generation pipeline
    // ... test implementation
}
```

2. **Environment Setup**

```bash
# Run integration tests
go test -tags=integration ./...

# Skip integration tests
go test -short ./...
```

### Performance Testing

1. **Benchmark Critical Paths**

```go
func BenchmarkDocumentGeneration(b *testing.B) {
    generator := NewGenerator(&Config{})
    testData := loadTestData()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := generator.GenerateWiki(testData)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

2. **Memory Profiling**

```bash
# Run with memory profiling
go test -bench=. -memprofile=mem.prof

# Analyze profile
go tool pprof mem.prof
```

## 📋 Pull Request Guidelines

### PR Requirements

1. **Description**

- Clear title describing the change
- Detailed description of what was changed and why
- Link to related issues
- Screenshots for UI changes (if applicable)

2. **Code Quality**

- All tests pass
- Code coverage maintained or improved
- No linting errors
- Documentation updated

3. **Commit Messages**
   Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add support for custom output templates

Add functionality to load and use custom Markdown templates
for documentation generation. This allows users to customize
the appearance and structure of generated documentation.

- Add template loading from filesystem
- Support for Go template syntax
- Add validation for template syntax
- Include comprehensive tests
- Update documentation with examples

Fixes #123
Closes #124
```

Types:

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

### PR Template

When creating a pull request, include:

```markdown
## Description

Brief description of changes

## Type of Change

- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to change)
- [ ] Documentation update

## Testing

- [ ] Unit tests pass
- [ ] Integration tests pass (if applicable)
- [ ] Manual testing completed

## Checklist

- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Code is commented where necessary
- [ ] Documentation updated
- [ ] Tests added for new functionality
- [ ] All tests pass

## Related Issues

Fixes #(issue number)
```

## 🔍 Code Review Process

### For Contributors

1. **Self-Review**

- Review your own code before submitting
- Check for typos and formatting issues
- Ensure tests cover new functionality
- Verify documentation is updated

2. **Respond to Feedback**

- Address all review comments
- Ask questions if feedback is unclear
- Make requested changes promptly
- Re-request review after changes

### For Reviewers

1. **Review Criteria**

- Code correctness and logic
- Test coverage and quality
- Documentation completeness
- Performance implications
- Security considerations

2. **Feedback Guidelines**

- Be constructive and specific
- Explain the reasoning behind suggestions
- Distinguish between required changes and suggestions
- Acknowledge good practices

## 🐛 Bug Reports

### Bug Report Template

````markdown
## Bug Description

A clear description of the bug

## Steps to Reproduce

1. Step one
2. Step two
3. Step three

## Expected Behavior

What should have happened

## Actual Behavior

What actually happened

## Environment

- OS: [e.g., macOS 12.0, Ubuntu 20.04]
- Go Version: [e.g., 1.21.0]
- DeepWiki CLI Version: [e.g., 1.0.0]
- OpenAI Model: [e.g., gpt-4o]

## Configuration

```yaml
# Include relevant configuration (redact API keys)
```
````

## Additional Context

Any other relevant information

````

## 💡 Feature Requests

### Feature Request Template

```markdown
## Feature Description
Clear description of the proposed feature

## Use Case
Describe the problem this feature would solve

## Proposed Solution
How you envision this feature working

## Alternatives Considered
Other approaches you've thought about

## Additional Context
Any other relevant information
````

## 🏷️ Release Process

### Versioning

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR**: Incompatible API changes
- **MINOR**: Backwards-compatible functionality additions
- **PATCH**: Backwards-compatible bug fixes

### Release Checklist

1. **Pre-Release**

- [ ] All tests pass
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Version bumped in appropriate files

2. **Release**

- [ ] Tag created with proper version
- [ ] Release notes written
- [ ] Binaries built and uploaded
- [ ] Documentation deployed

## 🤔 Questions and Support

### Getting Help

1. **Documentation**: Check the [README](README.md) and [docs/](docs/) first
2. **Search Issues**: Look for existing issues and discussions
3. **Ask Questions**: Use [GitHub Discussions](https://github.com/your-org/deepwiki-cli/discussions)
4. **Report Bugs**: Create a [GitHub Issue](https://github.com/your-org/deepwiki-cli/issues)

### Communication Channels

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: Questions, ideas, and general discussion
- **Pull Requests**: Code contributions and reviews

## 📜 License

By contributing to DeepWiki CLI, you agree that your contributions will be licensed under the same license as the project (MIT License).

## 🙏 Recognition

Contributors will be recognized in:

- Release notes for significant contributions
- README.md contributors section
- GitHub contributor graphs

Thank you for contributing to DeepWiki CLI! 🚀
