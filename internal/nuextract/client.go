package nuextract

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
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

// ExtractAndEnrich sends a PDF to NuExtract, then feeds its JSON into your OpenAI Agent
// via the Responses API, returning the enriched CV JSON.
func (c *Client) ExtractAndEnrich(file []byte) ([]byte, error) {
    // 1) Call NuExtract
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

    // 2) Call OpenAI Responses API
    if c.openAIAPIKey == "" {
        return nil, fmt.Errorf("OPENAI_API_KEY not set")
    }

    promptObj := map[string]string{
        "id":      "pmpt_68930fac64248193b98138eec93c9593095b4a2e570c9476",
    }
    payload := map[string]interface{}{
        "prompt":            promptObj,
        "input":             string(raw),
        "max_output_tokens": 50000,
    }
    bodyBytes, err := json.Marshal(payload)
    if err != nil {
        return nil, err
    }

    oaReq, err := http.NewRequest(http.MethodPost, "https://api.openai.com/v1/responses", bytes.NewReader(bodyBytes))
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

    // 3) Unwrap the assistant's JSON from the response envelope
    var wrap struct {
        Output []struct {
            Content []struct {
                Text string `json:"text"`
            } `json:"content"`
        } `json:"output"`
    }
    if err := json.Unmarshal(respBytes, &wrap); err != nil {
        return nil, err
    }
    if len(wrap.Output) == 0 || len(wrap.Output[0].Content) == 0 {
        return nil, fmt.Errorf("no content in OpenAI response")
    }

    return []byte(wrap.Output[0].Content[0].Text), nil
}