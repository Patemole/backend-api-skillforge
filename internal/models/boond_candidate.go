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

// BoondAgenciesRequest représente la requête pour récupérer les agences
type BoondAgenciesRequest struct {
	BoondJwt string `json:"boondJwt" binding:"required"`
}

// BoondAgency représente une agence Boond
type BoondAgency struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// BoondResourcesRequest représente la requête pour récupérer les ressources (filtrable)
type BoondResourcesRequest struct {
	BoondJwt   string `json:"boondJwt" binding:"required"`
	MaxResults *int   `json:"maxResults,omitempty"` // 1..500, défaut 500 par page

	// Filtres optionnels
	TypeOf    []int `json:"typeOf,omitempty"`    // Filtre par types de rôle (ex: [2,4,5] pour manager, direction, RH)
	IsVisible *bool `json:"isVisible,omitempty"` // Filtre par visibilité (true = visible, false = caché)
}

// BoondOrgChartRequest permet de demander l'organigramme (et optionnellement les liens RH)
type BoondOrgChartRequest struct {
	BoondJwt       string `json:"boondJwt" binding:"required"`
	MaxResults     *int   `json:"maxResults,omitempty"`
	IncludeHREdges *bool  `json:"includeHREdges,omitempty"`
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
