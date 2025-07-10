# Configuration Guide

DeepWiki offers flexible configuration through multiple sources with a clear precedence order.

## Configuration Sources (Precedence Order)

1. **Command Line Flags** (highest priority)
2. **Environment Variables**
3. **Configuration Files**
4. **Default Values** (lowest priority)

## Configuration File

### File Locations

DeepWiki automatically searches for configuration files in these locations:

1. `--config` flag specified file
2. `deepwiki.yaml` in current directory
3. `deepwiki.yml` in current directory
4. `.deepwiki.yaml` in current directory
5. `.deepwiki.yml` in current directory
6. `~/.deepwiki.yaml` in home directory
7. `~/.deepwiki.yml` in home directory

### Complete Configuration Reference

```yaml
# DeepWiki Configuration File
# All settings are optional - defaults will be used for missing values

# Provider Configuration
providers:
  # LLM Provider Configuration
  llm:
    # Provider type: "openai", "anthropic", or "ollama"
    provider: "openai"

    # API key (required for OpenAI and Anthropic)
    api_key: "${OPENAI_API_KEY}"

    # Model name
    # OpenAI: gpt-4o, gpt-4o-mini, gpt-4-turbo, gpt-4, gpt-3.5-turbo
    # Anthropic: claude-3-5-sonnet-20241022, claude-3-haiku-20240307
    # Ollama: llama3.1, llama3.2, codellama, mistral, etc.
    model: "gpt-4o"

    # Maximum tokens per API request
    # Range: 1-32768 (depending on model)
    max_tokens: 4000

    # Temperature for generation (creativity vs consistency)
    # Range: 0.0-2.0 (0.0 = deterministic, 2.0 = very creative)
    temperature: 0.1

    # Request timeout (duration string like "3m")
    request_timeout: "3m"

    # Maximum retry attempts
    max_retries: 3

    # Retry delay (duration string like "1s")
    retry_delay: "1s"

    # Rate limiting (requests per second)
    rate_limit_rps: 2.0

    # Custom base URL (for Ollama or custom OpenAI-compatible endpoints)
    # OpenAI: https://api.openai.com/v1 (default)
    # Anthropic: https://api.anthropic.com/v1 (default)
    # Ollama: http://localhost:11434 (default)
    base_url: ""

  # Embedding Provider Configuration
  embedding:
    # Provider type: "openai", "voyage", or "ollama"
    provider: "openai"

    # API key (required for OpenAI and Voyage)
    api_key: "${OPENAI_API_KEY}"

    # Model name
    # OpenAI: text-embedding-3-large, text-embedding-3-small, text-embedding-ada-002
    # Voyage: voyage-3-large, voyage-3-small
    # Ollama: nomic-embed-text, all-minilm, etc.
    model: "text-embedding-3-small"

    # Request timeout (duration string like "30s")
    request_timeout: "30s"

    # Maximum retry attempts
    max_retries: 3

    # Retry delay (duration string like "1s")
    retry_delay: "1s"

    # Rate limiting (requests per second)
    rate_limit_rps: 10.0

    # Custom base URL (for Ollama)
    # OpenAI: https://api.openai.com/v1 (default)
    # Voyage: https://api.voyageai.com/v1 (default)
    # Ollama: http://localhost:11434 (default)
    base_url: ""

    # Embedding dimensions (auto-detected if not specified)
    dimensions: 0

# Text Processing Configuration
processing:
  # Size of text chunks for embedding
  # Range: 100-1000 words
  chunk_size: 350

  # Overlap between consecutive chunks
  # Range: 0-chunk_size/2
  chunk_overlap: 100

  # Maximum number of files to process
  # Set to 0 for unlimited
  max_files: 1000

# File Filtering Configuration
filters:
  # File extensions to include (case-insensitive)
  include_extensions:
    # Programming Languages
    - ".go" # Go
    - ".py" # Python
    - ".js" # JavaScript
    - ".ts" # TypeScript
    - ".jsx" # React JSX
    - ".tsx" # React TSX
    - ".java" # Java
    - ".cpp" # C++
    - ".c" # C
    - ".h" # C/C++ Headers
    - ".hpp" # C++ Headers
    - ".cs" # C#
    - ".php" # PHP
    - ".rb" # Ruby
    - ".rs" # Rust
    - ".swift" # Swift
    - ".kt" # Kotlin
    - ".scala" # Scala
    - ".clj" # Clojure
    - ".hs" # Haskell
    - ".ml" # OCaml
    - ".fs" # F#
    - ".dart" # Dart
    - ".lua" # Lua
    - ".r" # R
    - ".R" # R
    - ".m" # Objective-C/MATLAB
    - ".mm" # Objective-C++
    - ".pl" # Perl
    - ".sh" # Shell scripts
    - ".bash" # Bash scripts
    - ".zsh" # Zsh scripts
    - ".fish" # Fish scripts
    - ".ps1" # PowerShell
    - ".bat" # Windows Batch
    - ".cmd" # Windows Command

    # Web Technologies
    - ".html" # HTML
    - ".htm" # HTML
    - ".css" # CSS
    - ".scss" # Sass
    - ".sass" # Sass
    - ".less" # Less
    - ".vue" # Vue.js
    - ".svelte" # Svelte

    # Configuration & Data
    - ".yaml" # YAML
    - ".yml" # YAML
    - ".json" # JSON
    - ".toml" # TOML
    - ".ini" # INI
    - ".cfg" # Config
    - ".conf" # Config
    - ".xml" # XML
    - ".proto" # Protocol Buffers
    - ".graphql" # GraphQL
    - ".gql" # GraphQL

    # Documentation
    - ".md" # Markdown
    - ".mdx" # MDX
    - ".txt" # Plain text
    - ".rst" # reStructuredText
    - ".org" # Org mode
    - ".tex" # LaTeX
    - ".adoc" # AsciiDoc

    # Database
    - ".sql" # SQL
    - ".psql" # PostgreSQL
    - ".mysql" # MySQL

    # Build & CI/CD
    - ".dockerfile" # Dockerfile
    - ".makefile" # Makefile
    - ".mk" # Makefile
    - ".gradle" # Gradle
    - ".maven" # Maven
    - ".ant" # Apache Ant

  # Directories to exclude (relative to project root)
  exclude_dirs:
    # Dependencies
    - "node_modules" # npm/yarn packages
    - "vendor" # Go/PHP vendor directory
    - ".venv" # Python virtual environment
    - "venv" # Python virtual environment
    - "env" # Python virtual environment
    - ".env" # Environment directory
    - "virtualenv" # Python virtual environment
    - "__pycache__" # Python cache
    - ".tox" # Python tox
    - "site-packages" # Python packages
    - ".bundle" # Ruby bundle
    - "gems" # Ruby gems
    - ".cargo" # Rust cargo
    - "target" # Rust/Maven target
    - ".gradle" # Gradle cache
    - ".mvn" # Maven wrapper

    # Build Outputs
    - "dist" # Distribution files
    - "build" # Build output
    - "out" # Output directory
    - "bin" # Binary files
    - "obj" # Object files
    - "lib" # Library files (sometimes)
    - ".build" # Build directory
    - "cmake-build-debug" # CMake debug build
    - "cmake-build-release" # CMake release build

    # Development Tools
    - ".git" # Git repository
    - ".svn" # Subversion
    - ".hg" # Mercurial
    - ".bzr" # Bazaar
    - ".idea" # JetBrains IDEs
    - ".vscode" # Visual Studio Code
    - ".vs" # Visual Studio
    - ".eclipse" # Eclipse IDE
    - ".settings" # IDE settings
    - ".project" # Project files
    - ".classpath" # Java classpath
    - ".factorypath" # Eclipse factory path

    # Temporary & Cache
    - "tmp" # Temporary files
    - "temp" # Temporary files
    - ".tmp" # Temporary files
    - ".temp" # Temporary files
    - "cache" # Cache directory
    - ".cache" # Cache directory
    - ".next" # Next.js cache
    - ".nuxt" # Nuxt.js cache
    - ".angular" # Angular cache
    - ".turbo" # Turborepo cache

    # Logs & Data
    - "logs" # Log files
    - "log" # Log files
    - ".logs" # Log files
    - "data" # Data directory (sometimes)
    - ".data" # Data directory
    - "backup" # Backup files
    - "backups" # Backup files

    # Testing (optional - remove if you want test documentation)
    - "test" # Test directory
    - "tests" # Test directory
    - "__tests__" # Jest tests
    - "spec" # Spec directory
    - ".pytest_cache" # Pytest cache
    - "coverage" # Coverage reports
    - ".nyc_output" # NYC coverage
    - "htmlcov" # HTML coverage

    # Documentation (optional - remove if you want to include existing docs)
    - "docs" # Documentation (to avoid conflicts)
    - "doc" # Documentation
    - ".docs" # Documentation
    - "documentation" # Documentation
    - "wiki" # Wiki files

    # Platform Specific
    - ".DS_Store" # macOS
    - "Thumbs.db" # Windows
    - "Desktop.ini" # Windows

  # File patterns to exclude (glob patterns)
  exclude_files:
    # Compiled & Binary Files
    - "*.min.js" # Minified JavaScript
    - "*.min.css" # Minified CSS
    - "*.bundle.js" # JavaScript bundles
    - "*.chunk.js" # JavaScript chunks
    - "*.pyc" # Python compiled
    - "*.pyo" # Python optimized
    - "*.class" # Java compiled
    - "*.jar" # Java archives
    - "*.war" # Java web archives
    - "*.ear" # Java enterprise archives
    - "*.exe" # Windows executables
    - "*.dll" # Windows libraries
    - "*.so" # Linux libraries
    - "*.dylib" # macOS libraries
    - "*.a" # Static libraries
    - "*.o" # Object files
    - "*.obj" # Object files
    - "*.lib" # Library files
    - "*.exp" # Export files
    - "*.pdb" # Debug files

    # Lock Files
    - "package-lock.json" # npm lock file
    - "yarn.lock" # Yarn lock file
    - "pnpm-lock.yaml" # pnpm lock file
    - "composer.lock" # Composer lock file
    - "Gemfile.lock" # Ruby lock file
    - "Pipfile.lock" # Python lock file
    - "poetry.lock" # Poetry lock file
    - "cargo.lock" # Rust lock file (optional)
    - "go.sum" # Go sum file (optional)

    # Generated Files
    - "*.generated.*" # Generated files
    - "*.gen.*" # Generated files
    - "*_generated.go" # Go generated files
    - "*_gen.go" # Go generated files
    - "*.pb.go" # Protocol buffer generated
    - "*.pb.cc" # Protocol buffer generated
    - "*.pb.h" # Protocol buffer generated
    - "*_pb2.py" # Python protobuf
    - "*_pb2_grpc.py" # Python gRPC

    # IDE & Editor Files
    - "*.swp" # Vim swap files
    - "*.swo" # Vim swap files
    - "*~" # Backup files
    - ".#*" # Emacs lock files
    - "#*#" # Emacs backup files
    - ".*.rej" # Rejected patches
    - ".*.orig" # Original files

    # System Files
    - ".DS_Store" # macOS
    - "Thumbs.db" # Windows
    - "desktop.ini" # Windows
    - "*.lnk" # Windows shortcuts

    # Logs & Temporary
    - "*.log" # Log files
    - "*.tmp" # Temporary files
    - "*.temp" # Temporary files
    - "*.bak" # Backup files
    - "*.backup" # Backup files
    - "core" # Core dumps
    - "*.dump" # Dump files

# Output Configuration
output:
  # Output format: "markdown" or "json"
  format: "markdown"

  # Output directory (relative to current directory or absolute path)
  directory: "./docs"

  # Output language
  # Supported: "en", "ja", "zh", "es", "kr", "vi"
  language: "en"

# Embeddings Configuration
embeddings:
  # Enable embedding generation and vector search
  enabled: true

  # Embedding dimensions (depends on model)
  # text-embedding-3-small: 1536, text-embedding-3-large: 3072
  dimensions: 1536

  # Number of top relevant chunks to retrieve
  top_k: 20

# Logging Configuration
logging:
  # Log level: "debug", "info", "warn", "error"
  level: "info"

  # Log format: "text" or "json"
  format: "text"

  # Log output: "stdout", "stderr", or file path
  output: "stderr"

  # Add source code location to logs
  add_source: false

  # Time format for logs (Go time format)
  time_format: "2006-01-02 15:04:05"
```

