package boond

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log"
	"strings"
	"time"

	"backend-api-skillforge/internal/supabase"
)

// BoondIntegration represents the Boond integration tokens
type BoondIntegration struct {
	TokenClient string `json:"tokenClient"`
	KeyClient   string `json:"keyClient"`
	TokenUser   string `json:"tokenUser"`
}

// GetIntegrationForUserOrOrganization fetches Boond tokens from the database
// It first tries to get tokens from the manager's profile (profiles.boond_manager),
// then falls back to organization tokens (organizations.boond_integration)
func GetIntegrationForUserOrOrganization(userID string, organizationID string) (*BoondIntegration, error) {
	// Try manager's tokens first if userID is provided
	if strings.TrimSpace(userID) != "" {
		log.Printf("🔎 [Boond] Tentative récupération tokens manager (user_id=%s)", userID)
		managerIntegration, err := getIntegrationForUser(userID)
		if err == nil && managerIntegration != nil {
			log.Printf("✅ [Boond] Tokens manager trouvés")
			return managerIntegration, nil
		}
		if err != nil {
			log.Printf("⚠️  [Boond] Erreur récupération tokens manager: %v", err)
		}
		log.Printf("ℹ️  [Boond] Aucun token manager trouvé, utilisation organisation")
	}

	// Fallback to organization tokens
	log.Printf("🔎 [Boond] Récupération tokens organisation (organization_id=%s)", organizationID)
	orgIntegration, err := getIntegrationForOrganization(organizationID)
	if err != nil {
		log.Printf("❌ [Boond] Erreur récupération tokens organisation: %v", err)
		return nil, err
	}
	if orgIntegration != nil {
		log.Printf("✅ [Boond] Tokens organisation trouvés")
		return orgIntegration, nil
	}

	log.Printf("⚠️  [Boond] Aucune intégration Boond trouvée (ni manager ni organisation)")
	return nil, nil
}

// getIntegrationForUser fetches Boond tokens from profiles.boond_manager
func getIntegrationForUser(userID string) (*BoondIntegration, error) {
	data, _, err := supabase.Client.
		From("profiles").
		Select("boond_manager", "exact", false).
		Eq("user_id", userID).
		Limit(1, "").
		Execute()

	if err != nil {
		return nil, err
	}

	var profiles []map[string]any
	if err := json.Unmarshal(data, &profiles); err != nil {
		return nil, err
	}

	if len(profiles) == 0 {
		return nil, nil
	}

	boondManagerRaw := profiles[0]["boond_manager"]
	if boondManagerRaw == nil {
		return nil, nil
	}

	// Convert to JSON and unmarshal
	boondManagerBytes, err := json.Marshal(boondManagerRaw)
	if err != nil {
		return nil, err
	}

	var integration BoondIntegration
	if err := json.Unmarshal(boondManagerBytes, &integration); err != nil {
		return nil, err
	}

	// Validate that all required fields are present
	if strings.TrimSpace(integration.TokenClient) == "" ||
		strings.TrimSpace(integration.KeyClient) == "" ||
		strings.TrimSpace(integration.TokenUser) == "" {
		return nil, nil
	}

	return &integration, nil
}

// getIntegrationForOrganization fetches Boond tokens from organizations.boond_integration
func getIntegrationForOrganization(organizationID string) (*BoondIntegration, error) {
	data, _, err := supabase.Client.
		From("organizations").
		Select("boond_integration", "exact", false).
		Eq("id", organizationID).
		Limit(1, "").
		Execute()

	if err != nil {
		return nil, err
	}

	var orgs []map[string]any
	if err := json.Unmarshal(data, &orgs); err != nil {
		return nil, err
	}

	if len(orgs) == 0 {
		return nil, nil
	}

	boondIntegrationRaw := orgs[0]["boond_integration"]
	if boondIntegrationRaw == nil {
		return nil, nil
	}

	// Convert to JSON and unmarshal
	boondIntegrationBytes, err := json.Marshal(boondIntegrationRaw)
	if err != nil {
		return nil, err
	}

	var integration BoondIntegration
	if err := json.Unmarshal(boondIntegrationBytes, &integration); err != nil {
		return nil, err
	}

	// Validate that all required fields are present
	if strings.TrimSpace(integration.TokenClient) == "" ||
		strings.TrimSpace(integration.KeyClient) == "" ||
		strings.TrimSpace(integration.TokenUser) == "" {
		return nil, nil
	}

	return &integration, nil
}

// GenerateJWT generates a Boond JWT (HS256) from integration tokens
func GenerateJWT(integration *BoondIntegration) (string, error) {
	if integration == nil {
		return "", nil
	}

	// Header
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)

	// Payload
	payload := map[string]any{
		"userToken":   integration.TokenUser,
		"clientToken": integration.TokenClient,
		"time":        time.Now().Unix(),
		"mode":        "normal",
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	payloadEncoded := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// Signature
	data := headerEncoded + "." + payloadEncoded
	mac := hmac.New(sha256.New, []byte(integration.KeyClient))
	mac.Write([]byte(data))
	signature := mac.Sum(nil)
	signatureEncoded := base64.RawURLEncoding.EncodeToString(signature)

	jwt := data + "." + signatureEncoded
	return jwt, nil
}
