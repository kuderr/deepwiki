# Configuration Guide

DeepWiki CLI offers flexible configuration through multiple sources with a clear precedence order.

## Configuration Sources (Precedence Order)

1. **Command Line Flags** (highest priority)
2. **Environment Variables**
3. **Configuration Files**
4. **Default Values** (lowest priority)

## Configuration File

### File Locations

DeepWiki CLI automatically searches for configuration files in these locations:

1. `--config` flag specified file
2. `deepwiki.yaml` in current directory
3. `deepwiki.yml` in current directory
4. `.deepwiki.yaml` in current directory
5. `.deepwiki.yml` in current directory
6. `~/.deepwiki.yaml` in home directory
7. `~/.deepwiki.yml` in home directory

### Complete Configuration Reference

```yaml
# DeepWiki CLI Configuration File
# All settings are optional - defaults will be used for missing values

# OpenAI API Configuration
openai:
  # Required: Your OpenAI API key
  api_key: "${OPENAI_API_KEY}"

  # The main model for content generation
  # Options: gpt-4o, gpt-4o-mini, gpt-4-turbo, gpt-4, gpt-3.5-turbo
  model: "gpt-4o"

  # Model for embedding generation
  # Options: text-embedding-3-large, text-embedding-3-small, text-embedding-ada-002
  embedding_model: "text-embedding-3-small"

  # Maximum tokens per API request
  # Range: 1-32768 (depending on model)
  max_tokens: 4000

  # Temperature for generation (creativity vs consistency)
  # Range: 0.0-2.0 (0.0 = deterministic, 2.0 = very creative)
  temperature: 0.1

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

# Rate Limiting Configuration (Advanced)
rate_limiting:
  # Requests per second for OpenAI API
  requests_per_second: 2

  # Maximum burst size
  burst_size: 5

  # Retry configuration
  max_retries: 3
  retry_delay: "1s"
  max_retry_delay: "30s"

# Caching Configuration (Advanced)
cache:
  # Enable caching
  enabled: true

  # Cache directory
  directory: "./.deepwiki-cache"

  # Maximum cache age
  max_age: "24h"

  # Maximum cache size
  max_size: "100MB"

# Advanced Processing Configuration
advanced:
  # Maximum concurrent workers
  max_workers: 4

  # Memory limit for processing (in MB)
  memory_limit: 512

  # Enable garbage collection optimization
  gc_optimization: true

  # Batch size for processing
  batch_size: 10
```

## Environment Variables

All configuration options can be overridden using environment variables with the `DEEPWIKI_` prefix:

### OpenAI Configuration

```bash
export OPENAI_API_KEY="sk-your-api-key"
export DEEPWIKI_MODEL="gpt-4o"
export DEEPWIKI_EMBEDDING_MODEL="text-embedding-3-small"
export DEEPWIKI_MAX_TOKENS="4000"
export DEEPWIKI_TEMPERATURE="0.1"
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
openai:
  api_key: "${OPENAI_API_KEY}"
```

### Development Configuration

```yaml
openai:
  api_key: "${OPENAI_API_KEY}"
  model: "gpt-4o-mini" # Faster, cheaper for development

output:
  directory: "./dev-docs"

logging:
  level: "debug" # Detailed logging
  format: "text"

cache:
  enabled: true # Speed up repeated runs
```

### Production Configuration

```yaml
openai:
  api_key: "${OPENAI_API_KEY}"
  model: "gpt-4o" # High quality
  temperature: 0.1 # Consistent output

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

cache:
  enabled: true
  max_age: "1h" # Fresh cache
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

advanced:
  max_workers: 8 # More concurrent processing
  memory_limit: 1024 # Higher memory limit
  batch_size: 20 # Larger batches

cache:
  enabled: true
  max_age: "6h" # Longer cache

rate_limiting:
  requests_per_second: 5 # Faster API calls (if your plan allows)
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
deepwiki-cli config validate deepwiki.yaml

# Validate current configuration
deepwiki-cli config validate

# Validate and show resolved configuration
deepwiki-cli config validate --show-resolved
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

This comprehensive configuration guide should help you customize DeepWiki CLI for your specific needs and environments.
