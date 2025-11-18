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
func ExtractAndEnrichWithFilenameAnthropic(file []byte, filename string, language string, needHaiku bool) ([]byte, error) {
	startTime := time.Now()
	apiKey := strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY"))
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY not set")
	}
	// Inverser needHaiku pour GetAnthropicConfig qui attend needSonnet
	// needHaiku=true → needSonnet=false (utiliser Haiku)
	// needHaiku=false → needSonnet=true (utiliser Sonnet, par défaut)
	acfg := GetAnthropicConfig(!needHaiku)
	model := acfg.Model
	if needHaiku {
		log.Printf("🔧 [Anthropic] needHaiku=true → modèle sélectionné: %s (Haiku)", model)
	} else {
		log.Printf("🔧 [Anthropic] needHaiku=false → modèle sélectionné: %s (Sonnet, par défaut)", model)
	}
	log.Printf("DEBUG: [Anthropic] Démarrage extraction (model=%s, needHaiku=%v)", model, needHaiku)

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

	// Désécaper les newlines dans le JSON pour améliorer la lisibilité pour l'AI
	// json.Marshal échappe les newlines comme \n, mais quand ce JSON est ensuite
	// intégré dans le prompt et re-marshallé, cela peut créer des \\n (double échappement).
	// On remplace donc \\n par \n pour que l'AI voie de vrais retours à la ligne.
	// Note: On utilise ReplaceAll pour gérer tous les cas (\\n, \\r\\n, etc.)
	rawStr = strings.ReplaceAll(rawStr, "\\n", "\n")
	rawStr = strings.ReplaceAll(rawStr, "\\r", "\r")
	rawStr = strings.ReplaceAll(rawStr, "\\t", "\t")

	prompt := GetExtractionPromptProductionWithLanguage(rawStr, language)

	// Ajouter une instruction stricte de sortie JSON
	prompt = prompt + "\n\nIMPORTANT: Réponds UNIQUEMENT par un objet JSON valide. Pas de markdown, pas de prose. Si une info manque, laisse une chaîne vide ou un tableau vide."

	// 2) Appel Anthropic Messages API
	payload := map[string]any{
		"model":      model,
		"max_tokens": acfg.MaxTokens,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	body, _ := json.Marshal(payload)
	bodySize := len(body)
	log.Printf("📤 [Anthropic] Envoi requête à l'API (taille payload: %d bytes, modèle: %s, max_tokens: %d)", bodySize, model, acfg.MaxTokens)

	req, err := http.NewRequest(http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	httpClient := &http.Client{Timeout: 10 * time.Minute}
	requestStart := time.Now()
	log.Printf("⏳ [Anthropic] Attente de la réponse (timeout: 10 minutes)...")
	resp, err := httpClient.Do(req)
	requestDuration := time.Since(requestStart)
	if err != nil {
		log.Printf("❌ [Anthropic] Erreur après %v: %v", requestDuration, err)
		return nil, err
	}
	defer resp.Body.Close()
	log.Printf("📥 [Anthropic] Réponse reçue après %v (status: %d)", requestDuration, resp.StatusCode)

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ [Anthropic] Erreur lecture du body: %v", err)
		return nil, err
	}
	log.Printf("📥 [Anthropic] Body lu (taille: %d bytes)", len(respBytes))
	previewLen := 500
	if len(respBytes) < previewLen {
		previewLen = len(respBytes)
	}
	log.Printf("📥 [Anthropic] Réponse API - Status: %d, Body preview: %s", resp.StatusCode, string(respBytes[:previewLen]))
	if resp.StatusCode >= 400 {
		log.Printf("❌ [Anthropic] Erreur API: %d - %s", resp.StatusCode, string(respBytes))
		return nil, fmt.Errorf("anthropic error %d: %s", resp.StatusCode, string(respBytes))
	}

	// Parse basique du format Messages (content[0].text)
	var aResp struct {
		Content []struct {
			Text string `json:"text"`
			Type string `json:"type"`
		} `json:"content"`
		StopReason string `json:"stop_reason"` // "end_turn", "max_tokens", "stop_sequence"
	}
	if err := json.Unmarshal(respBytes, &aResp); err != nil {
		return nil, err
	}
	if len(aResp.Content) == 0 || strings.TrimSpace(aResp.Content[0].Text) == "" {
		return nil, fmt.Errorf("anthropic returned empty content")
	}

	final := []byte(aResp.Content[0].Text)

	// Vérifier si la réponse a été tronquée
	if aResp.StopReason == "max_tokens" {
		log.Printf("⚠️ [Anthropic] Réponse tronquée (stop_reason=max_tokens, len=%d). Augmentez ANTHROPIC_MAX_TOKENS si nécessaire.", len(final))
		// Vérifier si le JSON est incomplet en essayant de le parser
		var testJSON map[string]any
		if err := json.Unmarshal(final, &testJSON); err != nil {
			log.Printf("❌ [Anthropic] JSON incomplet détecté (erreur: %v). La réponse a été tronquée avant la fin du JSON.", err)
			return nil, fmt.Errorf("anthropic response truncated: JSON incomplete (stop_reason=max_tokens, len=%d). Increase ANTHROPIC_MAX_TOKENS or reduce input size", len(final))
		}
	}

	duration := time.Since(startTime)
	log.Printf("DEBUG: [Anthropic] Terminé en %v (len=%d, stop_reason=%s)", duration, len(final), aResp.StopReason)
	return final, nil
}
