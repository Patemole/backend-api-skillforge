package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"backend-api-skillforge/internal/boond"
	"backend-api-skillforge/internal/nuextract"

	"github.com/gin-gonic/gin"
)

// ExtractCV traite l'upload d'un CV, renvoie le JSON NuExtract
// et, si un JWT Boond est fourni, déclenche la création du candidat dans Boond Manager.
func ExtractCV(c *gin.Context) {
	log.Printf("DEBUG: Début de l'extraction CV")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		log.Printf("ERROR: Erreur récupération fichier: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "file not provided"})
		return
	}
	defer file.Close()

	log.Printf("DEBUG: Fichier reçu - Nom: %s, Taille: %d", header.Filename, header.Size)

	data, err := io.ReadAll(file)
	if err != nil {
		log.Printf("ERROR: Erreur lecture fichier: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot read file"})
		return
	}

	log.Printf("DEBUG: Données lues - Taille: %d bytes", len(data))

	// Récupération du paramètre language du FormData
	language := strings.TrimSpace(c.PostForm("language"))
	if language == "" {
		language = "fr" // Valeur par défaut si non fournie
	}
	log.Printf("🌍 Langue sélectionnée: %s", language)

	client := nuextract.New()
	log.Printf("DEBUG: Client NuExtract créé, début de l'extraction...")

	result, err := client.ExtractAndEnrichWithFilename(data, header.Filename, language)
	if err != nil {
		log.Printf("ERROR: Erreur extraction NuExtract: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	log.Printf("DEBUG: Extraction réussie, taille résultat: %d bytes", len(result))

	// Lecture optionnelle du JWT pour Boond
	boondJWT := strings.TrimSpace(c.PostForm("boondJwt"))
	log.Printf("🔐 BOOND JWT reçu ? %t", boondJWT != "")
	if boondJWT != "" {
		// ATTENTION: log du JWT complet uniquement pour debug local. À retirer en prod.
		log.Printf("🔐 BOOND JWT (DEV - à retirer en prod) : %s", boondJWT)
		log.Printf("🔐 BOOND JWT aperçu: len=%d prefix=%s… suffix=…%s", len(boondJWT), maskToken(boondJWT, 10), tailToken(boondJWT, 10))
	}
	if boondJWT == "" {
		// Pas d'intégration Boond demandée → tenter une normalisation des dates avant retour
		cleaned := sanitizeJSONResult(result)
		var cv nuextract.CVExtractionSchema
		if err := json.Unmarshal(cleaned, &cv); err == nil {
			normalizeExperienceDates(&cv, language)
			if normalized, err := json.Marshal(cv); err == nil {
				result = normalized
			} else {
				log.Printf("WARNING: Échec du marshalling après normalisation des dates (no JWT): %v", err)
			}
		} else {
			log.Printf("WARNING: JSON non parsable pour normalisation (no JWT): %v", err)
		}
		log.Printf("📤 Réponse normale envoyée au frontend (pas de JWT Boond) - taille: %d bytes", len(result))
		c.Data(http.StatusOK, "application/json", result)
		return
	}

	log.Printf("🚀 boondJwt détecté → tentative de création de candidat dans Boond")

	// On parse le JSON d'extraction pour mapper vers Boond
	// Certains modèles OpenAI peuvent renvoyer le JSON entouré de ```json ... ``` → on nettoie
	cleaned := sanitizeJSONResult(result)
	if len(cleaned) != len(result) {
		log.Printf("🧼 JSON nettoyé des backticks: avant=%d bytes, après=%d bytes", len(result), len(cleaned))
	}

	var cv nuextract.CVExtractionSchema
	if err := json.Unmarshal(cleaned, &cv); err != nil {
		// Afficher un extrait pour debug
		snippet := string(cleaned)
		if len(snippet) > 200 {
			snippet = snippet[:200] + "…"
		}
		log.Printf("❌ Parser JSON extraction → erreur: %v | extrait: %s", err, snippet)
		// On renvoie quand même le résultat d'extraction au front, sans bloquer
		c.Data(http.StatusOK, "application/json", result)
		return
	}

	// Normaliser les dates d'expériences au format MM/YY
	normalizeExperienceDates(&cv, language)

	// Construire les attributs Boond à partir du CV extrait
	attributes := boond.BuildCandidateAttributesFromCV(cv)
	if b, _ := json.MarshalIndent(attributes, "", "  "); len(b) > 0 {
		log.Printf("🧩 Attributs Boond construits:\n%s", string(b))
	}

	// TODO: Déduplication temporairement désactivée - l'API Boond ne filtre pas correctement par email
	// Elle retourne le dernier candidat de la liste au lieu de faire une vraie recherche
	//
	// Stratégie de déduplication: chercher par email si disponible
	// ctx := c.Request.Context()
	// bClient := boond.New(boondJWT)
	//
	// existingID := ""
	//
	// // 1. Recherche par email d'abord
	// if strings.TrimSpace(cv.Email) != "" {
	// 	log.Printf("🔎 Dédup Boond: recherche par email '%s'", cv.Email)
	// 	if foundID, err := bClient.SearchCandidateByEmail(ctx, cv.Email); err != nil {
	// 		log.Printf("WARNING: Erreur lors de la recherche de doublon Boond par email: %v", err)
	// 	} else if foundID != "" {
	// 		existingID = foundID
	// 		log.Printf("✅ Doublon détecté par email: candidat déjà présent dans Boond (id=%s) — pas de création", existingID)
	// 	}
	// }
	//
	// // 2. Si pas trouvé par email, essayer par nom+prénom
	// if existingID == "" && strings.TrimSpace(cv.Prenom) != "" && strings.TrimSpace(cv.Nom) != "" {
	// 	log.Printf("🔎 Dédup Boond: recherche par nom+prénom '%s %s'", cv.Prenom, cv.Nom)
	// 	if foundID, err := bClient.SearchCandidateByName(ctx, cv.Prenom, cv.Nom); err != nil {
	// 		log.Printf("WARNING: Erreur lors de la recherche de doublon Boond par nom: %v", err)
	// 	} else if foundID != "" {
	// 		existingID = foundID
	// 		log.Printf("✅ Doublon détecté par nom: candidat déjà présent dans Boond (id=%s) — pas de création", existingID)
	// 	}
	// }
	//
	// // Si pas de doublon détecté, tenter la création
	// if existingID == "" {
	// 	log.Printf("📡 Création candidat Boond → POST /api/candidates …")
	// 	createdID, raw, err := bClient.CreateCandidate(ctx, attributes)
	// 	if err != nil {
	// 		log.Printf("❌ Échec de création du candidat Boond: %v", err)
	// 		// On n'échoue pas la route /extract: on continue à renvoyer le JSON NuExtract
	// 	} else {
	// 		log.Printf("🎉 Candidat Boond créé avec succès (id=%s)", createdID)
	// 		if len(raw) > 0 {
	// 			body := string(raw)
	// 			if len(body) > 600 {
	// 				body = body[:600] + "…(tronqué)"
	// 			}
	// 			log.Printf("📦 Réponse Boond (tronquée): %s", body)
	// 		}
	// 	}
	// }

	// Création directe du candidat (déduplication désactivée temporairement)
	ctx := c.Request.Context()
	bClient := boond.New(boondJWT)

	log.Printf("📡 Création candidat Boond → POST /api/candidates …")
	createdID, raw, err := bClient.CreateCandidate(ctx, attributes)
	if err != nil {
		log.Printf("❌ Échec de création du candidat Boond: %v", err)
		// On n'échoue pas la route /extract: on continue à renvoyer le JSON NuExtract
	} else {
		log.Printf("🎉 Candidat Boond créé avec succès (id=%s)", createdID)
		if len(raw) > 0 {
			body := string(raw)
			if len(body) > 600 {
				body = body[:600] + "…(tronqué)"
			}
			log.Printf("📦 Réponse Boond (tronquée): %s", body)
		}

		// Upload du CV après création du candidat
		log.Printf("📄 Upload du CV → POST /api/documents …")
		docID, err := bClient.UploadDocument(ctx, createdID, data, header.Filename)
		if err != nil {
			log.Printf("❌ Échec de l'upload du CV: %v", err)
			// On n'échoue pas la route /extract: le candidat est créé même si le CV échoue
		} else {
			log.Printf("🎉 CV uploadé avec succès (doc_id=%s)", docID)
		}

		// Enrichir la réponse avec l'ID du candidat Boond pour la synchronisation future
		log.Printf("🔄 Enrichissement de la réponse avec boond_candidate_id=%s", createdID)

		// Repartir de la version normalisée
		var response map[string]any
		normalizedCleaned, mErr := json.Marshal(cv)
		if mErr != nil {
			log.Printf("❌ Erreur marshalling CV normalisé pour enrichissement: %v", mErr)
			// fallback: tenter avec cleaned
			if err := json.Unmarshal(cleaned, &response); err != nil {
				log.Printf("❌ Erreur parsing JSON nettoyé pour enrichissement: %v", err)
				c.Data(http.StatusOK, "application/json", result)
				return
			}
		} else {
			if err := json.Unmarshal(normalizedCleaned, &response); err != nil {
				log.Printf("❌ Erreur parsing CV normalisé en map pour enrichissement: %v", err)
				// fallback: tenter avec cleaned
				if err := json.Unmarshal(cleaned, &response); err != nil {
					log.Printf("❌ Erreur parsing JSON nettoyé pour enrichissement (fallback): %v", err)
					c.Data(http.StatusOK, "application/json", result)
					return
				}
			}
		}

		// Ajouter l'ID Boond à la réponse
		response["boond_candidate_id"] = createdID
		log.Printf("✅ boond_candidate_id ajouté à la réponse: %s", createdID)

		// Renvoyer la réponse enrichie
		enrichedResult, err := json.Marshal(response)
		if err != nil {
			log.Printf("❌ Erreur marshalling réponse enrichie: %v", err)
			c.Data(http.StatusOK, "application/json", result)
			return
		}

		log.Printf("📤 Réponse enrichie envoyée au frontend (taille: %d bytes)", len(enrichedResult))
		previewLen := 200
		if len(enrichedResult) < previewLen {
			previewLen = len(enrichedResult)
		}
		log.Printf("🔍 Aperçu réponse enrichie: %s", string(enrichedResult[:previewLen]))
		c.Data(http.StatusOK, "application/json", enrichedResult)
		return
	}

	// Si pas de JWT ou échec de création Boond, renvoyer le JSON d'extraction normal
	log.Printf("📤 Réponse normale envoyée au frontend (sans boond_candidate_id) - taille: %d bytes", len(result))
	c.Data(http.StatusOK, "application/json", result)
}

// sanitizeJSONResult retire d'éventuels blocs de code ```json ... ``` et extrait uniquement l'objet JSON
func sanitizeJSONResult(b []byte) []byte {
	s := strings.TrimSpace(string(b))
	// S'il n'y a pas de backticks, renvoyer tel quel
	if !strings.Contains(s, "```") {
		return b
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

// maskToken retourne les n premiers caractères du token
func maskToken(t string, n int) string {
	if n <= 0 || len(t) <= n {
		return t
	}
	return t[:n]
}

// tailToken retourne les n derniers caractères du token
func tailToken(t string, n int) string {
	if n <= 0 || len(t) <= n {
		return t
	}
	return t[len(t)-n:]
}

// normalizeExperienceDates force les dates des expériences au format MM/YY
func normalizeExperienceDates(cv *nuextract.CVExtractionSchema, language string) {
	if cv == nil || len(cv.Experiences) == 0 {
		return
	}
	for i := range cv.Experiences {
		start := strings.TrimSpace(cv.Experiences[i].DateDebut)
		end := strings.TrimSpace(cv.Experiences[i].DateFin)

		if start != "" {
			if m, y, ok := parseMonthYear(start); ok {
				cv.Experiences[i].DateDebut = fmt.Sprintf("%02d/%02d", m, y%100)
			} else {
				// tentative sur formats numériques courants
				if formatted, ok2 := normalizeNumericDate(start); ok2 {
					cv.Experiences[i].DateDebut = formatted
				}
			}
		}

		if end != "" {
			// valeurs de type "En cours" / "In progress"
			if isPresent(end) {
				if strings.ToLower(strings.TrimSpace(language)) == "fr" {
					cv.Experiences[i].DateFin = "En cours"
				} else {
					cv.Experiences[i].DateFin = "In progress"
				}
			} else if m, y, ok := parseMonthYear(end); ok {
				cv.Experiences[i].DateFin = fmt.Sprintf("%02d/%02d", m, y%100)
			} else {
				if formatted, ok2 := normalizeNumericDate(end); ok2 {
					cv.Experiences[i].DateFin = formatted
				}
			}
		}
	}
}

// isPresent détecte des variantes de "en cours" / "present"
func isPresent(s string) bool {
	v := strings.ToLower(strings.TrimSpace(s))
	switch v {
	case "en cours", "encours", "actuel", "actuellement", "present", "présent", "in progress", "ongoing", "current", "now":
		return true
	}
	return false
}

// normalizeNumericDate gère des formats type MM/YYYY, YYYY-MM, MM-YY, etc.
func normalizeNumericDate(s string) (string, bool) {
	str := strings.TrimSpace(s)
	// MM/YYYY ou M/YYYY ou MM/YY
	re1 := regexp.MustCompile(`^(?i)\s*(\d{1,2})[\-/\.](\d{2,4})\s*$`)
	if m := re1.FindStringSubmatch(str); len(m) == 3 {
		mm, _ := strconv.Atoi(m[1])
		yy, _ := strconv.Atoi(m[2])
		if yy >= 100 { // YYYY → YY
			yy = yy % 100
		}
		if mm >= 1 && mm <= 12 {
			return fmt.Sprintf("%02d/%02d", mm, yy), true
		}
	}
	// YYYY-MM ou YYYY/MM
	re2 := regexp.MustCompile(`^(?i)\s*(\d{4})[\-/\.](\d{1,2})\s*$`)
	if m := re2.FindStringSubmatch(str); len(m) == 3 {
		yy, _ := strconv.Atoi(m[1])
		mm, _ := strconv.Atoi(m[2])
		if mm >= 1 && mm <= 12 {
			return fmt.Sprintf("%02d/%02d", mm, yy%100), true
		}
	}
	return "", false
}

// parseMonthYear gère des libellés avec mois en FR/EN et année, ex: "Février 2020", "Aug 2018"
func parseMonthYear(s string) (int, int, bool) {
	if s == "" {
		return 0, 0, false
	}
	cleaned := strings.ToLower(strings.TrimSpace(s))
	cleaned = strings.ReplaceAll(cleaned, ".", "")
	cleaned = strings.ReplaceAll(cleaned, ",", "")
	// Regex: <mois> <annee>
	re := regexp.MustCompile(`^([a-zàâäéèêëîïôöùûüç]{3,})\s+(\d{4})$`)
	mm := re.FindStringSubmatch(cleaned)
	if len(mm) == 3 {
		monName := strings.TrimSpace(mm[1])
		year, _ := strconv.Atoi(mm[2])
		if monthNum, ok := monthNameToNumber(monName); ok {
			return monthNum, year, true
		}
	}
	return 0, 0, false
}

func monthNameToNumber(name string) (int, bool) {
	m := map[string]int{
		// Français complets
		"janvier": 1, "février": 2, "fevrier": 2, "mars": 3, "avril": 4, "mai": 5, "juin": 6, "juillet": 7, "août": 8, "aout": 8, "septembre": 9, "octobre": 10, "novembre": 11, "décembre": 12, "decembre": 12,
		// Français abréviations courantes
		"janv": 1, "fev": 2, "févr": 2, "fevr": 2, "avr": 4, "sept": 9, "oct": 10, "nov": 11, "déc": 12, "dec": 12,
		// Anglais complets
		"january": 1, "february": 2, "march": 3, "april": 4, "may": 5, "june": 6, "july": 7, "august": 8, "september": 9, "october": 10, "november": 11, "december": 12,
		// Anglais abréviations
		"jan": 1, "feb": 2, "mar": 3, "apr": 4, "jun": 6, "jul": 7, "aug": 8, "sep": 9,
	}
	if v, ok := m[name]; ok {
		return v, true
	}
	return 0, false
}
