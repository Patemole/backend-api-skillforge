package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"backend-api-skillforge/internal/models"
	"backend-api-skillforge/internal/nuextract"
)

// min retourne le minimum entre deux entiers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// DossierVersionningService gère l'adaptation d'un dossier de compétences
// en fonction d'un besoin client, sans inventer d'informations.
type DossierVersionningService struct {
	apiKey          string
	config          nuextract.OpenAIConfig
	anthropicConfig nuextract.AnthropicConfig
	useAnthropic    bool
}

// NewDossierVersionningService crée une nouvelle instance du service.
func NewDossierVersionningService() *DossierVersionningService {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		panic("OPENAI_API_KEY not set")
	}
	cfg := nuextract.GetOpenAIConfig()
	return &DossierVersionningService{apiKey: apiKey, config: cfg, useAnthropic: false}
}

// NewDossierVersionningServiceAnthropic crée une nouvelle instance du service utilisant Anthropic.
func NewDossierVersionningServiceAnthropic() *DossierVersionningService {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		panic("ANTHROPIC_API_KEY not set")
	}
	aCfg := nuextract.GetAnthropicConfig(false)
	return &DossierVersionningService{apiKey: apiKey, anthropicConfig: aCfg, useAnthropic: true}
}

// GenerateVersionnedDossier appelle OpenAI ou Anthropic pour produire un nouveau dossier de compétences
// en conservant strictement le même schéma que models.CompetenceDossier.
func (s *DossierVersionningService) GenerateVersionnedDossier(candidateID string, competenceDossier models.CompetenceDossier, need *string, language string) (*models.CompetenceDossier, error) {
	if s.useAnthropic {
		return s.generateVersionnedDossierAnthropic(candidateID, competenceDossier, need, language)
	}
	return s.generateVersionnedDossierOpenAI(candidateID, competenceDossier, need, language)
}