## Environment Variables

All configuration options can be overridden using environment variables with the `DEEPWIKI_` prefix:

### Provider Configuration

```bash
# LLM Provider Configuration
export DEEPWIKI_LLM_PROVIDER="openai"        # or "anthropic" or "ollama"
export OPENAI_API_KEY="sk-your-api-key"      # for OpenAI
export ANTHROPIC_API_KEY="sk-ant-api-key"    # for Anthropic
export DEEPWIKI_LLM_BASE_URL="http://localhost:11434"
export DEEPWIKI_LLM_MODEL="gpt-4o"           # model name

# Embedding Provider Configuration
export DEEPWIKI_EMBEDDING_PROVIDER="openai"   # or "voyage" or "ollama"
export VOYAGE_API_KEY="pa-your-voyage-key"    # for Voyage AI
export DEEPWIKI_EMBEDDING_MODEL="text-embedding-3-small"  # model name
export DEEPWIKI_EMBEDDING_BASE_URL="http://localhost:11434"
```

### Output Configuration

```bash
export DEEPWIKI_OUTPUT_DIR="./documentation"
export DEEPWIKI_FORMAT="markdown"
export DEEPWIKI_LANGUAGE="en"
```

### Processing Configuration

```bash
export DEEPWIKI_CHUNK_SIZE="350"
export DEEPWIKI_CHUNK_OVERLAP="100"
export DEEPWIKI_MAX_FILES="1000"
```

