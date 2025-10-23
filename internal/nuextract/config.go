package nuextract

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

// GetOpenAIConfig retourne la configuration optimisée pour la vitesse
func GetOpenAIConfig() OpenAIConfig {
	return OpenAIConfig{
		Model:               "gpt-5", // Modèle GPT-5
		MaxCompletionTokens: 13000,   // Nouveau paramètre pour GPT-5
		// Temperature et TopP non supportés par GPT-5 (utilise les valeurs par défaut)
		FrequencyPenalty: 0,
		PresencePenalty:  0,
	}
}
