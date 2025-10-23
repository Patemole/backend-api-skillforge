package gladia

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

const (
	uploadURL        = "https://api.gladia.io/v2/upload"
	transcriptionURL = "https://api.gladia.io/v2/transcription/"
)

type Client struct {
	httpClient *http.Client
	apiKey     string
}

func NewClient() (*Client, error) {
	key := os.Getenv("GLADIA_KEY")
	if key == "" {
		return nil, errors.New("GLADIA_KEY manquant")
	}
	return &Client{
		httpClient: &http.Client{Timeout: 60 * time.Second},
		apiKey:     key,
	}, nil
}

type UploadResponse struct {
	AudioURL string `json:"audio_url"`
}

// UploadAudio envoie un flux binaire audio/vidéo à Gladia et retourne l'URL hébergée
func (c *Client) UploadAudio(ctx context.Context, filename string, reader io.Reader) (string, *http.Response, []byte, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("audio", filename)
	if err != nil {
		return "", nil, nil, err
	}
	if _, err := io.Copy(part, reader); err != nil {
		return "", nil, nil, err
	}
	if err := writer.Close(); err != nil {
		return "", nil, nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, &body)
	if err != nil {
		return "", nil, nil, err
	}
	req.Header.Set("x-gladia-key", c.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", resp, nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", resp, respBody, nil
	}
	var ur UploadResponse
	if err := json.Unmarshal(respBody, &ur); err != nil {
		return "", resp, respBody, err
	}
	return ur.AudioURL, resp, respBody, nil
}

type StartResponse struct {
	ResultURL string `json:"result_url"`
}

// StartTranscription déclenche la transcription et retourne l'URL de résultat
func (c *Client) StartTranscription(ctx context.Context, payload map[string]any) (string, *http.Response, []byte, error) {
	buf, err := json.Marshal(payload)
	if err != nil {
		return "", nil, nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, transcriptionURL, bytes.NewReader(buf))
	if err != nil {
		return "", nil, nil, err
	}
	req.Header.Set("x-gladia-key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", resp, nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", resp, respBody, nil
	}
	var sr StartResponse
	if err := json.Unmarshal(respBody, &sr); err != nil {
		return "", resp, respBody, err
	}
	return sr.ResultURL, resp, respBody, nil
}

// PollResult appelle l'URL de résultat (GET) avec la clé Gladia et retourne le JSON brut
func (c *Client) PollResult(ctx context.Context, resultURL string) (*http.Response, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, resultURL, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("x-gladia-key", c.apiKey)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return resp, nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp, body, nil
}
