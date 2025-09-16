package handlers

import (
	"net/http"

	"backend-api-skillforge/internal/models"
	"backend-api-skillforge/internal/openai"

	"github.com/gin-gonic/gin"
)

// GenerateTemplate génère un template d'email basé sur des exemples
func GenerateTemplate(c *gin.Context) {
	var req models.TemplateGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.TemplateGenerateResponse{
			Success: false,
			Error:   "Invalid request payload: " + err.Error(),
		})
		return
	}

	// Valider les contraintes supplémentaires
	if len(req.EmailExamples) < 3 {
		c.JSON(http.StatusBadRequest, models.TemplateGenerateResponse{
			Success: false,
			Error:   "At least 3 email examples are required",
		})
		return
	}

	// Valider que chaque exemple a un sujet et un contenu non vides
	for i, example := range req.EmailExamples {
		if example.Subject == "" {
			c.JSON(http.StatusBadRequest, models.TemplateGenerateResponse{
				Success: false,
				Error:   "Email example " + string(rune(i+1)) + " must have a non-empty subject",
			})
			return
		}
		if example.Content == "" {
			c.JSON(http.StatusBadRequest, models.TemplateGenerateResponse{
				Success: false,
				Error:   "Email example " + string(rune(i+1)) + " must have a non-empty content",
			})
			return
		}
	}

	// Valider que les variables disponibles ne sont pas vides
	if len(req.AvailableVariables) == 0 {
		c.JSON(http.StatusBadRequest, models.TemplateGenerateResponse{
			Success: false,
			Error:   "At least one available variable is required",
		})
		return
	}

	// Créer le service de génération de templates
	templateService := openai.NewTemplateGeneratorService()

	// Générer le template
	response, err := templateService.GenerateTemplate(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.TemplateGenerateResponse{
			Success: false,
			Error:   "Failed to generate template: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
