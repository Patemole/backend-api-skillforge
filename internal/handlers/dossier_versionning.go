package handlers

import (
	"log"
	"net/http"

	"backend-api-skillforge/internal/models"
	"backend-api-skillforge/internal/openai"

	"github.com/gin-gonic/gin"
)

// DossierVersionningRequest représente la requête pour créer une nouvelle version
type DossierVersionningRequest struct {
	NewTitle          string                   `json:"new_title"`
	Need              *string                  `json:"need,omitempty"`
	CandidateID       string                   `json:"candidate_id" binding:"required"`
	CompetenceDossier models.CompetenceDossier `json:"competence_dossier" binding:"required"`
}

// DossierVersionningResponse représente la réponse
type DossierVersionningResponse struct {
	CandidateID       string                   `json:"candidate_id"`
	CompetenceDossier models.CompetenceDossier `json:"competence_dossier"`
}

// CreateDossierVersion gère la création d'une nouvelle version de dossier de compétences.
// Payload attendu:
//
//	{
//	  "new_title": "Dossier v3 - ...",
//	  "need": "Mettre en avant ...", // optionnel
//	  "candidate_id": "uuid",
//	  "competence_dossier": { /* dossier source complet */ }
//	}
//
// Réponse:
//
//	{
//	  "candidate_id": "uuid",
//	  "competence_dossier": { /* dossier versionné */ }
//	}
func CreateDossierVersion(c *gin.Context) {
	log.Printf("🔄 DOSSIER VERSIONNING - Début de la requête")

	var req DossierVersionningRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ DOSSIER VERSIONNING - Erreur validation payload: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload: " + err.Error()})
		return
	}

	log.Printf("✅ DOSSIER VERSIONNING - Payload validé:")
	log.Printf("   - Candidate ID: %s", req.CandidateID)
	log.Printf("   - New Title: %s", req.NewTitle)
	log.Printf("   - Need: %v", req.Need)
	log.Printf("   - Dossier source reçu: %d expériences, %d formations",
		len(req.CompetenceDossier.Experiences), len(req.CompetenceDossier.Formations))

	service := openai.NewDossierVersionningService()
	log.Printf("🤖 DOSSIER VERSIONNING - Appel OpenAI en cours...")

	newDossier, err := service.GenerateVersionnedDossier(req.CandidateID, req.CompetenceDossier, req.Need)
	if err != nil {
		log.Printf("❌ DOSSIER VERSIONNING - Erreur OpenAI: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	log.Printf("✅ DOSSIER VERSIONNING - Réponse OpenAI reçue:")
	log.Printf("   - Dossier versionné: %d expériences, %d formations",
		len(newDossier.Experiences), len(newDossier.Formations))
	log.Printf("   - Poste: %s", newDossier.Poste)
	log.Printf("   - Nombre de réalisations total: %d",
		len(newDossier.Experiences)+len(newDossier.Logiciels))

	resp := DossierVersionningResponse{
		CandidateID:       req.CandidateID,
		CompetenceDossier: *newDossier,
	}

	c.JSON(http.StatusOK, resp)
}
