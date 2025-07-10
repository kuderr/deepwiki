package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kuderr/deepwiki/internal/logging"
	"github.com/kuderr/deepwiki/pkg/embedding"
	embeddingfactory "github.com/kuderr/deepwiki/pkg/embedding/factory"
	"github.com/kuderr/deepwiki/pkg/llm"
	llmfactory "github.com/kuderr/deepwiki/pkg/llm/factory"
	"github.com/kuderr/deepwiki/pkg/types"
	"gopkg.in/yaml.v3"
)

// Config represents the complete configuration for deepwiki
type Config struct {
	Providers  ProviderConfig    `yaml:"providers"`
	Processing ProcessingConfig  `yaml:"processing"`
	Filters    FiltersConfig     `yaml:"filters"`
	Output     OutputConfig      `yaml:"output"`
	Embeddings EmbeddingsConfig  `yaml:"embeddings"`
	Logging    logging.LogConfig `yaml:"logging"`
}

// ProcessingConfig contains text processing configuration
type ProcessingConfig struct {
	ChunkSize    int `yaml:"chunk_size"`
	ChunkOverlap int `yaml:"chunk_overlap"`
	MaxFiles     int `yaml:"max_files"`
}

// FiltersConfig contains file filtering configuration
type FiltersConfig struct {
	IncludeExtensions []string `yaml:"include_extensions"`
	ExcludeDirs       []string `yaml:"exclude_dirs"`
	ExcludeFiles      []string `yaml:"exclude_files"`
}

// OutputConfig contains output generation configuration
type OutputConfig struct {
	Format    string         `yaml:"format"`
	Directory string         `yaml:"directory"`
	Language  types.Language `yaml:"language"`
}

