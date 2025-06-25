# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Essential Commands

### Building and Testing

```bash
# Build the CLI
go build -o deepwiki-cli

# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run integration tests
go test -tags=integration ./...

# Run specific package tests
go test ./pkg/scanner/ -v
go test ./internal/config/ -v

# Format code
gofmt -w .
goimports -w .
```

### Running the CLI

```bash
# Generate docs for current directory
./deepwiki-cli generate

# Generate with dry run (no actual generation)
./deepwiki-cli generate --dry-run --verbose

# Generate with custom options
./deepwiki-cli generate --output-dir ./docs --language ja --comprehensive

# Generate Docusaurus v2 documentation
./deepwiki-cli generate --format docusaurus2 --output-dir ./docusaurus-v2-site

# Generate Docusaurus v3 documentation
./deepwiki-cli generate --format docusaurus3 --output-dir ./docusaurus-v3-site

# Test basic functionality (requires OPENAI_API_KEY)
export OPENAI_API_KEY="your-test-key"
./deepwiki-cli version
```

## Architecture Overview

DeepWiki CLI is a 6-phase AI-powered documentation generator built in Go that transforms local codebases into comprehensive wiki-style documentation.

### Core Pipeline (6 Phases)

1. **File Scanner** (`pkg/scanner/`): Discovers and filters files with language detection and content analysis
2. **Text Processor** (`pkg/processor/`): Chunks code into embeddings-ready segments using language-aware splitting
3. **Embeddings Generator** (`pkg/embeddings/`): Creates vector embeddings using OpenAI's embedding models
4. **RAG System** (`pkg/rag/`): Indexes documents in BoltDB vector database for semantic search
5. **Wiki Generator** (`pkg/generator/`): Uses GPT-4 with RAG to generate structured wiki pages
6. **Output Manager** (`pkg/output/`): Creates final documentation files in markdown, JSON, Docusaurus v2, or Docusaurus v3 format

### Key Components

**Configuration System** (`internal/config/`): Multi-source config loading (YAML files, env vars, CLI flags) with validation. Config files searched in: `deepwiki.yaml`, `.deepwiki.yaml`, `~/.deepwiki.yaml`

**Logging System** (`internal/logging/`): Structured logging with slog, component-based loggers, and multiple output formats

**OpenAI Client** (`pkg/openai/`): Rate-limited API client with retry logic, token counting, and cost tracking

**Vector Database** (`pkg/embeddings/vectordb.go`): BoltDB-based persistent vector storage with similarity search (cosine, euclidean, dot product)

**CLI Commands** (`cmd/`): Cobra-based CLI with generate, config, and version subcommands

### Main Entry Point Flow

`main.go` → `cmd/root.go` → `cmd/generate.go` → 6-phase pipeline execution with progress tracking

### Language Support

40+ programming languages with automatic detection. Default include extensions: `.go`, `.py`, `.js`, `.ts`, `.java`, `.cpp`, `.rs`, `.jsx`, `.tsx`, `.html`, `.css`, `.php`, `.swift`, `.cs`, `.md`, `.txt`, `.json`, `.yaml`

### Environment Variables

- `OPENAI_API_KEY`: Required for OpenAI API access
- `DEEPWIKI_MODEL`: OpenAI model (default: gpt-4o)
- `DEEPWIKI_EMBEDDING_MODEL`: Embedding model (default: text-embedding-3-small)
- `DEEPWIKI_OUTPUT_DIR`: Output directory
- `DEEPWIKI_FORMAT`: Output format (markdown, json, docusaurus2, docusaurus3)
- `DEEPWIKI_LANGUAGE`: Output language (en, ja, zh, es, kr, vi)

### Concurrent Processing

The system uses concurrent processing throughout:

- File scanning with configurable worker pools (default: 4 workers)
- Text processing with parallel chunking
- Batch embedding generation with rate limiting
- Vector database operations optimized for concurrent access

### Output Formats

The system supports four output formats:

- **Markdown**: Traditional markdown files with organized directory structure
- **JSON**: Structured JSON format for programmatic consumption
- **Docusaurus v2.4.x** (`docusaurus2`): Ready-to-deploy Docusaurus v2 site with:
  - JavaScript configuration files (sidebars.js, docusaurus.config.js)
  - React v17 and classic theme support
  - Mermaid diagrams support
  - package.json with v2.4.3 dependencies
- **Docusaurus v3.8.x** (`docusaurus3`): Modern Docusaurus v3 site with:
  - TypeScript configuration files (sidebars.ts, docusaurus.config.ts)
  - React v18 and enhanced performance features
  - Future flags for experimental optimizations
  - ESM and TypeScript support
  - package.json with v3.8.1 dependencies and Node.js >=18 requirement
  - Enhanced frontmatter with `displayed_sidebar` support

### Testing Strategy

- Unit tests for all packages with 80%+ coverage target
- Integration tests using `go test -tags=integration`
- Benchmark tests for performance-critical components
- Test data in `testdata/` directories
- Mock implementations for external dependencies (OpenAI API)
