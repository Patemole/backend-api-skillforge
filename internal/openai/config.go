package openai

// EmailConfig contient la configuration pour le générateur d'emails de présentation
type EmailConfig struct {
	Model               string  `json:"model"`
	MaxTokens           int     `json:"max_tokens,omitempty"`            // Pour les anciens modèles
	MaxCompletionTokens int     `json:"max_completion_tokens,omitempty"` // Pour GPT-5 et modèles récents
	Temperature         float64 `json:"temperature"`
	TopP                float64 `json:"top_p"`
	FrequencyPenalty    float64 `json:"frequency_penalty"`
	PresencePenalty     float64 `json:"presence_penalty"`
}

// GetEmailConfig retourne la configuration pour le générateur d'emails
func GetEmailConfig() EmailConfig {
	return EmailConfig{
		Model:               "gpt-4o-2024-08-06", // Modèle spécifique pour les emails
		MaxTokens:           4000,                // Limite de tokens pour les emails
		MaxCompletionTokens: 0,                   // Non utilisé avec gpt-4o
		Temperature:         0.3,                 // Température basse pour plus de cohérence
		TopP:                1.0,
		FrequencyPenalty:    0,
		PresencePenalty:     0,
	}
}
