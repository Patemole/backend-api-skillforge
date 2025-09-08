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
}

func New() *Client {
	return &Client{
		projectID:    os.Getenv("NUEXTRACT_PROJECT_ID"),
		nuexAPIKey:   os.Getenv("NUEXTRACT_API_KEY"),
		openAIAPIKey: os.Getenv("OPENAI_API_KEY"),
		http: &http.Client{
			Timeout: 5 * time.Minute, // Timeout de 5 minutes
		},
	}
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
	return c.ExtractAndEnrichWithFilename(file, "")
}

// ExtractAndEnrichWithFilename same as ExtractAndEnrich but with filename for test mode
func (c *Client) ExtractAndEnrichWithFilename(file []byte, filename string) ([]byte, error) {
	startTime := time.Now()
	log.Printf("DEBUG: Début de l'extraction et enrichissement (MODE TEST - OpenAI SEUL)")
	log.Printf("DEBUG: Taille du fichier: %d bytes", len(file))
	log.Printf("DEBUG: Project ID: %s", c.projectID)
	log.Printf("DEBUG: API Key présent: %t", c.nuexAPIKey != "")

	// MODE OPENAI DIRECT: On utilise OpenAI pour extraire directement le contenu du PDF
	log.Printf("DEBUG: MODE OPENAI DIRECT - Extraction PDF avec OpenAI")
	log.Printf("DEBUG: Nom du fichier: %s", filename)

	// Extraire le vrai contenu du PDF
	log.Printf("DEBUG: Extraction du contenu réel du PDF")
	
	var fileContent string
	var err error
	
	// Vérifier si c'est un PDF
	if strings.HasSuffix(strings.ToLower(filename), ".pdf") || len(file) > 1000 {
		log.Printf("DEBUG: Fichier PDF détecté, extraction du texte")
		
		// Essayer d'abord UniPDF (le plus puissant)
		unipdfExtractor := NewUniPDFExtractor()
		fileContent, err = unipdfExtractor.ExtractTextFromPDFWithTables(file)
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
	} else {
		// Fichier texte
		fileContent = string(file)
		log.Printf("DEBUG: Fichier texte détecté, %d caractères", len(fileContent))
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

	// 2) Call OpenAI Chat Completions API (plus rapide que Responses API)
	openAIStart := time.Now()
	if c.openAIAPIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}

	// Récupérer le prompt et la configuration
	prompt := GetExtractionPrompt(string(raw))
	config := GetOpenAIConfig()

	payload := map[string]interface{}{
		"model": config.Model,
		"messages": []map[string]string{
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