### Filtering Configuration

```bash
export DEEPWIKI_EXCLUDE_DIRS="node_modules,vendor,.git"
export DEEPWIKI_EXCLUDE_FILES="*.min.js,*.pyc"
```

### Logging Configuration

```bash
export DEEPWIKI_LOG_LEVEL="info"
export DEEPWIKI_LOG_FORMAT="text"
export DEEPWIKI_LOG_OUTPUT="stderr"
```

## Command Line Flags

All configuration can be overridden with command line flags:

### Basic Flags

```bash
--config string           # Configuration file path
--output-dir string       # Output directory
--format string          # Output format (markdown|json)
--language string        # Output language
--verbose                # Verbose output
--dry-run               # Preview without generating
```

### OpenAI Flags

```bash
--openai-key string      # OpenAI API key
--model string          # OpenAI model name
--chunk-size int        # Text chunk size
```

### Filtering Flags

```bash
--exclude-dirs string    # Comma-separated directories to exclude
--exclude-files string   # Comma-separated file patterns to exclude
```

## Configuration Examples

### Minimal Configuration

```yaml
# OpenAI Configuration
providers:
  llm:
    provider: "openai"
    api_key: "${OPENAI_API_KEY}"
  embedding:
    provider: "openai"
    api_key: "${OPENAI_API_KEY}"
```

