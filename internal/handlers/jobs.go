package handlers

import (
	"log"
	"net/http"
	"strings"
	"time"

	"backend-api-skillforge/internal/models"
	"backend-api-skillforge/internal/supabase"

	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// extractFirstName extracts only the first name from a full name for DC anonymization.
// If the last name (nom) is provided separately, we assume prenom is already just the first name.
// Otherwise, we extract the first name by taking the first part of the full name.
func extractFirstName(prenom, nom string) string {
	prenomTrimmed := strings.TrimSpace(prenom)
	if prenomTrimmed == "" {
		return ""
	}

	// If the last name is provided separately, prenom should already be just the first name
	if strings.TrimSpace(nom) != "" {
		return prenomTrimmed
	}

	// Otherwise, prenom probably contains the full name, we extract just the first name
	// We take the first part (first word) of the full name
	parts := strings.Fields(prenomTrimmed)
	if len(parts) == 1 {
		// Single word, it's probably already just the first name
		return prenomTrimmed
	}

	// Multiple words: we take the first word as the first name
	// (for cases like "Jean DUPONT" or "Manuel de OLIVEIRA", we take "Jean" or "Manuel")
	return parts[0]
}

// CreateJobRequest defines the expected request body for creating a job.
type CreateJobRequest struct {
	Type    string         `json:"type" binding:"required"`
	Payload map[string]any `json:"payload" binding:"required"`
	UserID  string         `json:"user_id" binding:"required"`
}

// CreateJob handles the creation of a new job.
func CreateJob(c *gin.Context) {
	var req CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":           false,
			"error":             "Données de requête invalides",
			"error_code":        "INVALID_REQUEST_PAYLOAD",
			"details":           "Vérifiez que tous les champs requis sont présents et correctement formatés",
			"technical_details": err.Error(),
		})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error":      "Format d'ID utilisateur invalide",
			"error_code": "INVALID_USER_ID_FORMAT",
			"details":    "L'ID utilisateur doit être un UUID valide",
			"user_id":    req.UserID,
		})
		return
	}

	// --- Transformation du Payload ---

	// 1. Extraire les données brutes du payload de la requête.
	rawPayload := req.Payload
	rawDossier, dossierExists := rawPayload["competence_dossier"]
	templateURL, _ := rawPayload["template_url"]
	organizationName, _ := rawPayload["organization_name"]
	dossierID, _ := rawPayload["dossier_id"]

	// Log détaillé du payload reçu du frontend
	log.Printf("🔍 DEBUG JOBS - Payload reçu du frontend:")
	log.Printf("   - Type: %s", req.Type)
	log.Printf("   - UserID: %s", req.UserID)
	log.Printf("   - TemplateURL: %v", templateURL)
	log.Printf("   - OrganizationName: %v", organizationName)
	log.Printf("   - DossierID: %v", dossierID)
	log.Printf("   - DossierExists: %t", dossierExists)

	if dossierExists {
		// Log du dossier brut reçu
		dossierJSON, _ := json.MarshalIndent(rawDossier, "   ", "  ")
		log.Printf("   - RawDossier (JSON brut):\n%s", string(dossierJSON))

		// Log spécifique des langues si présentes
		if rawDossierMap, ok := rawDossier.(map[string]interface{}); ok {
			if languages, exists := rawDossierMap["languages"]; exists {
				log.Printf("   - Languages détectées dans rawDossier: %+v (type: %T)", languages, languages)
			} else {
				log.Printf("   - ⚠️ Aucun champ 'languages' trouvé dans rawDossier")
			}

			// Log et transformation des compétences fonctionnelles si format objet
			if competenceFonct, exists := rawDossierMap["competence_fonctionnelle"]; exists {
				log.Printf("   - competence_fonctionnelle détectée: %+v (type: %T)", competenceFonct, competenceFonct)
				// Si c'est un objet avec word_list, extraire word_list
				if compMap, ok := competenceFonct.(map[string]interface{}); ok {
					if wordList, ok := compMap["word_list"].([]interface{}); ok {
						log.Printf("   - Format objet détecté, extraction de word_list: %+v", wordList)
						var wordListStrings []string
						for _, item := range wordList {
							if str, ok := item.(string); ok {
								wordListStrings = append(wordListStrings, str)
							}
						}
						rawDossierMap["competence_fonctionnelle"] = wordListStrings
						log.Printf("   - competence_fonctionnelle transformée en tableau: %+v", wordListStrings)
					}
				}
			}

			// Log et transformation des technical_skills si format objet
			if techSkills, exists := rawDossierMap["technical_skills"]; exists {
				log.Printf("   - technical_skills détectées: %+v (type: %T)", techSkills, techSkills)
				// Si c'est un objet avec word_list, extraire word_list
				if techMap, ok := techSkills.(map[string]interface{}); ok {
					if wordList, ok := techMap["word_list"].([]interface{}); ok {
						log.Printf("   - Format objet détecté, extraction de word_list: %+v", wordList)
						var wordListStrings []string
						for _, item := range wordList {
							if str, ok := item.(string); ok {
								wordListStrings = append(wordListStrings, str)
							}
						}
						rawDossierMap["technical_skills"] = wordListStrings
						log.Printf("   - technical_skills transformées en tableau: %+v", wordListStrings)
					}
				}
			}

			// Log et transformation des logiciels si format objet
			if logiciels, exists := rawDossierMap["logiciels"]; exists {
				log.Printf("   - logiciels détectés dans rawDossier: %+v (type: %T)", logiciels, logiciels)

				// Si c'est un tableau, vérifier le contenu de chaque logiciel
				if logicielsArray, ok := logiciels.([]interface{}); ok {
					log.Printf("   - Format tableau détecté: %d logiciels", len(logicielsArray))
					for i, logiciel := range logicielsArray {
						if logicielMap, ok := logiciel.(map[string]interface{}); ok {
							logicielName, _ := logicielMap["logiciel"].(string)
							tempsUtil, _ := logicielMap["temps_utilisation"].(string)
							log.Printf("   - Logiciel #%d (%s): temps_utilisation = '%s'", i+1, logicielName, tempsUtil)
						}
					}
				}

				// Si c'est un objet avec word_list, extraire word_list
				if logicMap, ok := logiciels.(map[string]interface{}); ok {
					if wordList, ok := logicMap["word_list"].([]interface{}); ok {
						log.Printf("   - Format objet détecté, extraction de word_list: %+v", wordList)
						// Vérifier chaque logiciel dans word_list pour temps_utilisation
						for i, logiciel := range wordList {
							if logicielMap, ok := logiciel.(map[string]interface{}); ok {
								logicielName, _ := logicielMap["logiciel"].(string)
								tempsUtil, _ := logicielMap["temps_utilisation"].(string)
								log.Printf("   - Logiciel word_list #%d (%s): temps_utilisation = '%s'", i+1, logicielName, tempsUtil)
							}
						}
						rawDossierMap["logiciels"] = wordList
						log.Printf("   - logiciels transformés en tableau: %+v", wordList)
					}
				}
			} else {
				log.Printf("   - ⚠️ Aucun champ 'logiciels' trouvé dans rawDossier")
			}

			// Log et transformation des expériences pour projets_name et result
			if experiences, exists := rawDossierMap["expériences"]; exists {
				if expList, ok := experiences.([]interface{}); ok {
					log.Printf("   - %d expériences détectées dans rawDossier", len(expList))
					for i, exp := range expList {
						if expMap, ok := exp.(map[string]interface{}); ok {
							// Lister toutes les clés de l'expérience pour debug
							var keys []string
							for k := range expMap {
								keys = append(keys, k)
							}
							log.Printf("   - Expérience %d - Clés disponibles: %v", i, keys)

							// Vérifier projets_name (avec variations possibles)
							if projetsName, exists := expMap["projets_name"]; exists {
								log.Printf("   - Expérience %d: projets_name = %+v (type: %T)", i, projetsName, projetsName)
							} else if projetsName, exists := expMap["projetsName"]; exists {
								log.Printf("   - Expérience %d: projetsName (camelCase) = %+v (type: %T)", i, projetsName, projetsName)
								expMap["projets_name"] = projetsName // Normaliser
							} else if projetsName, exists := expMap["project_name"]; exists {
								log.Printf("   - Expérience %d: project_name = %+v (type: %T)", i, projetsName, projetsName)
								expMap["projets_name"] = projetsName // Normaliser
							} else {
								log.Printf("   - Expérience %d: ⚠️ projets_name NON TROUVÉ", i)
							}

							// Vérifier result (avec variations possibles)
							if result, exists := expMap["result"]; exists {
								log.Printf("   - Expérience %d: result = '%+v' (type: %T, empty=%t)", i, result, result, result == "" || result == nil)
							} else if result, exists := expMap["resultat"]; exists {
								log.Printf("   - Expérience %d: resultat (français) = '%+v' (type: %T)", i, result, result)
								expMap["result"] = result // Normaliser
							} else {
								log.Printf("   - Expérience %d: ⚠️ result NON TROUVÉ", i)
							}
						}
					}
				}
			}
		}
	}

	// 2. Préparer le dossier de compétences structuré.
	dossier := models.CompetenceDossier{}

	// 3. Si des données de dossier existent, on tente de les mapper sur notre structure.
	if dossierExists {
		// On passe par JSON pour convertir la map générique (map[string]any)
		// en notre structure fortement typée (models.CompetenceDossier).
		// C'est l'étape qui "nettoie" et applique le template.
		dossierBytes, err := json.Marshal(rawDossier)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success":           false,
				"error":             "Échec du traitement des données du dossier",
				"error_code":        "DOSSIER_PROCESSING_FAILED",
				"details":           "Impossible de sérialiser les données du dossier en JSON",
				"technical_details": err.Error(),
			})
			return
		}

		// Structure intermédiaire pour gérer les langues flexibles
		var tempDossier struct {
			Prenom                  string              `json:"prenom"`
			Nom                     string              `json:"nom"`
			Email                   string              `json:"email"`
			Phone                   string              `json:"phone"`
			Summary                 string              `json:"summary"`
			Age                     string              `json:"age"`
			Poste                   string              `json:"poste"`
			Diplome                 string              `json:"diplome"`
			Experience              string              `json:"expérience"`
			Mobilite                string              `json:"mobilité"`
			Disponibilite           string              `json:"disponibilité"`
			PermisB                 string              `json:"permis_B"`
			Hobbies                 []string            `json:"hobbies"`
			Languages               []interface{}       `json:"languages"` // Interface{} pour accepter strings et objets
			SecteursActivites       []string            `json:"secteurs_activites"`
			DomainesExpertise       []string            `json:"domaines_expertise"`
			Formations              []models.Formation  `json:"formations"`
			Experiences             []models.Experience `json:"expériences"`
			Logiciels               []models.Logiciel   `json:"logiciels"`
			Certifications          []string            `json:"certifications"`
			TechnicalSkills         []string            `json:"technical_skills"`
			CompetenceFonctionnelle []string            `json:"competence_fonctionnelle"`
		}

		// Si un champ manque dans les données brutes, il sera laissé à sa valeur
		// par défaut (zéro value) dans la structure `dossier` (ex: "" pour un string).
		if err := json.Unmarshal(dossierBytes, &tempDossier); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success":           false,
				"error":             "Échec du mapping des données du dossier",
				"error_code":        "DOSSIER_MAPPING_FAILED",
				"details":           "Impossible de mapper les données du dossier vers la structure attendue. Vérifiez le format des données envoyées.",
				"technical_details": err.Error(),
				"dossier_data":      rawDossier,
			})
			return
		}

		// Log détaillé des expériences après unmarshal pour debug
		log.Printf("🔍 DEBUG JOBS - Expériences après unmarshal: %d expériences", len(tempDossier.Experiences))
		for i, exp := range tempDossier.Experiences {
			log.Printf("   - Expérience #%d (%s @ %s):", i+1, exp.Poste, exp.Entreprise)
			log.Printf("      * projets_name: %+v (len=%d)", exp.ProjetsName, len(exp.ProjetsName))
			log.Printf("      * result: '%s' (empty=%t)", exp.Result, exp.Result == "")
		}

		// Log détaillé des logiciels après unmarshal pour vérifier temps_utilisation
		log.Printf("🔍 DEBUG JOBS - Logiciels après unmarshal: %d logiciels", len(tempDossier.Logiciels))
		for i, logiciel := range tempDossier.Logiciels {
			log.Printf("   - Logiciel #%d: %s", i+1, logiciel.Logiciel)
			log.Printf("      * level: '%s'", logiciel.Level)
			log.Printf("      * temps_utilisation: '%s' (empty=%t)", logiciel.TempsUtilisation, logiciel.TempsUtilisation == "")
		}

		// Convertir les langues du format flexible vers []string
		var languages []string
		for _, lang := range tempDossier.Languages {
			switch v := lang.(type) {
			case string:
				// Langue simple (ex: "Français")
				languages = append(languages, v)
			case map[string]interface{}:
				// Langue avec niveau (ex: {"language": "Chinois", "level": "Intermediaire"})
				if langName, ok := v["language"].(string); ok {
					if level, ok := v["level"].(string); ok && level != "" {
						languages = append(languages, langName+" ("+level+")")
					} else {
						languages = append(languages, langName)
					}
				}
			}
		}

		// Mapper vers la structure finale
		// ✅ ANONYMIZATION: Extraire uniquement le prénom (pas le nom complet)
		firstName := extractFirstName(tempDossier.Prenom, tempDossier.Nom)
		dossier = models.CompetenceDossier{
			Prenom:                  firstName,
			Nom:                     tempDossier.Nom,
			Email:                   tempDossier.Email,
			Phone:                   tempDossier.Phone,
			Summary:                 tempDossier.Summary,
			Age:                     tempDossier.Age,
			Poste:                   tempDossier.Poste,
			Diplome:                 tempDossier.Diplome,
			Experience:              tempDossier.Experience,
			Mobilite:                tempDossier.Mobilite,
			Disponibilite:           tempDossier.Disponibilite,
			PermisB:                 tempDossier.PermisB,
			Hobbies:                 tempDossier.Hobbies,
			Languages:               languages,
			SecteursActivites:       tempDossier.SecteursActivites,
			DomainesExpertise:       tempDossier.DomainesExpertise,
			Formations:              tempDossier.Formations,
			Experiences:             tempDossier.Experiences,
			Logiciels:               tempDossier.Logiciels,
			Certifications:          tempDossier.Certifications,
			TechnicalSkills:         tempDossier.TechnicalSkills,
			CompetenceFonctionnelle: tempDossier.CompetenceFonctionnelle,
		}

		// Log du dossier après transformation
		log.Printf("🔄 DEBUG JOBS - Dossier après transformation:")
		log.Printf("   - CompetenceFonctionnelle: %+v (len=%d)", dossier.CompetenceFonctionnelle, len(dossier.CompetenceFonctionnelle))
		log.Printf("   - Nombre d'expériences: %d", len(dossier.Experiences))
		for i, exp := range dossier.Experiences {
			log.Printf("   - Expérience #%d - projets_name: %+v, result: '%s'", i+1, exp.ProjetsName, exp.Result)
		}
		log.Printf("   - Nombre de logiciels: %d", len(dossier.Logiciels))
		for i, logiciel := range dossier.Logiciels {
			log.Printf("   - Logiciel #%d (%s): level='%s', temps_utilisation='%s'", i+1, logiciel.Logiciel, logiciel.Level, logiciel.TempsUtilisation)
		}
		dossierJSON, _ := json.MarshalIndent(dossier, "   ", "  ")
		log.Printf("   - Dossier structuré (JSON complet):\n%s", string(dossierJSON))

		// Log spécifique des langues après transformation
		log.Printf("   - Languages après transformation: %+v (type: %T)", dossier.Languages, dossier.Languages)
		if len(dossier.Languages) == 0 {
			log.Printf("   - ⚠️ Aucune langue dans le dossier structuré")
		} else {
			log.Printf("   - ✅ %d langues détectées: %v", len(dossier.Languages), dossier.Languages)
		}
	}

	// 4. On reconstruit un payload propre et final pour le worker.
	finalPayload := map[string]any{
		"competence_dossier": dossier,
		"template_url":       templateURL,
		"organization_name":  organizationName,
		"dossier_id":         dossierID,
	}

	// Log du payload final envoyé au worker
	log.Printf("📤 DEBUG JOBS - Payload final envoyé au worker:")
	finalPayloadJSON, _ := json.MarshalIndent(finalPayload, "   ", "  ")
	log.Printf("   - FinalPayload:\n%s", string(finalPayloadJSON))

	// Log spécifique des langues dans le payload final
	if finalDossier, ok := finalPayload["competence_dossier"].(models.CompetenceDossier); ok {
		log.Printf("   - Languages dans payload final: %+v (type: %T)", finalDossier.Languages, finalDossier.Languages)
		if len(finalDossier.Languages) == 0 {
			log.Printf("   - ⚠️ Aucune langue transmise au worker")
		} else {
			log.Printf("   - ✅ %d langues transmises au worker: %v", len(finalDossier.Languages), finalDossier.Languages)
		}
	}

	// --- Fin de la Transformation ---

	newJob := models.Job{
		Type:      req.Type,
		UserID:    userID,
		Payload:   finalPayload, // On utilise le payload nettoyé
		Status:    "pending",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	// The key was to add the "representation" preference to get the inserted row back.
	data, _, err := supabase.Client.From("jobs").Insert(newJob, false, "representation", "", "").Execute()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":           false,
			"error":             "Échec de la création du job",
			"error_code":        "JOB_CREATION_FAILED",
			"details":           "Impossible de créer le job dans la base de données",
			"technical_details": err.Error(),
		})
		return
	}

	var results []models.Job
	if err = json.Unmarshal(data, &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":           false,
			"error":             "Échec du parsing du résultat de création du job",
			"error_code":        "JOB_PARSING_FAILED",
			"details":           "Impossible de parser la réponse de la base de données",
			"technical_details": err.Error(),
		})
		return
	}

	if len(results) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error":      "Aucun résultat retourné lors de la création du job",
			"error_code": "JOB_NO_RESULT",
			"details":    "La base de données n'a retourné aucun résultat après la création du job",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"job_id": results[0].ID})
}

