package worker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"backend-api-skillforge/internal/boond"
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

	// Extraction avec retry intelligent si sortie vide/illisible
	resultBytes, err := client.ExtractAndEnrichWithFilename(fileBytes, filename, language)
	if err != nil {
		log.Printf("❌ [worker extract_cv] Échec extraction initiale: %v", err)
	}

	// Helper pour vérifier vide après nettoyage
	isEmptyAfterClean := func(b []byte) bool {
		cleaned := sanitizeJSONResponse(b)
		return len(strings.TrimSpace(string(cleaned))) == 0
	}

	if err != nil || isEmptyAfterClean(resultBytes) {
		log.Printf("🔁 [worker extract_cv] Retry extraction (reason=%s)", func() string {
			if err != nil {
				return "error"
			}
			return "empty_json"
		}())
		time.Sleep(800 * time.Millisecond)
		resultBytes, err = client.ExtractAndEnrichWithFilename(fileBytes, filename, language)
		if err != nil {
			log.Printf("❌ [worker extract_cv] Échec extraction retry: %v", err)
		} else {
			log.Printf("✅ [worker extract_cv] Retry extraction réussi (len=%d)", len(resultBytes))
		}
	}

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

	// ANALYSE COMPLÈTE DE LA RÉPONSE D'OPENAI
	log.Printf("🔍 [worker extract_cv] ANALYSE RÉPONSE OpenAI pour job %s:", idStr)
	log.Printf("📏 Taille brute: %d bytes", len(resultBytes))

	// Log des premiers caractères pour debug
	previewRaw := 1000
	if len(resultBytes) < previewRaw {
		previewRaw = len(resultBytes)
	}
	log.Printf("👀 Aperçu brut (premiers %d chars): %s", previewRaw, string(resultBytes[:previewRaw]))

	// Nettoyer la réponse d'OpenAI (retirer les backticks ```json ... ```)
	cleanedBytes := sanitizeJSONResponse(resultBytes)
	log.Printf("📏 Taille après nettoyage: %d bytes", len(cleanedBytes))

	// Aperçu après nettoyage
	previewClean := 1000
	if len(cleanedBytes) < previewClean {
		previewClean = len(cleanedBytes)
	}
	log.Printf("👀 Aperçu nettoyé (premiers %d chars): %s", previewClean, string(cleanedBytes[:previewClean]))

	var result map[string]any
	if err := json.Unmarshal(cleanedBytes, &result); err != nil {
		// DEBUG: Si non JSON, on log TOUTE la réponse pour comprendre
		log.Printf("❌ [worker extract_cv] ÉCHEC parsing JSON pour job %s:", idStr)
		log.Printf("❌ Erreur: %v", err)
		log.Printf("❌ Taille réponse: %d bytes", len(cleanedBytes))
		log.Printf("❌ Contenu complet: %s", string(cleanedBytes))

		// Si le contenu est vide, c'est un problème grave
		if len(strings.TrimSpace(string(cleanedBytes))) == 0 {
			log.Printf("💥 ERREUR CRITIQUE: Réponse vide ou None!")
			result = map[string]any{
				"raw":   "",
				"error": "Extraction échouée : aucune donnée extraite",
			}
		} else {
			// Envelopper dans une clé raw pour garder l'info
			result = map[string]any{"raw": string(cleanedBytes)}
		}
	} else {
		log.Printf("✅ [worker extract_cv] JSON parsé avec succès pour job %s (champs: %d)", idStr, len(result))
		// Log les premières clés pour vérifier la structure
		keys := make([]string, 0, len(result))
		for k := range result {
			keys = append(keys, k)
		}
		if len(keys) > 0 {
			maxKeys := 10
			if len(keys) < maxKeys {
				maxKeys = len(keys)
			}
			log.Printf("📋 Champs extraits: %v", keys[:maxKeys])
		}
	}

	// Intégration Boond si un JWT est présent dans le payload (uniquement si parsing ok)
	boondJWT, _ := job.Payload["boondJwt"].(string)
	if strings.TrimSpace(boondJWT) != "" && len(result) > 0 {
		log.Printf("🚀 [worker extract_cv] boondJwt détecté → création candidat Boond")
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		bClient := boond.New(boondJWT)

		// Construire les attributs Boond à partir du résultat
		var cv nuextract.CVExtractionSchema
		if raw, err := json.Marshal(result); err == nil {
			_ = json.Unmarshal(raw, &cv)
		}
		attributes := boond.BuildCandidateAttributesFromCV(cv)

		createdID, _, err := bClient.CreateCandidate(ctx, attributes)
		if err != nil {
			log.Printf("❌ [worker extract_cv] Échec de création du candidat Boond: %v", err)
		} else if strings.TrimSpace(createdID) != "" {
			log.Printf("🎉 [worker extract_cv] Candidat Boond créé (id=%s)", createdID)
			// Upload du CV
			if _, err := bClient.UploadDocument(ctx, createdID, fileBytes, filename); err != nil {
				log.Printf("❌ [worker extract_cv] Échec upload du CV vers Boond: %v", err)
			} else {
				log.Printf("📄 [worker extract_cv] CV uploadé avec succès pour candidat %s", createdID)
			}
			// Enrichir le résultat du job
			result["boond_candidate_id"] = createdID
		}
	}

	// Ajouter toujours la langue utilisée dans le résultat
	if strings.TrimSpace(language) != "" {
		result["DC_language"] = language
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

	// Si la réponse est vide
	if len(s) == 0 {
		return b
	}

	// Nettoyer les backticks de code markdown ```json ... ``` ou ``` ... ```
	if strings.Contains(s, "```") {
		// Trouver le début du JSON (après ```json ou ```)
		start := 0
		if idx := strings.Index(s, "```json"); idx >= 0 {
			start = idx + 7 // Position après ```json
		} else if idx := strings.Index(s, "```"); idx >= 0 {
			start = idx + 3 // Position après ```
		}

		// Trouver la fin
		end := len(s)
		if idx := strings.LastIndex(s, "```"); idx > start {
			end = idx
		}

		if start > 0 && end > start {
			s = s[start:end]
		}
	}

	// Nettoyer les retours à la ligne et espaces en début/fin
	s = strings.TrimSpace(s)

	// Si c'est déjà un objet JSON valide, renvoyer
	if (strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")) ||
		(strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]")) {
		return []byte(s)
	}

	// Chercher le premier '{' et le dernier '}' pour extraire le JSON
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")

	if start >= 0 && end > start {
		inner := s[start : end+1]
		return []byte(strings.TrimSpace(inner))
	}

	// Chercher aussi des tableaux
	start = strings.Index(s, "[")
	end = strings.LastIndex(s, "]")

	if start >= 0 && end > start {
		inner := s[start : end+1]
		return []byte(strings.TrimSpace(inner))
	}

	// Si rien ne fonctionne, renvoyer le contenu original
	return []byte(s)
}
