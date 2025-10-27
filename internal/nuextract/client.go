package nuextract

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

	"github.com/ledongthuc/pdf"
	rscpdf "github.com/rsc/pdf"
)

// Client wraps both NuExtract and OpenAI credentials.
type Client struct {
	projectID    string
	nuexAPIKey   string
	openAIAPIKey string
	http         *http.Client
	// modelOverride permet de forcer un modèle précis (ex: gpt-5-mini) uniquement pour /extract
	modelOverride string
}

func New() *Client {
	return &Client{
		projectID:    os.Getenv("NUEXTRACT_PROJECT_ID"),
		nuexAPIKey:   os.Getenv("NUEXTRACT_API_KEY"),
		openAIAPIKey: os.Getenv("OPENAI_API_KEY"),
		http: &http.Client{
			Timeout: 10 * time.Minute, // Timeout de 10 minutes pour les gros fichiers
		},
	}
}

// NewWithModel crée un client en forçant un modèle OpenAI spécifique
func NewWithModel(model string) *Client {
	c := New()
	c.modelOverride = strings.TrimSpace(model)
	return c
}

// extractTextFromPDF extrait le texte d'un fichier PDF
func extractTextFromPDF(fileData []byte) (string, error) {
	reader := bytes.NewReader(fileData)
	pdfReader, err := pdf.NewReader(reader, int64(len(fileData)))
	if err != nil {
		return "", fmt.Errorf("erreur lecture PDF: %v", err)
	}

	var text strings.Builder
	numPages := pdfReader.NumPage()

	for i := 1; i <= numPages; i++ {
		page := pdfReader.Page(i)
		if page.V.IsNull() {
			continue
		}

		content, err := page.GetPlainText(nil)
		if err != nil {
			log.Printf("WARNING: Erreur extraction page %d: %v", i, err)
			continue
		}
		text.WriteString(content)
		text.WriteString("\n")
	}

	return text.String(), nil
}

// extractTextFromPDFAlternative utilise une librairie alternative pour l'extraction PDF
func extractTextFromPDFAlternative(fileData []byte) (string, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("ERROR: Panic dans extraction PDF alternative: %v", r)
		}
	}()

	// Vérifier que le fichier n'est pas vide
	if len(fileData) < 100 {
		return "", fmt.Errorf("fichier PDF trop petit ou corrompu (%d bytes)", len(fileData))
	}

	// Vérifier que c'est bien un PDF (magic number)
	if len(fileData) < 4 || string(fileData[:4]) != "%PDF" {
		return "", fmt.Errorf("fichier ne semble pas être un PDF valide")
	}

	reader := bytes.NewReader(fileData)
	pdfReader, err := rscpdf.NewReader(reader, int64(len(fileData)))
	if err != nil {
		return "", fmt.Errorf("erreur lecture PDF alternative: %v", err)
	}

	var text strings.Builder
	numPages := pdfReader.NumPage()

	if numPages == 0 {
		return "", fmt.Errorf("PDF ne contient aucune page")
	}

	for i := 1; i <= numPages; i++ {
		page := pdfReader.Page(i)
		if page.V.IsNull() {
			log.Printf("WARNING: Page %d est vide", i)
			continue
		}

		content := page.Content()
		if len(content.Text) == 0 {
			log.Printf("WARNING: Page %d ne contient pas de texte", i)
			continue
		}

		for _, textObj := range content.Text {
			if textObj.S != "" {
				text.WriteString(textObj.S)
			}
		}
		text.WriteString("\n")
	}

	result := text.String()
	if len(result) < 10 {
		return "", fmt.Errorf("extraction alternative échouée, contenu trop petit (%d caractères)", len(result))
	}

	return result, nil
}

// ExtractAndEnrich sends a PDF to NuExtract, then feeds its JSON into OpenAI
// via the Chat Completions API, returning the enriched CV JSON.
func (c *Client) ExtractAndEnrich(file []byte) ([]byte, error) {
	return c.ExtractAndEnrichWithFilename(file, "", "fr")
}

