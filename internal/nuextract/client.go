package nuextract

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Client struct {
	projectID string
	apiKey    string
	http      *http.Client
}

func New() *Client {
	return &Client{
		projectID: os.Getenv("NUEXTRACT_PROJECT_ID"),
		apiKey:    os.Getenv("NUEXTRACT_API_KEY"),
		http:      &http.Client{},
	}
}

// Extract envoie un fichier binaire (PDF) à NuExtract et renvoie la réponse brute.
func (c *Client) Extract(file []byte) ([]byte, error) {
	url := fmt.Sprintf("https://nuextract.ai/api/projects/%s/extract", c.projectID)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(file))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
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
	return io.ReadAll(resp.Body)
}
