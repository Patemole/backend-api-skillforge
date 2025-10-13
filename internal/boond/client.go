package boond

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Client gère les appels à l'API Boond Manager
type Client struct {
	BaseURL string
	JWT     string
	HTTP    *http.Client
}

// New crée un nouveau client Boond avec un timeout raisonnable
func New(jwt string) *Client {
	baseURL := os.Getenv("BOOND_API_BASE_URL")
	if strings.TrimSpace(baseURL) == "" {
		// URL par défaut corrigée - Boond Manager utilise ui.boondmanager.com
		baseURL = "https://ui.boondmanager.com"
	}

	// Log de l'URL utilisée pour debug
	fmt.Printf("🌐 [Boond] Base URL configurée: %s\n", baseURL)

	return &Client{
		BaseURL: baseURL,
		JWT:     jwt,
		HTTP:    &http.Client{Timeout: 12 * time.Second},
	}
}

// jsonAPICandidateList représente une réponse JSON:API de liste
type jsonAPICandidateList struct {
	Data []struct {
		ID   string         `json:"id"`
		Type string         `json:"type"`
		Attr map[string]any `json:"attributes"`
	} `json:"data"`
}

// jsonAPICreateResp représente une réponse JSON:API de création
type jsonAPICreateResp struct {
	Data struct {
		ID   string         `json:"id"`
		Type string         `json:"type"`
		Attr map[string]any `json:"attributes"`
	} `json:"data"`
}

// SearchCandidateByEmail tente de trouver un candidat par email.
// Retourne l'ID s'il existe, sinon une chaîne vide.
func (c *Client) SearchCandidateByEmail(ctx context.Context, email string) (string, error) {
	if strings.TrimSpace(email) == "" {
		return "", nil
	}

	// On essaye plusieurs stratégies de recherche pour maximiser les chances
	endpoints := []string{
		fmt.Sprintf("%s/api/candidates?email=%s", c.BaseURL, url.QueryEscape(email)),
		fmt.Sprintf("%s/api/candidates?search=%s", c.BaseURL, url.QueryEscape(email)),
		fmt.Sprintf("%s/api/candidates?q=%s", c.BaseURL, url.QueryEscape(email)),
	}

	for _, ep := range endpoints {
		fmt.Printf("🔎 [Boond] GET %s\n", ep)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, ep, nil)
		if err != nil {
			return "", err
		}
		addStdHeaders(req, c.JWT)

		resp, err := c.HTTP.Do(req)
		if err != nil {
			fmt.Printf("❌ [Boond] GET error: %v\n", err)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		fmt.Printf("📥 [Boond] GET status=%d body_len=%d\n", resp.StatusCode, len(body))
		if resp.StatusCode >= 400 {
			// On essaie l'endpoint suivant
			if len(body) > 0 {
				b := string(body)
				if len(b) > 400 {
					b = b[:400] + "…(tronqué)"
				}
				fmt.Printf("⚠️  [Boond] GET error body: %s\n", b)
			}
			continue
		}

		var list jsonAPICandidateList
		if err := json.Unmarshal(body, &list); err != nil {
			fmt.Printf("⚠️  [Boond] Unmarshal list error: %v\n", err)
			continue
		}
		if len(list.Data) > 0 {
			// Log détaillé du candidat trouvé pour vérifier si c'est le bon
			candidate := list.Data[0]
			fmt.Printf("✅ [Boond] Found candidate id=%s\n", candidate.ID)

			// Vérifier l'email pour s'assurer que c'est le bon candidat
			candidateEmail := ""
			if attrs, ok := candidate.Attr["email1"].(string); ok {
				candidateEmail = attrs
				fmt.Printf("   Email trouvé: %s\n", attrs)
			}
			if attrs, ok := candidate.Attr["firstName"].(string); ok {
				fmt.Printf("   Prénom: %s\n", attrs)
			}
			if attrs, ok := candidate.Attr["lastName"].(string); ok {
				fmt.Printf("   Nom: %s\n", attrs)
			}

			// Vérifier que l'email correspond vraiment
			if strings.EqualFold(candidateEmail, email) {
				fmt.Printf("✅ [Boond] Email correspond - doublon confirmé\n")
				return candidate.ID, nil
			} else {
				fmt.Printf("⚠️  [Boond] Email ne correspond pas (%s != %s) - pas un doublon\n", candidateEmail, email)
				// Continuer la recherche avec les autres endpoints
			}
		}
	}

	return "", nil
}

// SearchCandidateByName recherche un candidat par nom et prénom
func (c *Client) SearchCandidateByName(ctx context.Context, firstName, lastName string) (string, error) {
	if strings.TrimSpace(firstName) == "" || strings.TrimSpace(lastName) == "" {
		return "", nil
	}

	// Recherche par nom et prénom avec les bons paramètres
	ep := fmt.Sprintf("%s/api/candidates?firstName=%s&lastName=%s", c.BaseURL, url.QueryEscape(firstName), url.QueryEscape(lastName))

	fmt.Printf("🔎 [Boond] GET %s\n", ep)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ep, nil)
	if err != nil {
		return "", err
	}
	addStdHeaders(req, c.JWT)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		fmt.Printf("❌ [Boond] GET error: %v\n", err)
		return "", err
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	fmt.Printf("📥 [Boond] GET status=%d body_len=%d\n", resp.StatusCode, len(body))
	if resp.StatusCode >= 400 {
		if len(body) > 0 {
			b := string(body)
			if len(b) > 400 {
				b = b[:400] + "…(tronqué)"
			}
			fmt.Printf("⚠️  [Boond] GET error body: %s\n", b)
		}
		return "", nil
	}

	var list jsonAPICandidateList
	if err := json.Unmarshal(body, &list); err != nil {
		fmt.Printf("⚠️  [Boond] Unmarshal list error: %v\n", err)
		return "", nil
	}

	// Chercher un candidat avec le bon nom et prénom
	for _, candidate := range list.Data {
		candFirstName, _ := candidate.Attr["firstName"].(string)
		candLastName, _ := candidate.Attr["lastName"].(string)

		if strings.EqualFold(candFirstName, firstName) && strings.EqualFold(candLastName, lastName) {
			fmt.Printf("✅ [Boond] Found candidate by name id=%s (%s %s)\n", candidate.ID, candFirstName, candLastName)
			return candidate.ID, nil
		}
	}

	return "", nil
}

