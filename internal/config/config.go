package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/deepwiki-cli/deepwiki-cli/internal/logging"
	"gopkg.in/yaml.v3"
)

// Config represents the complete configuration for deepwiki-cli
type Config struct {
	OpenAI     OpenAIConfig      `yaml:"openai"`
	Processing ProcessingConfig  `yaml:"processing"`
	Filters    FiltersConfig     `yaml:"filters"`
	Output     OutputConfig      `yaml:"output"`
	Embeddings EmbeddingsConfig  `yaml:"embeddings"`
	Logging    logging.LogConfig `yaml:"logging"`
}

// OpenAIConfig contains OpenAI API configuration
type OpenAIConfig struct {
	APIKey         string  `yaml:"api_key"`
	Model          string  `yaml:"model"`
	EmbeddingModel string  `yaml:"embedding_model"`
	MaxTokens      int     `yaml:"max_tokens"`
	Temperature    float32 `yaml:"temperature"`
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
	Format        string `yaml:"format"`
	Directory     string `yaml:"directory"`
	Comprehensive bool   `yaml:"comprehensive"`
	Language      string `yaml:"language"`
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
		OpenAI: OpenAIConfig{
			APIKey:         "",
			Model:          "gpt-4o",
			EmbeddingModel: "text-embedding-3-small",
			MaxTokens:      4000,
			Temperature:    0.1,
		},
		Processing: ProcessingConfig{
			ChunkSize:    350,
			ChunkOverlap: 100,
			MaxFiles:     1000,
		},
		Filters: FiltersConfig{
			IncludeExtensions: []string{
				".go", ".py", ".js", ".ts", ".java", ".cpp", ".c", ".h", ".hpp",
				".rs", ".jsx", ".tsx", ".html", ".css", ".php", ".swift", ".cs",
				".md", ".txt", ".rst", ".json", ".yaml", ".yml",
			},
			ExcludeDirs: []string{
				"node_modules", ".git", "dist", "build", "target", ".next",
				"__pycache__", ".venv", "venv", "vendor", ".cargo", "bin", "obj",
			},
			ExcludeFiles: []string{
				"*.min.js", "*.pyc", "*.class", "package-lock.json", "yarn.lock",
				"*.exe", "*.dll", "*.so", "*.dylib", "*.a", "*.o",
			},
		},
		Output: OutputConfig{
			Format:        "markdown",
			Directory:     "./docs",
			Comprehensive: true,
			Language:      "en",
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
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		config.OpenAI.APIKey = apiKey
	}

	if model := os.Getenv("DEEPWIKI_MODEL"); model != "" {
		config.OpenAI.Model = model
	}

	if embModel := os.Getenv("DEEPWIKI_EMBEDDING_MODEL"); embModel != "" {
		config.OpenAI.EmbeddingModel = embModel
	}

	if outputDir := os.Getenv("DEEPWIKI_OUTPUT_DIR"); outputDir != "" {
		config.Output.Directory = outputDir
	}

	if format := os.Getenv("DEEPWIKI_FORMAT"); format != "" {
		config.Output.Format = format
	}

	if lang := os.Getenv("DEEPWIKI_LANGUAGE"); lang != "" {
		config.Output.Language = lang
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
	// Validate OpenAI configuration
	if config.OpenAI.Model == "" {
		return fmt.Errorf("OpenAI model is required")
	}

	if config.OpenAI.EmbeddingModel == "" {
		return fmt.Errorf("OpenAI embedding model is required")
	}

	if config.OpenAI.MaxTokens <= 0 {
		return fmt.Errorf("max tokens must be positive")
	}

	if config.OpenAI.Temperature < 0 || config.OpenAI.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
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

	validLanguages := map[string]bool{
		"en": true, "ja": true, "zh": true, "es": true, "kr": true, "vi": true,
	}
	if !validLanguages[config.Output.Language] {
		return fmt.Errorf("invalid language: %s (valid: en, ja, zh, es, kr, vi)", config.Output.Language)
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

	template := `# DeepWiki CLI Configuration File
# This file contains default values for all configuration options.
# You can override any value here or use environment variables.

` + string(data) + `
# Environment variables that can be used:
# OPENAI_API_KEY - OpenAI API key
# DEEPWIKI_MODEL - OpenAI model name
# DEEPWIKI_EMBEDDING_MODEL - OpenAI embedding model name
# DEEPWIKI_OUTPUT_DIR - Output directory
# DEEPWIKI_FORMAT - Output format (markdown, json)
# DEEPWIKI_LANGUAGE - Output language (en, ja, zh, es, kr, vi)
# DEEPWIKI_EXCLUDE_DIRS - Additional directories to exclude (comma-separated)
# DEEPWIKI_EXCLUDE_FILES - Additional files to exclude (comma-separated)
`

	return template
}