// EmbeddingsConfig contains embedding generation configuration
type EmbeddingsConfig struct {
	Enabled    bool `yaml:"enabled"`
	Dimensions int  `yaml:"dimensions"`
	TopK       int  `yaml:"top_k"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Providers: *DefaultProviderConfig(),
		Processing: ProcessingConfig{
			ChunkSize:    350,
			ChunkOverlap: 100,
			MaxFiles:     1000,
		},
		Filters: FiltersConfig{
			IncludeExtensions: []string{
				// Programming Languages
				".go", ".py", ".js", ".ts", ".jsx", ".tsx", ".java", ".cpp", ".c", ".h", ".hpp",
				".cs", ".php", ".rb", ".rs", ".swift", ".kt", ".scala", ".clj", ".hs", ".ml", ".fs",
				".dart", ".lua", ".r", ".R", ".m", ".mm", ".pl", ".sh", ".bash", ".zsh", ".fish",
				".ps1", ".bat", ".cmd",
				// Web Technologies
				".html", ".htm", ".css", ".scss", ".sass", ".less", ".vue", ".svelte",
				// Configuration & Data
				".yaml", ".yml", ".json", ".toml", ".ini", ".cfg", ".conf", ".xml", ".proto",
				".graphql", ".gql",
				// Documentation
				".md", ".mdx", ".txt", ".rst", ".org", ".tex", ".adoc",
				// Database
				".sql", ".psql", ".mysql",
				// Build & CI/CD
				".dockerfile", ".makefile", ".mk", ".gradle", ".maven", ".ant",
			},
			ExcludeDirs: []string{
				// Dependencies
				"node_modules", "vendor", ".venv", "venv", "env", ".env", "virtualenv",
				"__pycache__", ".tox", "site-packages", ".bundle", "gems", ".cargo", "target",
				".gradle", ".mvn",
				// Build Outputs
				"dist", "build", "out", "bin", "obj", "lib", ".build", "cmake-build-debug",
				"cmake-build-release",
				// Development Tools
				".git", ".svn", ".hg", ".bzr", ".idea", ".vscode", ".vs", ".eclipse",
				".settings", ".project", ".classpath", ".factorypath",
				// Temporary & Cache
				"tmp", "temp", ".tmp", ".temp", "cache", ".cache", ".next", ".nuxt",
				".angular", ".turbo",
				// Logs & Data
				"logs", "log", ".logs", "data", ".data", "backup", "backups",
				// Testing (optional)
				"test", "tests", "__tests__", "spec", ".pytest_cache", "coverage",
				".nyc_output", "htmlcov",
				// Documentation (optional)
				"docs", "doc", ".docs", "documentation", "wiki",
				// Platform Specific
				".DS_Store", "Thumbs.db", "Desktop.ini",
			},
			ExcludeFiles: []string{
				// Compiled & Binary Files
				"*.min.js", "*.min.css", "*.bundle.js", "*.chunk.js", "*.pyc", "*.pyo",
				"*.class", "*.jar", "*.war", "*.ear", "*.exe", "*.dll", "*.so", "*.dylib",
				"*.a", "*.o", "*.obj", "*.lib", "*.exp", "*.pdb",
				// Lock Files
				"package-lock.json", "yarn.lock", "pnpm-lock.yaml", "composer.lock",
				"Gemfile.lock", "Pipfile.lock", "poetry.lock", "cargo.lock", "go.sum",
				// Generated Files
				"*.generated.*", "*.gen.*", "*_generated.go", "*_gen.go", "*.pb.go",
				"*.pb.cc", "*.pb.h", "*_pb2.py", "*_pb2_grpc.py",
				// IDE & Editor Files
				"*.swp", "*.swo", "*~", ".#*", "#*#", ".*.rej", ".*.orig",
				// System Files
				".DS_Store", "Thumbs.db", "desktop.ini", "*.lnk",
				// Logs & Temporary
				"*.log", "*.tmp", "*.temp", "*.bak", "*.backup", "core", "*.dump",
			},
		},
		Output: OutputConfig{
			Format:    "markdown",
			Directory: "./docs",
			Language:  types.LanguageEnglish,
		},
		Embeddings: EmbeddingsConfig{
			Enabled:    true,
			Dimensions: 256,
			TopK:       20,
		},
		Logging: *logging.DefaultLogConfig(),
	}
}

// LoadConfig loads configuration from file, environment variables, and CLI flags
func LoadConfig(configFile string) (*Config, error) {
	config := DefaultConfig()

	// Load from config file if specified
	if configFile != "" {
		if err := loadFromFile(config, configFile); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	} else {
		// Try to find config file in common locations
		configPaths := []string{
			"deepwiki.yaml",
			"deepwiki.yml",
			".deepwiki.yaml",
			".deepwiki.yml",
		}

		// Also check home directory
		if homeDir, err := os.UserHomeDir(); err == nil {
			configPaths = append(configPaths,
				filepath.Join(homeDir, ".deepwiki.yaml"),
				filepath.Join(homeDir, ".deepwiki.yml"),
			)
		}

		for _, path := range configPaths {
			if _, err := os.Stat(path); err == nil {
				if err := loadFromFile(config, path); err != nil {
					return nil, fmt.Errorf("failed to load config file %s: %w", path, err)
				}
				break
			}
		}
	}

	// Override with environment variables
	loadFromEnv(config)

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// loadFromFile loads configuration from a YAML file
func loadFromFile(config *Config, filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, config)
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(config *Config) {
	// LLM Provider configuration
	if provider := os.Getenv("DEEPWIKI_LLM_PROVIDER"); provider != "" {
		config.Providers.LLM.Provider = provider
	}

	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" && config.Providers.LLM.Provider == "openai" {
		config.Providers.LLM.APIKey = apiKey
	}

	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" && config.Providers.LLM.Provider == "anthropic" {
		config.Providers.LLM.APIKey = apiKey
	}

	if baseURL := os.Getenv("DEEPWIKI_LLM_BASE_URL"); baseURL != "" {
		config.Providers.LLM.BaseURL = baseURL
	}

	if model := os.Getenv("DEEPWIKI_LLM_MODEL"); model != "" {
		config.Providers.LLM.Model = model
	}

	// Embedding Provider configuration
	if provider := os.Getenv("DEEPWIKI_EMBEDDING_PROVIDER"); provider != "" {
		config.Providers.Embedding.Provider = provider
	}

	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" && config.Providers.Embedding.Provider == "openai" {
		config.Providers.Embedding.APIKey = apiKey
	}

	if apiKey := os.Getenv("VOYAGE_API_KEY"); apiKey != "" && config.Providers.Embedding.Provider == "voyage" {
		config.Providers.Embedding.APIKey = apiKey
	}

	if model := os.Getenv("DEEPWIKI_EMBEDDING_MODEL"); model != "" {
		config.Providers.Embedding.Model = model
	}

	if baseURL := os.Getenv("DEEPWIKI_EMBEDDING_BASE_URL"); baseURL != "" {
		config.Providers.Embedding.BaseURL = baseURL
	}

	if outputDir := os.Getenv("DEEPWIKI_OUTPUT_DIR"); outputDir != "" {
		config.Output.Directory = outputDir
	}

	if format := os.Getenv("DEEPWIKI_FORMAT"); format != "" {
		config.Output.Format = format
	}

	if lang := os.Getenv("DEEPWIKI_LANGUAGE"); lang != "" {
		if parsedLang, err := types.ParseLanguageWithCode(lang); err == nil {
			config.Output.Language = parsedLang
		}
	}

	if excludeDirs := os.Getenv("DEEPWIKI_EXCLUDE_DIRS"); excludeDirs != "" {
		config.Filters.ExcludeDirs = append(config.Filters.ExcludeDirs,
			strings.Split(excludeDirs, ",")...)
	}

	if excludeFiles := os.Getenv("DEEPWIKI_EXCLUDE_FILES"); excludeFiles != "" {
		config.Filters.ExcludeFiles = append(config.Filters.ExcludeFiles,
			strings.Split(excludeFiles, ",")...)
	}
}

// validateConfig validates the configuration values
func validateConfig(config *Config) error {
	// Validate LLM provider configuration
	if config.Providers.LLM.Provider == "" {
		return fmt.Errorf("LLM provider is required")
	}

	if config.Providers.LLM.Model == "" {
		return fmt.Errorf("LLM model is required")
	}

	if config.Providers.LLM.MaxTokens <= 0 {
		return fmt.Errorf("LLM max tokens must be positive")
	}

	if config.Providers.LLM.Temperature < 0 || config.Providers.LLM.Temperature > 2 {
		return fmt.Errorf("LLM temperature must be between 0 and 2")
	}

	// Validate embedding provider configuration
	if config.Providers.Embedding.Provider == "" {
		return fmt.Errorf("embedding provider is required")
	}

	if config.Providers.Embedding.Model == "" {
		return fmt.Errorf("embedding model is required")
	}

	// Validate processing configuration
	if config.Processing.ChunkSize <= 0 {
		return fmt.Errorf("chunk size must be positive")
	}

	if config.Processing.ChunkOverlap < 0 {
		return fmt.Errorf("chunk overlap cannot be negative")
	}

	if config.Processing.ChunkOverlap >= config.Processing.ChunkSize {
		return fmt.Errorf("chunk overlap must be less than chunk size")
	}

	// Validate output configuration
	validFormats := map[string]bool{
		"markdown":           true,
		"json":               true,
		"docusaurus2":        true,
		"docusaurus3":        true,
		"simple-docusaurus2": true,
		"simple-docusaurus3": true,
	}
	if !validFormats[config.Output.Format] {
		return fmt.Errorf(
			"invalid output format: %s (valid: markdown, json, docusaurus2, docusaurus3, simple-docusaurus2, simple-docusaurus3)",
			config.Output.Format,
		)
	}

	if !config.Output.Language.IsValid() {
		return fmt.Errorf(
			"invalid language: %s (valid: %s)",
			config.Output.Language,
			strings.Join(types.AllLanguageCodes(), ", "),
		)
	}

	// Validate embeddings configuration
	if config.Embeddings.TopK <= 0 {
		return fmt.Errorf("embeddings top_k must be positive")
	}

	if config.Embeddings.Dimensions <= 0 {
		return fmt.Errorf("embeddings dimensions must be positive")
	}

	return nil
}

// SaveConfig saves the configuration to a YAML file
func SaveConfig(config *Config, filename string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0o644)
}

// GenerateTemplate generates a template configuration file
func GenerateTemplate() string {
	config := DefaultConfig()
	data, _ := yaml.Marshal(config)

	template := `# DeepWiki Configuration File
# This file contains default values for all configuration options.
# You can override any value here or use environment variables.

` + string(data) + `

# Environment variables that can be used:
# LLM Provider Configuration:
# DEEPWIKI_LLM_PROVIDER - LLM provider (openai, anthropic, ollama)
# OPENAI_API_KEY - OpenAI API key
# ANTHROPIC_API_KEY - Anthropic API key
# DEEPWIKI_LLM_BASE_URL - Base URL
# DEEPWIKI_LLM_MODEL - LLM model name
# 
# Embedding Provider Configuration:
# DEEPWIKI_EMBEDDING_PROVIDER - Embedding provider (openai, voyage, ollama)
# DEEPWIKI_EMBEDDING_MODEL - Embedding model name
# OPENAI_API_KEY - OpenAI API key (also used for embeddings)
# VOYAGE_API_KEY - Voyage AI API key
# DEEPWIKI_EMBEDDING_BASE_URL - Base URL
# 
# Output Configuration:
# DEEPWIKI_OUTPUT_DIR - Output directory
# DEEPWIKI_FORMAT - Output format (markdown, json)
# DEEPWIKI_LANGUAGE - Output language (English/en, Russian/ru)
# DEEPWIKI_EXCLUDE_DIRS - Additional directories to exclude (comma-separated)
# DEEPWIKI_EXCLUDE_FILES - Additional files to exclude (comma-separated)
`

	return template
}

// GetLLMProvider creates and returns an LLM provider from the configuration
func (c *Config) GetLLMProvider() (llm.Provider, error) {
	llmConfig, err := c.Providers.LLM.ToLLMConfig()
	if err != nil {
		return nil, fmt.Errorf("invalid LLM configuration: %w", err)
	}

	return llmfactory.NewLLMProvider(llmConfig)
}

// GetEmbeddingProvider creates and returns an embedding provider from the configuration
func (c *Config) GetEmbeddingProvider() (embedding.Provider, error) {
	embeddingConfig, err := c.Providers.Embedding.ToEmbeddingConfig()
	if err != nil {
		return nil, fmt.Errorf("invalid embedding configuration: %w", err)
	}

	return embeddingfactory.NewEmbeddingProvider(embeddingConfig)
}
