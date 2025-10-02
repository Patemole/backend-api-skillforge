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
	
	// Organiser les variables par catégories pour une meilleure compréhension
	variablesByCategory := s.organizeVariablesByCategory(req.AvailableVariables)
	variablesJSON, _ := json.MarshalIndent(variablesByCategory, "", "  ")

	return fmt.Sprintf(`Tu es un expert en génération de templates d'emails professionnels. Tu dois analyser les exemples d'emails suivants et générer un template réutilisable en identifiant les patterns et en mappant les variables appropriées.

**Contexte :**
- Nom du template : %s
- Organisation ID : %s

**Exemples d'emails à analyser :**
%s

**Variables disponibles (organisées par catégorie) :**
%s

**Instructions détaillées :**

1. **Analyse des patterns :**
   - Identifie les structures communes dans les exemples (salutation, présentation, détails, conclusion)
   - Repère les informations variables qui changent entre les exemples
   - Note le ton et le style professionnel utilisé

2. **Mapping des variables :**
   - Utilise les variables de base ({{prenom}}, {{titre_poste}}, etc.) pour les informations principales
   - Pour les expériences, utilise les variables granulaires ({{experiences.poste}}, {{experiences.entreprise}}, etc.) pour plus de précision
   - Mappe intelligemment selon le contexte (ex: "{{experiences.entreprise}}" pour le nom de l'entreprise, "{{experiences.poste}}" pour le poste)
   - Utilise les variables de comptage ({{experiences_count}}, {{logiciels_count}}) pour des phrases comme "avec {{experiences_count}} expériences"

3. **Génération du template :**
   - Crée un template cohérent qui respecte la structure des exemples
   - Utilise les variables appropriées pour chaque section
   - Maintient le ton professionnel et la logique des exemples
   - Assure-toi que le template est générique mais détaillé

4. **Variables hiérarchiques :**
   - Les variables comme {{experiences.poste}} font référence aux détails des expériences
   - Les variables comme {{experiences}} peuvent être utilisées pour un résumé global
   - Utilise les variables spécifiques pour plus de précision quand c'est pertinent

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

**Exemples de mapping intelligent :**
- "Jean Dupont" → {{prenom}}
- "Développeur Full-Stack" → {{titre_poste}}
- "5 ans d'expérience" → {{nombre_experience}}
- "chez Google" → {{experiences.entreprise}}
- "en tant que Senior Developer" → {{experiences.poste}}
- "React, Node.js" → {{logiciels}} ou {{competences_techniques}}

**Important :**
- Utilise UNIQUEMENT les variables de la liste "available_variables"
- Privilégie les variables granulaires pour plus de précision
- Le confidence_score doit être entre 0.0 et 1.0
- Assure-toi que le template est cohérent avec les exemples fournis`,
		req.TemplateName,
		req.OrganizationID,
		string(examplesJSON),
		string(variablesJSON),
	)
}

// organizeVariablesByCategory organise les variables par catégories pour une meilleure compréhension
func (s *TemplateGeneratorService) organizeVariablesByCategory(variables []string) map[string][]string {
	categories := map[string][]string{
		"Informations de base": {},
		"Expériences (globales)": {},
		"Expériences (détails)": {},
		"Logiciels & Compétences": {},
		"Projets": {},
		"Formations": {},
		"Autres": {},
	}

	for _, variable := range variables {
		switch {
		case contains(variable, []string{"{{prenom}}", "{{titre_poste}}", "{{nombre_experience}}", "{{disponibilite}}", "{{mobilite}}", "{{diplome}}", "{{age}}", "{{permis_b}}"}):
			categories["Informations de base"] = append(categories["Informations de base"], variable)
		case contains(variable, []string{"{{experiences}}", "{{experiences_count}}"}):
			categories["Expériences (globales)"] = append(categories["Expériences (globales)"], variable)
		case contains(variable, []string{"{{experiences.poste}}", "{{experiences.entreprise}}", "{{experiences.duree}}", "{{experiences.date_debut}}", "{{experiences.date_fin}}", "{{experiences.projet}}", "{{experiences.contexte}}", "{{experiences.realisations}}", "{{experiences.logiciels}}"}):
			categories["Expériences (détails)"] = append(categories["Expériences (détails)"], variable)
		case contains(variable, []string{"{{logiciel}}", "{{logiciels}}", "{{logiciels_count}}", "{{competences_techniques}}", "{{competences_fonctionnelles}}"}):
			categories["Logiciels & Compétences"] = append(categories["Logiciels & Compétences"], variable)
		case contains(variable, []string{"{{projets}}", "{{projets_count}}"}):
			categories["Projets"] = append(categories["Projets"], variable)
		case contains(variable, []string{"{{formations}}", "{{formations_count}}", "{{formations.diplome}}", "{{formations.etablissement}}", "{{formations.annee}}"}):
			categories["Formations"] = append(categories["Formations"], variable)
		default:
			categories["Autres"] = append(categories["Autres"], variable)
		}
	}

	// Nettoyer les catégories vides
	for category, vars := range categories {
		if len(vars) == 0 {
			delete(categories, category)
		}
	}

	return categories
}

// contains vérifie si une variable est dans une liste
func contains(variable string, list []string) bool {
	for _, item := range list {
		if variable == item {
			return true
		}
	}
	return false
}
