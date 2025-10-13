package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
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

	client := nuextract.New()
	log.Printf("DEBUG: Client NuExtract créé, début de l'extraction...")

	result, err := client.ExtractAndEnrichWithFilename(data, header.Filename)
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
		// Pas d'intégration Boond demandée → retour immédiat du JSON
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

		var response map[string]any
		if err := json.Unmarshal(cleaned, &response); err != nil {
			log.Printf("❌ Erreur parsing JSON nettoyé pour enrichissement: %v", err)
			// Si erreur de parsing, renvoyer le JSON original
			c.Data(http.StatusOK, "application/json", result)
			return
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