### Ollama Configuration (Local AI)

```yaml
# Use Ollama for both LLM and embeddings (completely offline)
providers:
  llm:
    provider: "ollama"
    model: "llama3.1"
    base_url: "http://localhost:11434"
  embedding:
    provider: "ollama"
    model: "nomic-embed-text"
    base_url: "http://localhost:11434"
```

### Hybrid Configuration

```yaml
# Use Claude for LLM and Ollama for embeddings (cost-effective)
providers:
  llm:
    provider: "anthropic"
    api_key: "${ANTHROPIC_API_KEY}"
    model: "claude-3-5-sonnet-20241022"
  embedding:
    provider: "ollama"
    model: "nomic-embed-text"
    base_url: "http://localhost:11434"
```

### Development Configuration

```yaml
providers:
  llm:
    provider: "openai"
    api_key: "${OPENAI_API_KEY}"
    model: "gpt-4o-mini" # Faster, cheaper for development
  embedding:
    provider: "ollama" # Free local embeddings
    model: "nomic-embed-text"

output:
  directory: "./dev-docs"

logging:
  level: "debug" # Detailed logging
  format: "text"
```

### Production Configuration

```yaml
providers:
  llm:
    provider: "openai"
    api_key: "${OPENAI_API_KEY}"
    model: "gpt-4o" # High quality
    temperature: 0.1 # Consistent output
  embedding:
    provider: "openai"
    api_key: "${OPENAI_API_KEY}"
    model: "text-embedding-3-small"

output:
  directory: "./docs"
  format: "markdown"

filters:
  exclude_dirs:
    - "node_modules"
    - "vendor"
    - "tests"
    - ".git"

processing:
  max_files: 500 # Limit for large projects
  chunk_size: 350

logging:
  level: "info"
  format: "json" # Structured logs for production
```

