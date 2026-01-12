package handlers

import (
	"log"
	"net/http"
	"time"

	"backend-api-skillforge/internal/models"
	"backend-api-skillforge/internal/supabase"

	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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
			Prenom            string              `json:"prenom"`
			Nom               string              `json:"nom"`
			Email             string              `json:"email"`
			Phone             string              `json:"phone"`
			Summary           string              `json:"summary"`
			Age               string              `json:"age"`
			Poste             string              `json:"poste"`
			Diplome           string              `json:"diplome"`
			Experience        string              `json:"expérience"`
			Mobilite          string              `json:"mobilité"`
			Disponibilite     string              `json:"disponibilité"`
			PermisB           string              `json:"permis_B"`
			Hobbies           []string            `json:"hobbies"`
			Languages         []interface{}       `json:"languages"` // Interface{} pour accepter strings et objets
			SecteursActivites []string            `json:"secteurs_activites"`
			DomainesExpertise []string            `json:"domaines_expertise"`
			Formations        []models.Formation  `json:"formations"`
			Experiences       []models.Experience `json:"expériences"`
			Logiciels         []models.Logiciel   `json:"logiciels"`
			Certifications    []string            `json:"certifications"`
			TechnicalSkills   []string            `json:"technical_skills"`
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
		dossier = models.CompetenceDossier{
			Prenom:            tempDossier.Prenom,
			Nom:               tempDossier.Nom,
			Email:             tempDossier.Email,
			Phone:             tempDossier.Phone,
			Summary:           tempDossier.Summary,
			Age:               tempDossier.Age,
			Poste:             tempDossier.Poste,
			Diplome:           tempDossier.Diplome,
			Experience:        tempDossier.Experience,
			Mobilite:          tempDossier.Mobilite,
			Disponibilite:     tempDossier.Disponibilite,
			PermisB:           tempDossier.PermisB,
			Hobbies:           tempDossier.Hobbies,
			Languages:         languages,
			SecteursActivites: tempDossier.SecteursActivites,
			DomainesExpertise: tempDossier.DomainesExpertise,
			Formations:        tempDossier.Formations,
			Experiences:       tempDossier.Experiences,
			Logiciels:         tempDossier.Logiciels,
			Certifications:    tempDossier.Certifications,
			TechnicalSkills:   tempDossier.TechnicalSkills,
		}

		// Log du dossier après transformation
		log.Printf("🔄 DEBUG JOBS - Dossier après transformation:")
		dossierJSON, _ := json.MarshalIndent(dossier, "   ", "  ")
		log.Printf("   - Dossier structuré:\n%s", string(dossierJSON))

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