// generateVersionnedDossierOpenAI appelle OpenAI pour produire un nouveau dossier de compétences
func (s *DossierVersionningService) generateVersionnedDossierOpenAI(candidateID string, competenceDossier models.CompetenceDossier, need *string, language string) (*models.CompetenceDossier, error) {
	log.Printf("🤖 OPENAI VERSIONNING - Début de la génération (langue: %s)", language)

	// Sérialiser le dossier source pour le mettre dans le prompt
	sourceJSON, err := json.MarshalIndent(competenceDossier, "", "  ")
	if err != nil {
		log.Printf("❌ OPENAI VERSIONNING - Erreur sérialisation: %v", err)
		return nil, fmt.Errorf("échec de la sérialisation du dossier source: %v", err)
	}

	log.Printf("📝 OPENAI VERSIONNING - Prompt construit (taille: %d chars)", len(sourceJSON))
	prompt := s.buildVersionningPrompt(string(sourceJSON), need, language)

	payload := map[string]any{
		"model": s.config.Model, // Utilise la même config que /extract
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature":       s.config.Temperature, // Utilise la même température que /extract (0.1)
		"top_p":             s.config.TopP,
		"frequency_penalty": s.config.FrequencyPenalty,
		"presence_penalty":  s.config.PresencePenalty,
		// Pas de response_format: json_object - on utilise le prompt strict comme /extract
	}

	// Ajouter les paramètres de tokens selon le modèle
	if s.config.MaxCompletionTokens > 0 {
		payload["max_completion_tokens"] = s.config.MaxCompletionTokens
	} else if s.config.MaxTokens > 0 {
		payload["max_tokens"] = s.config.MaxTokens
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("❌ OPENAI VERSIONNING - Erreur sérialisation payload: %v", err)
		return nil, fmt.Errorf("erreur lors de la sérialisation de la requête: %v", err)
	}

	log.Printf("🚀 OPENAI VERSIONNING - Envoi requête à OpenAI (taille: %d bytes)", len(bodyBytes))
	httpReq, err := http.NewRequest(http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("❌ OPENAI VERSIONNING - Erreur création requête: %v", err)
		return nil, fmt.Errorf("erreur lors de la création de la requête: %v", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("❌ OPENAI VERSIONNING - Erreur envoi requête: %v", err)
		return nil, fmt.Errorf("erreur lors de l'envoi de la requête: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("📡 OPENAI VERSIONNING - Réponse reçue (status: %d)", resp.StatusCode)

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ OPENAI VERSIONNING - Erreur lecture réponse: %v", err)
		return nil, fmt.Errorf("erreur lors de la lecture de la réponse: %v", err)
	}
	if resp.StatusCode >= 400 {
		log.Printf("❌ OPENAI VERSIONNING - Erreur API OpenAI %d: %s", resp.StatusCode, string(respBytes))
		return nil, fmt.Errorf("openai error %d: %s", resp.StatusCode, string(respBytes))
	}

	log.Printf("✅ OPENAI VERSIONNING - Réponse lue (taille: %d bytes)", len(respBytes))

	// Structure minimale pour parser la réponse OpenAI
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
		log.Printf("❌ OPENAI VERSIONNING - Erreur parsing réponse: %v", err)
		return nil, fmt.Errorf("erreur lors du parsing de la réponse: %v", err)
	}
	if openAIResp.Error != nil {
		log.Printf("❌ OPENAI VERSIONNING - Erreur OpenAI: %s", openAIResp.Error.Message)
		return nil, fmt.Errorf("erreur OpenAI: %s", openAIResp.Error.Message)
	}
	if len(openAIResp.Choices) == 0 {
		log.Printf("❌ OPENAI VERSIONNING - Aucune réponse générée")
		return nil, fmt.Errorf("aucune réponse générée par OpenAI")
	}

	content := strings.TrimSpace(openAIResp.Choices[0].Message.Content)
	log.Printf("📄 OPENAI VERSIONNING - Contenu brut reçu (taille: %d chars)", len(content))

	// Dé-fencer au cas où
	if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
		log.Printf("🔧 OPENAI VERSIONNING - Contenu nettoyé (taille: %d chars)", len(content))
	}

	// Nettoyer les caractères d'échappement problématiques
	content = strings.ReplaceAll(content, "\\\"", "\"")
	content = strings.ReplaceAll(content, "\\n", "\n")
	content = strings.ReplaceAll(content, "\\t", "\t")
	log.Printf("🧹 OPENAI VERSIONNING - Contenu nettoyé des échappements (taille: %d chars)", len(content))

	// Normaliser le JSON pour corriger les types de données
	normalizedContent, err := s.normalizeJSONTypes(content)
	if err != nil {
		log.Printf("❌ OPENAI VERSIONNING - Erreur normalisation JSON: %v", err)
		return nil, fmt.Errorf("erreur lors de la normalisation du JSON: %v", err)
	}
	log.Printf("🔧 OPENAI VERSIONNING - JSON normalisé (taille: %d chars)", len(normalizedContent))

	var newDossier models.CompetenceDossier
	if err := json.Unmarshal([]byte(normalizedContent), &newDossier); err != nil {
		log.Printf("❌ OPENAI VERSIONNING - Erreur parsing JSON final: %v", err)
		log.Printf("❌ OPENAI VERSIONNING - Contenu problématique (premiers 500 chars): %s", normalizedContent[:min(500, len(normalizedContent))])
		return nil, fmt.Errorf("réponse OpenAI invalide (JSON): %v; contenu: %s", err, normalizedContent)
	}

	log.Printf("✅ OPENAI VERSIONNING - Dossier versionné parsé avec succès:")
	log.Printf("   - Nom: %s %s", newDossier.Prenom, newDossier.Nom)
	log.Printf("   - Poste: %s", newDossier.Poste)
	log.Printf("   - Expériences: %d", len(newDossier.Experiences))
	log.Printf("   - Formations: %d", len(newDossier.Formations))

	return &newDossier, nil
}

