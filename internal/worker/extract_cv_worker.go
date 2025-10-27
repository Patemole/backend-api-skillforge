package worker

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"backend-api-skillforge/internal/models"
	"backend-api-skillforge/internal/nuextract"
	"backend-api-skillforge/internal/supabase"
)

// StartExtractCVWorker lance une boucle qui traite les jobs extract_cv
func StartExtractCVWorker() {
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		log.Printf("✅ [worker extract_cv] Worker démarré et en attente de jobs...")

		for range ticker.C {
			if err := processOneExtractJob(); err != nil {
				log.Printf("[worker extract_cv] erreur cycle: %v", err)
			}
		}
	}()
}

func processOneExtractJob() error {
	// Essayez de "réserver" un job pending via un CAS (compare-and-swap):
	// 1) Lire un pending
	// 2) Tenter Update où status = 'pending' (si 0 ligne, quelqu'un d'autre l'a pris → recommencer)
	log.Printf("🔍 [worker extract_cv] Tentative de récupération d'un job pending...")

	var (
		job        models.Job
		idStr      string
		claimed    bool
		tryCounter int
	)

	for tryCounter = 0; tryCounter < 5; tryCounter++ {
		data, _, err := supabase.Client.
			From("jobs").
			Select("*", "exact", false).
			Eq("status", "pending").
			Eq("type", "extract_cv").
			Limit(1, "").
			Execute()
		if err != nil {
			log.Printf("❌ [worker extract_cv] Erreur requête: %v", err)
			return err
		}
		var jobs []models.Job
		if err := json.Unmarshal(data, &jobs); err != nil {
			log.Printf("❌ [worker extract_cv] Erreur parsing: %v", err)
			return err
		}
		if len(jobs) == 0 {
			// Rien à traiter pour le moment
			log.Printf("⏸️  [worker extract_cv] Aucun job pending trouvé (tentative %d/5)", tryCounter+1)
			return nil
		}

		job = jobs[0]
		idStr = strconv.FormatInt(job.ID, 10)

		// CAS: ne passer en processing que si toujours pending
		udata, _, _ := supabase.Client.
			From("jobs").
			Update(map[string]any{
				"status":     "processing",
				"updated_at": time.Now().UTC().Format(time.RFC3339),
			}, "representation", "").
			Eq("id", idStr).
			Eq("status", "pending").
			Execute()

		var updated []models.Job
		_ = json.Unmarshal(udata, &updated)
		if len(updated) == 0 {
			// Perdu la course, réessayer immédiatement pour prendre le suivant
			log.Printf("⚠️  [worker extract_cv] Job %s déjà pris, on réessaie...", idStr)
			continue
		}

		// Réservation réussie
		claimed = true
		log.Printf("✅ [worker extract_cv] Job réservé id=%s → processing", idStr)
		break
	}

	if !claimed {
		// Rien réservé après plusieurs tentatives
		return nil
	}

	// Extraire les champs utiles
	filename, _ := job.Payload["filename"].(string)
	fileB64, _ := job.Payload["file_base64"].(string)
	language, _ := job.Payload["language"].(string)
	generationMode, _ := job.Payload["generationMode"].(string)
	if language == "" {
		language = "fr"
	}

	// Décodage
	fileBytes, err := base64.StdEncoding.DecodeString(fileB64)
	if err != nil {
		// Échec → failed
		_, _, _ = supabase.Client.
			From("jobs").
			Update(map[string]any{
				"status":     "failed",
				"error":      "invalid base64 file",
				"updated_at": time.Now().UTC().Format(time.RFC3339),
			}, "representation", "").
			Eq("id", idStr).
			Execute()
		return nil
	}

	// Choix du modèle
	model := "gpt-5"
	if generationMode == "fast" {
		model = "gpt-5-mini"
	}
	log.Printf("⚡ [worker extract_cv] Modèle choisi: %s (generationMode=%s)", model, generationMode)
	client := nuextract.NewWithModel(model)

	// Extraction
	resultBytes, err := client.ExtractAndEnrichWithFilename(fileBytes, filename, language)
	if err != nil {
		_, _, _ = supabase.Client.
			From("jobs").
			Update(map[string]any{
				"status":     "failed",
				"error":      err.Error(),
				"updated_at": time.Now().UTC().Format(time.RFC3339),
			}, "representation", "").
			Eq("id", idStr).
			Execute()
		return nil
	}

	// Nettoyer la réponse d'OpenAI (retirer les backticks ```json ... ```)
	cleanedBytes := sanitizeJSONResponse(resultBytes)

	var result map[string]any
	if err := json.Unmarshal(cleanedBytes, &result); err != nil {
		// Si non JSON, envelopper dans une clé raw
		previewLen := 500
		if len(cleanedBytes) < previewLen {
			previewLen = len(cleanedBytes)
		}
		log.Printf("⚠️  [worker extract_cv] Réponse JSON invalide pour job %s: %v | Premiers %d chars: %s", idStr, err, previewLen, string(cleanedBytes[:previewLen]))
		result = map[string]any{"raw": string(cleanedBytes)}
	} else {
		log.Printf("✅ [worker extract_cv] JSON parsé avec succès pour job %s", idStr)
	}

	// Mettre à jour en done
	_, _, _ = supabase.Client.
		From("jobs").
		Update(map[string]any{
			"status":     "done",
			"result":     result,
			"updated_at": time.Now().UTC().Format(time.RFC3339),
		}, "representation", "").
		Eq("id", idStr).
		Execute()

	log.Printf("[worker extract_cv] job id=%d done", job.ID)
	return nil
}

// sanitizeJSONResponse retire d'éventuels blocs de code ```json ... ``` et extrait uniquement l'objet JSON
func sanitizeJSONResponse(b []byte) []byte {
	s := strings.TrimSpace(string(b))

	// S'il n'y a pas de backticks, renvoyer tel quel
	if !strings.Contains(s, "```") {
		return []byte(s)
	}

	// Essayer de trouver le premier '{' et le dernier '}'
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		inner := s[start : end+1]
		return []byte(inner)
	}
	return b
}
