package models

// MemberInviteRequest définit la structure de la requête pour inviter un membre
type MemberInviteRequest struct {
	InviterEmail     string `json:"inviter_email" binding:"required,email"`
	RecipientEmail   string `json:"recipient_email" binding:"required,email"`
	MemberLink       string `json:"member_link" binding:"required,url"`
	OrganizationID   string `json:"organization_id" binding:"required"`
	Role             string `json:"role" binding:"required,oneof=admin member"`
	InviterName      string `json:"inviter_name,omitempty"`
	OrganizationName string `json:"organization_name,omitempty"`
}

// MemberInviteResponse définit la structure de la réponse de succès
type MemberInviteResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// MemberInviteErrorResponse définit la structure de la réponse d'erreur
type MemberInviteErrorResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	ErrorCode string `json:"error_code,omitempty"`
}

// MemberInviteEmailData contient les données pour le template d'email d'invitation
type MemberInviteEmailData struct {
	InviterEmail     string
	RecipientEmail   string
	MemberLink       string
	OrganizationID   string
	Role             string
	InviterName      string
	OrganizationName string
}
