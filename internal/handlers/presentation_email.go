package handlers

import (
	"net/http"

	"backend-api-skillforge/internal/models"
	"backend-api-skillforge/internal/openai"

	"github.com/gin-gonic/gin"
)

// GeneratePresentationEmailV2 génère un email de présentation avec reformatage et sélection d'expériences
func GeneratePresentationEmailV2(c *gin.Context) {
	var req models.PresentationEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.PresentationEmailResponse{
			EmailContent: "",
			Success:      false,
			Error:        "Invalid request payload: " + err.Error(),
		})
		return
	}

	// Valider les champs obligatoires
	if req.CandidateData.Prenom == "" {
		c.JSON(http.StatusBadRequest, models.PresentationEmailResponse{
			EmailContent: "",
			Success:      false,
			Error:        "Le prénom du candidat est obligatoire",
		})
		return
	}

	if req.CandidateData.TitrePoste == "" {
		c.JSON(http.StatusBadRequest, models.PresentationEmailResponse{
			EmailContent: "",
			Success:      false,
			Error:        "Le titre du poste est obligatoire",
		})
		return
	}

	if req.CandidateData.NombreExperience < 0 {
		c.JSON(http.StatusBadRequest, models.PresentationEmailResponse{
			EmailContent: "",
			Success:      false,
			Error:        "Le nombre d'années d'expérience doit être positif",
		})
		return
	}

	if req.CandidateData.Disponibilite == "" {
		c.JSON(http.StatusBadRequest, models.PresentationEmailResponse{
			EmailContent: "",
			Success:      false,
			Error:        "La disponibilité est obligatoire",
		})
		return
	}

	if req.CandidateData.Mobilite == "" {
		c.JSON(http.StatusBadRequest, models.PresentationEmailResponse{
			EmailContent: "",
			Success:      false,
			Error:        "La mobilité est obligatoire",
		})
		return
	}

	if req.CandidateData.Diplome == "" {
		c.JSON(http.StatusBadRequest, models.PresentationEmailResponse{
			EmailContent: "",
			Success:      false,
			Error:        "Le diplôme est obligatoire",
		})
		return
	}

	// Créer le service de génération d'emails de présentation
	emailService := openai.NewPresentationEmailGeneratorService()

	// Générer l'email de présentation
	response, err := emailService.GeneratePresentationEmail(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.PresentationEmailResponse{
			EmailContent: "",
			Success:      false,
			Error:        "Failed to generate presentation email: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
