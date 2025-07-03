package embedding

import (
	"fmt"
)

// GetSupportedProviders returns a list of supported embedding provider types
func GetSupportedProviders() []ProviderType {
	return []ProviderType{
		ProviderOpenAI,
		ProviderVoyage,
		ProviderOllama,
	}
}

// ValidateConfig validates the embedding provider configuration
func ValidateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.Model == "" {
		return fmt.Errorf("model is required")
	}

	switch config.Provider {
	case ProviderOpenAI, ProviderVoyage:
		if config.APIKey == "" {
			return fmt.Errorf("API key is required for provider %s", config.Provider)
		}
	case ProviderOllama:
		// Ollama doesn't require API key
		if config.BaseURL == "" {
			return fmt.Errorf("base_url is required for Ollama provider")
		}
	case "":
		return fmt.Errorf("provider type is required")
	default:
		return fmt.Errorf("unsupported provider: %s", config.Provider)
	}

	if config.MaxRetries < 0 {
		return fmt.Errorf("max_retries must be non-negative")
	}

	if config.RateLimitRPS <= 0 {
		return fmt.Errorf("rate_limit_rps must be positive")
	}

	return nil
}