// generateVersionnedDossierAnthropic appelle Anthropic pour produire un nouveau dossier de compétences
func (s *DossierVersionningService) generateVersionnedDossierAnthropic(candidateID string, competenceDossier models.CompetenceDossier, need *string, language string) (*models.CompetenceDossier, error) {
	log.Printf("🤖 ANTHROPIC VERSIONNING - Début de la génération (langue: %s)", language)

	// Sérialiser le dossier source pour le mettre dans le prompt
	sourceJSON, err := json.MarshalIndent(competenceDossier, "", "  ")
	if err != nil {
		log.Printf("❌ ANTHROPIC VERSIONNING - Erreur sérialisation: %v", err)
		return nil, fmt.Errorf("échec de la sérialisation du dossier source: %v", err)
	}

	log.Printf("📝 ANTHROPIC VERSIONNING - Prompt construit (taille: %d chars)", len(sourceJSON))
	prompt := s.buildVersionningPrompt(string(sourceJSON), need, language)

	// Appel Anthropic Messages API
	payload := map[string]any{
		"model":      s.anthropicConfig.Model,
		"max_tokens": s.anthropicConfig.MaxTokens,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("❌ ANTHROPIC VERSIONNING - Erreur sérialisation payload: %v", err)
		return nil, fmt.Errorf("erreur lors de la sérialisation de la requête: %v", err)
	}

	log.Printf("🚀 ANTHROPIC VERSIONNING - Envoi requête à Anthropic (taille: %d bytes)", len(bodyBytes))
	httpReq, err := http.NewRequest(http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("❌ ANTHROPIC VERSIONNING - Erreur création requête: %v", err)
		return nil, fmt.Errorf("erreur lors de la création de la requête: %v", err)
	}
	httpReq.Header.Set("x-api-key", s.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("content-type", "application/json")

	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("❌ ANTHROPIC VERSIONNING - Erreur envoi requête: %v", err)
		return nil, fmt.Errorf("erreur lors de l'envoi de la requête: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("📡 ANTHROPIC VERSIONNING - Réponse reçue (status: %d)", resp.StatusCode)

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ ANTHROPIC VERSIONNING - Erreur lecture réponse: %v", err)
		return nil, fmt.Errorf("erreur lors de la lecture de la réponse: %v", err)
	}
	if resp.StatusCode >= 400 {
		log.Printf("❌ ANTHROPIC VERSIONNING - Erreur API Anthropic %d: %s", resp.StatusCode, string(respBytes))
		return nil, fmt.Errorf("anthropic error %d: %s", resp.StatusCode, string(respBytes))
	}

	log.Printf("✅ ANTHROPIC VERSIONNING - Réponse lue (taille: %d bytes)", len(respBytes))

	// Parse basique du format Messages (content[0].text)
	var aResp struct {
		Content []struct {
			Text string `json:"text"`
			Type string `json:"type"`
		} `json:"content"`
	}
	if err := json.Unmarshal(respBytes, &aResp); err != nil {
		log.Printf("❌ ANTHROPIC VERSIONNING - Erreur parsing réponse: %v", err)
		return nil, fmt.Errorf("erreur lors du parsing de la réponse: %v", err)
	}
	if len(aResp.Content) == 0 || strings.TrimSpace(aResp.Content[0].Text) == "" {
		log.Printf("❌ ANTHROPIC VERSIONNING - Contenu vide")
		return nil, fmt.Errorf("anthropic returned empty content")
	}

	content := strings.TrimSpace(aResp.Content[0].Text)
	log.Printf("📄 ANTHROPIC VERSIONNING - Contenu brut reçu (taille: %d chars)", len(content))

	// Dé-fencer au cas où
	if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
		log.Printf("🔧 ANTHROPIC VERSIONNING - Contenu nettoyé (taille: %d chars)", len(content))
	}

	// Nettoyer les caractères d'échappement problématiques
	content = strings.ReplaceAll(content, "\\\"", "\"")
	content = strings.ReplaceAll(content, "\\n", "\n")
	content = strings.ReplaceAll(content, "\\t", "\t")
	log.Printf("🧹 ANTHROPIC VERSIONNING - Contenu nettoyé des échappements (taille: %d chars)", len(content))

	// Normaliser le JSON pour corriger les types de données
	normalizedContent, err := s.normalizeJSONTypes(content)
	if err != nil {
		log.Printf("❌ ANTHROPIC VERSIONNING - Erreur normalisation JSON: %v", err)
		return nil, fmt.Errorf("erreur lors de la normalisation du JSON: %v", err)
	}
	log.Printf("🔧 ANTHROPIC VERSIONNING - JSON normalisé (taille: %d chars)", len(normalizedContent))

	var newDossier models.CompetenceDossier
	if err := json.Unmarshal([]byte(normalizedContent), &newDossier); err != nil {
		log.Printf("❌ ANTHROPIC VERSIONNING - Erreur parsing JSON final: %v", err)
		log.Printf("❌ ANTHROPIC VERSIONNING - Contenu problématique (premiers 500 chars): %s", normalizedContent[:min(500, len(normalizedContent))])
		return nil, fmt.Errorf("réponse Anthropic invalide (JSON): %v; contenu: %s", err, normalizedContent)
	}

	log.Printf("✅ ANTHROPIC VERSIONNING - Dossier versionné parsé avec succès:")
	log.Printf("   - Nom: %s %s", newDossier.Prenom, newDossier.Nom)
	log.Printf("   - Poste: %s", newDossier.Poste)
	log.Printf("   - Expériences: %d", len(newDossier.Experiences))
	log.Printf("   - Formations: %d", len(newDossier.Formations))

	return &newDossier, nil
}

