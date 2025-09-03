package nuextract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
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
		http:         &http.Client{},
	}
}

// ExtractAndEnrich sends a PDF to NuExtract, then feeds its JSON into OpenAI
// via the Chat Completions API, returning the enriched CV JSON.
func (c *Client) ExtractAndEnrich(file []byte) ([]byte, error) {
	startTime := time.Now()
	log.Printf("DEBUG: Début de l'extraction et enrichissement")

	// 1) Call NuExtract
	nuexStart := time.Now()
	nuexURL := fmt.Sprintf("https://nuextract.ai/api/projects/%s/extract", c.projectID)
	req, err := http.NewRequest(http.MethodPost, nuexURL, bytes.NewReader(file))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.nuexAPIKey)
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("nuextract error %d: %s", resp.StatusCode, body)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	nuexDuration := time.Since(nuexStart)
	log.Printf("DEBUG: NuExtract terminé en %v", nuexDuration)
	log.Printf("DEBUG: Réponse brute de NuExtract:\n%s\n", string(raw))

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
	log.Printf("DEBUG: JSON final après traitement OpenAI:\n%s\n", string(finalJSON))
	log.Printf("DEBUG: Usage tokens - Prompt: %d, Completion: %d, Total: %d",
		openAIResp.Usage.PromptTokens,
		openAIResp.Usage.CompletionTokens,
		openAIResp.Usage.TotalTokens)
	log.Printf("DEBUG: Temps total d'extraction et enrichissement: %v", totalDuration)

	return finalJSON, nil
}
