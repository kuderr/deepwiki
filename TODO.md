# DeepWiki CLI Implementation TODO

## Project Overview

Go CLI tool reimplementation of deepwiki for local directory documentation generation using OpenAI.

## 🛠️ Technical Debt & Improvements

### High Priority TODOs

#### Text Processing & Chunking

- [ ] **Integrate tiktoken-go for accurate token counting** in `pkg/processor/processor.go:485`
  - Replace rough estimation (~4 chars per token) with proper tokenization
  - Affects chunking accuracy and token limits
- [ ] **Use more accurate file size limits** in `pkg/processor/processor.go:458`
  - Base limits on content type rather than arbitrary multiplier
  - Consider language-specific token densities
- [ ] **Fix citations** in ExtractSourceCitations and prompts: they are always with no ref in `[api.ts:108-120](#)`
  - [ ] And point somehow to real code?

#### RAG & Retrieval

- [ ] **Implement sophisticated tag matching** in `pkg/rag/retriever.go:151`
  - Add fuzzy search, synonyms, and relevance scoring
  - Consider using embedding-based tag similarity
- [ ] **Use LanguageSpecificProcessor`s from pkg/processor parsing for code structure detection** in `pkg/rag/retriever.go:696` ?
- [ ] **Implement proper context enrichment** in `pkg/rag/retriever.go:771`
  - Add surrounding chunks, function context, class hierarchy
  - Include line numbers and code structure information
- [ ] **Use hash-based cache keys** in `pkg/rag/retriever.go:755`
  - Include all relevant context parameters in cache key
  - Prevent cache collisions and improve hit rates

#### Wiki Generation & Content Processing

- [ ] **Actually read and parse README file content** in `pkg/generator/generator.go:275`
  - Replace placeholder detection with actual file reading and parsing
  - Extract meaningful content for wiki structure generation
- [ ] **Add file tree context for better page generation** in `pkg/generator/generator.go:130`
  - Include project file tree in page content prompts
  - Help AI understand project structure for better documentation
- [ ] **Improve file tree structure with proper hierarchy** in `pkg/generator/generator.go:251`
  - Build proper tree structure with correct indentation
  - Sort directories and files logically
- [ ] **Enhance ID generation with UUIDs or collision detection** in `pkg/generator/parser.go:151`
  - Replace simple string sanitization with more robust ID generation
  - Ensure uniqueness across wiki structures and prevent collisions
- [ ] Make pages count configurable

#### Code Quality & Architecture

- [ ] **Implement proper word truncation** in `pkg/embeddings/generator.go:315`
  - Respect character boundaries and word integrity
  - Handle Unicode properly
- [ ] **Add comprehensive error handling** for edge cases
  - Improve error messages with context
  - Add recovery mechanisms for partial failures
- [ ] **Layer separation** by services/dbs and stuff ?
  - Better file layering in packages: interfaces, structs/types, implementation, tests
  - Use services as entrypoint to packages as in pkg/embeddings/search_service.go

### Medium Priority TODOs

#### Performance Optimizations

- [ ] **Implement concurrent chunk processing** where applicable
- [ ] **Add connection pooling** for database operations
- [ ] **Optimize memory usage** for large file processing
- [ ] **Implement result streaming** for large search results

#### Feature Enhancements

- [ ] **Add support for more programming languages** in semantic chunking
- [ ] **Implement configurable similarity metrics** beyond cosine
- [ ] **Add query suggestion and auto-completion**
- [ ] **Implement result ranking and relevance feedback**

#### Testing & Quality Assurance

- [ ] **Add integration tests** for complete workflows
- [ ] **Implement performance benchmarks** for large codebases
- [ ] **Add property-based testing** for edge cases
- [ ] **Create stress tests** for concurrent operations

### Low Priority TODOs

#### Documentation & Usability

- [ ] **Add comprehensive API documentation**
- [ ] **Create usage examples** for different scenarios
- [ ] **Implement debugging and profiling tools**
- [ ] **Add metrics and monitoring capabilities**
- [ ] **Sample config** deepwiki-cli config init creates full default config

#### Advanced Features

- [ ] **Implement custom embedding models** support
- [ ] **Add multi-language query support**
- [ ] **Create plugin system** for custom processors
- [ ] **Implement distributed processing** for large projects

---

## 📝 Additional Features (Future Enhancements)

### Nice-to-Have Features

- [ ] HTML output with static site generation
- [ ] Interactive CLI with prompts for configuration
- [ ] Plugin system for custom processors
- [ ] Integration with popular documentation sites
- [ ] Watch mode for automatic regeneration
- [ ] Custom template support
- [ ] Diff mode to show changes between generations
- [ ] docusaurus likec4 support

### Advanced Features

- [ ] Support for additional LLM providers (if needed)
- [ ] Custom embedding models
- [ ] Advanced filtering with regex patterns
- [ ] Integration with git for change tracking
- [ ] API server mode for team usage
- [ ] Docker containerization

## 🚀 Success Criteria

### Performance Targets

- [ ] Process 1000+ files in under 5 minutes
- [ ] Memory usage under 1GB for large projects
- [ ] Accurate file filtering (no false positives/negatives)
- [ ] High-quality documentation generation

### Quality Targets

- [ ] Comprehensive documentation matching original quality
- [ ] Clean, maintainable Go code
- [ ] Robust error handling and recovery
- [ ] Clear, helpful error messages

### Usability Targets

- [ ] Simple one-command operation
- [ ] Intuitive CLI interface
- [ ] Good documentation and examples
- [ ] Cross-platform compatibility
