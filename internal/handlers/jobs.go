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
	
	// Log pour debug
	if organizationName != nil && organizationName != "" {
		log.Printf("✅ Organization name found in payload: %v", organizationName)
	} else {
		log.Printf("⚠️ No organization_name found in payload, will use default")
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
	}

	// 4. On reconstruit un payload propre et final pour le worker.
	finalPayload := map[string]any{
		"competence_dossier": dossier,
		"template_url":       templateURL,
		"organization_name":  organizationName,
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
