package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"backend-api-skillforge/internal/models"
	"backend-api-skillforge/internal/nuextract"

	"github.com/gin-gonic/gin"
)

// GeneratePresentationEmail génère un email de présentation de candidat avec OpenAI
func GeneratePresentationEmail(c *gin.Context) {
	log.Printf("[EMAIL_GENERATION] Début de la génération d'email")
	
	// Lire le body brut pour debugging
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("[EMAIL_GENERATION] ❌ Erreur lecture body: %v", err)
		c.JSON(http.StatusBadRequest, models.GenerateEmailResponse{
			EmailContent: "",
			Success:      false,
			Error:        "Failed to read request body: " + err.Error(),
		})
		return
	}
	log.Printf("[EMAIL_GENERATION] 📄 Body brut reçu (premiers 500 chars): %.500s...", string(bodyBytes))
	
	// Recréer le reader pour le parsing
	c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	
	var req models.GenerateEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[EMAIL_GENERATION] ❌ Erreur de parsing JSON: %v", err)
		log.Printf("[EMAIL_GENERATION] 📄 Body qui a causé l'erreur: %s", string(bodyBytes))
		c.JSON(http.StatusBadRequest, models.GenerateEmailResponse{
			EmailContent: "",
			Success:      false,
			Error:        "Invalid request payload: " + err.Error(),
		})
		return
	}

	log.Printf("[EMAIL_GENERATION] Requête reçue - TemplateID: %v, Template: %v", req.TemplateID, req.Template != nil)
	log.Printf("[EMAIL_GENERATION] 🔍 Détails Template - TemplateID: %+v", req.TemplateID)
	log.Printf("[EMAIL_GENERATION] 🔍 Détails Template - Template: %+v", req.Template)
	if req.Template != nil {
		log.Printf("[EMAIL_GENERATION] 🔍 Template détails - ID: %s, Name: %s, Subject: %s", req.Template.ID, req.Template.Name, req.Template.Subject)
	}

	// Convertir les données du candidat en JSON pour le prompt
	candidateDataJSON, err := json.MarshalIndent(req.CandidateData, "", "  ")
	if err != nil {
		log.Printf("[EMAIL_GENERATION] Erreur de sérialisation des données candidat: %v", err)
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

	log.Printf("[EMAIL_GENERATION] Données candidat sérialisées, besoin: %s", need)

	var emailContent string
	var generationErr error

	// Vérifier si un template est fourni
	if req.Template != nil {
		log.Printf("[EMAIL_GENERATION] ✅ Template détecté - Nom: %s, ID: %s", req.Template.Name, req.Template.ID)
		log.Printf("[EMAIL_GENERATION] 📝 Template Subject: %s", req.Template.Subject)
		log.Printf("[EMAIL_GENERATION] 📄 Template Content (premiers 200 chars): %.200s...", req.Template.Content)
		log.Printf("[EMAIL_GENERATION] 🔄 Démarrage génération avec template")
		emailContent, generationErr = generateEmailWithTemplate(string(candidateDataJSON), need, req.Template)
	} else {
		log.Printf("[EMAIL_GENERATION] ⚠️  Aucun template fourni - Mode classique activé")
		log.Printf("[EMAIL_GENERATION] 🔄 Démarrage génération classique")
		emailContent, generationErr = generateEmailWithOpenAI(string(candidateDataJSON), need)
	}

	if generationErr != nil {
		log.Printf("[EMAIL_GENERATION] ❌ Erreur de génération: %v", generationErr)
		c.JSON(http.StatusInternalServerError, models.GenerateEmailResponse{
			EmailContent: "",
			Success:      false,
			Error:        "Failed to generate email: " + generationErr.Error(),
		})
		return
	}

	log.Printf("[EMAIL_GENERATION] ✅ Email généré avec succès")
	log.Printf("[EMAIL_GENERATION] 📊 Taille de l'email généré: %d caractères", len(emailContent))
	log.Printf("[EMAIL_GENERATION] 📧 Aperçu de l'email (premiers 300 chars): %.300s...", emailContent)
	c.JSON(http.StatusOK, models.GenerateEmailResponse{
		EmailContent: emailContent,
		Success:      true,
	})
}

