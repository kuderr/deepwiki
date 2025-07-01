# Go CLI Tool Implementation Plan for DeepWiki

## Project Overview

**Goal**: Create a Go CLI tool that generates documentation for local directories using the same AI-powered logic as deepwiki, but simplified for local-only processing with OpenAI as the sole LLM provider.

## Core Architecture

### 1. CLI Interface

```bash
deepwiki-cli [flags] [directory]

Flags:
  --output-dir, -o     Output directory for generated docs (default: ./docs)
  --format, -f         Output format: markdown|json (default: markdown)
  --language, -l       Language for generation: en|ja|zh|es|kr|vi (default: en)
  --openai-key         OpenAI API key (or use OPENAI_API_KEY env var)
  --model, -m          OpenAI model (default: gpt-4o)
  --exclude-dirs       Comma-separated list of directories to exclude
  --exclude-files      Comma-separated patterns for files to exclude
  --chunk-size         Text chunk size for embeddings (default: 350)
  --help, -h           Show help
  --version, -v        Show version
```

### 2. Core Components

#### A. File Scanner (`pkg/scanner/`)

- **Purpose**: Recursively scan directory and filter files
- **Key Functions**:
  - `ScanDirectory(path string, filters FileFilters) ([]FileInfo, error)`
  - `ApplyFilters(files []FileInfo, filters FileFilters) []FileInfo`
  - Default exclusions: `.git/`, `node_modules/`, `dist/`, etc.
  - Support for code files: `.go`, `.py`, `.js`, `.ts`, `.java`, etc.
  - Support for docs: `.md`, `.txt`, `.json`, `.yaml`

#### B. Text Processing (`pkg/processor/`)

- **Purpose**: Chunk files and prepare for embedding
- **Key Functions**:
  - `ChunkText(content string, chunkSize int, overlap int) []TextChunk`
  - `ProcessFiles(files []FileInfo) ([]Document, error)`
  - Token counting for OpenAI models using tiktoken-go

#### C. Embedding Service (`pkg/embeddings/`)

- **Purpose**: Generate embeddings using OpenAI
- **Key Functions**:
  - `GenerateEmbeddings(texts []string) ([][]float32, error)`
  - `CreateVectorDB(documents []Document) (*VectorDB, error)`
  - Use OpenAI `text-embedding-3-small` (256 dimensions)
  - Local vector storage using bbolt or similar

#### D. Wiki Generator (`pkg/generator/`)

- **Purpose**: Generate wiki structure and content using OpenAI
- **Key Functions**:
  - `GenerateWikiStructure(fileTree string, readme string) (*WikiStructure, error)`
  - `GeneratePageContent(page WikiPage, context []Document) (string, error)`
  - Implement the same prompts as deepwiki

#### E. RAG System (`pkg/rag/`)

- **Purpose**: Retrieve relevant context for each wiki page
- **Key Functions**:
  - `RetrieveRelevantDocs(query string, topK int) ([]Document, error)`
  - `SearchSimilar(embedding []float32, topK int) ([]Document, error)`

#### F. Output Manager (`pkg/output/`)

- **Purpose**: Generate final documentation files
- **Key Functions**:
  - `GenerateMarkdown(wiki WikiStructure, pages map[string]WikiPage) error`
  - `GenerateJSON(wiki WikiStructure, pages map[string]WikiPage) error`
  - `CreateIndex(wiki WikiStructure) error`

### 3. Data Structures

```go
type WikiStructure struct {
    ID           string       `json:"id"`
    Title        string       `json:"title"`
    Description  string       `json:"description"`
    Pages        []WikiPage   `json:"pages"`
    Sections     []WikiSection `json:"sections,omitempty"`
    RootSections []string     `json:"rootSections,omitempty"`
}

type WikiPage struct {
    ID           string   `json:"id"`
    Title        string   `json:"title"`
    Content      string   `json:"content"`
    FilePaths    []string `json:"filePaths"`
    Importance   string   `json:"importance"` // high, medium, low
    RelatedPages []string `json:"relatedPages"`
    ParentID     string   `json:"parentId,omitempty"`
}

type Document struct {
    ID       string            `json:"id"`
    Text     string            `json:"text"`
    FilePath string            `json:"filePath"`
    Metadata map[string]string `json:"metadata"`
    Embedding []float32        `json:"embedding,omitempty"`
}

type FileFilters struct {
    ExcludedDirs  []string `json:"excludedDirs"`
    ExcludedFiles []string `json:"excludedFiles"`
    IncludeExts   []string `json:"includeExts"`
}
```

### 4. Implementation Phases

#### Phase 1: Core Infrastructure (Week 1)

- CLI framework setup using cobra
- File scanner implementation
- Configuration management
- Basic project structure

#### Phase 2: OpenAI Integration (Week 2)

- OpenAI client setup
- Embedding generation
- Text completion for wiki generation
- Rate limiting and error handling

#### Phase 3: Text Processing & RAG (Week 3)

- Text chunking implementation
- Vector database (local storage)
- Similarity search
- Document retrieval system

#### Phase 4: Wiki Generation (Week 4)

- Wiki structure generation (using same prompts as original)
- Page content generation
- Progress tracking and concurrent processing

#### Phase 5: Output & Polish (Week 5)

- Markdown/JSON output generation
- Error handling improvements
- Documentation and testing
- Performance optimization

#### Phase 6: Testing & Documentation (Week 6)

- Comprehensive testing
- Example projects
- User documentation
- Performance benchmarks

### 5. Key Implementation Details

#### OpenAI Configuration