// normalizeJSONTypes corrige les types de données dans le JSON pour éviter les erreurs de parsing
func (s *DossierVersionningService) normalizeJSONTypes(jsonContent string) (string, error) {
	// Parser le JSON en interface{} pour pouvoir le manipuler
	var data interface{}
	if err := json.Unmarshal([]byte(jsonContent), &data); err != nil {
		return "", fmt.Errorf("erreur parsing JSON initial: %v", err)
	}

	// Normaliser les types de données
	normalizedData := s.normalizeDataTypes(data)

	// Re-sérialiser en JSON
	normalizedJSON, err := json.Marshal(normalizedData)
	if err != nil {
		return "", fmt.Errorf("erreur sérialisation JSON normalisé: %v", err)
	}

	return string(normalizedJSON), nil
}

// normalizeDataTypes parcourt récursivement les données pour corriger les types
func (s *DossierVersionningService) normalizeDataTypes(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		// Pour les objets, normaliser récursivement
		result := make(map[string]interface{})
		for key, value := range v {
			// Normaliser les champs spécifiques qui doivent être des arrays
			if key == "réalisations" || key == "logiciels" || key == "hobbies" || key == "languages" ||
				key == "secteurs_activites" || key == "domaines_expertise" || key == "formations" ||
				key == "expériences" || key == "AI_suggest" {
				result[key] = s.ensureArray(value)
			} else {
				result[key] = s.normalizeDataTypes(value)
			}
		}
		return result
	case []interface{}:
		// Pour les arrays, normaliser chaque élément
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = s.normalizeDataTypes(item)
		}
		return result
	default:
		return v
	}
}

// ensureArray s'assure qu'une valeur est un array, convertit les strings en arrays si nécessaire
func (s *DossierVersionningService) ensureArray(value interface{}) interface{} {
	switch v := value.(type) {
	case []interface{}:
		// Déjà un array, retourner tel quel
		return v
	case string:
		// String, la convertir en array avec un seul élément
		if v == "" {
			return []interface{}{}
		}
		return []interface{}{v}
	case nil:
		// Nil, retourner un array vide
		return []interface{}{}
	default:
		// Autre type, essayer de le convertir en string puis en array
		return []interface{}{fmt.Sprintf("%v", v)}
	}
}

// buildVersionningPrompt construit le prompt de transformation en insistant sur
// la conservation stricte du schéma de sortie, comme dans /extract.
func (s *DossierVersionningService) buildVersionningPrompt(sourceDossierJSON string, need *string, language string) string {
	needText := ""
	if need != nil {
		needText = *need
	}

	// Adapter le prompt selon la langue
	if language == "en" {
		return s.buildVersionningPromptEnglish(sourceDossierJSON, needText)
	}
	return s.buildVersionningPromptFrench(sourceDossierJSON, needText)
}

// buildVersionningPromptFrench construit le prompt en français
func (s *DossierVersionningService) buildVersionningPromptFrench(sourceDossierJSON string, needText string) string {
	return fmt.Sprintf(`Adapte ce dossier de compétences selon le besoin client. RÈGLES CRITIQUES:

1. CONSERVE TOUT: prenom, nom, email, phone, age, diplome, expérience, mobilité, disponibilité, permis_B
2. RENVOIE TOUTES les réalisations (même si tu ne les modifies pas)
3. MODIFIE SEULEMENT: contexte, projet, réalisations, summary, poste
4. RÉORGANISE par pertinence si besoin

BESOIN: %s

DOSSIER:
%s

Réponds UNIQUEMENT avec le JSON, sans texte avant/après.`, needText, sourceDossierJSON)
}

// buildVersionningPromptEnglish construit le prompt en anglais
func (s *DossierVersionningService) buildVersionningPromptEnglish(sourceDossierJSON string, needText string) string {
	return fmt.Sprintf(`Adapt this competence portfolio according to the client's need. CRITICAL RULES:

1. PRESERVE EVERYTHING: prenom, nom, email, phone, age, diplome, expérience, mobilité, disponibilité, permis_B
2. RETURN ALL achievements (even if you don't modify them)
3. MODIFY ONLY: contexte, projet, réalisations, summary, poste
4. REORGANIZE by relevance if needed

NEED: %s

PORTFOLIO:
%s

Respond ONLY with the JSON, without text before/after.`, needText, sourceDossierJSON)
}
