# Troubleshooting Guide

This guide helps you diagnose and resolve common issues with DeepWiki CLI.

## Quick Diagnostics

### 1. Check Installation

```bash
# Verify DeepWiki CLI is installed and accessible
deepwiki-cli version

# Expected output:
# DeepWiki CLI version 1.0.0
# Built with Go 1.24+
```

### 2. Validate Configuration

```bash
# Check your configuration
deepwiki-cli config validate

# Show resolved configuration (combines all sources)
deepwiki-cli config validate --show-resolved
```

### 3. Test with Dry Run

```bash
# Preview what would be generated without actually doing it
deepwiki-cli generate --dry-run --verbose
```

## Common Issues

### 🔑 OpenAI API Issues

#### Issue: "API key not provided"

**Symptoms:**

```
Error: OpenAI API key is required. Set --openai-key flag or OPENAI_API_KEY environment variable
```

**Solutions:**

```bash
# Option 1: Environment variable (recommended)
export OPENAI_API_KEY="sk-your-actual-api-key-here"

# Option 2: Command line flag
deepwiki-cli generate --openai-key "sk-your-actual-api-key-here"

# Option 3: Configuration file
echo "openai:\n  api_key: \"sk-your-actual-api-key-here\"" > deepwiki.yaml
```

**Verification:**

```bash
# Test API connectivity
echo $OPENAI_API_KEY  # Should show your key
deepwiki-cli generate --dry-run  # Should not show API key error
```

#### Issue: "Invalid API key"

**Symptoms:**

```
Error: HTTP 401 - Invalid API key provided
```

**Solutions:**