```go
type OpenAIConfig struct {
    APIKey      string `yaml:"api_key"`
    Model       string `yaml:"model"`        // Default: "gpt-4o"
    EmbedModel  string `yaml:"embed_model"`  // Default: "text-embedding-3-small"
    MaxTokens   int    `yaml:"max_tokens"`   // Default: 4096
    Temperature float32 `yaml:"temperature"` // Default: 0.1
}
```

#### File Processing Logic

```go
// Priority file extensions (same as original)
codeExtensions := []string{".py", ".js", ".ts", ".java", ".cpp", ".c", ".h", ".hpp", ".go", ".rs", ".jsx", ".tsx", ".html", ".css", ".php", ".swift", ".cs"}
docExtensions := []string{".md", ".txt", ".rst", ".json", ".yaml", ".yml"}

// Default exclusions
excludedDirs := []string{".venv", "node_modules", ".git", "dist", "build", "target", ".next", "__pycache__"}
excludedFiles := []string{"*.min.js", "*.pyc", "*.class", "package-lock.json", "yarn.lock"}
```

#### Core Prompts (From Original Analysis)

##### Wiki Structure Generation Prompt:

```
Analyze this local directory and create a wiki structure for it.

1. The complete file tree:
<file_tree>
{fileTree}
</file_tree>

2. The README file:
<readme>
{readme}
</readme>

Create a structured wiki with 8-12 pages covering:
- Overview, System Architecture, Core Features
- Data Management/Flow, Components
- API/Services, Deployment, Extensibility

Return in XML format: <wiki_structure>...</wiki_structure>
```

##### Page Generation Prompt:

```
You are an expert technical writer generating comprehensive documentation.

Generate a wiki page about "{pageTitle}" based ONLY on these source files:
{relevantFiles}

Requirements:
1. Start with <details> block listing ALL source files (minimum 5)
2. Use extensive Mermaid diagrams (flowchart TD, sequenceDiagram, classDiagram)
3. Include code snippets with proper formatting
4. Cite sources: [filename.ext:line_numbers]()
5. Use markdown tables for structured data
6. Ensure technical accuracy from source files only
```

#### Concurrency & Performance

- Process embeddings in batches (500 texts per batch)
- Generate wiki pages with controlled concurrency (max 2-3 concurrent)
- Progress bars using appropriate Go library
- Caching of embeddings and generated content

#### Project Structure

```
deepwiki-cli/
├── cmd/
│   └── root.go           # CLI commands
├── pkg/
│   ├── scanner/         # File scanning
│   ├── processor/       # Text processing
│   ├── embeddings/      # OpenAI embeddings
│   ├── generator/       # Wiki generation
│   ├── rag/            # RAG system
│   └── output/         # Output generation
├── internal/
│   ├── config/         # Configuration
│   └── prompts/        # Prompt templates
├── examples/           # Example outputs
├── docs/              # Documentation
├── go.mod
├── go.sum
└── main.go            # Entry point
```

### 6. Dependencies

- **CLI**: `github.com/spf13/cobra`
- **OpenAI**: `github.com/sashabaranov/go-openai`
- **Vector DB**: `go.etcd.io/bbolt` (local storage)
- **Progress**: `github.com/schollz/progressbar/v3`
- **YAML**: `gopkg.in/yaml.v3`
- **Tokenizer**: Go port of tiktoken

### 7. Configuration File Example (`deepwiki.yaml`)

```yaml
openai:
  api_key: "${OPENAI_API_KEY}"
  model: "gpt-4o"
  embedding_model: "text-embedding-3-small"
  max_tokens: 4000
  temperature: 0.7

processing:
  chunk_size: 350
  chunk_overlap: 100
  max_files: 1000

filters:
  include_extensions: [".go", ".py", ".js", ".ts", ".md", ".yaml", ".json"]
  exclude_dirs: ["node_modules", ".git", "dist", "build", "vendor"]
  exclude_files: ["*.min.js", "*.pyc", "package-lock.json"]

output:
  format: "markdown" # markdown, html, json
  directory: "./wiki"
  language: "en"

embeddings:
  enabled: true
  dimensions: 256
  top_k: 20
```

### 8. Output Structure

```
docs/
├── index.md                    # Main wiki index
├── pages/
│   ├── overview.md
│   ├── architecture.md
│   ├── core-features.md
│   └── ...
├── wiki-structure.json         # Machine-readable structure
└── assets/
    └── diagrams/              # Generated Mermaid diagrams
```

## Key Differences from Original

### Simplified Architecture:

- **No WebSocket/HTTP server** - Pure CLI tool
- **No chat interface** - Focus on documentation generation
- **No caching system** - Generate fresh each time
- **No multi-provider support** - OpenAI only
- **No web UI** - Terminal-based with progress indicators

### Enhanced Local Processing:

- **Direct directory access** - No git cloning needed
- **Efficient file processing** - Streaming and batching
- **Configurable output** - Multiple format support
- **Better error handling** - Robust file system operations

### Core Features Maintained:

- **Same AI prompts** - Proven documentation generation logic
- **Same file filtering** - Battle-tested exclusion rules
- **Same chunking strategy** - 350 words, 100 overlap
- **Same embedding approach** - OpenAI text-embedding-3-small
- **Same output quality** - Comprehensive technical documentation

## Success Metrics

- **Performance**: Process 1000+ files in under 5 minutes
- **Quality**: Generate comprehensive documentation matching original quality
- **Usability**: Simple one-command operation for any project
- **Reliability**: Handle edge cases and provide clear error messages
- **Extensibility**: Clean architecture for future enhancements

This plan maintains the proven AI-powered documentation generation capabilities of deepwiki while creating a streamlined, local-first CLI tool perfectly suited for developers who want to generate documentation for their current working directory.