// CreateCandidate crée un candidat Boond avec les attributs fournis.
// Retourne l'ID créé et le payload brut de la réponse.
func (c *Client) CreateCandidate(ctx context.Context, attributes map[string]any) (string, json.RawMessage, error) {
	payload := map[string]any{
		"data": map[string]any{
			"type":       "candidate",
			"attributes": attributes,
		},
	}
	buf, err := json.Marshal(payload)
	if err != nil {
		return "", nil, err
	}

	ep := fmt.Sprintf("%s/api/candidates", c.BaseURL)
	fmt.Printf("📡 [Boond] POST %s (payload_len=%d)\n", ep, len(buf))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ep, bytes.NewReader(buf))
	if err != nil {
		return "", nil, err
	}
	addStdHeaders(req, c.JWT)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		fmt.Printf("❌ [Boond] POST error: %v\n", err)
		return "", nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("📥 [Boond] POST status=%d body_len=%d\n", resp.StatusCode, len(body))
	if resp.StatusCode >= 400 {
		if len(body) > 0 {
			b := string(body)
			if len(b) > 600 {
				b = b[:600] + "…(tronqué)"
			}
			fmt.Printf("⚠️  [Boond] POST error body: %s\n", b)
		}
		return "", body, fmt.Errorf("boond create candidate failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var created jsonAPICreateResp
	if err := json.Unmarshal(body, &created); err == nil && created.Data.ID != "" {
		fmt.Printf("🎉 [Boond] Candidate created id=%s\n", created.Data.ID)
		return created.Data.ID, body, nil
	}

	// Fallback: essayer d'extraire un id à la racine si non JSON:API strict
	var generic map[string]any
	if err := json.Unmarshal(body, &generic); err == nil {
		if idVal, ok := generic["id"].(string); ok && idVal != "" {
			fmt.Printf("🎉 [Boond] Candidate created (fallback) id=%s\n", idVal)
			return idVal, body, nil
		}
	}

	return "", body, nil
}

// UploadDocument upload un fichier (CV) et l'associe à un candidat
func (c *Client) UploadDocument(ctx context.Context, candidateID string, fileData []byte, filename string) (string, error) {
	// Créer un buffer multipart pour l'upload
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Ajouter les champs requis
	writer.WriteField("parentId", candidateID)
	writer.WriteField("parentType", "candidateResume")
	writer.WriteField("parsing", "true") // Activer le parsing IA

	// Ajouter le fichier
	fileWriter, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", err
	}
	fileWriter.Write(fileData)
	writer.Close()

	// Créer la requête
	ep := fmt.Sprintf("%s/api/documents", c.BaseURL)
	fmt.Printf("📄 [Boond] POST %s (parentId=%s, filename=%s)\n", ep, candidateID, filename)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ep, &buf)
	if err != nil {
		return "", err
	}

	// Headers pour multipart
	req.Header.Set("X-Jwt-Client-Boondmanager", c.JWT)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")

	fmt.Printf("🔑 [Boond] Headers upload:\n")
	fmt.Printf("   X-Jwt-Client-Boondmanager: %s...\n", maskToken(c.JWT, 20))
	fmt.Printf("   Content-Type: %s\n", req.Header.Get("Content-Type"))

	resp, err := c.HTTP.Do(req)
	if err != nil {
		fmt.Printf("❌ [Boond] Upload error: %v\n", err)
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("📥 [Boond] Upload status=%d body_len=%d\n", resp.StatusCode, len(body))
	if resp.StatusCode >= 400 {
		if len(body) > 0 {
			b := string(body)
			if len(b) > 400 {
				b = b[:400] + "…(tronqué)"
			}
			fmt.Printf("⚠️  [Boond] Upload error body: %s\n", b)
		}
		return "", fmt.Errorf("boond upload document failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	// Parser la réponse pour récupérer l'ID du document
	var docResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &docResp); err != nil {
		return "", fmt.Errorf("failed to parse upload response: %w", err)
	}

	fmt.Printf("🎉 [Boond] Document uploaded id=%s\n", docResp.Data.ID)
	return docResp.Data.ID, nil
}

func addStdHeaders(req *http.Request, jwt string) {
	// Format correct pour Boond Manager (basé sur le code qui fonctionne)
	req.Header.Set("X-Jwt-Client-Boondmanager", jwt)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Log des headers pour debug
	fmt.Printf("🔑 [Boond] Headers envoyés:\n")
	fmt.Printf("   X-Jwt-Client-Boondmanager: %s...\n", maskToken(jwt, 20))
	fmt.Printf("   Content-Type: %s\n", req.Header.Get("Content-Type"))
	fmt.Printf("   Accept: %s\n", req.Header.Get("Accept"))
}

// maskToken retourne les n premiers caractères du token
func maskToken(t string, n int) string {
	if n <= 0 || len(t) <= n {
		return t
	}
	return t[:n]
}