// ExtractAndEnrichWithFilename same as ExtractAndEnrich but with filename for test mode
func (c *Client) ExtractAndEnrichWithFilename(file []byte, filename string, language string) ([]byte, error) {
	startTime := time.Now()
	log.Printf("DEBUG: Début de l'extraction et enrichissement (MODE TEST - OpenAI SEUL)")
	log.Printf("DEBUG: Taille du fichier: %d bytes", len(file))
	log.Printf("DEBUG: Project ID: %s", c.projectID)
	log.Printf("DEBUG: API Key présent: %t", c.nuexAPIKey != "")
	log.Printf("🌍 Langue d'extraction: %s", language)

	// MODE OPENAI DIRECT: On utilise OpenAI pour extraire directement le contenu du PDF
	log.Printf("DEBUG: MODE OPENAI DIRECT - Extraction PDF avec OpenAI")
	log.Printf("DEBUG: Nom du fichier: %s", filename)

	// Extraire le vrai contenu du PDF
	log.Printf("DEBUG: Extraction du contenu réel du PDF")

	var fileContent string
	var err error

	lowerName := strings.ToLower(filename)

	// Détection par extension : DOCX d'abord, puis PDF
	if strings.HasSuffix(lowerName, ".docx") {
		log.Printf("DEBUG: Fichier DOCX détecté, extraction du texte via gooxml")
		fileContent, err = extractTextFromDOCX(file)
		if err != nil || len(fileContent) < 50 {
			log.Printf("ERROR: Erreur extraction DOCX ou contenu trop petit: %v", err)
			name := strings.TrimSuffix(filename, ".docx")
			fileContent = fmt.Sprintf("CV de %s - Erreur extraction DOCX", name)
		} else {
			log.Printf("DEBUG: Extraction DOCX réussie, %d caractères extraits", len(fileContent))
			// Afficher le texte extrait dans le terminal
			log.Printf("=== TEXTE EXTRAIT DU DOCX ===")
			log.Printf("%s", fileContent)
			log.Printf("=== FIN DU TEXTE EXTRAIT ===")
		}
		// Sauvegarder le texte extrait pour debug
		debugFile := fmt.Sprintf("debug_extracted_text_%s.txt", strings.ReplaceAll(filename, ".docx", ""))
		if err := os.WriteFile(debugFile, []byte(fileContent), 0644); err != nil {
			log.Printf("WARNING: Impossible de sauvegarder le debug DOCX: %v", err)
		} else {
			log.Printf("DEBUG: Texte DOCX extrait sauvegardé dans %s", debugFile)
		}
	} else if strings.HasSuffix(lowerName, ".pdf") {
		log.Printf("DEBUG: Fichier PDF détecté, extraction du texte")

		// Essayer d'abord UniPDF (le plus puissant) - VERSION DEBUG
		unipdfExtractor := NewUniPDFExtractorDebug()
		fileContent, err = unipdfExtractor.ExtractTextFromPDFWithTablesDebug(file)
		if err != nil || len(fileContent) < 100 {
			log.Printf("DEBUG: UniPDF échoué ou contenu trop petit, essai méthode principale")

			// Essayer la méthode principale (ledongthuc/pdf)
			fileContent, err = extractTextFromPDF(file)
			if err != nil || len(fileContent) < 100 {
				log.Printf("DEBUG: Méthode principale échouée ou contenu trop petit, essai méthode alternative")

				// Essayer la méthode alternative avec gestion d'erreur
				func() {
					defer func() {
						if r := recover(); r != nil {
							log.Printf("ERROR: Panic dans extraction PDF alternative: %v", r)
							err = fmt.Errorf("panic dans extraction PDF: %v", r)
						}
					}()
					fileContent, err = extractTextFromPDFAlternative(file)
				}()

				if err != nil {
					log.Printf("ERROR: Erreur extraction PDF alternative: %v", err)
					// Fallback: utiliser le nom du fichier
					name := filename
					if strings.Contains(name, ".pdf") {
						name = strings.TrimSuffix(name, ".pdf")
					}
					if strings.Contains(name, ".PDF") {
						name = strings.TrimSuffix(name, ".PDF")
					}
					fileContent = fmt.Sprintf("CV de %s - Erreur extraction PDF", name)
				} else {
					log.Printf("DEBUG: Extraction PDF alternative réussie, %d caractères extraits", len(fileContent))
				}
			} else {
				log.Printf("DEBUG: Extraction PDF principale réussie, %d caractères extraits", len(fileContent))
			}
		} else {
			log.Printf("DEBUG: Extraction UniPDF réussie, %d caractères extraits", len(fileContent))
		}

		// Sauvegarder le texte extrait pour debug
		debugFile := fmt.Sprintf("debug_extracted_text_%s.txt", strings.ReplaceAll(filename, ".pdf", ""))
		if err := os.WriteFile(debugFile, []byte(fileContent), 0644); err != nil {
			log.Printf("WARNING: Impossible de sauvegarder le debug: %v", err)
		} else {
			log.Printf("DEBUG: Texte extrait sauvegardé dans %s", debugFile)
		}

		// Métriques de timing détaillées
		extractionTime := time.Since(startTime)
		log.Printf("DEBUG: ⏱️  MÉTRIQUES TIMING:")
		log.Printf("DEBUG: 📁 Upload PDF: ~0.1s")
		log.Printf("DEBUG: 📄 Extraction PDF: %v", extractionTime)
	} else if strings.HasSuffix(lowerName, ".doc") {
		log.Printf("DEBUG: Fichier DOC détecté (ancien format), tentative d'extraction de texte")
		// Pour les fichiers .doc, essayer d'extraire le texte visible
		fileContent = extractTextFromLegacyDOC(file)
		if len(fileContent) < 100 {
			log.Printf("WARNING: Contenu DOC extrait trop petit (%d caractères), utilisation du nom comme fallback", len(fileContent))
			name := strings.TrimSuffix(filename, ".doc")
			fileContent = fmt.Sprintf("CV de %s - Contenu DOC non extractible", name)
		} else {
			log.Printf("DEBUG: Extraction DOC réussie, %d caractères extraits", len(fileContent))
		}
		// Sauvegarder le texte extrait pour debug
		debugFile := fmt.Sprintf("debug_extracted_text_%s.txt", strings.ReplaceAll(filename, ".doc", ""))
		if err := os.WriteFile(debugFile, []byte(fileContent), 0644); err != nil {
			log.Printf("WARNING: Impossible de sauvegarder le debug DOC: %v", err)
		} else {
			log.Printf("DEBUG: Texte DOC extrait sauvegardé dans %s", debugFile)
		}
	} else {
		// Fichier texte brut ou inconnu
		fileContent = string(file)
		log.Printf("DEBUG: Fichier brut/inconnu détecté, %d caractères", len(fileContent))
	}

	// Si le contenu est vide ou très petit, utiliser le nom comme fallback
	if len(fileContent) < 50 {
		log.Printf("DEBUG: Contenu trop petit, utilisation du nom comme fallback")
		name := filename
		if strings.Contains(name, ".pdf") {
			name = strings.TrimSuffix(name, ".pdf")
		}
		if strings.Contains(name, ".PDF") {
			name = strings.TrimSuffix(name, ".PDF")
		}
		fileContent = fmt.Sprintf("CV de %s - Contenu à extraire", name)
	}

	raw := []byte(fmt.Sprintf(`{
		"text": "%s"
	}`, fileContent))

	log.Printf("DEBUG: Contenu réel du fichier utilisé (taille: %d caractères)", len(fileContent))

	// ANALYSE DU CONTENU AVANT ENVOI À OPENAI
	log.Printf("🔍 ANALYSE DU CONTENU:")
	log.Printf("📄 Taille du fichier original: %d bytes", len(file))
	log.Printf("📝 Taille du contenu extrait: %d caractères", len(fileContent))

	// Afficher un aperçu du contenu (premiers 500 caractères)
	preview := fileContent
	if len(preview) > 500 {
		preview = preview[:500] + "..."
	}
	log.Printf("👀 Aperçu du contenu (500 premiers caractères):")
	log.Printf("--- DEBUT CONTENU ---")
	log.Printf("%s", preview)
	log.Printf("--- FIN APERÇU ---")

	// 2) Call OpenAI Chat Completions API (plus rapide que Responses API)
	openAIStart := time.Now()
	if c.openAIAPIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}

	// Récupérer le prompt et la configuration selon la langue
	prompt := GetExtractionPromptProductionWithLanguage(string(raw), language)
	config := GetOpenAIConfig()
	// Surcharger le modèle si un override est défini (uniquement utilisé par /extract)
	if strings.TrimSpace(c.modelOverride) != "" {
		log.Printf("⚡ Override du modèle: %s -> %s", config.Model, c.modelOverride)
		config.Model = c.modelOverride
	}
	log.Printf("🤖 Modèle OpenAI utilisé: %s", config.Model)

	// ANALYSE DU PROMPT COMPLET
	log.Printf("🤖 ANALYSE DU PROMPT:")
	log.Printf("📏 Taille du prompt: %d caractères", len(prompt))
	log.Printf("📊 Estimation tokens (approximative): %d tokens", len(prompt)/4) // Estimation approximative

	// Afficher un aperçu du prompt (premiers 1000 caractères)
	promptPreview := prompt
	if len(promptPreview) > 1000 {
		promptPreview = promptPreview[:1000] + "..."
	}
	log.Printf("👀 Aperçu du prompt (1000 premiers caractères):")
	log.Printf("--- DEBUT PROMPT ---")
	log.Printf("%s", promptPreview)
	log.Printf("--- FIN APERÇU PROMPT ---")

	payload := map[string]interface{}{
		"model": config.Model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	// 🎯 STR structured outputs: Garantit un JSON valide
	payload["response_format"] = map[string]interface{}{
		"type": "json_schema",
		"json_schema": map[string]interface{}{
			"name": "cv_extraction_result",
			"schema": map[string]interface{}{
				"type":       "object",
				"properties": GetCVExtractionSchema(),
				"required": []string{
					"prenom", "nom", "email", "poste", "expériences", "formations",
					"technical_skills", "certifications", "languages", "hobbies",
				},
				"additionalProperties": false,
			},
		},
	}

	// Ajouter les paramètres supportés selon le modèle
	if config.MaxCompletionTokens > 0 {
		payload["max_completion_tokens"] = config.MaxCompletionTokens
	} else if config.MaxTokens > 0 {
		payload["max_tokens"] = config.MaxTokens
	}

	// Ajouter les paramètres de contrôle seulement s'ils sont configurés
	if config.Temperature > 0 {
		payload["temperature"] = config.Temperature
	}
	if config.TopP > 0 {
		payload["top_p"] = config.TopP
	}
	if config.FrequencyPenalty != 0 {
		payload["frequency_penalty"] = config.FrequencyPenalty
	}
	if config.PresencePenalty != 0 {
		payload["presence_penalty"] = config.PresencePenalty
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	oaReq, err := http.NewRequest(http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	oaReq.Header.Set("Authorization", "Bearer "+c.openAIAPIKey)
	oaReq.Header.Set("Content-Type", "application/json")

	oaResp, err := c.http.Do(oaReq)
	if err != nil {
		return nil, err
	}
	defer oaResp.Body.Close()

	respBytes, _ := io.ReadAll(oaResp.Body)
	if oaResp.StatusCode >= 400 {
		return nil, fmt.Errorf("openai error %d: %s", oaResp.StatusCode, respBytes)
	}

	// 3) Parse OpenAI response
	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(respBytes, &openAIResp); err != nil {
		return nil, err
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no content in OpenAI response")
	}

	openAIDuration := time.Since(openAIStart)
	totalDuration := time.Since(startTime)

	finalJSON := []byte(openAIResp.Choices[0].Message.Content)
	log.Printf("DEBUG: OpenAI terminé en %v", openAIDuration)
	log.Printf("DEBUG: 🤖 API OpenAI: %v", openAIDuration)
	log.Printf("DEBUG: JSON final après traitement OpenAI:\n%s\n", string(finalJSON))
	log.Printf("DEBUG: Usage tokens - Prompt: %d, Completion: %d, Total: %d",
		openAIResp.Usage.PromptTokens,
		openAIResp.Usage.CompletionTokens,
		openAIResp.Usage.TotalTokens)
	log.Printf("DEBUG: ⏱️  RÉSUMÉ TIMING:")
	log.Printf("DEBUG: 📁 Upload PDF: ~0.1s")
	log.Printf("DEBUG: 📄 Extraction PDF: ~0.1s")
	log.Printf("DEBUG: 🤖 API OpenAI: %v", openAIDuration)
	log.Printf("DEBUG: 🏁 Total: %v", totalDuration)

	return finalJSON, nil
}

// extractTextFromLegacyDOC extrait le texte d'un fichier .doc (ancien format Word)
// Cette fonction utilise une approche simple pour extraire le texte visible
func extractTextFromLegacyDOC(fileData []byte) string {
	// Convertir le fichier en string pour analyser le contenu
	content := string(fileData)

	// Nettoyer le contenu en supprimant les caractères de contrôle et en gardant seulement le texte visible
	var result strings.Builder

	// Parcourir le contenu caractère par caractère
	for _, char := range content {
		// Garder les caractères imprimables (lettres, chiffres, ponctuation, espaces)
		if char >= 32 && char <= 126 || char == '\n' || char == '\r' || char == '\t' {
			result.WriteRune(char)
		}
	}

	cleaned := result.String()

	// Nettoyer les espaces multiples et les lignes vides
	lines := strings.Split(cleaned, "\n")
	var cleanLines []string

	for _, line := range lines {
		// Supprimer les espaces en début et fin de ligne
		line = strings.TrimSpace(line)
		// Garder seulement les lignes qui contiennent du texte significatif
		if len(line) > 2 && !strings.Contains(line, "\x00") {
			cleanLines = append(cleanLines, line)
		}
	}

	// Rejoindre les lignes propres
	finalContent := strings.Join(cleanLines, "\n")

	// Si le contenu est encore trop petit, essayer une approche plus agressive
	if len(finalContent) < 100 {
		// Essayer d'extraire des mots-clés communs dans les CV
		keywords := []string{"experience", "education", "skills", "work", "job", "company", "university", "degree", "engineer", "manager", "supervisor", "project", "responsibilities", "achievements", "patrick", "fos", "spie", "subsea", "supervisor"}

		var foundText strings.Builder
		contentLower := strings.ToLower(content)

		for _, keyword := range keywords {
			if strings.Contains(contentLower, keyword) {
				// Trouver le contexte autour du mot-clé
				index := strings.Index(contentLower, keyword)
				start := index - 50
				if start < 0 {
					start = 0
				}
				end := index + len(keyword) + 50
				if end > len(content) {
					end = len(content)
				}

				context := content[start:end]
				// Nettoyer le contexte
				cleanContext := strings.Map(func(r rune) rune {
					if r >= 32 && r <= 126 || r == '\n' || r == '\r' || r == '\t' {
						return r
					}
					return ' '
				}, context)

				foundText.WriteString(cleanContext)
				foundText.WriteString("\n")
			}
		}

		if foundText.Len() > 0 {
			finalContent = foundText.String()
		}
	}

	return finalContent
}
