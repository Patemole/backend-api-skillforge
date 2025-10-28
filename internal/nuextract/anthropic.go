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
)

// ExtractAndEnrichWithFilenameAnthropic effectue l'extraction et l'enrichissement via Anthropic (fallback)
func ExtractAndEnrichWithFilenameAnthropic(file []byte, filename string, language string) ([]byte, error) {
	startTime := time.Now()
	apiKey := strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY"))
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY not set")
	}
	aCfg := GetAnthropicConfig()
	model := aCfg.Model

	log.Printf("DEBUG: [Anthropic] Démarrage fallback (model=%s)", model)

	// 1) Extraire le texte du fichier (mêmes heuristiques que le client OpenAI)
	lowerName := strings.ToLower(filename)
	var fileContent string
	var err error

	if strings.HasSuffix(lowerName, ".docx") {
		log.Printf("DEBUG: [Anthropic] DOCX détecté, extraction via gooxml")
		fileContent, err = extractTextFromDOCX(file)
		if err != nil || len(fileContent) < 50 {
			log.Printf("ERROR: [Anthropic] Erreur extraction DOCX ou contenu trop petit: %v", err)
			name := strings.TrimSuffix(filename, ".docx")
			fileContent = fmt.Sprintf("CV de %s - Erreur extraction DOCX", name)
		}
	} else if strings.HasSuffix(lowerName, ".pdf") {
		log.Printf("DEBUG: [Anthropic] PDF détecté, extraction via UniPDF puis fallback")
		unipdfExtractor := NewUniPDFExtractorDebug()
		fileContent, err = unipdfExtractor.ExtractTextFromPDFWithTablesDebug(file)
		if err != nil || len(fileContent) < 100 {
			log.Printf("DEBUG: [Anthropic] UniPDF échoué/trop petit, essai méthode principale")
			fileContent, err = extractTextFromPDF(file)
			if err != nil || len(fileContent) < 100 {
				log.Printf("DEBUG: [Anthropic] Méthode principale échouée/trop petit, essai alternative")
				fileContent, err = extractTextFromPDFAlternative(file)
				if err != nil {
					log.Printf("ERROR: [Anthropic] Erreur extraction PDF alternative: %v", err)
					name := filename
					if strings.Contains(name, ".pdf") {
						name = strings.TrimSuffix(name, ".pdf")
					}
					if strings.Contains(name, ".PDF") {
						name = strings.TrimSuffix(name, ".PDF")
					}
					fileContent = fmt.Sprintf("CV de %s - Erreur extraction PDF", name)
				}
			}
		}
	} else if strings.HasSuffix(lowerName, ".doc") {
		log.Printf("DEBUG: [Anthropic] DOC détecté (legacy), extraction basique")
		fileContent = extractTextFromLegacyDOC(file)
		if len(fileContent) < 100 {
			name := strings.TrimSuffix(filename, ".doc")
			fileContent = fmt.Sprintf("CV de %s - Contenu DOC non extractible", name)
		}
	} else {
		fileContent = string(file)
	}

	if len(fileContent) < 50 {
		name := filename
		if strings.Contains(name, ".pdf") {
			name = strings.TrimSuffix(name, ".pdf")
		}
		if strings.Contains(name, ".PDF") {
			name = strings.TrimSuffix(name, ".PDF")
		}
		fileContent = fmt.Sprintf("CV de %s - Contenu à extraire", name)
	}

	// Construire l'input JSON brut de la même manière
	rawObj := map[string]string{"text": fileContent}
	rawBytes, _ := json.Marshal(rawObj)
	rawStr := string(rawBytes)

	prompt := GetExtractionPromptProductionWithLanguage(rawStr, language)

	// Ajouter une instruction stricte de sortie JSON
	prompt = prompt + "\n\nIMPORTANT: Réponds UNIQUEMENT par un objet JSON valide. Pas de markdown, pas de prose. Si une info manque, laisse une chaîne vide ou un tableau vide."

	// 2) Appel Anthropic Messages API
	payload := map[string]any{
		"model":      model,
		"max_tokens": aCfg.MaxTokens,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	httpClient := &http.Client{Timeout: 10 * time.Minute}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("anthropic error %d: %s", resp.StatusCode, string(respBytes))
	}

	// Parse basique du format Messages (content[0].text)
	var aResp struct {
		Content []struct {
			Text string `json:"text"`
			Type string `json:"type"`
		} `json:"content"`
	}
	if err := json.Unmarshal(respBytes, &aResp); err != nil {
		return nil, err
	}
	if len(aResp.Content) == 0 || strings.TrimSpace(aResp.Content[0].Text) == "" {
		return nil, fmt.Errorf("anthropic returned empty content")
	}

	final := []byte(aResp.Content[0].Text)
	duration := time.Since(startTime)
	log.Printf("DEBUG: [Anthropic] Terminé en %v (len=%d)", duration, len(final))
	return final, nil
}
