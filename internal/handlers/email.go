package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"backend-api-skillforge/internal/models"
	"backend-api-skillforge/internal/nuextract"

	"github.com/gin-gonic/gin"
)

// GeneratePresentationEmail génère un email de présentation de candidat avec OpenAI
func GeneratePresentationEmail(c *gin.Context) {
	var req models.GenerateEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.GenerateEmailResponse{
			EmailContent: "",
			Success:      false,
			Error:        "Invalid request payload: " + err.Error(),
		})
		return
	}

	// Convertir les données du candidat en JSON pour le prompt
	candidateDataJSON, err := json.MarshalIndent(req.CandidateData, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.GenerateEmailResponse{
			EmailContent: "",
			Success:      false,
			Error:        "Failed to process candidate data: " + err.Error(),
		})
		return
	}

	// Préparer le besoin (peut être nil)
	need := ""
	if req.Need != nil {
		need = *req.Need
	}

	// Générer l'email avec OpenAI en utilisant le client existant
	emailContent, err := generateEmailWithOpenAI(string(candidateDataJSON), need)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.GenerateEmailResponse{
			EmailContent: "",
			Success:      false,
			Error:        "Failed to generate email with OpenAI: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.GenerateEmailResponse{
		EmailContent: emailContent,
		Success:      true,
	})
}

// generateEmailWithOpenAI génère un email en utilisant l'API OpenAI
func generateEmailWithOpenAI(candidateData, need string) (string, error) {
	// Récupérer la configuration OpenAI existante
	config := nuextract.GetOpenAIConfig()

	// Construire le prompt pour l'email
	prompt := nuextract.GetEmailPrompt(candidateData, need)

	// Préparer la requête OpenAI (même structure que dans nuextract/client.go)
	payload := map[string]interface{}{
		"model": config.Model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "Tu es un expert en recrutement et en rédaction d'emails professionnels. Tu écris des emails de présentation de candidats pour les entreprises.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens":        config.MaxTokens,
		"temperature":       config.Temperature,
		"top_p":             config.TopP,
		"frequency_penalty": config.FrequencyPenalty,
		"presence_penalty":  config.PresencePenalty,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("erreur lors de la sérialisation de la requête: %v", err)
	}

	// Faire l'appel à l'API OpenAI
	req, err := http.NewRequest(http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("erreur lors de la création de la requête: %v", err)
	}

	// Utiliser la même clé API que le client NuExtract
	openAIAPIKey := os.Getenv("OPENAI_API_KEY")
	if openAIAPIKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY not set")
	}

	req.Header.Set("Authorization", "Bearer "+openAIAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("erreur lors de l'envoi de la requête: %v", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("erreur lors de la lecture de la réponse: %v", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("openai error %d: %s", resp.StatusCode, respBytes)
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
		return "", fmt.Errorf("erreur lors du parsing de la réponse: %v", err)
	}

	if openAIResp.Error != nil {
		return "", fmt.Errorf("erreur OpenAI: %s", openAIResp.Error.Message)
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("aucune réponse générée par OpenAI")
	}

	return openAIResp.Choices[0].Message.Content, nil
}
