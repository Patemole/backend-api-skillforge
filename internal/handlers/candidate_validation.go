package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"backend-api-skillforge/internal/models"
	"backend-api-skillforge/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CandidateValidationHandler gère les notifications de validation de candidat
type CandidateValidationHandler struct {
	resendService *services.ResendService
}

// NewCandidateValidationHandler crée une nouvelle instance du handler
func NewCandidateValidationHandler() *CandidateValidationHandler {
	return &CandidateValidationHandler{
		resendService: services.NewResendService(),
	}
}

// HandleCandidateValidation traite les requêtes de notification de validation candidat
func (h *CandidateValidationHandler) HandleCandidateValidation(c *gin.Context) {
	var req models.CandidateValidationRequest
	
	// Parser et valider la requête
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("ERROR: Invalid request payload: %v", err)
		c.JSON(http.StatusBadRequest, models.CandidateValidationResponse{
			Success: false,
			Message: "Invalid request payload",
			Error:   err.Error(),
		})
		return
	}

	// Validation supplémentaire des données
	if err := h.validateRequest(&req); err != nil {
		log.Printf("ERROR: Request validation failed: %v", err)
		c.JSON(http.StatusBadRequest, models.CandidateValidationResponse{
			Success: false,
			Message: "Request validation failed",
			Error:   err.Error(),
		})
		return
	}

	// Vérifier que le pourcentage de completion est à 100%
	if req.CompletionPercentage != 100 {
		log.Printf("WARNING: Completion percentage is not 100%%: %d", req.CompletionPercentage)
		c.JSON(http.StatusBadRequest, models.CandidateValidationResponse{
			Success: false,
			Message: "Notification can only be sent for 100% completion",
			Error:   fmt.Sprintf("Completion percentage is %d%%, expected 100%%", req.CompletionPercentage),
		})
		return
	}

	// Parser la date de validation
	validationDate, err := time.Parse(time.RFC3339, req.ValidationDate)
	if err != nil {
		log.Printf("ERROR: Invalid validation date format: %v", err)
		c.JSON(http.StatusBadRequest, models.CandidateValidationResponse{
			Success: false,
			Message: "Invalid validation date format",
			Error:   "Validation date must be in RFC3339 format (e.g., 2024-01-05T10:30:00.000Z)",
		})
		return
	}

	// Formater la date en français
	formattedDate := h.formatDateInFrench(validationDate)

	// Préparer les données pour le template email
	emailData := models.EmailTemplateData{
		CandidateName:    req.CandidateName,
		CandidateEmail:   req.CandidateEmail,
		InviterEmail:     req.InviterEmail,
		OrganizationName: req.OrganizationName,
		ValidationDate:   formattedDate,
		DossierURL:       req.DossierURL,
	}

	// Envoyer l'email de notification
	emailResp, err := h.resendService.SendCandidateValidationEmail(emailData)
	if err != nil {
		log.Printf("ERROR: Failed to send candidate validation email: %v", err)
		
		// Logger l'erreur pour audit
		h.logValidationAttempt(req, false, err.Error(), "")
		
		c.JSON(http.StatusInternalServerError, models.CandidateValidationResponse{
			Success: false,
			Message: "Failed to send notification email",
			Error:   err.Error(),
		})
		return
	}

	// Générer un ID de notification
	notificationID := uuid.New().String()

	// Logger le succès pour audit
	h.logValidationAttempt(req, true, "", notificationID)

	// Log de succès
	log.Printf("SUCCESS: Candidate validation notification sent - Candidate: %s, Inviter: %s, Email ID: %s", 
		req.CandidateName, req.InviterEmail, emailResp.ID)

	// Réponse de succès
	c.JSON(http.StatusOK, models.CandidateValidationResponse{
		Success:        true,
		Message:        "Notification envoyée avec succès",
		NotificationID: notificationID,
	})
}

// validateRequest effectue une validation supplémentaire des données
func (h *CandidateValidationHandler) validateRequest(req *models.CandidateValidationRequest) error {
	// Vérifier que les UUIDs sont valides
	if !h.isValidUUID(req.CandidateID) {
		return fmt.Errorf("invalid candidate_id format")
	}
	if !h.isValidUUID(req.OrganizationID) {
		return fmt.Errorf("invalid organization_id format")
	}
	if !h.isValidUUID(req.InviterID) {
		return fmt.Errorf("invalid inviter_id format")
	}
	if !h.isValidUUID(req.DossierID) {
		return fmt.Errorf("invalid dossier_id format")
	}

	// Vérifier que les emails sont valides
	if !h.resendService.ValidateEmailAddress(req.CandidateEmail) {
		return fmt.Errorf("invalid candidate_email format")
	}
	if !h.resendService.ValidateEmailAddress(req.InviterEmail) {
		return fmt.Errorf("invalid inviter_email format")
	}

	// Vérifier que l'URL du dossier est valide
	if !h.isValidURL(req.DossierURL) {
		return fmt.Errorf("invalid dossier_url format")
	}

	return nil
}

// isValidUUID vérifie si une chaîne est un UUID valide
func (h *CandidateValidationHandler) isValidUUID(str string) bool {
	_, err := uuid.Parse(str)
	return err == nil
}

// isValidURL vérifie si une chaîne est une URL valide (validation basique)
func (h *CandidateValidationHandler) isValidURL(str string) bool {
	return len(str) > 0 && 
		   (str[:7] == "http://" || str[:8] == "https://")
}

// formatDateInFrench formate une date en français
func (h *CandidateValidationHandler) formatDateInFrench(date time.Time) string {
	// Mois en français
	months := []string{
		"janvier", "février", "mars", "avril", "mai", "juin",
		"juillet", "août", "septembre", "octobre", "novembre", "décembre",
	}

	day := date.Day()
	month := months[date.Month()-1]
	year := date.Year()
	hour := date.Hour()
	minute := date.Minute()

	// Formater la date et l'heure
	if minute == 0 {
		return fmt.Sprintf("%d %s %d à %dh", day, month, year, hour)
	}
	return fmt.Sprintf("%d %s %d à %dh%02d", day, month, year, hour, minute)
}

// logValidationAttempt enregistre une tentative d'envoi de notification
func (h *CandidateValidationHandler) logValidationAttempt(req models.CandidateValidationRequest, success bool, errorMsg, notificationID string) {
	logEntry := models.ValidationLogEntry{
		Timestamp:      time.Now(),
		CandidateID:    req.CandidateID,
		InviterEmail:   req.InviterEmail,
		DossierID:      req.DossierID,
		OrganizationID: req.OrganizationID,
		Success:        success,
		ErrorMessage:   errorMsg,
		NotificationID: notificationID,
	}

	// Convertir en JSON pour le log
	logData, err := json.Marshal(logEntry)
	if err != nil {
		log.Printf("ERROR: Failed to marshal validation log entry: %v", err)
		return
	}

	// Logger selon le statut
	if success {
		log.Printf("VALIDATION_LOG_SUCCESS: %s", string(logData))
	} else {
		log.Printf("VALIDATION_LOG_ERROR: %s", string(logData))
	}
}

// GetValidationLogs retourne les logs de validation (pour debug/admin)
func (h *CandidateValidationHandler) GetValidationLogs(c *gin.Context) {
	// Cette fonction pourrait être implémentée pour retourner les logs
	// Pour l'instant, on retourne juste un message
	c.JSON(http.StatusOK, gin.H{
		"message": "Validation logs endpoint - to be implemented",
		"note":    "Check server logs for validation attempts",
	})
}
