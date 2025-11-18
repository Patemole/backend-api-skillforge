package worker

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"

	"backend-api-skillforge/internal/boond"
	"backend-api-skillforge/internal/models"
	"backend-api-skillforge/internal/nuextract"
	"backend-api-skillforge/internal/services"
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
	//log.Printf("🔍 [worker extract_cv] Tentative de récupération d'un job pending...")

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
			//log.Printf("⏸️  [worker extract_cv] Aucun job pending trouvé (tentative %d/5)", tryCounter+1)
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
	linkedinURL, _ := job.Payload["linkedin_url"].(string)

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

	var numPages int
	reader := bytes.NewReader(fileBytes)
	pdfReader, err := pdf.NewReader(reader, int64(len(fileBytes)))
	if err == nil {
		numPages = pdfReader.NumPage()
		log.Printf("📄 [worker extract_cv] PDF détecté: %d pages (fichier: %s)", numPages, filename)
		log.Printf("🔍 [worker extract_cv] Condition numPages > 10: %v (numPages=%d > 10=%v)", numPages > 10, numPages, numPages > 10)
	} else {
		log.Printf("❌ [worker extract_cv] Erreur lecture PDF: %v (fichier: %s)", err, filename)
		numPages = 0
		log.Printf("⚠️  [worker extract_cv] numPages défini à 0 à cause de l'erreur, needHaiku sera false (Sonnet par défaut)")
	}

	// Démarrer l'appel Apify en parallèle si linkedin_url et config présents
	var (
		apifyRes  map[string]any
		apifyErr  error
		apifyDone chan struct{}
	)
	if strings.TrimSpace(linkedinURL) != "" &&
		(strings.TrimSpace(os.Getenv("APIFY_TOKEN")) != "" || strings.TrimSpace(os.Getenv("APIFY_API_TOKEN")) != "") &&
		strings.TrimSpace(os.Getenv("APIFY_ACTOR_ID")) != "" {
		apifyDone = make(chan struct{})
		go func(url string) {
			defer close(apifyDone)
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
			defer cancel()
			res, err := services.FetchLinkedInProfile(ctx, url)
			if err != nil {
				apifyErr = err
				return
			}
			apifyRes = res
		}(linkedinURL)
	}

	// Sélection moteur selon generationMode
	var (
		resultBytes  []byte
		providerUsed string
		attempts     []map[string]any
		totalStart   = time.Now()
	)
	if generationMode == "fast" && strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY")) != "" {
		// Mode rapide: tenter Anthropic en premier
		// Par défaut, utiliser Sonnet. Si le PDF a moins de 10 pages, utiliser Haiku.
		needHaiku := false
		if numPages < 10 {
			needHaiku = true
			log.Printf("ℹ️  [worker extract_cv] Fichier petit (%d pages < 10) → needHaiku = true → utilisation de Haiku", numPages)
		} else {
			log.Printf("✅ [worker extract_cv] Fichier volumineux (%d pages >= 10) → needHaiku = false → utilisation de claude-sonnet-4-5-20250929", numPages)
		}
		acfg := nuextract.GetAnthropicConfig(!needHaiku)
		log.Printf("⚡ [worker extract_cv] Modèle OpenAI (fallback potentiel): gpt-5-mini (generationMode=%s)", generationMode)
		log.Printf("🚀 [worker extract_cv] Essai 1: Anthropic (modèle sélectionné: %s)", acfg.Model)
		log.Printf("🔧 [worker extract_cv] Appel Anthropic avec paramètres: needHaiku=%v, numPages=%d, filename=%s", needHaiku, numPages, filename)
		callStart := time.Now()
		resultBytes, err = nuextract.ExtractAndEnrichWithFilenameAnthropic(fileBytes, filename, language, needHaiku)
		attempts = append(attempts, map[string]any{
			"provider":    "anthropic",
			"model":       acfg.Model,
			"duration_ms": time.Since(callStart).Milliseconds(),
			"error":       boolToString(err != nil),
		})
		if err != nil || len(strings.TrimSpace(string(resultBytes))) == 0 {
			if err != nil {
				log.Printf("❌ [worker extract_cv] Anthropic a échoué: %v", err)
			} else {
				log.Printf("❌ [worker extract_cv] Anthropic a renvoyé un contenu vide")
			}
			// Fallback OpenAI (gpt-5-mini)
			model := "gpt-5-mini"
			log.Printf("🛟 [worker extract_cv] Fallback OpenAI (%s)", model)
			client := nuextract.NewWithModel(model)
			callStart2 := time.Now()
			resultBytes, err = client.ExtractAndEnrichWithFilename(fileBytes, filename, language)
			attempts = append(attempts, map[string]any{
				"provider":    "openai",
				"model":       model,
				"duration_ms": time.Since(callStart2).Milliseconds(),
				"error":       boolToString(err != nil),
			})
			if err != nil {
				log.Printf("❌ [worker extract_cv] Échec fallback OpenAI: %v", err)
			}
			providerUsed = "openai"
		}
		if err == nil && len(strings.TrimSpace(string(resultBytes))) > 0 && providerUsed == "" {
			providerUsed = "anthropic"
		}
	} else {
		// Mode détaillé (ou absence de clé Anthropic): OpenAI d'abord
		model := "gpt-5"
		if generationMode == "fast" {
			model = "gpt-5-mini"
		}
		log.Printf("⚡ [worker extract_cv] Modèle choisi: %s (generationMode=%s)", model, generationMode)
		client := nuextract.NewWithModel(model)

		// Extraction initiale
		callStart := time.Now()
		resultBytes, err = client.ExtractAndEnrichWithFilename(fileBytes, filename, language)
		attempts = append(attempts, map[string]any{
			"provider":    "openai",
			"model":       model,
			"duration_ms": time.Since(callStart).Milliseconds(),
			"error":       boolToString(err != nil),
		})
		if err != nil {
			log.Printf("❌ [worker extract_cv] Échec extraction initiale: %v", err)
		}
		providerUsed = "openai"
	}

	// Helper pour vérifier vide après nettoyage
	isEmptyAfterClean := func(b []byte) bool {
		cleaned := sanitizeJSONResponse(b)
		return len(strings.TrimSpace(string(cleaned))) == 0
	}

	if err != nil || isEmptyAfterClean(resultBytes) {
		reason := "empty_json"
		if err != nil {
			reason = "error"
		}
		log.Printf("🔁 [worker extract_cv] Retry extraction (reason=%s)", reason)
		time.Sleep(800 * time.Millisecond)

		// 2) Fallback Anthropic direct si ANTHROPIC_API_KEY configuré
		if strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY")) != "" {
			log.Printf("🛟 [worker extract_cv] Fallback Anthropic déclenché (essai 2)")
			time.Sleep(400 * time.Millisecond)
			needHaikuFallback := false
			if numPages < 10 {
				needHaikuFallback = true
				log.Printf("ℹ️  [worker extract_cv] [FALLBACK] Fichier petit (%d pages < 10) → needHaiku = true → utilisation de Haiku", numPages)
			} else {
				log.Printf("✅ [worker extract_cv] [FALLBACK] Fichier volumineux (%d pages >= 10) → needHaiku = false → utilisation de claude-sonnet-4-5-20250929", numPages)
			}
			log.Printf("🔧 [worker extract_cv] [FALLBACK] Appel Anthropic avec paramètres: needHaiku=%v, numPages=%d, filename=%s", needHaikuFallback, numPages, filename)
			resultBytes, err = nuextract.ExtractAndEnrichWithFilenameAnthropic(fileBytes, filename, language, needHaikuFallback)
			if err != nil {
				log.Printf("❌ [worker extract_cv] Échec fallback Anthropic: %v", err)
			} else {
				log.Printf("✅ [worker extract_cv] Fallback Anthropic réussi (len=%d)", len(resultBytes))
			}
		}
	}

	if err != nil {
		// Inclure la langue même en échec, pour cohérence côté front
		failResult := map[string]any{}
		if strings.TrimSpace(language) != "" {
			failResult["DC_language"] = language
		}
		_, _, _ = supabase.Client.
			From("jobs").
			Update(map[string]any{
				"status":     "failed",
				"error":      err.Error(),
				"result":     failResult,
				"updated_at": time.Now().UTC().Format(time.RFC3339),
			}, "representation", "").
			Eq("id", idStr).
			Execute()
		return nil
	}

	// ANALYSE COMPLÈTE DE LA RÉPONSE DU PROVIDER
	log.Printf("⏱️  [worker extract_cv] Metrics job %s → provider=%s, total=%v, attempts=%d", idStr, providerUsed, time.Since(totalStart), len(attempts))
	log.Printf("🔍 [worker extract_cv] ANALYSE RÉPONSE provider pour job %s:", idStr)
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
	parsedOK := false
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
		parsedOK = true
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

	// Intégration Boond si un JWT est présent dans le payload (uniquement si parsing OK et CV valide)
	boondJWT, _ := job.Payload["boondJwt"].(string)
	invitedBy, _ := job.Payload["invited_by"].(string)
	if strings.TrimSpace(boondJWT) != "" || strings.TrimSpace(invitedBy) != "" {
		if !parsedOK {
			log.Printf("⏭️  [worker extract_cv] Boond ignoré: JSON non parsé (job %s)", idStr)
		} else {
			// Construire les attributs Boond à partir du résultat
			var cv nuextract.CVExtractionSchema
			if raw, err := json.Marshal(result); err == nil {
				_ = json.Unmarshal(raw, &cv)
			}

			hasIdentity := strings.TrimSpace(cv.Email) != "" ||
				(strings.TrimSpace(cv.Prenom) != "" && strings.TrimSpace(cv.Nom) != "")

			if !hasIdentity {
				log.Printf("⏭️  [worker extract_cv] Boond ignoré: identité incomplète (email ou prénom+nom requis) (job %s)", idStr)
			} else {
				// Si invited_by est présent, utiliser les tokens du manager
				finalJWT := boondJWT
				if strings.TrimSpace(invitedBy) != "" {
					log.Printf("🔎 [worker extract_cv] invited_by détecté → tentative récupération tokens manager (user_id=%s)", invitedBy)

					// Récupérer l'organization_id depuis le profil de l'utilisateur
					orgID := ""
					userProfileData, _, err := supabase.Client.
						From("profiles").
						Select("organization_id", "exact", false).
						Eq("user_id", job.UserID.String()).
						Limit(1, "").
						Execute()
					if err == nil {
						var profiles []map[string]any
						if err := json.Unmarshal(userProfileData, &profiles); err == nil && len(profiles) > 0 {
							if orgIDRaw, ok := profiles[0]["organization_id"].(string); ok {
								orgID = orgIDRaw
							}
						}
					}

					if orgID != "" {
						integration, err := boond.GetIntegrationForUserOrOrganization(invitedBy, orgID)
						if err == nil && integration != nil {
							generatedJWT, err := boond.GenerateJWT(integration)
							if err == nil && strings.TrimSpace(generatedJWT) != "" {
								finalJWT = generatedJWT
								log.Printf("✅ [worker extract_cv] JWT généré depuis tokens manager (invited_by=%s)", invitedBy)
							} else {
								log.Printf("⚠️  [worker extract_cv] Échec génération JWT depuis tokens manager, utilisation JWT frontend")
							}
						} else {
							log.Printf("⚠️  [worker extract_cv] Aucun token manager trouvé pour invited_by=%s, utilisation JWT frontend", invitedBy)
						}
					} else {
						log.Printf("⚠️  [worker extract_cv] Impossible de récupérer organization_id, utilisation JWT frontend")
					}
				}

				if strings.TrimSpace(finalJWT) == "" {
					log.Printf("⏭️  [worker extract_cv] Aucun JWT disponible (ni frontend ni manager), Boond ignoré")
				} else {
					log.Printf("🚀 [worker extract_cv] boondJwt détecté → création candidat Boond")
					ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
					defer cancel()

					bClient := boond.New(finalJWT)
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
			}
		}
	}

	// Ajouter toujours la langue utilisée dans le résultat
	if strings.TrimSpace(language) != "" {
		result["DC_language"] = language
	}

	// Attendre la fin d'Apify (si déclenché) avant mise à jour du job, puis merger
	if apifyDone != nil {
		<-apifyDone
		if result == nil {
			result = map[string]any{}
		}
		if apifyErr != nil {
			result["_linkedin_error"] = apifyErr.Error()
		} else if apifyRes != nil {
			result["linkedin_profile"] = apifyRes
		}
	}

	// Ajouter métriques au résultat
	metrics := map[string]any{
		"provider":        providerUsed,
		"generation_mode": generationMode,
		"total_s":         time.Since(totalStart).Seconds(),
		"attempts":        attempts,
	}
	if result == nil {
		result = map[string]any{}
	}
	result["_metrics"] = metrics

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

// boolToString convertit un booléen en chaîne "true"/"false" pour homogénéité JSON
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
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
