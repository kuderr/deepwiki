# 🤖 DeepWiki CLI

**AI-Powered Documentation Generator for Local Codebases**

DeepWiki CLI is a powerful Go-based command-line tool that automatically generates comprehensive, structured documentation for your local projects using AI. It analyzes your codebase, understands the architecture, and creates detailed wiki-style documentation with minimal manual effort.

[![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)](https://golang.org)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/kuderr/deepwiki)

## ✨ Features

### ✅ **Complete Documentation Generation Pipeline**

- **Multi-language support**: Go, Python, JavaScript, TypeScript, Java, C++, and 40+ more
- **AI-Powered Analysis**: Uses GPT-4, Claude, or other LLMs for intelligent code understanding
- **Structured Output**: Generates organized wiki pages with clear hierarchy and navigation
- **Multiple Formats**: Markdown and JSON output with customizable templates

### ✅ **Flexible Provider Architecture**

- **Multiple LLM Providers**: OpenAI (GPT-4o, GPT-3.5-turbo) and Anthropic (Claude 3.5 Sonnet/Haiku)
- **Multiple Embedding Providers**: OpenAI, Voyage AI, and **Local Ollama** support
- **Mix & Match**: Use Claude for LLM + local Ollama for embeddings to minimize costs
- **Local-First Option**: Full offline capability with Ollama embeddings
- **Cost Optimization**: Choose providers based on your budget and privacy needs

### ✅ **Advanced Processing Capabilities**

- **Vector Embeddings**: Semantic search with multiple embedding providers
- **RAG System**: Retrieval-Augmented Generation for accurate, contextual documentation
- **Concurrent Processing**: Optimized performance with configurable worker pools
- **Memory Management**: Efficient processing of large codebases with memory limits
- **Local Embeddings**: Run embeddings locally with Ollama - no API costs!

### ✅ **Production-Ready Features**

- **Simple CLI**: One-command operation with comprehensive configuration options
- **Visual Progress**: Real-time progress bars with ETA and processing statistics
- **Error Recovery**: Robust error handling with detailed diagnostic information
- **Dry Run Mode**: Preview generation without creating files
- **Integration Ready**: Perfect for CI/CD pipelines and automated workflows

## Installation

```bash
# Clone and build
git clone <repository>
cd deepwiki-cli
go build -o deepwiki-cli

# Or install directly
go install github.com/kuderr/deepwiki@latest
```

## Quick Start

### 1. Generate Documentation

```bash
# Generate docs for current directory
deepwiki-cli generate

# Generate docs for specific directory
deepwiki-cli generate /path/to/project

# With custom options
deepwiki-cli generate --output-dir ./docs --language ja
```

### 2. Configuration Examples

```bash
# Use Claude LLM + local Ollama embeddings (cost-effective)
export ANTHROPIC_API_KEY="your-key"
export DEEPWIKI_LLM_PROVIDER="anthropic"
export DEEPWIKI_EMBEDDING_PROVIDER="ollama"
deepwiki-cli generate

# Use OpenAI for everything (simple setup)
export OPENAI_API_KEY="your-key"
export DEEPWIKI_LLM_PROVIDER="openai"
export DEEPWIKI_EMBEDDING_PROVIDER="openai"
deepwiki-cli generate

# Use Voyage AI for high-quality embeddings
export OPENAI_API_KEY="your-llm-key"
export VOYAGE_API_KEY="your-embedding-key"
export DEEPWIKI_LLM_PROVIDER="openai"
export DEEPWIKI_EMBEDDING_PROVIDER="voyage"
deepwiki-cli generate
```

### 3. Advanced Configuration

```bash
# Create configuration template
deepwiki-cli config init

# Validate configuration
deepwiki-cli config validate

# Use custom config file
deepwiki-cli generate --config my-config.yaml
```

### 4. Environment Setup

```bash
# LLM Provider Setup (choose one)
export OPENAI_API_KEY="your-openai-key"           # For OpenAI GPT models
export ANTHROPIC_API_KEY="your-anthropic-key"     # For Claude models

# Embedding Provider Setup (choose one)
export OPENAI_API_KEY="your-openai-key"           # For OpenAI embeddings
export VOYAGE_API_KEY="your-voyage-key"           # For Voyage AI embeddings
# Ollama doesn't need API key, just make sure it's running locally

# Provider Selection
export DEEPWIKI_LLM_PROVIDER="openai"             # or "anthropic"
export DEEPWIKI_EMBEDDING_PROVIDER="ollama"       # or "openai" or "voyage"
export OLLAMA_BASE_URL="http://localhost:11434"   # For Ollama (optional)
```

## Configuration

### Example Configuration File

```yaml
# Provider Configuration
providers:
  llm:
    provider: "openai" # "openai" or "anthropic"
    api_key: "${OPENAI_API_KEY}"
    model: "gpt-4o"
    max_tokens: 4000
    temperature: 0.1
    request_timeout: "3m"
    max_retries: 3
    rate_limit_rps: 2.0

  embedding:
    provider: "ollama" # "openai", "voyage", or "ollama"
    model: "nomic-embed-text"
    base_url: "http://localhost:11434"
    request_timeout: "30s"
    max_retries: 3
    rate_limit_rps: 10.0

processing:
  chunk_size: 350
  chunk_overlap: 100
  max_files: 1000

filters:
  include_extensions: [".go", ".py", ".js", ".ts", ".md"]
  exclude_dirs: ["node_modules", ".git", "dist"]
  exclude_files: ["*.min.js", "*.pyc"]

output:
  format: "markdown"
  directory: "./docs"
  language: "en"

logging:
  level: "info"
  format: "text"
  output: "stderr"
```

### Provider Options

#### LLM Providers

**OpenAI:**

```yaml
providers:
  llm:
    provider: "openai"
    api_key: "${OPENAI_API_KEY}"
    model: "gpt-4o" # or "gpt-4o-mini", "gpt-3.5-turbo"
    base_url: "https://api.openai.com/v1"
```

**Anthropic:**

```yaml
providers:
  llm:
    provider: "anthropic"
    api_key: "${ANTHROPIC_API_KEY}"
    model: "claude-3-5-sonnet-20241022" # or "claude-3-5-haiku-20241022"
    base_url: "https://api.anthropic.com/v1"
```

#### Embedding Providers

**OpenAI Embeddings:**

```yaml
providers:
  embedding:
    provider: "openai"
    api_key: "${OPENAI_API_KEY}"
    model: "text-embedding-3-small" # or "text-embedding-3-large"
    dimensions: 1536
```

**Voyage AI Embeddings:**

```yaml
providers:
  embedding:
    provider: "voyage"
    api_key: "${VOYAGE_API_KEY}"
    model: "voyage-3-large" # or "voyage-3.5-lite", "voyage-code-3"
    dimensions: 1024
```

**Ollama Local Embeddings:**

```yaml
providers:
  embedding:
    provider: "ollama"
    model: "nomic-embed-text" # or "mxbai-embed-large", "all-minilm"
    base_url: "http://localhost:11434"
    dimensions: 768
```

### CLI Options

```bash
deepwiki-cli generate [flags]

Flags:
  -o, --output-dir string      Output directory (default "./docs")
  -f, --format string         Output format: markdown, json (default "markdown")
  -l, --language string       Language: en, ja, zh, es, kr, vi (default "en")
  -m, --model string          OpenAI model (default "gpt-4o")
      --openai-key string     OpenAI API key
      --exclude-dirs string   Directories to exclude (comma-separated)
      --exclude-files string  File patterns to exclude (comma-separated)
      --config string         Configuration file path
  -v, --verbose               Verbose output
      --dry-run              Show what would be done
```

## Current Capabilities

### 🔍 Phase 1: File Scanner & Analysis

- **Language Detection**: Supports 40+ programming languages with automatic detection
- **Smart Filtering**: Intelligent exclusion of build artifacts, dependencies, and binary files
- **Content Analysis**: Binary detection, line counting, test file identification, and importance scoring
- **Performance**: Concurrent processing with configurable worker pools and memory limits

### 📝 Phase 2: Text Processing & Chunking

- **Advanced Chunking**: Semantic code boundary detection for optimal text splitting
- **Language-Aware**: Language-specific chunking strategies for better context preservation
- **Token Management**: Accurate token counting and chunk size optimization for AI processing
- **Preprocessing**: Content normalization, comment removal, and whitespace handling

### 🧠 Phase 3: Vector Embeddings Generation

- **OpenAI Integration**: Uses text-embedding-3-small/large models for high-quality embeddings
- **Batch Processing**: Efficient batch embedding generation with retry logic and rate limiting
- **Multiple Models**: Support for ada-002, text-embedding-3-small, and text-embedding-3-large
- **Error Handling**: Robust error recovery with exponential backoff and request optimization

### 🔍 Phase 4: RAG System & Document Indexing

- **Vector Database**: BoltDB-based persistent vector storage with similarity search
- **Document Retrieval**: Advanced RAG (Retrieval-Augmented Generation) system for context-aware documentation
- **Similarity Search**: Cosine similarity, Euclidean distance, and dot product search methods
- **Metadata Management**: Rich metadata storage for enhanced search and filtering

### 🏗️ Phase 5: Wiki Structure Generation

- **AI-Powered Structure**: GPT-4 generates intelligent wiki hierarchies based on codebase analysis
- **Contextual Pages**: Automatic page generation with relevant code snippets and explanations
- **Progress Tracking**: Real-time progress monitoring with ETA calculations and task completion
- **Multiple Languages**: Support for documentation generation in English, Japanese, Chinese, Spanish, Korean, and Vietnamese

### 📄 Phase 6: Content Generation & Output

- **Multiple Formats**: Markdown and JSON output with structured organization
- **File Management**: Automated directory structure creation and file organization
- **Rich Content**: Generated pages include code examples, file references, and cross-links
- **Statistics**: Comprehensive generation statistics including word counts, processing time, and error reporting

### ⚙️ Configuration System

- **Multiple Sources**: YAML files, environment variables, CLI flags with priority handling
- **Validation**: Comprehensive config validation with helpful error messages and defaults
- **Templates**: Auto-generated configuration templates with best practices
- **Flexibility**: Override any setting at runtime with full configuration inheritance

### 📊 Logging & Progress System

- **Structured Logging**: JSON and text formats with slog integration
- **Component-Based**: Separate loggers for scanner, processor, embeddings, generator, and output
- **Progress Tracking**: Visual progress bars with phase-based tracking and ETA calculations
- **Error Management**: Detailed error reporting with context and recovery suggestions

### 🔧 CLI Management

- **Interactive Progress**: Real-time CLI progress display with phase completion status
- **Statistics Tracking**: Comprehensive operation statistics including API usage and token consumption
- **Dry Run Mode**: Preview operations without actual file generation or API calls
- **Verbose Logging**: Detailed operation logging for debugging and monitoring

## Testing

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./pkg/scanner -v
go test ./internal/config -v
go test ./internal/logging -v

# Test scanner with real project
OPENAI_API_KEY=dummy ./deepwiki-cli generate --verbose --dry-run
```

## Architecture

```
deepwiki/
├── cmd/                    # CLI commands (Cobra)
├── internal/
│   ├── config/            # Configuration management
│   ├── logging/           # Structured logging (slog)
│   └── prompts/           # AI prompt templates
├── pkg/
│   ├── llm/               # LLM provider interfaces
│   │   ├── openai/        # OpenAI LLM implementation
│   │   └── anthropic/     # Anthropic LLM implementation
│   ├── embedding/         # Embedding provider interfaces
│   │   ├── openai/        # OpenAI embeddings
│   │   ├── voyage/        # Voyage AI embeddings
│   │   └── ollama/        # Local Ollama embeddings
│   ├── scanner/           # File scanning and analysis
│   ├── embeddings/        # Embedding generation logic
│   ├── generator/         # Documentation generation
│   ├── rag/              # RAG system implementation
│   └── output/            # Output formatting
├── examples/              # Example configurations
└── docs/                  # Generated documentation
```

### Provider Architecture

The new separated provider architecture allows for:

- **Independent LLM and Embedding providers**: Mix and match based on your needs
- **Local embedding support**: Use Ollama for cost-free local embeddings
- **Easy extensibility**: Add new providers without changing core logic
- **Cost optimization**: Choose expensive LLMs for generation, cheap/local for embeddings

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass: `go test ./...`
5. Submit a pull request

---

**Built with Go 1.24+ • Coded with Claude code • Inspired by deepwiki**
