# DeepWiki CLI Tutorial

This tutorial will guide you through using DeepWiki CLI to generate comprehensive documentation for your projects.

## Prerequisites

Before starting, ensure you have:

1. **Go 1.24+** installed
2. **OpenAI API Key** (get one from [OpenAI Platform](https://platform.openai.com/api-keys))
3. **DeepWiki CLI** installed

## Tutorial 1: Basic Usage

### Step 1: Installation

```bash
# Clone and build
git clone https://github.com/your-org/deepwiki-cli.git
cd deepwiki-cli
go build -o deepwiki-cli

# Or install directly
go install github.com/your-org/deepwiki-cli@latest
```

### Step 2: Set Up Your Environment

```bash
# Set your OpenAI API key
export OPENAI_API_KEY="sk-your-actual-api-key-here"

# Verify installation
./deepwiki-cli version
```

### Step 3: Generate Your First Documentation

```bash
# Navigate to your project
cd /path/to/your/project

# Generate documentation (basic)
./deepwiki-cli generate

# This will:
# 1. Scan your project files
# 2. Generate embeddings for code understanding
# 3. Create wiki structure
# 4. Generate detailed documentation pages
# 5. Output to ./docs/ directory
```

### Step 4: Explore the Generated Documentation

```bash
# View the main index
cat docs/index.md

# Browse individual pages
ls docs/pages/

# Check the machine-readable structure
cat docs/wiki-structure.json
```

## Tutorial 2: Configuration and Customization

### Step 1: Create a Configuration File

```bash
# Generate a configuration template
./deepwiki-cli config init > deepwiki.yaml

# Edit the configuration
vim deepwiki.yaml
```

Example customization:

```yaml
openai:
  model: "gpt-4o-mini" # Use faster, cheaper model
  temperature: 0.2 # More focused output

output:
  directory: "./documentation"
  language: "ja" # Japanese documentation

filters:
  exclude_dirs:
    - "node_modules"
    - "vendor"
    - "tests"
  include_extensions:
    - ".go"
    - ".py"
    - ".js"
    - ".md"
```

### Step 2: Use Custom Configuration

```bash
# Validate your configuration
./deepwiki-cli config validate deepwiki.yaml

# Generate with custom config
./deepwiki-cli generate --config deepwiki.yaml
```

### Step 3: Use Command Line Overrides

```bash
# Override specific settings
./deepwiki-cli generate \
  --output-dir ./custom-docs \
  --language en \
  --model gpt-4o \
  --exclude-dirs "tests,examples" \
  --verbose
```

## Tutorial 3: Working with Large Projects

### Step 1: Preview with Dry Run

```bash
# See what would be generated without actually doing it
./deepwiki-cli generate --dry-run --verbose
```

### Step 2: Filter Large Projects

```bash
# For projects with many files, use filtering
./deepwiki-cli generate \
  --exclude-dirs "node_modules,vendor,.git,dist,build" \
  --exclude-files "*.min.js,*.pyc,*.class" \
  --chunk-size 300 \
  --verbose
```

### Step 3: Monitor Progress

```bash
# Use verbose mode to see detailed progress
./deepwiki-cli generate --verbose

# Output will show:
# - File scanning progress
# - Embedding generation status
# - Wiki page creation progress
# - Token usage and costs
# - Processing time estimates
```

## Tutorial 4: Different Output Formats

### Markdown Output (Default)

```bash
# Generate markdown documentation
./deepwiki-cli generate --format markdown --output-dir ./markdown-docs
```

Output structure:

```
markdown-docs/
├── index.md                 # Main navigation
├── pages/
│   ├── overview.md
│   ├── architecture.md
│   ├── api-documentation.md
│   └── deployment.md
├── wiki-structure.json
└── assets/
    └── diagrams/
```

### JSON Output for Integration

```bash
# Generate JSON for tool integration
./deepwiki-cli generate --format json --output-dir ./json-docs
```

Output structure:

```
json-docs/
├── wiki.json              # Complete wiki data
├── index.json             # Navigation structure
└── pages/
    ├── page1.json
    ├── page2.json
    └── page3.json
```

## Tutorial 5: Multi-Language Projects

### Step 1: Configure for Polyglot Projects

```yaml
# deepwiki-polyglot.yaml
filters:
  include_extensions:
    - ".go" # Go files
    - ".py" # Python files
    - ".js" # JavaScript files
    - ".ts" # TypeScript files
    - ".java" # Java files
    - ".cpp" # C++ files
    - ".rs" # Rust files
    - ".md" # Documentation
    - ".yaml" # Configuration
    - ".json" # Data files

processing:
  chunk_size: 400 # Larger chunks for complex projects
  max_files: 1500 # Handle more files

output:
  language: "en"
```

### Step 2: Generate Multi-Language Documentation

```bash
./deepwiki-cli generate \
  --config deepwiki-polyglot.yaml \
  --output-dir ./polyglot-docs \
  --verbose
```

## Tutorial 6: Integration with Development Workflow

### Git Hook Integration

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash
# Pre-commit hook to update documentation

echo "Updating documentation..."
./deepwiki-cli generate \
  --output-dir ./docs \
  --format markdown \
  --config .deepwiki.yaml

# Add generated docs to commit
git add docs/
echo "Documentation updated and staged for commit."
```

### CI/CD Integration (GitHub Actions)

Create `.github/workflows/documentation.yml`:

```yaml
name: Update Documentation

on:
  push:
    branches: [main, develop]
    paths:
      - "src/**"
      - "lib/**"
      - "cmd/**"

jobs:
  docs:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v3
        with:
          go-version: "1.21"

      - name: Install DeepWiki CLI
        run: go install github.com/your-org/deepwiki-cli@latest

      - name: Generate Documentation
        env:
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
        run: |
          deepwiki-cli generate \
            --output-dir ./docs \
            --format markdown \
            --verbose

      - name: Commit Documentation
        run: |
          git config --local user.email "action@github.com"
          git config --local user.name "GitHub Action"
          git add docs/
          git diff --staged --quiet || git commit -m "📚 Update documentation"
          git push
```

## Tutorial 7: Advanced Features

### Custom Prompt Engineering

Create custom prompts in your configuration:

```yaml
prompts:
  wiki_structure: |
    Analyze this codebase and create a comprehensive wiki structure.
    Focus on:
    1. System architecture and design patterns
    2. API endpoints and data models
    3. Configuration and deployment
    4. Developer guides and best practices

  page_content: |
    Generate detailed documentation for: {page_title}

    Requirements:
    - Include relevant code examples
    - Add mermaid diagrams where helpful
    - Cite specific source files
    - Focus on practical usage examples
```

### Performance Optimization

For large codebases:

```bash
# Optimize for speed
./deepwiki-cli generate \
  --model gpt-4o-mini \
  --chunk-size 250 \
  --exclude-dirs "tests,docs,examples,vendor" \
  --format json \
  --verbose

# Optimize for quality
./deepwiki-cli generate \
  --model gpt-4o \
  --chunk-size 400 \
  --format markdown \
  --verbose
```

### Caching for Development

Enable caching for repeated runs:

```yaml
cache:
  enabled: true
  directory: "./.deepwiki-cache"
  max_age: "24h"
  max_size: "100MB"
```

## Tutorial 8: Troubleshooting Common Issues

### Issue 1: API Rate Limits

```bash
# If you hit rate limits, use a slower model or add delays
./deepwiki-cli generate \
  --model gpt-4o-mini \
  --config-rate-limit 1 \
  --verbose
```

### Issue 2: Large Project Timeouts

```bash
# Break large projects into chunks
./deepwiki-cli generate \
  --exclude-dirs "frontend,mobile" \
  --output-dir ./backend-docs

./deepwiki-cli generate \
  --include-dirs "frontend" \
  --output-dir ./frontend-docs
```

### Issue 3: Memory Issues

```bash
# Reduce memory usage
export GOMAXPROCS=2
./deepwiki-cli generate \
  --chunk-size 200 \
  --exclude-dirs "large-data-directory" \
  --verbose
```

### Issue 4: Token Limit Errors

```bash
# Reduce chunk size to avoid token limits
./deepwiki-cli generate \
  --chunk-size 250 \
  --model gpt-4o-mini \
  --verbose
```

## Best Practices

### 1. Project Preparation

- Clean up your codebase before generation
- Ensure your README.md is comprehensive
- Add meaningful comments to complex code
- Remove or exclude generated/build files

### 2. Configuration Management

- Use project-specific configuration files
- Store API keys in environment variables
- Version control your deepwiki configuration
- Document your configuration choices

### 3. Output Management

- Add generated docs to `.gitignore` or commit them
- Use meaningful output directory names
- Consider separate documentation repositories
- Regularly regenerate to keep docs fresh

### 4. Cost Management

- Start with smaller models (gpt-4o-mini) for testing
- Use filtering to reduce the amount of code processed
- Monitor token usage in verbose mode
- Cache results during development

### 5. Quality Optimization

- Review and edit generated documentation
- Add custom prompts for project-specific needs
- Supplement AI-generated content with manual documentation
- Use the dry-run mode to preview before generation

## Next Steps

After completing these tutorials, you should be able to:

1. ✅ Generate basic documentation for any project
2. ✅ Customize output format and content
3. ✅ Integrate DeepWiki CLI into your development workflow
4. ✅ Troubleshoot common issues
5. ✅ Optimize for performance and cost

### Advanced Topics

- Custom prompt engineering
- Integration with documentation platforms
- Automated quality checking
- Multi-repository documentation
- Custom output templates

### Community Resources

- [GitHub Issues](https://github.com/your-org/deepwiki-cli/issues) - Report bugs and request features
- [Discussions](https://github.com/your-org/deepwiki-cli/discussions) - Ask questions and share tips
- [Wiki](https://github.com/your-org/deepwiki-cli/wiki) - Additional documentation and examples

Happy documenting! 🚀📚
