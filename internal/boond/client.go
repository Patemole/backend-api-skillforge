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
// Le paramètre managerID est optionnel pour lier le candidat à un manager spécifique.
func (c *Client) CreateCandidate(ctx context.Context, attributes map[string]any, managerID ...string) (string, json.RawMessage, error) {
	data := map[string]any{
		"type":       "candidate",
		"attributes": attributes,
	}

	// Ajouter la relation mainManager si un manager ID est fourni
	if len(managerID) > 0 && strings.TrimSpace(managerID[0]) != "" {
		data["relationships"] = map[string]any{
			"mainManager": map[string]any{
				"data": map[string]any{
					"type": "resource",
					"id":   managerID[0],
				},
			},
		}
		fmt.Printf("👤 [Boond] Liaison candidat au manager: %s\n", managerID[0])
	}

	payload := map[string]any{
		"data": data,
	}
	buf, err := json.Marshal(payload)
	if err != nil {
		return "", nil, err
	}

	// Log du payload complet pour debug (uniquement les 1000 premiers caractères)
	payloadStr := string(buf)
	if len(payloadStr) > 1000 {
		fmt.Printf("📋 [Boond] Payload (preview): %s...\n", payloadStr[:1000])
	} else {
		fmt.Printf("📋 [Boond] Payload complet: %s\n", payloadStr)
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

// GetCandidateDocuments récupère la liste des documents d'un candidat
func (c *Client) GetCandidateDocuments(ctx context.Context, candidateID string) ([]map[string]any, error) {
	ep := fmt.Sprintf("%s/api/candidates/%s/documents", c.BaseURL, candidateID)
	fmt.Printf("📋 [Boond] GET %s\n", ep)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ep, nil)
	if err != nil {
		return nil, err
	}
	addStdHeaders(req, c.JWT)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		fmt.Printf("❌ [Boond] GET documents error: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("📥 [Boond] GET documents status=%d body_len=%d\n", resp.StatusCode, len(body))
	if resp.StatusCode >= 400 {
		if len(body) > 0 {
			b := string(body)
			if len(b) > 400 {
				b = b[:400] + "…(tronqué)"
			}
			fmt.Printf("⚠️  [Boond] GET documents error body: %s\n", b)
		}
		// Si 404, retourner une liste vide (pas d'erreur)
		if resp.StatusCode == 404 {
			fmt.Printf("ℹ️  [Boond] Aucun document trouvé pour le candidat (404)\n")
			return []map[string]any{}, nil
		}
		return nil, fmt.Errorf("boond get documents failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	// Parser la réponse JSON:API
	var docListResp struct {
		Data []struct {
			ID   string         `json:"id"`
			Type string         `json:"type"`
			Attr map[string]any `json:"attributes"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &docListResp); err != nil {
		return nil, fmt.Errorf("failed to parse documents response: %w", err)
	}

	// Transformer en liste simple
	var documents []map[string]any
	for _, doc := range docListResp.Data {
		documents = append(documents, map[string]any{
			"id":   doc.ID,
			"type": doc.Type,
			"name": doc.Attr["name"],
		})
	}

	fmt.Printf("✅ [Boond] %d documents récupérés pour le candidat\n", len(documents))
	return documents, nil
}

// DeleteDocument supprime un document Boond par son ID
func (c *Client) DeleteDocument(ctx context.Context, documentID string) error {
	if strings.TrimSpace(documentID) == "" {
		return fmt.Errorf("document ID cannot be empty")
	}

	ep := fmt.Sprintf("%s/api/documents/%s", c.BaseURL, documentID)
	fmt.Printf("🗑️  [Boond] DELETE %s\n", ep)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, ep, nil)
	if err != nil {
		return err
	}
	addStdHeaders(req, c.JWT)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		fmt.Printf("❌ [Boond] DELETE document error: %v\n", err)
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("📥 [Boond] DELETE document status=%d body_len=%d\n", resp.StatusCode, len(body))
	if resp.StatusCode >= 400 {
		if len(body) > 0 {
			b := string(body)
			if len(b) > 400 {
				b = b[:400] + "…(tronqué)"
			}
			fmt.Printf("⚠️  [Boond] DELETE document error body: %s\n", b)
		}
		return fmt.Errorf("boond delete document failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	fmt.Printf("✅ [Boond] Document deleted id=%s\n", documentID)
	return nil
}

// UploadDocument upload un fichier (CV) et l'associe à un candidat
// Vérifie d'abord s'il existe déjà un document avec le même nom et le supprime si nécessaire
func (c *Client) UploadDocument(ctx context.Context, candidateID string, fileData []byte, filename string) (string, error) {
	// 🔧 FIX: Vérifier s'il existe déjà un document avec le même nom
	fmt.Printf("🔍 [Boond] Vérification des documents existants pour éviter les doublons (filename=%s)\n", filename)
	existingDocs, err := c.GetCandidateDocuments(ctx, candidateID)
	if err != nil {
		// Si l'erreur est non critique (ex: endpoint non disponible), continuer quand même
		fmt.Printf("⚠️  [Boond] Erreur lors de la récupération des documents existants (non bloquant): %v\n", err)
		existingDocs = []map[string]any{}
	}

	// Chercher un document avec le même nom
	for _, doc := range existingDocs {
		docName, ok := doc["name"].(string)
		if ok && strings.EqualFold(docName, filename) {
			docID, ok := doc["id"].(string)
			if ok && docID != "" {
				fmt.Printf("⚠️  [Boond] Document existant trouvé avec le même nom: %s (id=%s), suppression...\n", filename, docID)
				// Supprimer le document existant pour éviter les doublons
				if deleteErr := c.DeleteDocument(ctx, docID); deleteErr != nil {
					fmt.Printf("⚠️  [Boond] Erreur lors de la suppression du document existant (non bloquant): %v\n", deleteErr)
					// Continuer quand même l'upload - Boond peut gérer les doublons
				} else {
					fmt.Printf("✅ [Boond] Document existant supprimé avec succès\n")
				}
				break
			}
		}
	}

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
	fmt.Printf("📄 [Boond] POST %s (parentId=%s, filename=%s, size=%d bytes)\n", ep, candidateID, filename, len(fileData))

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

	// 🔧 FIX: Créer un client HTTP avec un timeout plus long pour les uploads volumineux
	// Le timeout par défaut de 12s est trop court pour les fichiers Word volumineux
	uploadClient := &http.Client{
		Timeout: 5 * time.Minute, // 5 minutes pour les uploads de fichiers volumineux
	}

	resp, err := uploadClient.Do(req)
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

// DeleteCandidate supprime un candidat Boond par son ID
func (c *Client) DeleteCandidate(ctx context.Context, candidateID string) error {
	if strings.TrimSpace(candidateID) == "" {
		return fmt.Errorf("candidate ID cannot be empty")
	}

	ep := fmt.Sprintf("%s/api/candidates/%s", c.BaseURL, candidateID)
	fmt.Printf("🗑️  [Boond] DELETE %s\n", ep)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, ep, nil)
	if err != nil {
		return err
	}
	addStdHeaders(req, c.JWT)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		fmt.Printf("❌ [Boond] DELETE error: %v\n", err)
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("📥 [Boond] DELETE status=%d body_len=%d\n", resp.StatusCode, len(body))
	if resp.StatusCode >= 400 {
		if len(body) > 0 {
			b := string(body)
			if len(b) > 400 {
				b = b[:400] + "…(tronqué)"
			}
			fmt.Printf("⚠️  [Boond] DELETE error body: %s\n", b)
		}
		return fmt.Errorf("boond delete candidate failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	fmt.Printf("✅ [Boond] Candidate deleted id=%s\n", candidateID)
	return nil
}

// UpdateCandidate met à jour un candidat Boond avec les attributs fournis
func (c *Client) UpdateCandidate(ctx context.Context, candidateID string, attributes map[string]any) (json.RawMessage, error) {
	if strings.TrimSpace(candidateID) == "" {
		return nil, fmt.Errorf("candidate ID cannot be empty")
	}

	payload := map[string]any{
		"data": map[string]any{
			"id":         candidateID,
			"type":       "candidate",
			"attributes": attributes,
		},
	}
	buf, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	ep := fmt.Sprintf("%s/api/candidates/%s/information", c.BaseURL, candidateID)
	fmt.Printf("📝 [Boond] PUT %s (payload_len=%d)\n", ep, len(buf))

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, ep, bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	addStdHeaders(req, c.JWT)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		fmt.Printf("❌ [Boond] PATCH error: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("📥 [Boond] PUT status=%d body_len=%d\n", resp.StatusCode, len(body))
	if resp.StatusCode >= 400 {
		if len(body) > 0 {
			b := string(body)
			if len(b) > 600 {
				b = b[:600] + "…(tronqué)"
			}
			fmt.Printf("⚠️  [Boond] PUT error body: %s\n", b)
		}
		return body, fmt.Errorf("boond update candidate failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	fmt.Printf("✅ [Boond] Candidate updated id=%s\n", candidateID)
	return body, nil
}

// GetAvailabilityTypes récupère la liste des types de disponibilité depuis Boond
func (c *Client) GetAvailabilityTypes(ctx context.Context) (map[string]string, error) {
	ep := fmt.Sprintf("%s/api/application/dictionary/setting/availability", c.BaseURL)
	fmt.Printf("📅 [Boond] GET %s\n", ep)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ep, nil)
	if err != nil {
		return nil, err
	}
	addStdHeaders(req, c.JWT)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		fmt.Printf("❌ [Boond] GET availability types error: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("📥 [Boond] GET availability types status=%d body_len=%d\n", resp.StatusCode, len(body))
	if resp.StatusCode >= 400 {
		// Les endpoints de dictionnaire peuvent ne pas exister (404) - ce n'est pas critique
		if resp.StatusCode == 404 {
			fmt.Printf("ℹ️  [Boond] Endpoint availability types non disponible (404) - ignoré\n")
			return make(map[string]string), nil // Retourner une map vide au lieu d'une erreur
		}
		if len(body) > 0 {
			b := string(body)
			if len(b) > 400 {
				b = b[:400] + "…(tronqué)"
			}
			fmt.Printf("⚠️  [Boond] GET availability types error body: %s\n", b)
		}
		return nil, fmt.Errorf("boond get availability types failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	// Parser la réponse pour extraire les types de disponibilité
	var dictResp struct {
		Data []struct {
			ID   string `json:"id"`
			Type string `json:"type"`
			Attr struct {
				Name string `json:"name"`
			} `json:"attributes"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &dictResp); err != nil {
		return nil, fmt.Errorf("failed to parse availability types response: %w", err)
	}

	// Créer un mapping nom -> ID
	availabilityMap := make(map[string]string)
	for _, avail := range dictResp.Data {
		if avail.Attr.Name != "" {
			availabilityMap[avail.Attr.Name] = avail.ID
			fmt.Printf("📅 [Boond] Type trouvé: %s -> ID %s\n", avail.Attr.Name, avail.ID)
		}
	}

	fmt.Printf("✅ [Boond] %d types de disponibilité récupérés\n", len(availabilityMap))
	return availabilityMap, nil
}

// GetMobilityAreas récupère la liste des zones de mobilité depuis Boond
func (c *Client) GetMobilityAreas(ctx context.Context) (map[string]string, error) {
	ep := fmt.Sprintf("%s/api/application/dictionary/setting/mobilityArea", c.BaseURL)
	fmt.Printf("🗺️  [Boond] GET %s\n", ep)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ep, nil)
	if err != nil {
		return nil, err
	}
	addStdHeaders(req, c.JWT)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		fmt.Printf("❌ [Boond] GET mobility areas error: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("📥 [Boond] GET mobility areas status=%d body_len=%d\n", resp.StatusCode, len(body))
	if resp.StatusCode >= 400 {
		// Les endpoints de dictionnaire peuvent ne pas exister (404) - ce n'est pas critique
		if resp.StatusCode == 404 {
			fmt.Printf("ℹ️  [Boond] Endpoint mobility areas non disponible (404) - ignoré\n")
			return make(map[string]string), nil // Retourner une map vide au lieu d'une erreur
		}
		if len(body) > 0 {
			b := string(body)
			if len(b) > 400 {
				b = b[:400] + "…(tronqué)"
			}
			fmt.Printf("⚠️  [Boond] GET mobility areas error body: %s\n", b)
		}
		return nil, fmt.Errorf("boond get mobility areas failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	// Parser la réponse pour extraire les zones de mobilité
	var dictResp struct {
		Data []struct {
			ID   string `json:"id"`
			Type string `json:"type"`
			Attr struct {
				Name string `json:"name"`
			} `json:"attributes"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &dictResp); err != nil {
		return nil, fmt.Errorf("failed to parse mobility areas response: %w", err)
	}

	// Créer un mapping nom -> ID
	mobilityMap := make(map[string]string)
	for _, area := range dictResp.Data {
		if area.Attr.Name != "" {
			mobilityMap[area.Attr.Name] = area.ID
			fmt.Printf("🗺️  [Boond] Zone trouvée: %s -> ID %s\n", area.Attr.Name, area.ID)
		}
	}

	fmt.Printf("✅ [Boond] %d zones de mobilité récupérées\n", len(mobilityMap))
	return mobilityMap, nil
}

// GetAgencies récupère la liste des agences depuis Boond
func (c *Client) GetAgencies(ctx context.Context) ([]map[string]string, error) {
	ep := fmt.Sprintf("%s/api/agencies", c.BaseURL)
	fmt.Printf("🏢 [Boond] GET %s\n", ep)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ep, nil)
	if err != nil {
		return nil, err
	}
	addStdHeaders(req, c.JWT)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		fmt.Printf("❌ [Boond] GET agencies error: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("📥 [Boond] GET agencies status=%d body_len=%d\n", resp.StatusCode, len(body))
	if resp.StatusCode >= 400 {
		if len(body) > 0 {
			b := string(body)
			if len(b) > 400 {
				b = b[:400] + "…(tronqué)"
			}
			fmt.Printf("⚠️  [Boond] GET agencies error body: %s\n", b)
		}
		return nil, fmt.Errorf("boond get agencies failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	// Parser la réponse pour extraire les agences
	var agenciesResp struct {
		Data []struct {
			ID   string `json:"id"`
			Type string `json:"type"`
			Attr struct {
				Name string `json:"name"`
			} `json:"attributes"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &agenciesResp); err != nil {
		return nil, fmt.Errorf("failed to parse agencies response: %w", err)
	}

	// Créer une liste d'agences avec ID et nom
	var agencies []map[string]string
	for _, agency := range agenciesResp.Data {
		if agency.Attr.Name != "" {
			agencies = append(agencies, map[string]string{
				"id":   agency.ID,
				"name": agency.Attr.Name,
			})
			fmt.Printf("🏢 [Boond] Agence trouvée: %s -> ID %s\n", agency.Attr.Name, agency.ID)
		}
	}

	fmt.Printf("✅ [Boond] %d agences récupérées\n", len(agencies))
	return agencies, nil
}

// GetAllResources récupère toutes les ressources (consultants, managers, RH, etc.) en paginant
func (c *Client) GetAllResources(ctx context.Context, maxResults int, typeOfFilter []int, isVisibleFilter *bool) ([]map[string]any, error) {
	if maxResults <= 0 || maxResults > 500 {
		maxResults = 500
	}

	var allResources []map[string]any
	page := 1

	for {
		// Construire l'URL avec pagination
		params := url.Values{}
		params.Set("page", fmt.Sprintf("%d", page))
		params.Set("maxResults", fmt.Sprintf("%d", maxResults))
		ep := fmt.Sprintf("%s/api/resources?%s", c.BaseURL, params.Encode())
		fmt.Printf("👥 [Boond] GET %s\n", ep)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, ep, nil)
		if err != nil {
			return nil, err
		}
		addStdHeaders(req, c.JWT)

		resp, err := c.HTTP.Do(req)
		if err != nil {
			fmt.Printf("❌ [Boond] GET resources error: %v\n", err)
			return nil, err
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		fmt.Printf("📥 [Boond] GET resources status=%d body_len=%d\n", resp.StatusCode, len(body))
		if resp.StatusCode >= 400 {
			if len(body) > 0 {
				b := string(body)
				if len(b) > 600 {
					b = b[:600] + "…(tronqué)"
				}
				fmt.Printf("⚠️  [Boond] GET resources error body: %s\n", b)
			}
			return nil, fmt.Errorf("boond get resources failed: status=%d body=%s", resp.StatusCode, string(body))
		}

		// Parser la page
		var pageResp struct {
			Data []struct {
				ID   string         `json:"id"`
				Type string         `json:"type"`
				Attr map[string]any `json:"attributes"`
				Rel  map[string]any `json:"relationships"`
			} `json:"data"`
		}

		if err := json.Unmarshal(body, &pageResp); err != nil {
			return nil, fmt.Errorf("failed to parse resources response: %w", err)
		}

		// Transformer en structure simple et appliquer les filtres
		for _, it := range pageResp.Data {
			res := map[string]any{
				"id":            it.ID,
				"type":          it.Type,
				"attributes":    it.Attr,
				"relationships": it.Rel,
			}

			// Appliquer les filtres
			if shouldIncludeResource(res, typeOfFilter, isVisibleFilter) {
				allResources = append(allResources, res)
			}
		}

		fmt.Printf("✅ [Boond] Page %d: %d ressources\n", page, len(pageResp.Data))

		if len(pageResp.Data) < maxResults {
			// Dernière page
			break
		}
		page++

		// Garde-fou pour éviter les boucles infinies
		if page > 2000 {
			fmt.Printf("⚠️  [Boond] Arrêt pagination de sécurité après %d pages\n", page)
			break
		}
	}

	fmt.Printf("👥 [Boond] Total ressources agrégées: %d\n", len(allResources))
	return allResources, nil
}

// shouldIncludeResource vérifie si une ressource doit être incluse selon les filtres
func shouldIncludeResource(resource map[string]any, typeOfFilter []int, isVisibleFilter *bool) bool {
	attrs, ok := resource["attributes"].(map[string]any)
	if !ok {
		return false
	}

	// Filtre par typeOf
	if len(typeOfFilter) > 0 {
		typeOf, ok := attrs["typeOf"]
		if !ok {
			return false
		}

		var typeOfInt int
		switch v := typeOf.(type) {
		case float64:
			typeOfInt = int(v)
		case int:
			typeOfInt = v
		default:
			return false
		}

		found := false
		for _, allowedType := range typeOfFilter {
			if typeOfInt == allowedType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Filtre par isVisible
	if isVisibleFilter != nil {
		isVisible, ok := attrs["isVisible"].(bool)
		if !ok {
			return false
		}
		if isVisible != *isVisibleFilter {
			return false
		}
	}

	return true
}

// FindTesterResourceByEmail trouve une ressource par email (mêmes filtres que "Tester Ressources")
// puis retourne l'ID du manager Boond rattaché à cette ressource.
func (c *Client) FindTesterResourceByEmail(ctx context.Context, userEmail string) (string, error) {
	if strings.TrimSpace(userEmail) == "" {
		return "", fmt.Errorf("user email cannot be empty")
	}

	// Reproduire les filtres utilisés par le bouton "Tester Ressources":
	// typeOf = [2,4,5] (managers, direction, RH) et isVisible = true.
	typeFilters := []int{2, 4, 5}
	isVisible := true
	resources, err := c.GetAllResources(ctx, 500, typeFilters, &isVisible)
	if err != nil {
		return "", fmt.Errorf("failed to get resources: %w", err)
	}

	userEmailLower := strings.ToLower(strings.TrimSpace(userEmail))
	fmt.Printf("🔍 [Boond] Recherche ressource pour email: %s (total: %d ressources)\n", userEmail, len(resources))

	// Parcourir les ressources pour trouver une correspondance
	for idx, resource := range resources {
		attrs, ok := resource["attributes"].(map[string]any)
		if !ok {
			continue
		}

		// 🔧 FIX: Extraire les emails de manière plus robuste
		// Vérifier plusieurs champs email potentiels avec gestion de différents types
		emailsToCheck := []string{}

		// Fonction helper pour extraire un email d'un champ (peut être string, array, etc.)
		extractEmail := func(key string) []string {
			var emails []string
			val, exists := attrs[key]
			if !exists {
				return emails
			}

			switch v := val.(type) {
			case string:
				if strings.TrimSpace(v) != "" {
					emails = append(emails, strings.ToLower(strings.TrimSpace(v)))
				}
			case []interface{}:
				// Si c'est un tableau, extraire tous les strings
				for _, item := range v {
					if str, ok := item.(string); ok && strings.TrimSpace(str) != "" {
						emails = append(emails, strings.ToLower(strings.TrimSpace(str)))
					}
				}
			case []string:
				// Si c'est un tableau de strings directement
				for _, str := range v {
					if strings.TrimSpace(str) != "" {
						emails = append(emails, strings.ToLower(strings.TrimSpace(str)))
					}
				}
			}
			return emails
		}

		// Vérifier tous les champs email possibles (case-insensitive)
		emailFields := []string{"email1", "email2", "email", "Email1", "Email2", "Email", "EMAIL1", "EMAIL2", "EMAIL"}
		for _, field := range emailFields {
			emailsToCheck = append(emailsToCheck, extractEmail(field)...)
		}

		// 🔧 DEBUG: Log les champs email trouvés pour debug (première ressource seulement)
		if idx == 0 {
			fmt.Printf("🔍 [Boond] DEBUG - Champs email trouvés dans la première ressource:\n")
			for key, val := range attrs {
				if strings.Contains(strings.ToLower(key), "email") {
					fmt.Printf("   %s: %v (type: %T)\n", key, val, val)
				}
			}
		}

		matched := false
		for _, email := range emailsToCheck {
			if email != "" && email == userEmailLower {
				matched = true
				break
			}
		}

		// Si l'email correspond, retourner l'ID de la ressource (manager)
		if matched {
			title, _ := attrs["title"].(string)
			typeOf := attrs["typeOf"]
			resID, _ := resource["id"].(string)
			fmt.Printf("✅ [Boond] Ressource trouvée: id=%s, email=%s, title=%s, typeOf=%v\n", resID, userEmail, title, typeOf)
			if strings.TrimSpace(resID) == "" {
				fmt.Printf("⚠️  [Boond] Ressource correspondante mais id manquant\n")
				return "", nil
			}

			return resID, nil
		}
	}

	fmt.Printf("⚠️  [Boond] Aucune ressource trouvée pour l'email: %s\n", userEmail)
	return "", nil
}

// getString helper
func getString(attrs map[string]any, key string) string {
	if v, ok := attrs[key].(string); ok {
		return v
	}
	return ""
}

// extractRelationshipID extrait l'ID d'une relation depuis les relationships
func extractRelationshipID(relationships map[string]any, relKey string) string {
	if relationships == nil {
		return ""
	}
	rel, _ := relationships[relKey].(map[string]any)
	if rel == nil {
		return ""
	}
	data := rel["data"]
	if data == nil {
		return ""
	}
	if m, ok := data.(map[string]any); ok {
		if id, ok := m["id"].(string); ok {
			return id
		}
	}
	return ""
}

// maskToken retourne les n premiers caractères du token
func maskToken(t string, n int) string {
	if n <= 0 || len(t) <= n {
		return t
	}
	return t[:n]
}