### Multi-Language Project Configuration

```yaml
openai:
  api_key: "${OPENAI_API_KEY}"
  model: "gpt-4o"

filters:
  include_extensions:
    - ".go"
    - ".py"
    - ".js"
    - ".ts"
    - ".java"
    - ".cpp"
    - ".rs"
    - ".php"
    - ".rb"
    - ".cs"
    - ".md"
    - ".yaml"
    - ".json"
  exclude_dirs:
    - "node_modules"
    - "vendor"
    - "__pycache__"
    - "target"
    - "build"
    - "dist"

processing:
  chunk_size: 400 # Larger chunks for complex projects
  max_files: 1500

output:
  format: "markdown"
```

### Performance-Optimized Configuration

```yaml
openai:
  api_key: "${OPENAI_API_KEY}"
  model: "gpt-4o-mini" # Faster model
  max_tokens: 2000 # Smaller responses

processing:
  chunk_size: 300 # Smaller chunks
  max_files: 800 # Limit file count
```

## Ollama Setup and Configuration

### Installing Ollama

```bash
# macOS/Linux
curl -fsSL https://ollama.ai/install.sh | sh

# Windows
# Download and install from https://ollama.ai

# Verify installation
ollama --version
```

### Setting Up Models

```bash
# Pull LLM models for text generation
ollama pull llama3.1        # Meta's Llama 3.1 (recommended)
ollama pull llama3.2        # Latest Llama 3.2
ollama pull codellama       # Code-focused model
ollama pull mistral         # Mistral 7B model

# Pull embedding models
ollama pull nomic-embed-text    # Recommended for embeddings
ollama pull all-minilm          # Alternative embedding model

# List available models
ollama list

# Check model info
ollama show llama3.1
```

### Ollama Configuration Options

```yaml
providers:
  llm:
    provider: "ollama"
    model: "llama3.1"
    base_url: "http://localhost:11434" # Default Ollama URL
    max_tokens: 4000
    temperature: 0.1
    request_timeout: "5m" # Longer timeout for local processing
    max_retries: 2
    rate_limit_rps: 1.0 # Lower rate for local processing

  embedding:
    provider: "ollama"
    model: "nomic-embed-text"
    base_url: "http://localhost:11434"
    request_timeout: "2m"
    max_retries: 2
    rate_limit_rps: 5.0
```

### Remote Ollama Configuration

```yaml
# If running Ollama on a different machine
providers:
  llm:
    provider: "ollama"
    model: "llama3.1"
    base_url: "http://192.168.1.100:11434" # Remote Ollama server

  embedding:
    provider: "ollama"
    model: "nomic-embed-text"
    base_url: "http://192.168.1.100:11434"
```

