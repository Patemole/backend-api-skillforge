package boond

import (
	"os"
	"strconv"
	"strings"
)

// BoondConfig contient la configuration pour les appels Boond
type BoondConfig struct {
	// DefaultCandidateState est l'ID de l'étape par défaut lors de la création d'un candidat
	// Par défaut: 4 (E1 - Posé)
	DefaultCandidateState int `json:"default_candidate_state"`
}

// GetBoondConfig retourne la configuration Boond (avec overrides via env)
func GetBoondConfig() BoondConfig {
	cfg := BoondConfig{
		DefaultCandidateState: 4, // Par défaut: E1 - Posé (ID 4)
	}

	// Permettre l'override via variable d'environnement
	if v := strings.TrimSpace(os.Getenv("BOOND_DEFAULT_CANDIDATE_STATE")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.DefaultCandidateState = n
		}
	}

	return cfg
}

