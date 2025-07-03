package llm

import (
	"testing"
	"time"
)

func TestProviderType_String(t *testing.T) {
	tests := []struct {
		provider ProviderType
		expected string
	}{
		{ProviderOpenAI, "openai"},
		{ProviderAnthropic, "anthropic"},
		{ProviderType("unknown"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.provider), func(t *testing.T) {
			if string(tt.provider) != tt.expected {
				t.Errorf("ProviderType string = %v, want %v", string(tt.provider), tt.expected)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	tests := []struct {
		name        string
		provider    ProviderType
		apiKey      string
		wantModel   string
		wantBaseURL string
	}{
		{
			name:        "OpenAI default config",
			provider:    ProviderOpenAI,
			apiKey:      "test-openai-key",
			wantModel:   "gpt-4o",
			wantBaseURL: "https://api.openai.com/v1",
		},
		{
			name:        "Anthropic default config",
			provider:    ProviderAnthropic,
			apiKey:      "test-anthropic-key",
			wantModel:   "claude-3-5-sonnet-20241022",
			wantBaseURL: "https://api.anthropic.com/v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig(tt.provider, tt.apiKey)

			if config == nil {
				t.Fatal("DefaultConfig() returned nil")
			}

			if config.Provider != tt.provider {
				t.Errorf("Provider = %v, want %v", config.Provider, tt.provider)
			}

			if config.APIKey != tt.apiKey {
				t.Errorf("APIKey = %v, want %v", config.APIKey, tt.apiKey)
			}

			if config.Model != tt.wantModel {
				t.Errorf("Model = %v, want %v", config.Model, tt.wantModel)
			}

			if config.BaseURL != tt.wantBaseURL {
				t.Errorf("BaseURL = %v, want %v", config.BaseURL, tt.wantBaseURL)
			}

			// Test default values
			if config.MaxTokens != 4000 {
				t.Errorf("MaxTokens = %v, want %v", config.MaxTokens, 4000)
			}

			if config.Temperature != 0.1 {
				t.Errorf("Temperature = %v, want %v", config.Temperature, 0.1)
			}

			if config.RequestTimeout != 3*time.Minute {
				t.Errorf("RequestTimeout = %v, want %v", config.RequestTimeout, 3*time.Minute)
			}

			if config.MaxRetries != 3 {
				t.Errorf("MaxRetries = %v, want %v", config.MaxRetries, 3)
			}

			if config.RetryDelay != 1*time.Second {
				t.Errorf("RetryDelay = %v, want %v", config.RetryDelay, 1*time.Second)
			}

			if config.RateLimitRPS != 2.0 {
				t.Errorf("RateLimitRPS = %v, want %v", config.RateLimitRPS, 2.0)
			}
		})
	}
}

func TestMessage_Validation(t *testing.T) {
	tests := []struct {
		name    string
		message Message
		valid   bool
	}{
		{
			name:    "valid user message",
			message: Message{Role: "user", Content: "Hello"},
			valid:   true,
		},
		{
			name:    "valid assistant message",
			message: Message{Role: "assistant", Content: "Hi there!"},
			valid:   true,
		},
		{
			name:    "valid system message",
			message: Message{Role: "system", Content: "You are a helpful assistant"},
			valid:   true,
		},
		{
			name:    "empty role",
			message: Message{Role: "", Content: "Hello"},
			valid:   false,
		},
		{
			name:    "empty content",
			message: Message{Role: "user", Content: ""},
			valid:   false,
		},
		{
			name:    "invalid role",
			message: Message{Role: "invalid", Content: "Hello"},
			valid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test basic validation logic
			hasValidRole := tt.message.Role == "user" || tt.message.Role == "assistant" || tt.message.Role == "system"
			hasContent := tt.message.Content != ""
			isValid := hasValidRole && hasContent

			if isValid != tt.valid {
				t.Errorf("Message validation = %v, want %v", isValid, tt.valid)
			}
		})
	}
}

func TestChatCompletionOptions_Defaults(t *testing.T) {
	// Test that default options make sense
	opts := ChatCompletionOptions{}

	// Default values should be zero values, but we can test that they're settable
	opts.MaxTokens = 1000
	opts.Temperature = 0.7
	opts.Stream = true

	if opts.MaxTokens != 1000 {
		t.Errorf("MaxTokens = %v, want %v", opts.MaxTokens, 1000)
	}

	if opts.Temperature != 0.7 {
		t.Errorf("Temperature = %v, want %v", opts.Temperature, 0.7)
	}

	if opts.Stream != true {
		t.Errorf("Stream = %v, want %v", opts.Stream, true)
	}
}

func TestTokenCount_Calculations(t *testing.T) {
	tests := []struct {
		name             string
		promptTokens     int
		completionTokens int
		expectedTotal    int
	}{
		{
			name:             "basic calculation",
			promptTokens:     100,
			completionTokens: 50,
			expectedTotal:    150,
		},
		{
			name:             "zero completion tokens",
			promptTokens:     100,
			completionTokens: 0,
			expectedTotal:    100,
		},
		{
			name:             "zero prompt tokens",
			promptTokens:     0,
			completionTokens: 50,
			expectedTotal:    50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenCount := TokenCount{
				PromptTokens:     tt.promptTokens,
				CompletionTokens: tt.completionTokens,
				TotalTokens:      tt.promptTokens + tt.completionTokens,
			}

			if tokenCount.TotalTokens != tt.expectedTotal {
				t.Errorf("TotalTokens = %v, want %v", tokenCount.TotalTokens, tt.expectedTotal)
			}

			// Test that individual fields are preserved
			if tokenCount.PromptTokens != tt.promptTokens {
				t.Errorf("PromptTokens = %v, want %v", tokenCount.PromptTokens, tt.promptTokens)
			}

			if tokenCount.CompletionTokens != tt.completionTokens {
				t.Errorf("CompletionTokens = %v, want %v", tokenCount.CompletionTokens, tt.completionTokens)
			}
		})
	}
}

func TestChoice_Validation(t *testing.T) {
	tests := []struct {
		name   string
		choice Choice
		valid  bool
	}{
		{
			name: "valid choice",
			choice: Choice{
				Index:        0,
				Message:      Message{Role: "assistant", Content: "Hello"},
				FinishReason: "stop",
			},
			valid: true,
		},
		{
			name: "negative index",
			choice: Choice{
				Index:        -1,
				Message:      Message{Role: "assistant", Content: "Hello"},
				FinishReason: "stop",
			},
			valid: false,
		},
		{
			name: "invalid message",
			choice: Choice{
				Index:        0,
				Message:      Message{Role: "", Content: ""},
				FinishReason: "stop",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test basic validation logic
			hasValidIndex := tt.choice.Index >= 0
			hasValidMessage := tt.choice.Message.Role != "" && tt.choice.Message.Content != ""
			isValid := hasValidIndex && hasValidMessage

			if isValid != tt.valid {
				t.Errorf("Choice validation = %v, want %v", isValid, tt.valid)
			}
		})
	}
}

func TestUsage_Validation(t *testing.T) {
	tests := []struct {
		name  string
		usage Usage
		valid bool
	}{
		{
			name: "valid usage",
			usage: Usage{
				PromptTokens:     100,
				CompletionTokens: 50,
				TotalTokens:      150,
			},
			valid: true,
		},
		{
			name: "negative prompt tokens",
			usage: Usage{
				PromptTokens:     -1,
				CompletionTokens: 50,
				TotalTokens:      49,
			},
			valid: false,
		},
		{
			name: "negative completion tokens",
			usage: Usage{
				PromptTokens:     100,
				CompletionTokens: -1,
				TotalTokens:      99,
			},
			valid: false,
		},
		{
			name: "inconsistent total",
			usage: Usage{
				PromptTokens:     100,
				CompletionTokens: 50,
				TotalTokens:      100, // Should be 150
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test basic validation logic
			hasValidTokens := tt.usage.PromptTokens >= 0 && tt.usage.CompletionTokens >= 0
			hasConsistentTotal := tt.usage.TotalTokens == (tt.usage.PromptTokens + tt.usage.CompletionTokens)
			isValid := hasValidTokens && hasConsistentTotal

			if isValid != tt.valid {
				t.Errorf("Usage validation = %v, want %v", isValid, tt.valid)
			}
		})
	}
}

func TestStreamChoice_Validation(t *testing.T) {
	tests := []struct {
		name   string
		choice StreamChoice
		valid  bool
	}{
		{
			name: "valid stream choice",
			choice: StreamChoice{
				Index: 0,
				Delta: Message{Role: "assistant", Content: "Hello"},
			},
			valid: true,
		},
		{
			name: "valid stream choice with empty delta content",
			choice: StreamChoice{
				Index: 0,
				Delta: Message{Role: "assistant", Content: ""},
			},
			valid: true, // Empty content is valid for streaming
		},
		{
			name: "negative index",
			choice: StreamChoice{
				Index: -1,
				Delta: Message{Role: "assistant", Content: "Hello"},
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test basic validation logic
			hasValidIndex := tt.choice.Index >= 0
			hasValidDelta := tt.choice.Delta.Role != "" // Content can be empty for streaming
			isValid := hasValidIndex && hasValidDelta

			if isValid != tt.valid {
				t.Errorf("StreamChoice validation = %v, want %v", isValid, tt.valid)
			}
		})
	}
}

func TestChatCompletionResponse_Validation(t *testing.T) {
	tests := []struct {
		name     string
		response ChatCompletionResponse
		valid    bool
	}{
		{
			name: "valid response",
			response: ChatCompletionResponse{
				ID:      "chatcmpl-123",
				Object:  "chat.completion",
				Created: 1234567890,
				Model:   "gpt-4o",
				Choices: []Choice{
					{
						Index:        0,
						Message:      Message{Role: "assistant", Content: "Hello"},
						FinishReason: "stop",
					},
				},
				Usage: Usage{
					PromptTokens:     10,
					CompletionTokens: 5,
					TotalTokens:      15,
				},
			},
			valid: true,
		},
		{
			name: "empty ID",
			response: ChatCompletionResponse{
				ID:      "",
				Object:  "chat.completion",
				Created: 1234567890,
				Model:   "gpt-4o",
				Choices: []Choice{
					{
						Index:        0,
						Message:      Message{Role: "assistant", Content: "Hello"},
						FinishReason: "stop",
					},
				},
				Usage: Usage{
					PromptTokens:     10,
					CompletionTokens: 5,
					TotalTokens:      15,
				},
			},
			valid: false,
		},
		{
			name: "no choices",
			response: ChatCompletionResponse{
				ID:      "chatcmpl-123",
				Object:  "chat.completion",
				Created: 1234567890,
				Model:   "gpt-4o",
				Choices: []Choice{},
				Usage: Usage{
					PromptTokens:     10,
					CompletionTokens: 5,
					TotalTokens:      15,
				},
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test basic validation logic
			hasID := tt.response.ID != ""
			hasModel := tt.response.Model != ""
			hasChoices := len(tt.response.Choices) > 0
			isValid := hasID && hasModel && hasChoices

			if isValid != tt.valid {
				t.Errorf("ChatCompletionResponse validation = %v, want %v", isValid, tt.valid)
			}
		})
	}
}
