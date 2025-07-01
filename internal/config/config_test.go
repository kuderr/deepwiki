package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	// Test OpenAI defaults
	if config.OpenAI.Model != "gpt-4o" {
		t.Errorf("Expected default model 'gpt-4o', got '%s'", config.OpenAI.Model)
	}

	if config.OpenAI.EmbeddingModel != "text-embedding-3-small" {
		t.Errorf("Expected default embedding model 'text-embedding-3-small', got '%s'", config.OpenAI.EmbeddingModel)
	}

	if config.OpenAI.MaxTokens != 4000 {
		t.Errorf("Expected default max tokens 4000, got %d", config.OpenAI.MaxTokens)
	}

	// Test Processing defaults
	if config.Processing.ChunkSize != 350 {
		t.Errorf("Expected default chunk size 350, got %d", config.Processing.ChunkSize)
	}

	if config.Processing.ChunkOverlap != 100 {
		t.Errorf("Expected default chunk overlap 100, got %d", config.Processing.ChunkOverlap)
	}

	// Test Output defaults
	if config.Output.Format != "markdown" {
		t.Errorf("Expected default format 'markdown', got '%s'", config.Output.Format)
	}

	if config.Output.Language != "en" {
		t.Errorf("Expected default language 'en', got '%s'", config.Output.Language)
	}

	// Test Filters defaults
	if len(config.Filters.IncludeExtensions) == 0 {
		t.Error("Expected some default include extensions")
	}

	if len(config.Filters.ExcludeDirs) == 0 {
		t.Error("Expected some default exclude directories")
	}

	// Test Embeddings defaults
	if !config.Embeddings.Enabled {
		t.Error("Expected embeddings to be enabled by default")
	}

	if config.Embeddings.TopK != 20 {
		t.Errorf("Expected default top_k 20, got %d", config.Embeddings.TopK)
	}

	// Test Logging defaults
	if config.Logging.Level != "info" {
		t.Errorf("Expected default log level 'info', got '%s'", config.Logging.Level)
	}
}

func TestLoadConfig_NoFile(t *testing.T) {
	// Test loading config with no file (should use defaults)
	config, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if config == nil {
		t.Fatal("LoadConfig returned nil config")
	}

	// Should have default values
	if config.OpenAI.Model != "gpt-4o" {
		t.Errorf("Expected default model, got '%s'", config.OpenAI.Model)
	}
}