### Model Recommendations

| Use Case             | LLM Model      | Embedding Model    | Memory Required |
| -------------------- | -------------- | ------------------ | --------------- |
| General purpose      | `llama3.1`     | `nomic-embed-text` | 8GB+            |
| Code documentation   | `codellama`    | `nomic-embed-text` | 8GB+            |
| Resource constrained | `llama3.2`     | `all-minilm`       | 4GB+            |
| High quality         | `llama3.1:70b` | `nomic-embed-text` | 40GB+           |

### Performance Tuning

```yaml
providers:
  llm:
    provider: "ollama"
    model: "llama3.1"
    request_timeout: "10m" # Increase for complex generations
    max_retries: 1 # Reduce retries for faster failure
    rate_limit_rps: 0.5 # Lower rate for stability

processing:
  chunk_size: 250 # Smaller chunks for faster processing
  max_files: 500 # Limit files for memory usage
```

### Troubleshooting Ollama

#### Common Issues

```bash
# Check Ollama service status
ollama ps

# Restart Ollama service
sudo systemctl restart ollama  # Linux
brew services restart ollama   # macOS

# Check available models
ollama list

# Test model directly
ollama run llama3.1 "Hello, how are you?"

# Check Ollama logs
journalctl -u ollama          # Linux
brew services info ollama     # macOS
```

#### Configuration Issues

```yaml
# Issue: Connection refused
# Solution: Ensure Ollama is running and base_url is correct
providers:
  llm:
    provider: "ollama"
    base_url: "http://localhost:11434"  # Check this URL

# Issue: Model not found
# Solution: Pull the model first
# ollama pull llama3.1

# Issue: Slow performance
# Solution: Adjust timeouts and reduce concurrency
providers:
  llm:
    provider: "ollama"
    request_timeout: "10m"
    rate_limit_rps: 0.5

```

### CI/CD Configuration

```yaml
openai:
  api_key: "${OPENAI_API_KEY}"
  model: "gpt-4o-mini" # Cost-effective for CI

output:
  directory: "./generated-docs"
  format: "markdown"

filters:
  exclude_dirs:
    - "node_modules"
    - "vendor"
    - "tests"
    - "examples"
    - ".git"

processing:
  max_files: 300 # Limit for CI time constraints

logging:
  level: "info"
  format: "json" # Structured logs for CI

cache:
  enabled: false # Don't cache in CI (clean builds)
```

## Configuration Validation

### Validate Configuration File

```bash
# Validate specific file
deepwiki config validate deepwiki.yaml

# Validate current configuration
deepwiki config validate

# Validate and show resolved configuration
deepwiki config validate --show-resolved
```

### Common Validation Errors

```bash
# Invalid model name
Error: invalid model "gpt-5" - supported models: gpt-4o, gpt-4o-mini, gpt-4-turbo

# Invalid language
Error: invalid language "fr" - supported languages: en, ja, zh, es, kr, vi

# Invalid chunk size
Error: chunk_size must be between 100 and 1000, got 50

# Invalid temperature
Error: temperature must be between 0.0 and 2.0, got 3.0

# Missing API key
Error: OpenAI API key is required - set OPENAI_API_KEY or openai.api_key
```

## Best Practices

### 1. Configuration Management

- Store sensitive data (API keys) in environment variables
- Use different configurations for development and production
- Version control your configuration files (excluding secrets)
- Validate configuration before deployment

### 2. Security

- Never commit API keys to version control
- Use environment variables or secure secret management
- Set appropriate file permissions on config files
- Rotate API keys regularly

### 3. Performance

- Use appropriate models for your use case
- Configure filtering to reduce processing time
- Enable caching for development workflows
- Monitor token usage and costs

### 4. Maintenance

- Keep configurations simple and well-documented
- Use comments to explain non-obvious settings
- Regular review and update of exclusion patterns
- Test configuration changes in development first

This comprehensive configuration guide should help you customize DeepWiki for your specific needs and environments.
