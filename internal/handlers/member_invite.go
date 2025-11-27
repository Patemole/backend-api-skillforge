when package handlers

import (
	"log"
	"net/http"

	"backend-api-skillforge/internal/models"
	"backend-api-skillforge/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MemberInviteHandler gère les invitations de membres
type MemberInviteHandler struct {
	resendService *services.ResendService
}

// NewMemberInviteHandler crée une nouvelle instance du handler
func NewMemberInviteHandler() *MemberInviteHandler {
	return &MemberInviteHandler{
		resendService: services.NewResendService(),
	}
}

// HandleMemberInvite traite les requêtes d'invitation de membres
func (h *MemberInviteHandler) HandleMemberInvite(c *gin.Context) {
	var req models.MemberInviteRequest

	// Binding et validation de la requête
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("MEMBER_INVITE_ERROR: Erreur de validation de la requête: %v", err)
		c.JSON(http.StatusBadRequest, models.MemberInviteErrorResponse{
			Success:   false,
			Message:   "Données de requête invalides",
			ErrorCode: "INVALID_REQUEST",
		})
		return
	}

	// Générer un ID de requête unique
	requestID := uuid.New().String()

	// Préparer les données pour l'email
	emailData := models.MemberInviteEmailData{
		InviterEmail:     req.InviterEmail,
		RecipientEmail:   req.RecipientEmail,
		MemberLink:       req.MemberLink,
		OrganizationID:   req.OrganizationID,
		Role:             req.Role,
		InviterName:      req.InviterName,
		OrganizationName: req.OrganizationName,
	}

	// Envoyer l'email d'invitation
	emailResp, err := h.resendService.SendMemberInviteEmail(emailData)
	if err != nil {
		log.Printf("MEMBER_INVITE_ERROR: Erreur envoi email (request_id: %s): %v", requestID, err)
		c.JSON(http.StatusInternalServerError, models.MemberInviteErrorResponse{
			Success:   false,
			Message:   "Erreur lors de l'envoi de l'email",
			ErrorCode: "EMAIL_SEND_FAILED",
		})
		return
	}

	// Log de succès
	log.Printf("MEMBER_INVITE_SUCCESS: Email envoyé avec succès (request_id: %s, email_id: %s, inviter: %s, recipient: %s, org: %s, role: %s)",
		requestID, emailResp.ID, req.InviterEmail, req.RecipientEmail, req.OrganizationID, req.Role)

	// Réponse de succès
	c.JSON(http.StatusOK, models.MemberInviteResponse{
		Success:   true,
		Message:   "Email d'invitation envoyé avec succès",
		RequestID:  requestID,
	})
}

