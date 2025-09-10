package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"backend-api-skillforge/internal/models"
	"backend-api-skillforge/internal/services"
)

// CandidateInviteHandler gère les invitations de candidats
type CandidateInviteHandler struct {
	resendService *services.ResendService
}

// NewCandidateInviteHandler crée une nouvelle instance du handler
func NewCandidateInviteHandler() *CandidateInviteHandler {
	return &CandidateInviteHandler{
		resendService: services.NewResendService(),
	}
}

// HandleCandidateInvite traite les requêtes d'invitation de candidats
func (h *CandidateInviteHandler) HandleCandidateInvite(c *gin.Context) {
	var req models.CandidateInviteRequest
	
	// Binding et validation de la requête
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("CANDIDATE_INVITE_ERROR: Erreur de validation de la requête: %v", err)
		c.JSON(http.StatusBadRequest, models.CandidateInviteErrorResponse{
			Success:   false,
			Message:   "Données de requête invalides",
			ErrorCode: "INVALID_REQUEST",
		})
		return
	}

	// Générer un ID de requête unique
	requestID := uuid.New().String()

	// Préparer les données pour l'email
	emailData := models.CandidateInviteEmailData{
		InviterEmail:   req.InviterEmail,
		RecipientEmail: req.RecipientEmail,
		CandidateLink:  req.CandidateLink,
		DossierID:      req.DossierID,
		CandidateID:    req.CandidateID,
	}

	// Envoyer l'email d'invitation
	emailResp, err := h.resendService.SendCandidateInviteEmail(emailData)
	if err != nil {
		log.Printf("CANDIDATE_INVITE_ERROR: Erreur envoi email (request_id: %s): %v", requestID, err)
		c.JSON(http.StatusInternalServerError, models.CandidateInviteErrorResponse{
			Success:   false,
			Message:   "Erreur lors de l'envoi de l'email",
			ErrorCode: "EMAIL_SEND_FAILED",
		})
		return
	}

	// Log de succès
	log.Printf("CANDIDATE_INVITE_SUCCESS: Email envoyé avec succès (request_id: %s, email_id: %s, inviter: %s, recipient: %s, dossier: %s)", 
		requestID, emailResp.ID, req.InviterEmail, req.RecipientEmail, req.DossierID)

	// Réponse de succès
	c.JSON(http.StatusOK, models.CandidateInviteResponse{
		Success:   true,
		Message:   "Email d'invitation envoyé avec succès",
		RequestID: requestID,
	})
}