1. Verify your API key is correct:

   - Go to [OpenAI Platform](https://platform.openai.com/api-keys)
   - Copy the key exactly (including `sk-` prefix)
   - Ensure no extra spaces or characters

2. Check key permissions:

   - Ensure the key has sufficient permissions
   - Some keys may be restricted to specific models

3. Verify key is active:
   - Old keys may have been deactivated
   - Generate a new key if needed

#### Issue: "Rate limit exceeded"

**Symptoms:**

```
Error: HTTP 429 - Rate limit exceeded. Please try again later
```

**Solutions:**

```bash
# Option 1: Wait and retry (rate limits reset over time)
sleep 60
deepwiki-cli generate

# Option 2: Use a slower request rate
deepwiki-cli generate --config rate-limited.yaml
```

Rate-limited configuration (`rate-limited.yaml`):

```yaml
openai:
  api_key: "${OPENAI_API_KEY}"
  model: "gpt-4o-mini" # Cheaper model

rate_limiting:
  requests_per_second: 1 # Slower rate
  max_retries: 5
  retry_delay: "2s"
```

#### Issue: "Token limit exceeded"

**Symptoms:**

```
Error: This model's maximum context length is 4096 tokens, however you requested 5000 tokens
```

**Solutions:**

```bash
# Option 1: Reduce chunk size
deepwiki-cli generate --chunk-size 250

# Option 2: Use a model with larger context
deepwiki-cli generate --model gpt-4o  # 128k context

# Option 3: Configure in file
```

Configuration for large contexts:

```yaml
openai:
  model: "gpt-4o"
  max_tokens: 4000

processing:
  chunk_size: 300
  chunk_overlap: 50
```

#### Issue: "Insufficient quota"

**Symptoms:**

```
Error: You exceeded your current quota, please check your plan and billing details
```

**Solutions:**

1. Check your OpenAI billing:

   - Go to [OpenAI Usage](https://platform.openai.com/usage)
   - Add payment method if needed
   - Upgrade plan if necessary

2. Use a cheaper model temporarily:

```bash
deepwiki-cli generate --model gpt-4o-mini
```

### 📁 File System Issues

#### Issue: "Permission denied"

**Symptoms:**

```
Error: failed to scan directory: permission denied
Error: failed to create output directory: permission denied
```

**Solutions:**

```bash
# Check directory permissions
ls -la /path/to/project
ls -la /path/to/output

# Fix permissions if you own the directory
chmod 755 /path/to/project
chmod 755 /path/to/output

# Use a different output directory
deepwiki-cli generate --output-dir ~/Documents/wiki-docs

# Run with appropriate user permissions
sudo deepwiki-cli generate  # Use with caution
```

#### Issue: "Directory does not exist"

**Symptoms:**

```
Error: directory does not exist: /path/to/project
```

**Solutions:**

```bash
# Verify the path exists
ls -la /path/to/project

# Use absolute path
deepwiki-cli generate $(pwd)

# Create the directory if needed
mkdir -p /path/to/project
```

#### Issue: "Output directory not empty"

**Symptoms:**

```
Warning: Output directory contains existing files
```

**Solutions:**

```bash
# Clean output directory
rm -rf ./docs/*

# Use a new output directory
deepwiki-cli generate --output-dir ./fresh-docs

# Or backup existing files
mv ./docs ./docs-backup
```

### 💾 Memory Issues

#### Issue: "Out of memory"

**Symptoms:**

```
Error: runtime: out of memory
fatal error: runtime: out of memory
```

**Solutions:**

```bash
# Option 1: Reduce memory usage
export GOMAXPROCS=2  # Limit Go processes
deepwiki-cli generate \
  --chunk-size 200 \
  --exclude-dirs "large-directory" \
  --max-files 500

# Option 2: Use memory-efficient configuration
```

Memory-efficient configuration:

```yaml
processing:
  chunk_size: 250
  max_files: 300

advanced:
  memory_limit: 256 # MB
  max_workers: 2
  gc_optimization: true

filters:
  exclude_dirs:
    - "node_modules"
    - "vendor"
    - "build"
    - "dist"
    - "data" # Large data directories
```

#### Issue: "Process killed (OOM)"

**Symptoms:**

```
Killed
(Process terminated without error message)
```

**Solutions:**

```bash
# Check available memory
free -h  # Linux
vm_stat | grep free  # macOS

# Monitor memory usage during execution
top -p $(pgrep deepwiki-cli)

# Use smaller batches
deepwiki-cli generate \
  --chunk-size 150 \
  --exclude-dirs "tests,docs,examples" \
  --verbose
```

### 🐌 Performance Issues

#### Issue: "Generation takes too long"

**Symptoms:**

- Process runs for hours without completing
- Very slow progress updates

**Solutions:**

```bash
# Option 1: Use faster model
deepwiki-cli generate --model gpt-4o-mini

# Option 2: Reduce scope
deepwiki-cli generate \
  --exclude-dirs "tests,docs,vendor" \
  --max-files 200

# Option 3: Use performance config
```

Performance configuration:

```yaml
openai:
  model: "gpt-4o-mini"
  max_tokens: 2000

processing:
  chunk_size: 300
  max_files: 500

advanced:
  max_workers: 6
  batch_size: 15

rate_limiting:
  requests_per_second: 3 # If your plan allows
```

#### Issue: "High token usage/costs"

**Symptoms:**

- Unexpectedly high API costs
- Many tokens being used

**Solutions:**

```bash
# Monitor token usage
deepwiki-cli generate --verbose  # Shows token counts

# Use cost-effective configuration
```

Cost-effective configuration:

```yaml
openai:
  model: "gpt-4o-mini" # Much cheaper than gpt-4o
  max_tokens: 1500
  temperature: 0.0 # More consistent, less retries

processing:
  chunk_size: 250 # Smaller chunks = fewer tokens
  max_files: 300 # Limit scope

filters:
  exclude_dirs:
    - "tests"
    - "examples"
    - "docs"
    - "vendor"
```

### 🗂️ File Processing Issues

#### Issue: "No files found to process"

**Symptoms:**

```
Warning: No files found matching the specified criteria
Generated 0 pages
```

**Solutions:**

```bash
# Check file filtering
deepwiki-cli generate --dry-run --verbose

# Verify include extensions
deepwiki-cli generate \
  --include-extensions ".go,.py,.js,.md" \
  --dry-run

# Remove overly restrictive exclusions
deepwiki-cli generate \
  --exclude-dirs "node_modules" \
  --exclude-files "*.min.js"
```

Debug file filtering:

```yaml
filters:
  include_extensions:
    - ".go"
    - ".py"
    - ".js"
    - ".ts"
    - ".md"
    - ".txt"
  exclude_dirs:
    - "node_modules" # Only essential exclusions
  exclude_files:
    - "*.min.js"

logging:
  level: "debug" # Shows file filtering details
```

#### Issue: "Binary files being processed"

**Symptoms:**

```
Warning: Binary file detected: image.png
Error: Invalid UTF-8 in file: binary.exe
```

**Solutions:**

```bash
# Add binary file exclusions
deepwiki-cli generate \
  --exclude-files "*.png,*.jpg,*.gif,*.exe,*.dll,*.so"
```

Enhanced filtering configuration:

```yaml
filters:
  exclude_files:
    # Images
    - "*.png"
    - "*.jpg"
    - "*.jpeg"
    - "*.gif"
    - "*.bmp"
    - "*.tiff"
    - "*.ico"
    - "*.svg"

    # Videos
    - "*.mp4"
    - "*.avi"
    - "*.mov"
    - "*.wmv"

    # Audio
    - "*.mp3"
    - "*.wav"
    - "*.flac"

    # Archives
    - "*.zip"
    - "*.tar"
    - "*.gz"
    - "*.rar"
    - "*.7z"

    # Executables
    - "*.exe"
    - "*.dll"
    - "*.so"
    - "*.dylib"
    - "*.app"

    # Documents
    - "*.pdf"
    - "*.doc"
    - "*.docx"
    - "*.xls"
    - "*.xlsx"
    - "*.ppt"
    - "*.pptx"
```

### 🔧 Configuration Issues

#### Issue: "Configuration file not found"

**Symptoms:**

```
Warning: Configuration file not found, using defaults
```

**Solutions:**

```bash
# Create configuration file
deepwiki-cli config init > deepwiki.yaml

# Use specific config file
deepwiki-cli generate --config /path/to/config.yaml

# Check search paths
deepwiki-cli config validate --verbose
```

#### Issue: "Invalid configuration"

**Symptoms:**

```
Error: invalid configuration: temperature must be between 0.0 and 2.0, got 3.0
```

**Solutions:**

```bash
# Validate configuration
deepwiki-cli config validate deepwiki.yaml

# Fix common validation errors:
# - temperature: 0.0-2.0
# - chunk_size: 100-1000
# - max_tokens: 1-32768 (model dependent)
# - language: en, ja, zh, es, kr, vi
# - format: markdown, json
```

### 🌐 Network Issues

#### Issue: "Connection timeout"

**Symptoms:**

```
Error: connection timeout
Error: dial tcp: i/o timeout
```

**Solutions:**

```bash
# Check internet connectivity
ping api.openai.com

# Use longer timeout
export HTTP_TIMEOUT=120
deepwiki-cli generate

# Check proxy settings
echo $HTTP_PROXY
echo $HTTPS_PROXY

# Test with curl
curl -H "Authorization: Bearer $OPENAI_API_KEY" \
     https://api.openai.com/v1/models
```

#### Issue: "SSL certificate errors"

**Symptoms:**

```
Error: x509: certificate signed by unknown authority
```

**Solutions:**

```bash
# Update certificates (Linux)
sudo apt update && sudo apt install ca-certificates

# macOS - update system
sudo softwareupdate -i -a

# Bypass SSL (not recommended for production)
export SSL_VERIFY=false
```

### 📝 Output Issues

#### Issue: "Generated documentation is empty"

**Symptoms:**

- Files are created but have no content
- Pages contain only headers

**Solutions:**

```bash
# Check if files were actually processed
deepwiki-cli generate --verbose --dry-run

# Verify API responses
deepwiki-cli generate --verbose  # Shows API calls

# Try different model
deepwiki-cli generate --model gpt-4o
```

#### Issue: "Malformed markdown/JSON"

**Symptoms:**

- Broken markdown formatting
- Invalid JSON structure

**Solutions:**

```bash
# Use lower temperature for more consistent output
deepwiki-cli generate --temperature 0.0

# Check API responses in verbose mode
deepwiki-cli generate --verbose

# Regenerate specific pages
rm -rf docs/pages/problematic-page.md
deepwiki-cli generate
```

## Debugging Techniques

### 1. Enable Verbose Logging

```bash
# Maximum verbosity
deepwiki-cli generate --verbose

# Debug-level logging in config
```

```yaml
logging:
  level: "debug"
  format: "text"
  output: "stderr"
  add_source: true
```

### 2. Use Dry Run Mode

```bash
# See what would be processed without API calls
deepwiki-cli generate --dry-run --verbose

# Check file discovery
deepwiki-cli generate --dry-run | grep "Files found:"
```

### 3. Test API Connectivity

```bash
# Test OpenAI API directly
curl -H "Authorization: Bearer $OPENAI_API_KEY" \
     -H "Content-Type: application/json" \
     -d '{
       "model": "gpt-4o-mini",
       "messages": [{"role": "user", "content": "Hello"}],
       "max_tokens": 5
     }' \
     https://api.openai.com/v1/chat/completions
```

### 4. Monitor Resource Usage

```bash
# Monitor during execution
top -p $(pgrep deepwiki-cli)

# Check disk space
df -h

# Monitor network
netstat -an | grep :443
```

### 5. Profile Performance

```bash
# Build with profiling
go build -tags debug

# Run with profiling
deepwiki-cli generate --profile cpu

# Analyze profile
go tool pprof cpu.prof
```

## Getting Help

### 1. Collect Debug Information

Before asking for help, collect this information:

```bash
# Version information
deepwiki-cli version

# System information
go version
uname -a  # Linux/macOS
systeminfo  # Windows

# Configuration
deepwiki-cli config validate --show-resolved

# Error logs
deepwiki-cli generate --verbose 2>&1 | tee debug.log
```

### 2. Create Minimal Reproduction

Create the smallest possible example that reproduces the issue:

```bash
# Create test project
mkdir test-project
cd test-project
echo "package main\n\nfunc main() {}" > main.go
echo "# Test Project" > README.md

# Test with minimal config
export OPENAI_API_KEY="your-key"
deepwiki-cli generate --verbose
```

### 3. Report Issues

When reporting issues, include:

1. **Environment**: OS, Go version, DeepWiki CLI version
2. **Configuration**: Your configuration file (redact API key)
3. **Command**: Exact command that failed
4. **Error**: Complete error message and logs
5. **Context**: Project size, file types, etc.

### 4. Community Resources

- **GitHub Issues**: [Report bugs and request features](https://github.com/your-org/deepwiki-cli/issues)
- **Discussions**: [Ask questions and share tips](https://github.com/your-org/deepwiki-cli/discussions)
- **Wiki**: [Additional documentation](https://github.com/your-org/deepwiki-cli/wiki)

## Prevention Best Practices

### 1. Test Before Production

```bash
# Always test with dry run first
deepwiki-cli generate --dry-run

# Test with small subset
deepwiki-cli generate --max-files 10 --verbose

# Use development configuration
deepwiki-cli generate --config dev-config.yaml
```

### 2. Monitor Resource Usage

```bash
# Set resource limits
ulimit -v 1000000  # 1GB virtual memory limit

# Monitor API usage
deepwiki-cli generate --verbose | grep "Tokens used:"
```

### 3. Backup and Version Control

```bash
# Backup existing documentation
cp -r docs docs-backup

# Version control configuration
git add deepwiki.yaml
git commit -m "Add DeepWiki configuration"
```

### 4. Regular Maintenance

```bash
# Update DeepWiki CLI regularly
go install github.com/your-org/deepwiki-cli@latest

# Review and update configuration
deepwiki-cli config validate
```

This troubleshooting guide should help you resolve most common issues with DeepWiki CLI. If you encounter an issue not covered here, please report it on our GitHub issues page.
