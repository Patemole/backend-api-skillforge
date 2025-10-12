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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload: " + err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id format"})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process dossier data"})
			return
		}
		// Si un champ manque dans les données brutes, il sera laissé à sa valeur
		// par défaut (zéro value) dans la structure `dossier` (ex: "" pour un string).
		if err := json.Unmarshal(dossierBytes, &dossier); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to map dossier data to structure"})
			return
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create job: " + err.Error()})
		return
	}

	var results []models.Job
	if err = json.Unmarshal(data, &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse job creation result: " + err.Error()})
		return
	}

	if len(results) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create job: no result returned"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"job_id": results[0].ID})
}

// GetJobStatus checks the status of a specific job.
func GetJobStatus(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "job ID is required"})
		return
	}

	var results []models.Job
	// On ne sélectionne que les colonnes nécessaires pour le client
	query := supabase.Client.From("jobs").Select("status,result,error", "exact", false).Eq("id", jobID).Limit(1, "")

	data, _, err := query.Execute()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query job status: " + err.Error()})
		return
	}

	if err := json.Unmarshal(data, &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse job status result: " + err.Error()})
		return
	}

	if len(results) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	job := results[0]

	// Le worker mettra l'URL du fichier dans la colonne 'result'
	c.JSON(http.StatusOK, gin.H{
		"status": job.Status,
		"result": job.Result, // Important pour récupérer l'URL du fichier
		"error":  job.Error,
	})
}
