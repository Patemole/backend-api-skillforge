package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// FetchLinkedInProfile lance un actor Apify pour scrapper un profil LinkedIn
// et renvoie un objet JSON contenant les items du dataset et des métadonnées.
// Requiert APIFY_TOKEN et APIFY_ACTOR_ID dans l'environnement.
func FetchLinkedInProfile(ctx context.Context, profileURL string) (map[string]any, error) {
	profileURL = strings.TrimSpace(profileURL)
	if profileURL == "" {
		return nil, fmt.Errorf("linkedin_url manquant")
	}

	// Récupération des variables d'environnement
	token := strings.TrimSpace(os.Getenv("APIFY_TOKEN"))
	if token == "" {
		// fallback éventuel
		token = strings.TrimSpace(os.Getenv("APIFY_API_TOKEN"))
		log.Printf("📝 [Apify] APIFY_TOKEN non trouvé, utilisation de APIFY_API_TOKEN")
	}
	actorID := strings.TrimSpace(os.Getenv("APIFY_ACTOR_ID"))

	// Log des valeurs chargées (sans exposer le token complet)
	tokenPreview := "NON_TROUVÉ"
	tokenLen := 0
	if token != "" {
		if len(token) > 10 {
			tokenPreview = token[:10] + "..."
		} else {
			tokenPreview = "***"
		}
		tokenLen = len(token)
	}

	actorIDDisplay := actorID
	if actorID == "" {
		actorIDDisplay = "NON_TROUVÉ"
	}

	log.Printf("📝 [Apify] Variables d'environnement chargées:")
	log.Printf("   - APIFY_TOKEN: %s (longueur=%d)", tokenPreview, tokenLen)
	log.Printf("   - APIFY_ACTOR_ID: '%s' (longueur=%d)", actorIDDisplay, len(actorID))
	log.Printf("   - Profile URL: %s", profileURL)

	if token == "" || actorID == "" {
		return nil, fmt.Errorf("APIFY_TOKEN ou APIFY_ACTOR_ID non configuré (token=%t, actorID=%t)", token != "", actorID != "")
	}

	// Vérification du format de l'actor ID
	// Apify accepte soit un ID court (ex: "2SyF0bVxmgGr8IVCZ") soit un nom complet (ex: "username/actor-name")
	// Le format avec ~ n'est pas valide pour l'API
	if strings.Contains(actorID, "~") {
		log.Printf("⚠️  [Apify] ATTENTION: Actor ID contient '~' - format invalide!")
		log.Printf("   Format attendu: 'username/actor-name' ou 'actor-id'")
		log.Printf("   Valeur actuelle: '%s'", actorID)
		// On essaie de convertir dev_fusion~linkedin-profile-scraper en dev_fusion/linkedin-profile-scraper
		actorID = strings.ReplaceAll(actorID, "~", "/")
		log.Printf("   Tentative de correction: '%s'", actorID)
	}

	log.Printf("🔍 [Apify] Démarrage avec Actor ID: %s", actorID)

	base := "https://api.apify.com/v2"

	// 1) Start run
	body := map[string]any{
		"profileUrls": []string{profileURL},
	}
	jsonBody, _ := json.Marshal(body)
	startURL := fmt.Sprintf("%s/acts/%s/runs?token=%s", base, url.PathEscape(actorID), url.QueryEscape(token))
	log.Printf("🔍 [Apify] POST URL: %s/acts/%s/runs?token=%s...", base, actorID, tokenPreview)
	req, err := http.NewRequestWithContext(ctx, "POST", startURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		log.Printf("❌ [Apify] Erreur HTTP %d - Actor ID utilisé: %s", resp.StatusCode, actorID)
		log.Printf("❌ [Apify] Réponse complète: %s", string(respBytes))
		return nil, fmt.Errorf("apify start error %d: %s", resp.StatusCode, string(respBytes))
	}
	var start struct {
		Data map[string]any `json:"data"`
	}
	_ = json.Unmarshal(respBytes, &start)
	runID, _ := valueString(start.Data["id"])
	if strings.TrimSpace(runID) == "" {
		return nil, fmt.Errorf("apify: run id manquant dans la réponse")
	}

	// 2) Poll run status jusqu'à SUCCEEDED/FAILED/ABORTED/TIMED_OUT
	deadline := time.Now().Add(8 * time.Minute)
	var datasetID string
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("apify: timeout du run")
		}
		runURL := fmt.Sprintf("%s/actor-runs/%s?token=%s", base, url.PathEscape(runID), url.QueryEscape(token))
		rReq, _ := http.NewRequestWithContext(ctx, "GET", runURL, nil)
		rResp, err := client.Do(rReq)
		if err != nil {
			return nil, err
		}
		b, _ := io.ReadAll(rResp.Body)
		rResp.Body.Close()
		if rResp.StatusCode >= 400 {
			return nil, fmt.Errorf("apify run status error %d: %s", rResp.StatusCode, string(b))
		}
		var run struct {
			Data map[string]any `json:"data"`
		}
		_ = json.Unmarshal(b, &run)
		status, _ := valueString(run.Data["status"])
		// dataset id peut être "defaultDatasetId" ou "defaultDatasetId" selon payload
		datasetID, _ = valueString(run.Data["defaultDatasetId"]) // clé standard
		if datasetID == "" {
			// certains payloads utilisent datasetId
			datasetID, _ = valueString(run.Data["datasetId"])
		}
		s := strings.ToUpper(strings.TrimSpace(status))
		switch s {
		case "SUCCEEDED":
			if strings.TrimSpace(datasetID) == "" {
				return nil, fmt.Errorf("apify: datasetId manquant malgré SUCCEEDED")
			}
			goto FETCH
		case "FAILED", "ABORTED", "TIMED_OUT":
			return nil, fmt.Errorf("apify: run terminé avec statut %s", s)
		default:
			time.Sleep(2 * time.Second)
		}
	}

FETCH:
	// 3) Récupérer le dataset items
	itemsURL := fmt.Sprintf("%s/datasets/%s/items?format=json&clean=1&token=%s", base, url.PathEscape(datasetID), url.QueryEscape(token))
	itReq, _ := http.NewRequestWithContext(ctx, "GET", itemsURL, nil)
	itResp, err := client.Do(itReq)
	if err != nil {
		return nil, err
	}
	defer itResp.Body.Close()
	itBody, _ := io.ReadAll(itResp.Body)
	if itResp.StatusCode >= 400 {
		return nil, fmt.Errorf("apify dataset error %d: %s", itResp.StatusCode, string(itBody))
	}
	var items []map[string]any
	if err := json.Unmarshal(itBody, &items); err != nil {
		return nil, fmt.Errorf("apify: parse dataset items: %w", err)
	}

	result := map[string]any{
		"source":      "apify-linkedin",
		"profile_url": profileURL,
		"run_id":      runID,
		"dataset_id":  datasetID,
		"fetched_at":  time.Now().UTC().Format(time.RFC3339),
		"items":       items,
	}
	return result, nil
}

func valueString(v any) (string, bool) {
	switch t := v.(type) {
	case string:
		return t, true
	case fmt.Stringer:
		return t.String(), true
	default:
		b, err := json.Marshal(v)
		if err == nil && len(b) > 0 {
			return string(b), true
		}
	}
	return "", false
}
