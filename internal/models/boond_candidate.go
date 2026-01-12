package models

import "mime/multipart"

// BoondCandidateDeleteRequest représente la requête pour supprimer un candidat Boond
type BoondCandidateDeleteRequest struct {
	BoondCandidateId string `json:"boondCandidateId" binding:"required"`
	BoondJwt         string `json:"boondJwt" binding:"required"`
}

// BoondCandidateModifyRequest représente la requête pour modifier un candidat Boond
type BoondCandidateModifyRequest struct {
	BoondCandidateId string  `json:"boondCandidateId" binding:"required"`
	BoondJwt         string  `json:"boondJwt" binding:"required"`
	Email            *string `json:"email,omitempty"`
	Phone            *string `json:"phone,omitempty"`
	Mobilite         *string `json:"mobilite,omitempty"`
	Disponibilite    *string `json:"disponibilite,omitempty"`
	Status           *string `json:"status,omitempty"`
}

// BoondCandidateDCRequest représente la requête pour uploader un dossier de compétences
type BoondCandidateDCRequest struct {
	BoondCandidateId string                `form:"boondCandidateId" binding:"required"`
	BoondJwt         string                `form:"boondJwt" binding:"required"`
	File             *multipart.FileHeader `form:"file" binding:"required"`
	Filename         string                `form:"filename" binding:"required"`
}

// BoondCandidateResponse représente la réponse standard pour les opérations Boond
type BoondCandidateResponse struct {
	Success          bool   `json:"success"`
	Message          string `json:"message"`
	ErrorCode        string `json:"error_code,omitempty"`
	Details          string `json:"details,omitempty"`
	TechnicalDetails string `json:"technical_details,omitempty"`
	Data             any    `json:"data,omitempty"`
}
