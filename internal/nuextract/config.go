package nuextract

// OpenAIConfig contient la configuration pour les appels OpenAI
type OpenAIConfig struct {
	Model            string  `json:"model"`
	MaxTokens        int     `json:"max_tokens"`
	Temperature      float64 `json:"temperature"`
	TopP             float64 `json:"top_p"`
	FrequencyPenalty float64 `json:"frequency_penalty"`
	PresencePenalty  float64 `json:"presence_penalty"`
}

// GetOpenAIConfig retourne la configuration optimisée pour la vitesse
func GetOpenAIConfig() OpenAIConfig {
	return OpenAIConfig{
		Model:            "gpt-4o",       // Plus flexible que gpt-4-turbo
		MaxTokens:        8000,           // Limite GPT-4o: plus élevée
		Temperature:      0.1,           // Faible pour plus de cohérence et de vitesse
		TopP:             0.9,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
	}
}
