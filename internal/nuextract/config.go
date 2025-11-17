package nuextract

import (
	"os"
	"strconv"
	"strings"
)

// OpenAIConfig contient la configuration pour les appels OpenAI
type OpenAIConfig struct {
	Model               string  `json:"model"`
	MaxTokens           int     `json:"max_tokens,omitempty"`            // Pour les anciens modèles
	MaxCompletionTokens int     `json:"max_completion_tokens,omitempty"` // Pour GPT-5 et modèles récents
	Temperature         float64 `json:"temperature"`
	TopP                float64 `json:"top_p"`
	FrequencyPenalty    float64 `json:"frequency_penalty"`
	PresencePenalty     float64 `json:"presence_penalty"`
}

// AnthropicConfig contient la configuration pour Anthropic
type AnthropicConfig struct {
	Model     string `json:"model"`
	MaxTokens int    `json:"max_tokens"`
}

// GetOpenAIConfig retourne la configuration OpenAI (avec overrides via env)
func GetOpenAIConfig() OpenAIConfig {
	cfg := OpenAIConfig{
		Model:               "gpt-5",
		MaxCompletionTokens: 13000,
		Temperature:         1,
		TopP:                1,
		FrequencyPenalty:    0,
		PresencePenalty:     0,
	}
	if m := strings.TrimSpace(os.Getenv("OPENAI_MODEL")); m != "" {
		cfg.Model = m
	}
	if v := strings.TrimSpace(os.Getenv("OPENAI_MAX_COMPLETION_TOKENS")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.MaxCompletionTokens = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("OPENAI_TEMPERATURE")); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.Temperature = f
		}
	}
	if v := strings.TrimSpace(os.Getenv("OPENAI_TOP_P")); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.TopP = f
		}
	}
	return cfg
}

// GetAnthropicConfig retourne la configuration Anthropic (avec overrides via env)
func GetAnthropicConfig() AnthropicConfig {
	cfg := AnthropicConfig{
		Model:     "claude-sonnet-4-5",
		MaxTokens: 13000,
	}
	if m := strings.TrimSpace(os.Getenv("ANTHROPIC_MODEL")); m != "" {
		cfg.Model = m
	}
	if v := strings.TrimSpace(os.Getenv("ANTHROPIC_MAX_TOKENS")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.MaxTokens = n
		}
	}
	return cfg
}
