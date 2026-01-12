package handlers

import (
	"net/http"
	"strings"

	"backend-api-skillforge/internal/models"
	"backend-api-skillforge/internal/openai"

	"github.com/gin-gonic/gin"
)

// GeneratePresentationEmailV2 génère un email de présentation avec reformatage et sélection d'expériences
func GeneratePresentationEmailV2(c *gin.Context) {
	// Structure intermédiaire pour gérer les langues flexibles
	var tempReq struct {
		CandidateData struct {
			Prenom                    string `json:"prenom" binding:"required"`
			TitrePoste                string `json:"titre_poste" binding:"required"`
			NombreExperience          int    `json:"nombre_experience" binding:"required"`
			Disponibilite             string `json:"disponibilite" binding:"required"`
			Mobilite                  string `json:"mobilite" binding:"required"`
			Diplome                   string `json:"diplome" binding:"required"`
			Langues                   string `json:"langues"`
			Certifications            string `json:"certifications"`
			Hobbies                   string `json:"hobbies"`
			Experience                string `json:"experience"`
			ExperienceCount           int    `json:"experience_count"`
			Logiciel                  string `json:"logiciel"`
			Logiciels                 string `json:"logiciels"`
			LogicielsCount            int    `json:"logiciels_count"`
			CompetencesTechniques     string `json:"competences_techniques"`
			CompetencesFonctionnelles string `json:"competences_fonctionnelles"`
			Projets                   string `json:"projets"`
			ProjetsCount              int    `json:"projets_count"`
		} `json:"candidateData" binding:"required"`
		Need       *string                      `json:"need,omitempty"`
		TemplateID *string                      `json:"templateId,omitempty"`
		Template   *models.PresentationTemplate `json:"template,omitempty"`
	}

	if err := c.ShouldBindJSON(&tempReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":           false,
			"error":             "Données de requête invalides",
			"error_code":        "INVALID_REQUEST_PAYLOAD",
			"details":           "Vérifiez que tous les champs requis sont présents et correctement formatés",
			"technical_details": err.Error(),
		})
		return
	}

	// Les langues sont déjà formatées en string par le frontend

	// Construire la requête finale avec les langues converties
	req := models.PresentationEmailRequest{
		CandidateData: models.PresentationCandidateData{
			Prenom:                    tempReq.CandidateData.Prenom,
			TitrePoste:                tempReq.CandidateData.TitrePoste,
			NombreExperience:          tempReq.CandidateData.NombreExperience,
			Disponibilite:             tempReq.CandidateData.Disponibilite,
			Mobilite:                  tempReq.CandidateData.Mobilite,
			Diplome:                   tempReq.CandidateData.Diplome,
			Langues:                   tempReq.CandidateData.Langues,
			Certifications:            tempReq.CandidateData.Certifications,
			Hobbies:                   tempReq.CandidateData.Hobbies,
			Experience:                tempReq.CandidateData.Experience,
			ExperienceCount:           tempReq.CandidateData.ExperienceCount,
			Logiciel:                  tempReq.CandidateData.Logiciel,
			Logiciels:                 tempReq.CandidateData.Logiciels,
			LogicielsCount:            tempReq.CandidateData.LogicielsCount,
			CompetencesTechniques:     tempReq.CandidateData.CompetencesTechniques,
			CompetencesFonctionnelles: tempReq.CandidateData.CompetencesFonctionnelles,
			Projets:                   tempReq.CandidateData.Projets,
			ProjetsCount:              tempReq.CandidateData.ProjetsCount,
		},
		Need:       tempReq.Need,
		TemplateID: tempReq.TemplateID,
		Template:   tempReq.Template,
	}

	// Valider les champs obligatoires
	if strings.TrimSpace(req.CandidateData.Prenom) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error":      "Le prénom du candidat est obligatoire",
			"error_code": "MISSING_PRENOM",
			"details":    "Le champ 'prenom' ne peut pas être vide",
		})
		return
	}

	if strings.TrimSpace(req.CandidateData.TitrePoste) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error":      "Le titre du poste est obligatoire",
			"error_code": "MISSING_TITRE_POSTE",
			"details":    "Le champ 'titre_poste' ne peut pas être vide",
		})
		return
	}

	if req.CandidateData.NombreExperience < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error":      "Le nombre d'années d'expérience doit être positif ou nul",
			"error_code": "INVALID_EXPERIENCE_YEARS",
			"details":    "Le champ 'nombre_experience' doit être un nombre positif ou zéro",
		})
		return
	}

	if strings.TrimSpace(req.CandidateData.Disponibilite) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error":      "La disponibilité est obligatoire",
			"error_code": "MISSING_DISPONIBILITE",
			"details":    "Le champ 'disponibilite' ne peut pas être vide",
		})
		return
	}

	if strings.TrimSpace(req.CandidateData.Mobilite) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error":      "La mobilité est obligatoire",
			"error_code": "MISSING_MOBILITE",
			"details":    "Le champ 'mobilite' ne peut pas être vide",
		})
		return
	}

	if strings.TrimSpace(req.CandidateData.Diplome) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error":      "Le diplôme est obligatoire",
			"error_code": "MISSING_DIPLOME",
			"details":    "Le champ 'diplome' ne peut pas être vide",
		})
		return
	}

	// Créer le service de génération d'emails de présentation
	emailService := openai.NewPresentationEmailGeneratorService()

	// Générer l'email de présentation
	response, err := emailService.GeneratePresentationEmail(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":           false,
			"error":             "Échec de la génération de l'email de présentation",
			"error_code":        "EMAIL_GENERATION_FAILED",
			"details":           "Une erreur est survenue lors de la génération du contenu de l'email",
			"technical_details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
