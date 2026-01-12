package models

// CandidateInviteRequest définit la structure de la requête pour inviter un candidat
type CandidateInviteRequest struct {
	InviterEmail   string `json:"inviter_email" binding:"required,email"`
	RecipientEmail string `json:"recipient_email" binding:"required,email"`
	CandidateLink  string `json:"candidate_link" binding:"required,url"`
	DossierID      string `json:"dossier_id" binding:"required"`
	CandidateID    string `json:"candidate_id" binding:"required"`
}

// CandidateInviteResponse définit la structure de la réponse de succès
type CandidateInviteResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// CandidateInviteErrorResponse définit la structure de la réponse d'erreur
type CandidateInviteErrorResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	ErrorCode string `json:"error_code,omitempty"`
}

// CandidateInviteEmailData contient les données pour le template d'email d'invitation
type CandidateInviteEmailData struct {
	InviterEmail   string
	RecipientEmail string
	CandidateLink  string
	DossierID      string
	CandidateID    string
}
