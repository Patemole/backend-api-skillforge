package models

import "time"

// CandidateValidationRequest définit la structure de la requête pour notifier la validation d'un candidat
type CandidateValidationRequest struct {
	CandidateID         string `json:"candidate_id" binding:"required"`
	CandidateName       string `json:"candidate_name" binding:"required"`
	CandidateEmail      string `json:"candidate_email" binding:"required,email"`
	OrganizationID      string `json:"organization_id" binding:"required"`
	OrganizationName    string `json:"organization_name" binding:"required"`
	InviterID           string `json:"inviter_id" binding:"required"`
	InviterEmail        string `json:"inviter_email" binding:"required,email"`
	InviterName         string `json:"inviter_name" binding:"required"`
	DossierID           string `json:"dossier_id" binding:"required"`
	CompletionPercentage int   `json:"completion_percentage" binding:"required,min=0,max=100"`
	ValidationDate      string `json:"validation_date" binding:"required"`
	DossierURL          string `json:"dossier_url" binding:"required,url"`
}

// CandidateValidationResponse définit la structure de la réponse pour la notification de validation
type CandidateValidationResponse struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	NotificationID string `json:"notification_id,omitempty"`
	Error          string `json:"error,omitempty"`
}

// ResendEmailRequest définit la structure pour l'envoi d'email via Resend
type ResendEmailRequest struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	HTML    string `json:"html"`
}

// ResendEmailResponse définit la structure de la réponse de Resend
type ResendEmailResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	To      string `json:"to"`
	From    string `json:"from"`
	Subject string `json:"subject"`
	HTML    string `json:"html"`
	Text    string `json:"text"`
	Created string `json:"created_at"`
}

// ResendErrorResponse définit la structure d'erreur de Resend
type ResendErrorResponse struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	Status  int    `json:"statusCode"`
}

// EmailTemplateData contient les données pour le template d'email
type EmailTemplateData struct {
	CandidateName    string
	CandidateEmail   string
	InviterEmail     string
	OrganizationName string
	ValidationDate   string
	DossierURL       string
}

// ValidationLogEntry pour logger les notifications envoyées
type ValidationLogEntry struct {
	Timestamp      time.Time `json:"timestamp"`
	CandidateID    string    `json:"candidate_id"`
	InviterEmail   string    `json:"inviter_email"`
	DossierID      string    `json:"dossier_id"`
	OrganizationID string    `json:"organization_id"`
	Success        bool      `json:"success"`
	ErrorMessage   string    `json:"error_message,omitempty"`
	NotificationID string    `json:"notification_id,omitempty"`
}