// GetJobStatus checks the status of a specific job.
func GetJobStatus(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error":      "ID du job requis",
			"error_code": "MISSING_JOB_ID",
			"details":    "L'ID du job doit être fourni dans l'URL",
		})
		return
	}

	var results []models.Job
	// On ne sélectionne que les colonnes nécessaires pour le client
	query := supabase.Client.From("jobs").Select("status,result,error", "exact", false).Eq("id", jobID).Limit(1, "")

	data, _, err := query.Execute()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":           false,
			"error":             "Échec de la requête du statut du job",
			"error_code":        "JOB_STATUS_QUERY_FAILED",
			"details":           "Impossible de récupérer le statut du job depuis la base de données",
			"technical_details": err.Error(),
			"job_id":            jobID,
		})
		return
	}

	if err := json.Unmarshal(data, &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":           false,
			"error":             "Échec du parsing du statut du job",
			"error_code":        "JOB_STATUS_PARSING_FAILED",
			"details":           "Impossible de parser la réponse de la base de données",
			"technical_details": err.Error(),
			"job_id":            jobID,
		})
		return
	}

	if len(results) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success":    false,
			"error":      "Job non trouvé",
			"error_code": "JOB_NOT_FOUND",
			"details":    "Aucun job trouvé avec l'ID fourni",
			"job_id":     jobID,
		})
		return
	}

	job := results[0]

	// Le worker mettra l'URL du fichier dans la colonne 'result'
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"status":  job.Status,
		"result":  job.Result, // Important pour récupérer l'URL du fichier
		"error":   job.Error,
		"job_id":  jobID,
	})
}