// generateEmailWithTemplate génère un email en utilisant un template spécifique
func generateEmailWithTemplate(candidateData, need string, template *models.EmailTemplate) (string, error) {
	log.Printf("[TEMPLATE_GENERATION] 🚀 Début génération avec template: %s", template.Name)
	log.Printf("[TEMPLATE_GENERATION] 📋 Template ID: %s", template.ID)
	log.Printf("[TEMPLATE_GENERATION] 📝 Template Subject: %s", template.Subject)
	log.Printf("[TEMPLATE_GENERATION] 📄 Template Content (premiers 200 chars): %.200s...", template.Content)
	log.Printf("[TEMPLATE_GENERATION] 🔧 Template IsDefault: %v", template.IsDefault)
	
	// Récupérer la configuration OpenAI existante
	config := nuextract.GetOpenAIConfig()
	log.Printf("[TEMPLATE_GENERATION] ⚙️  Configuration OpenAI récupérée - Model: %s", config.Model)

	// Construire le prompt pour l'email avec template
	log.Printf("[TEMPLATE_GENERATION] 🔨 Construction du prompt avec template...")
	prompt := nuextract.GetEmailPromptWithTemplate(candidateData, need, template.Content, template.Subject)

	log.Printf("[TEMPLATE_GENERATION] ✅ Prompt généré, longueur: %d caractères", len(prompt))
	log.Printf("[TEMPLATE_GENERATION] 📝 Aperçu du prompt (premiers 300 chars): %.300s...", prompt)

	// Préparer la requête OpenAI
	log.Printf("[TEMPLATE_GENERATION] 🔧 Préparation du payload OpenAI...")
	payload := map[string]interface{}{
		"model": config.Model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "Tu es un expert en recrutement et en rédaction d'emails professionnels. Tu génères des emails en respectant exactement les templates fournis.",
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

	log.Printf("[TEMPLATE_GENERATION] 📊 Paramètres OpenAI - MaxTokens: %d, Temperature: %.2f, TopP: %.2f", 
		config.MaxTokens, config.Temperature, config.TopP)

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[TEMPLATE_GENERATION] ❌ Erreur sérialisation payload: %v", err)
		return "", fmt.Errorf("erreur lors de la sérialisation de la requête: %v", err)
	}
	log.Printf("[TEMPLATE_GENERATION] ✅ Payload sérialisé, taille: %d bytes", len(bodyBytes))

	// Faire l'appel à l'API OpenAI
	log.Printf("[TEMPLATE_GENERATION] 🌐 Création de la requête HTTP vers OpenAI...")
	req, err := http.NewRequest(http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("[TEMPLATE_GENERATION] ❌ Erreur création requête: %v", err)
		return "", fmt.Errorf("erreur lors de la création de la requête: %v", err)
	}

	// Utiliser la même clé API que le client NuExtract
	openAIAPIKey := os.Getenv("OPENAI_API_KEY")
	if openAIAPIKey == "" {
		log.Printf("[TEMPLATE_GENERATION] ❌ Erreur: OPENAI_API_KEY non définie")
		return "", fmt.Errorf("OPENAI_API_KEY not set")
	}
	log.Printf("[TEMPLATE_GENERATION] 🔑 Clé API OpenAI trouvée (longueur: %d)", len(openAIAPIKey))

	req.Header.Set("Authorization", "Bearer "+openAIAPIKey)
	req.Header.Set("Content-Type", "application/json")

	log.Printf("[TEMPLATE_GENERATION] 🚀 Envoi de la requête vers OpenAI...")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[TEMPLATE_GENERATION] ❌ Erreur envoi requête: %v", err)
		return "", fmt.Errorf("erreur lors de l'envoi de la requête: %v", err)
	}
	defer resp.Body.Close()
	log.Printf("[TEMPLATE_GENERATION] 📡 Réponse reçue - Status: %d", resp.StatusCode)

	log.Printf("[TEMPLATE_GENERATION] 📖 Lecture de la réponse OpenAI...")
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[TEMPLATE_GENERATION] ❌ Erreur lecture réponse: %v", err)
		return "", fmt.Errorf("erreur lors de la lecture de la réponse: %v", err)
	}
	log.Printf("[TEMPLATE_GENERATION] ✅ Réponse lue, taille: %d bytes", len(respBytes))

	if resp.StatusCode >= 400 {
		log.Printf("[TEMPLATE_GENERATION] ❌ Erreur OpenAI %d: %s", resp.StatusCode, string(respBytes))
		return "", fmt.Errorf("openai error %d: %s", resp.StatusCode, respBytes)
	}
	log.Printf("[TEMPLATE_GENERATION] ✅ Status code OK: %d", resp.StatusCode)

	// Parser la réponse OpenAI
	log.Printf("[TEMPLATE_GENERATION] 🔍 Parsing de la réponse JSON...")
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
		log.Printf("[TEMPLATE_GENERATION] ❌ Erreur parsing réponse: %v", err)
		log.Printf("[TEMPLATE_GENERATION] 📄 Réponse brute: %s", string(respBytes))
		return "", fmt.Errorf("erreur lors du parsing de la réponse: %v", err)
	}
	log.Printf("[TEMPLATE_GENERATION] ✅ Réponse JSON parsée avec succès")

	if openAIResp.Error != nil {
		log.Printf("[TEMPLATE_GENERATION] ❌ Erreur OpenAI: %s", openAIResp.Error.Message)
		return "", fmt.Errorf("erreur OpenAI: %s", openAIResp.Error.Message)
	}

	if len(openAIResp.Choices) == 0 {
		log.Printf("[TEMPLATE_GENERATION] ❌ Aucune réponse générée")
		return "", fmt.Errorf("aucune réponse générée par OpenAI")
	}
	log.Printf("[TEMPLATE_GENERATION] ✅ %d choix reçus de OpenAI", len(openAIResp.Choices))

	emailContent := openAIResp.Choices[0].Message.Content
	log.Printf("[TEMPLATE_GENERATION] 🎉 Email généré avec succès!")
	log.Printf("[TEMPLATE_GENERATION] 📊 Longueur de l'email: %d caractères", len(emailContent))
	log.Printf("[TEMPLATE_GENERATION] 📧 Aperçu de l'email (premiers 300 chars): %.300s...", emailContent)
	
	return emailContent, nil
}