func TestLoadConfig_FromFile(t *testing.T) {
	// Create temporary config file
	tempDir, err := os.MkdirTemp("", "deepwiki-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, "test-config.yaml")
	configContent := `
openai:
  model: "gpt-3.5-turbo"
  max_tokens: 2000
  temperature: 0.5

processing:
  chunk_size: 500
  chunk_overlap: 50

output:
  format: "json"
  language: "ja"

logging:
  level: "debug"
  format: "json"
`

	if err := os.WriteFile(configFile, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Load config from file
	config, err := LoadConfig(configFile)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify values from file
	if config.OpenAI.Model != "gpt-3.5-turbo" {
		t.Errorf("Expected model 'gpt-3.5-turbo', got '%s'", config.OpenAI.Model)
	}

	if config.OpenAI.MaxTokens != 2000 {
		t.Errorf("Expected max tokens 2000, got %d", config.OpenAI.MaxTokens)
	}

	if config.Processing.ChunkSize != 500 {
		t.Errorf("Expected chunk size 500, got %d", config.Processing.ChunkSize)
	}

	if config.Output.Format != "json" {
		t.Errorf("Expected format 'json', got '%s'", config.Output.Format)
	}

	if config.Output.Language != "ja" {
		t.Errorf("Expected language 'ja', got '%s'", config.Output.Language)
	}

	if config.Logging.Level != "debug" {
		t.Errorf("Expected log level 'debug', got '%s'", config.Logging.Level)
	}
}

func TestLoadConfig_EnvironmentVariables(t *testing.T) {
	// Set environment variables
	originalAPIKey := os.Getenv("OPENAI_API_KEY")
	originalModel := os.Getenv("DEEPWIKI_MODEL")
	originalFormat := os.Getenv("DEEPWIKI_FORMAT")

	os.Setenv("OPENAI_API_KEY", "test-api-key")
	os.Setenv("DEEPWIKI_MODEL", "gpt-4-turbo")
	os.Setenv("DEEPWIKI_FORMAT", "json")

	// Restore original values after test
	defer func() {
		if originalAPIKey != "" {
			os.Setenv("OPENAI_API_KEY", originalAPIKey)
		} else {
			os.Unsetenv("OPENAI_API_KEY")
		}
		if originalModel != "" {
			os.Setenv("DEEPWIKI_MODEL", originalModel)
		} else {
			os.Unsetenv("DEEPWIKI_MODEL")
		}
		if originalFormat != "" {
			os.Setenv("DEEPWIKI_FORMAT", originalFormat)
		} else {
			os.Unsetenv("DEEPWIKI_FORMAT")
		}
	}()

	// Load config (should pick up env vars)
	config, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify environment variables were loaded
	if config.OpenAI.APIKey != "test-api-key" {
		t.Errorf("Expected API key 'test-api-key', got '%s'", config.OpenAI.APIKey)
	}

	if config.OpenAI.Model != "gpt-4-turbo" {
		t.Errorf("Expected model 'gpt-4-turbo', got '%s'", config.OpenAI.Model)
	}

	if config.Output.Format != "json" {
		t.Errorf("Expected format 'json', got '%s'", config.Output.Format)
	}
}

func TestValidateConfig_Valid(t *testing.T) {
	config := DefaultConfig()

	err := validateConfig(config)
	if err != nil {
		t.Errorf("validateConfig failed for default config: %v", err)
	}
}

func TestValidateConfig_InvalidModel(t *testing.T) {
	config := DefaultConfig()
	config.OpenAI.Model = ""

	err := validateConfig(config)
	if err == nil {
		t.Error("Expected validation error for empty model")
	}
}

func TestValidateConfig_InvalidTemperature(t *testing.T) {
	config := DefaultConfig()
	config.OpenAI.Temperature = 3.0 // Too high

	err := validateConfig(config)
	if err == nil {
		t.Error("Expected validation error for temperature > 2")
	}
}

func TestValidateConfig_InvalidChunkSize(t *testing.T) {
	config := DefaultConfig()
	config.Processing.ChunkSize = 0

	err := validateConfig(config)
	if err == nil {
		t.Error("Expected validation error for chunk size <= 0")
	}
}

func TestValidateConfig_InvalidChunkOverlap(t *testing.T) {
	config := DefaultConfig()
	config.Processing.ChunkOverlap = 400 // Greater than chunk size

	err := validateConfig(config)
	if err == nil {
		t.Error("Expected validation error for chunk overlap >= chunk size")
	}
}

func TestValidateConfig_InvalidFormat(t *testing.T) {
	config := DefaultConfig()
	config.Output.Format = "invalid"

	err := validateConfig(config)
	if err == nil {
		t.Error("Expected validation error for invalid format")
	}
}

func TestValidateConfig_InvalidLanguage(t *testing.T) {
	config := DefaultConfig()
	config.Output.Language = "invalid"

	err := validateConfig(config)
	if err == nil {
		t.Error("Expected validation error for invalid language")
	}
}

func TestSaveConfig(t *testing.T) {
	config := DefaultConfig()
	config.OpenAI.Model = "test-model"

	// Create temporary file
	tempDir, err := os.MkdirTemp("", "deepwiki-save-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, "save-test.yaml")

	// Save config
	err = SaveConfig(config, configFile)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Load it back and verify
	loadedConfig, err := LoadConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loadedConfig.OpenAI.Model != "test-model" {
		t.Errorf("Expected model 'test-model', got '%s'", loadedConfig.OpenAI.Model)
	}
}

func TestGenerateTemplate(t *testing.T) {
	template := GenerateTemplate()

	if template == "" {
		t.Error("GenerateTemplate returned empty string")
	}

	// Should contain YAML content
	if !contains(template, "openai:") {
		t.Error("Template should contain 'openai:' section")
	}

	if !contains(template, "processing:") {
		t.Error("Template should contain 'processing:' section")
	}

	if !contains(template, "output:") {
		t.Error("Template should contain 'output:' section")
	}

	// Should contain comments
	if !contains(template, "# DeepWiki CLI Configuration File") {
		t.Error("Template should contain header comment")
	}

	if !contains(template, "OPENAI_API_KEY") {
		t.Error("Template should contain environment variable documentation")
	}
}

func TestLoadConfig_NonExistentFile(t *testing.T) {
	_, err := LoadConfig("/non/existent/file.yaml")

	if err == nil {
		t.Error("Expected error for non-existent config file")
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	// Create temporary file with invalid YAML
	tempDir, err := os.MkdirTemp("", "deepwiki-invalid-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, "invalid.yaml")
	invalidYAML := `
openai:
  model: "test"
  invalid yaml syntax here {{{ 
`

	if err := os.WriteFile(configFile, []byte(invalidYAML), 0o644); err != nil {
		t.Fatalf("Failed to write invalid config file: %v", err)
	}

	_, err = LoadConfig(configFile)
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsInMiddle(s, substr)))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
