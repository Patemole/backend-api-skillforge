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
	apiKey string
	config nuextract.OpenAIConfig
}

// NewDossierVersionningService crée une nouvelle instance du service.
func NewDossierVersionningService() *DossierVersionningService {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		panic("OPENAI_API_KEY not set")
	}
	cfg := nuextract.GetOpenAIConfig()
	return &DossierVersionningService{apiKey: apiKey, config: cfg}
}

// GenerateVersionnedDossier appelle OpenAI pour produire un nouveau dossier de compétences
// en conservant strictement le même schéma que models.CompetenceDossier.
func (s *DossierVersionningService) GenerateVersionnedDossier(candidateID string, competenceDossier models.CompetenceDossier, need *string) (*models.CompetenceDossier, error) {
	log.Printf("🤖 OPENAI VERSIONNING - Début de la génération")

	// Sérialiser le dossier source pour le mettre dans le prompt
	sourceJSON, err := json.MarshalIndent(competenceDossier, "", "  ")
	if err != nil {
		log.Printf("❌ OPENAI VERSIONNING - Erreur sérialisation: %v", err)
		return nil, fmt.Errorf("échec de la sérialisation du dossier source: %v", err)
	}

	log.Printf("📝 OPENAI VERSIONNING - Prompt construit (taille: %d chars)", len(sourceJSON))
	prompt := s.buildVersionningPrompt(string(sourceJSON), need)

	payload := map[string]any{
		"model": s.config.Model, // Utilise la même config que /extract
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens":        s.config.MaxTokens,
		"temperature":       s.config.Temperature, // Utilise la même température que /extract (0.1)
		"top_p":             s.config.TopP,
		"frequency_penalty": s.config.FrequencyPenalty,
		"presence_penalty":  s.config.PresencePenalty,
		// Pas de response_format: json_object - on utilise le prompt strict comme /extract
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

	client := &http.Client{Timeout: 60 * time.Second}
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

	var newDossier models.CompetenceDossier
	if err := json.Unmarshal([]byte(content), &newDossier); err != nil {
		log.Printf("❌ OPENAI VERSIONNING - Erreur parsing JSON final: %v", err)
		log.Printf("❌ OPENAI VERSIONNING - Contenu problématique (premiers 500 chars): %s", content[:min(500, len(content))])
		return nil, fmt.Errorf("réponse OpenAI invalide (JSON): %v; contenu: %s", err, content)
	}

	log.Printf("✅ OPENAI VERSIONNING - Dossier versionné parsé avec succès:")
	log.Printf("   - Nom: %s %s", newDossier.Prenom, newDossier.Nom)
	log.Printf("   - Poste: %s", newDossier.Poste)
	log.Printf("   - Expériences: %d", len(newDossier.Experiences))
	log.Printf("   - Formations: %d", len(newDossier.Formations))

	return &newDossier, nil
}

// buildVersionningPrompt construit le prompt de transformation en insistant sur
// la conservation stricte du schéma de sortie, comme dans /extract.
func (s *DossierVersionningService) buildVersionningPrompt(sourceDossierJSON string, need *string) string {
	needText := ""
	if need != nil {
		needText = *need
	}

	// Prompt strict comme dans /extract, avec instructions claires pour le JSON
	return fmt.Sprintf(`Tu es un expert RH spécialisé dans l'adaptation de dossiers de compétences. Je souhaite que tu adaptes un dossier de compétences existant selon un besoin client spécifique, en respectant STRICTEMENT le schéma JSON fourni.

NE CHANGE SURTOUT PAS LES CLÉS DE CE DICTIONNAIRE, CAR IL DOIT ÊTRE UTILISÉ AUTREMENT PAR LA SUITE.

OBJECTIF: Créer une NOUVELLE VERSION adaptée à un besoin client spécifique en:
- Reprenant UNIQUEMENT les informations présentes dans le dossier source (NE RIEN INVENTER).
- Mettant en avant les expériences, réalisations et formulations les plus pertinentes pour le besoin.
- Réordonnant si nécessaire les éléments (expériences) pour refléter la pertinence, et en reformulant pour plus d'impact.
- Conservant STRICTEMENT le même schéma JSON et les mêmes clés que l'objet d'entrée (y compris accents: "expériences", "réalisations", "mobilité", etc.).
- Conservant les types identiques (chaînes, tableaux, etc.).
- Ne jamais ajouter de nouveaux champs et ne jamais supprimer des champs existants; tu peux laisser des chaînes vides si nécessaire.

BESOIN CLIENT (optionnel): %s

DOSSIER SOURCE:
%s

INSTRUCTIONS CRITIQUES :

1. **CONSERVATION DU SCHÉMA** :
   - Respecte EXACTEMENT les clés suivantes: prenom, nom, age, poste, diplome, expérience, mobilité, disponibilité, permis_B, hobbies, languages, secteurs_activites, domaines_expertise, formations[], expériences[], logiciels[].
   - Dans chaque entrée de "expériences", tu peux réécrire "contexte", "projet", "réalisations" et réordonner; tu NE DOIS PAS inventer de nouvelles entreprises, dates, ou postes.
   - Conserve les types de données identiques (string, array, etc.).

2. **ADAPTATION SELON LE BESOIN** :
   - Si un besoin client est fourni, mets en avant les expériences et compétences les plus pertinentes.
   - Reformule les descriptions pour mieux correspondre au besoin.
   - Réordonne les expériences par pertinence si nécessaire.

3. **RÈGLES STRICTES** :
   - NE RIEN INVENTER : utilise uniquement les informations du dossier source.
   - Ne change pas les noms d'entreprises, les dates, ou les postes.
   - Tu peux reformuler les descriptions et réorganiser l'ordre.

L'output doit respecter EXACTEMENT le modèle ci-dessus. Si une information n'est pas présente et que tu ne peux pas l'estimer, laisse le champ vide (chaîne vide "").

Réponds UNIQUEMENT avec le JSON structuré, sans texte avant ou après.`, needText, sourceDossierJSON)
}
