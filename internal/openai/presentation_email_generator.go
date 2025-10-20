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

// PresentationEmailGeneratorService gère la génération d'emails de présentation avec reformatage
type PresentationEmailGeneratorService struct {
	apiKey string
	config nuextract.OpenAIConfig
}

// NewPresentationEmailGeneratorService crée une nouvelle instance du service
func NewPresentationEmailGeneratorService() *PresentationEmailGeneratorService {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		panic("OPENAI_API_KEY not set")
	}

	config := nuextract.GetOpenAIConfig()
	return &PresentationEmailGeneratorService{
		apiKey: apiKey,
		config: config,
	}
}

// GeneratePresentationEmail génère un email de présentation avec reformatage et sélection d'expériences
func (s *PresentationEmailGeneratorService) GeneratePresentationEmail(req models.PresentationEmailRequest) (*models.PresentationEmailResponse, error) {
	// Construire le prompt pour la génération d'email de présentation
	prompt := s.buildPresentationPrompt(req)

	// Préparer la requête OpenAI
	payload := map[string]interface{}{
		"model": "gpt-4o-2024-08-06",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "Tu es un expert en rédaction d'emails professionnels de présentation de candidats. Tu améliores la fluidité du français et sélectionnes les expériences les plus pertinentes selon le besoin.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens":        s.config.MaxTokens,
		"temperature":       0.3, // Température basse pour plus de cohérence
		"top_p":             s.config.TopP,
		"frequency_penalty": s.config.FrequencyPenalty,
		"presence_penalty":  s.config.PresencePenalty,
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

	return &models.PresentationEmailResponse{
		EmailContent: openAIResp.Choices[0].Message.Content,
		Success:      true,
	}, nil
}

// buildPresentationPrompt construit le prompt pour la génération d'email de présentation
func (s *PresentationEmailGeneratorService) buildPresentationPrompt(req models.PresentationEmailRequest) string {
	// Construire le contexte du candidat
	candidateContext := s.buildCandidateContext(req.CandidateData)

	// Construire le contexte du besoin si présent
	needContext := ""
	if req.Need != nil && *req.Need != "" {
		needContext = fmt.Sprintf("\n**BESOIN SPÉCIFIQUE :**\n%s\n", *req.Need)
	}

	// Construire le contexte du template si présent
	templateContext := ""
	if req.Template != nil {
		templateContext = fmt.Sprintf("\n**TEMPLATE À UTILISER :**\nSujet: %s\nContenu: %s\n", req.Template.Subject, req.Template.Content)
	}

	return fmt.Sprintf(`Tu dois améliorer la fluidité du français d'un email de présentation en faisant des MICRO-AJUSTEMENTS uniquement.

**DONNÉES DU CANDIDAT :**
%s

%s

%s

**INSTRUCTIONS ULTRA-STRICTES :**

1. **MICRO-AJUSTEMENTS UNIQUEMENT :**
   - Ajoute UNIQUEMENT les liaisons manquantes (ex: "un talent" → "un talent prometteur")
   - Corrige UNIQUEMENT les tournures de phrases pour un français naturel
   - **INTERDICTION ABSOLUE : Ne rajoute AUCUNE nouvelle phrase**
   - **INTERDICTION ABSOLUE : Ne change PAS le contenu, seulement la fluidité**

2. **SÉLECTION D'EXPÉRIENCES (si besoin présent) :**
   - Si un besoin spécifique est fourni, sélectionne 2-3 expériences les plus pertinentes
   - **INTERDICTION ABSOLUE : Ne rajoute AUCUNE nouvelle phrase**
   - Utilise UNIQUEMENT les expériences fournies, telles quelles

3. **SÉLECTION DE RÉALISATIONS (si besoin présent) :**
   - Si un besoin spécifique est fourni, sélectionne 3-5 réalisations les plus pertinentes parmi les expériences choisies
   - Les réalisations sont séparées par des virgules dans chaque expérience
   - **INTERDICTION ABSOLUE : Ne rajoute AUCUNE nouvelle phrase**
   - Utilise UNIQUEMENT les réalisations fournies, telles quelles
   - **IMPORTANT : Ne liste PAS toutes les réalisations d'une expérience, sélectionne seulement les 2-3 plus pertinentes par expérience**
   - Exemple : au lieu de "A, B, C, D, E, F" → sélectionne "A, C, E" si ce sont les plus pertinentes

4. **SÉLECTION DE LOGICIELS (si besoin présent) :**
   - Si un besoin spécifique est fourni, sélectionne 3-5 logiciels les plus pertinents
   - **INTERDICTION ABSOLUE : Ne rajoute AUCUNE nouvelle phrase**
   - Utilise UNIQUEMENT les logiciels fournis, telles quelles

5. **RESPECT DU TEMPLATE :**
   - Suis EXACTEMENT le template fourni
   - Remplace UNIQUEMENT les variables par les valeurs
   - **INTERDICTION ABSOLUE : Ne rajoute AUCUNE phrase supplémentaire**

6. **PRÉSERVATION DU FORMATAGE HTML :**
   - **PRÉSERVE TOUT LE FORMATAGE HTML EXISTANT** (couleurs, gras, italique, listes)
   - **NE NETTOIE PAS** les balises HTML comme <strong>, <em>, <span style="...">, <ul>, <li>, etc.
   - **RENVOIE LE CONTENU TEL QUEL** avec le formatage HTML intact
   - **AJOUTE UN SAUT DE LIGNE** <br><br> entre le sujet et le début du contenu de l'email
   - **SUPPRIME "Contenu:"** du template (ex: "Contenu: Bonjour," → "Bonjour,")

7. **RÈGLES D'OR :**
   - **MICRO-AJUSTEMENTS SEULEMENT** (1-2 mots max par phrase)
   - **AUCUNE NOUVELLE PHRASE**
   - **AUCUNE INFORMATION SUPPLEMENTAIRE**
   - **GARDE LE CONTENU IDENTIQUE**
   - **AMÉLIORE SEULEMENT LA FLUIDITÉ**

**EXEMPLE DE CE QUE TU PEUX FAIRE :**
- "un talent" → "un talent prometteur" ✅
- "il est au début" → "il est au début de sa carrière" ✅
- "sa capacité" → "sa capacité à s'adapter" ✅
- Sélectionner les réalisations : "A, B, C, D, E, F" → "A, C, E" (les plus pertinentes) ✅
- Préserver le HTML : "<strong>titre</strong>" → "<strong>titre</strong>" (inchangé) ✅
- Supprimer "Contenu:" : "Sujet: Titre\nContenu: Bonjour" → "Sujet: Titre<br><br>Bonjour" ✅

**EXEMPLE DE CE QUE TU NE PEUX PAS FAIRE :**
- Rajouter "J'espère que cet email vous trouve bien" ❌
- Rajouter "Au cours de ses expériences passées" ❌
- Rajouter des phrases complètes ❌

**FORMAT DE RÉPONSE :**
Retourne directement le contenu de l'email avec les micro-ajustements, sans JSON, sans formatage supplémentaire.`,
		candidateContext,
		needContext,
		templateContext,
	)
}

// buildCandidateContext construit le contexte du candidat à partir des données
func (s *PresentationEmailGeneratorService) buildCandidateContext(data models.PresentationCandidateData) string {
	context := fmt.Sprintf(`- Prénom: %s
- Titre du poste: %s
- Nombre d'années d'expérience: %d
- Disponibilité: %s
- Mobilité: %s
- Diplôme: %s`,
		data.Prenom, data.TitrePoste, data.NombreExperience, data.Disponibilite, data.Mobilite, data.Diplome)

	if data.Langues != "" {
		context += fmt.Sprintf("\n- Langues: %s", data.Langues)
	}
	if data.Certifications != "" {
		context += fmt.Sprintf("\n- Certifications: %s", data.Certifications)
	}
	if data.Hobbies != "" {
		context += fmt.Sprintf("\n- Hobbies: %s", data.Hobbies)
	}

	if data.Experience != "" {
		context += fmt.Sprintf("\n\n**EXPÉRIENCES (%d au total):**\n%s", data.ExperienceCount, data.Experience)
	}

	if data.Logiciels != "" {
		context += fmt.Sprintf("\n\n**LOGICIELS (%d au total):**\n%s", data.LogicielsCount, data.Logiciels)
	}

	if data.CompetencesTechniques != "" {
		context += fmt.Sprintf("\n- Compétences techniques: %s", data.CompetencesTechniques)
	}
	if data.CompetencesFonctionnelles != "" {
		context += fmt.Sprintf("\n- Compétences fonctionnelles: %s", data.CompetencesFonctionnelles)
	}

	if data.Projets != "" {
		context += fmt.Sprintf("\n\n**PROJETS (%d au total):**\n%s", data.ProjetsCount, data.Projets)
	}

	return context
}
