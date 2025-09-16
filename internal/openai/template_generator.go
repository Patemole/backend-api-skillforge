package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"backend-api-skillforge/internal/models"
	"backend-api-skillforge/internal/nuextract"
)

// TemplateGeneratorService gère la génération de templates d'emails avec OpenAI
type TemplateGeneratorService struct {
	apiKey string
	config nuextract.OpenAIConfig
}

// NewTemplateGeneratorService crée une nouvelle instance du service
func NewTemplateGeneratorService() *TemplateGeneratorService {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		panic("OPENAI_API_KEY not set")
	}

	config := nuextract.GetOpenAIConfig()
	return &TemplateGeneratorService{
		apiKey: apiKey,
		config: config,
	}
}

// GenerateTemplate génère un template d'email basé sur les exemples fournis
func (s *TemplateGeneratorService) GenerateTemplate(req models.TemplateGenerateRequest) (*models.TemplateGenerateResponse, error) {
	// Construire le prompt pour la génération de template
	prompt := s.buildTemplatePrompt(req)

	// Préparer la requête OpenAI avec JSON mode
	payload := map[string]interface{}{
		"model":       "gpt-4o-2024-08-06", // Utiliser le modèle le plus récent
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "Tu es un expert en génération de templates d'emails professionnels. Tu analyses des exemples d'emails et génères un template réutilisable en identifiant les patterns et en mappant les variables appropriées.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens":        s.config.MaxTokens,
		"temperature":       0.3, // Température plus basse pour plus de cohérence
		"top_p":             s.config.TopP,
		"frequency_penalty": s.config.FrequencyPenalty,
		"presence_penalty":  s.config.PresencePenalty,
		"response_format": map[string]string{
			"type": "json_object",
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la sérialisation de la requête: %v", err)
	}

	// Faire l'appel à l'API OpenAI
	httpReq, err := http.NewRequest(http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création de la requête: %v", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de l'envoi de la requête: %v", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la lecture de la réponse: %v", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("openai error %d: %s", resp.StatusCode, respBytes)
	}

	// Parser la réponse OpenAI
	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}

	if err := json.Unmarshal(respBytes, &openAIResp); err != nil {
		return nil, fmt.Errorf("erreur lors du parsing de la réponse: %v", err)
	}

	if openAIResp.Error != nil {
		return nil, fmt.Errorf("erreur OpenAI: %s", openAIResp.Error.Message)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("aucune réponse générée par OpenAI")
	}

	// Parser la réponse JSON générée par OpenAI
	var templateResp models.TemplateGenerateResponse
	if err := json.Unmarshal([]byte(openAIResp.Choices[0].Message.Content), &templateResp); err != nil {
		return nil, fmt.Errorf("erreur lors du parsing de la réponse template: %v", err)
	}

	// Générer un ID de template unique
	templateResp.TemplateID = fmt.Sprintf("template-%d", time.Now().UnixMilli())
	templateResp.Success = true

	return &templateResp, nil
}

// buildTemplatePrompt construit le prompt pour la génération de template
func (s *TemplateGeneratorService) buildTemplatePrompt(req models.TemplateGenerateRequest) string {
	// Convertir les exemples en JSON pour le prompt
	examplesJSON, _ := json.MarshalIndent(req.EmailExamples, "", "  ")
	variablesJSON, _ := json.MarshalIndent(req.AvailableVariables, "", "  ")

	return fmt.Sprintf(`Tu dois analyser les exemples d'emails suivants et générer un template réutilisable.

**Contexte :**
- Nom du template : %s
- Organisation ID : %s

**Exemples d'emails à analyser :**
%s

**Variables disponibles :**
%s

**Instructions :**
1. Analyse les patterns communs dans les exemples d'emails
2. Identifie les parties variables qui correspondent aux variables disponibles
3. Génère un template avec les variables appropriées mappées
4. Assure-toi que le template est cohérent et professionnel
5. Utilise uniquement les variables de la liste "available_variables"
6. Le template doit être générique mais maintenir la structure et le ton des exemples

**Format de réponse attendu (JSON) :**
{
  "success": true,
  "generated_template": {
    "subject": "Template du sujet avec {{variable}}",
    "content": "Template du contenu avec {{variable}}"
  },
  "detected_variables": ["{{variable1}}", "{{variable2}}", ...],
  "confidence_score": 0.95
}

**Important :**
- Utilise uniquement les variables de la liste "available_variables"
- Mappe intelligemment les variables selon le contexte
- Le confidence_score doit être entre 0.0 et 1.0
- Assure-toi que le template est cohérent avec les exemples fournis`,
		req.TemplateName,
		req.OrganizationID,
		string(examplesJSON),
		string(variablesJSON),
	)
}