// generateEmailWithOpenAI génère un email en utilisant l'API OpenAI (mode classique)
func generateEmailWithOpenAI(candidateData, need string) (string, error) {
	log.Printf("[CLASSIC_GENERATION] Début génération classique")
	
	// Récupérer la configuration OpenAI existante
	config := nuextract.GetOpenAIConfig()

	// Construire le prompt pour l'email
	prompt := nuextract.GetEmailPrompt(candidateData, need)

	log.Printf("[CLASSIC_GENERATION] Prompt généré, longueur: %d caractères", len(prompt))

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
		log.Printf("[CLASSIC_GENERATION] Erreur sérialisation payload: %v", err)
		return "", fmt.Errorf("erreur lors de la sérialisation de la requête: %v", err)
	}

	// Faire l'appel à l'API OpenAI
	req, err := http.NewRequest(http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("[CLASSIC_GENERATION] Erreur création requête: %v", err)
		return "", fmt.Errorf("erreur lors de la création de la requête: %v", err)
	}

	// Utiliser la même clé API que le client NuExtract
	openAIAPIKey := os.Getenv("OPENAI_API_KEY")
	if openAIAPIKey == "" {
		log.Printf("[CLASSIC_GENERATION] Erreur: OPENAI_API_KEY non définie")
		return "", fmt.Errorf("OPENAI_API_KEY not set")
	}

	req.Header.Set("Authorization", "Bearer "+openAIAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[CLASSIC_GENERATION] Erreur envoi requête: %v", err)
		return "", fmt.Errorf("erreur lors de l'envoi de la requête: %v", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[CLASSIC_GENERATION] Erreur lecture réponse: %v", err)
		return "", fmt.Errorf("erreur lors de la lecture de la réponse: %v", err)
	}

	if resp.StatusCode >= 400 {
		log.Printf("[CLASSIC_GENERATION] Erreur OpenAI %d: %s", resp.StatusCode, string(respBytes))
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
		log.Printf("[CLASSIC_GENERATION] Erreur parsing réponse: %v", err)
		return "", fmt.Errorf("erreur lors du parsing de la réponse: %v", err)
	}

	if openAIResp.Error != nil {
		log.Printf("[CLASSIC_GENERATION] Erreur OpenAI: %s", openAIResp.Error.Message)
		return "", fmt.Errorf("erreur OpenAI: %s", openAIResp.Error.Message)
	}

	if len(openAIResp.Choices) == 0 {
		log.Printf("[CLASSIC_GENERATION] Aucune réponse générée")
		return "", fmt.Errorf("aucune réponse générée par OpenAI")
	}

	log.Printf("[CLASSIC_GENERATION] Email généré avec succès, longueur: %d caractères", len(openAIResp.Choices[0].Message.Content))
	return openAIResp.Choices[0].Message.Content, nil
}
