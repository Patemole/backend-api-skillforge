package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"

	"backend-api-skillforge/internal/boond"
	"backend-api-skillforge/internal/models"
	"backend-api-skillforge/internal/orgchart"
	"backend-api-skillforge/internal/supabase"
)

// DeleteBoondCandidate gère la suppression d'un candidat Boond
func DeleteBoondCandidate(c *gin.Context) {
	var req models.BoondCandidateDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BoondCandidateResponse{
			Success: false,
			Message: "Données de requête invalides: " + err.Error(),
		})
		return
	}

	// Validation des champs requis
	if strings.TrimSpace(req.BoondCandidateId) == "" {
		c.JSON(http.StatusBadRequest, models.BoondCandidateResponse{
			Success: false,
			Message: "L'ID du candidat Boond est requis",
		})
		return
	}

	if strings.TrimSpace(req.BoondJwt) == "" {
		c.JSON(http.StatusBadRequest, models.BoondCandidateResponse{
			Success: false,
			Message: "Le JWT Boond est requis",
		})
		return
	}

	// Créer le client Boond
	client := boond.New(req.BoondJwt)

	// Supprimer le candidat
	if err := client.DeleteCandidate(c.Request.Context(), req.BoondCandidateId); err != nil {
		// Gestion d'erreurs spécifiques selon le type d'erreur
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, models.BoondCandidateResponse{
				Success:   false,
				Message:   "Candidat non trouvé",
				ErrorCode: "CANDIDATE_NOT_FOUND",
				Details:   "Le candidat avec l'ID '" + req.BoondCandidateId + "' n'existe pas dans Boond Manager",
			})
			return
		}

		if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusUnauthorized, models.BoondCandidateResponse{
				Success:   false,
				Message:   "Authentification échouée",
				ErrorCode: "AUTHENTICATION_FAILED",
				Details:   "Le token JWT fourni n'est pas valide ou a expiré",
			})
			return
		}

		if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "forbidden") {
			c.JSON(http.StatusForbidden, models.BoondCandidateResponse{
				Success:   false,
				Message:   "Accès refusé",
				ErrorCode: "INSUFFICIENT_PERMISSIONS",
				Details:   "Vous n'avez pas les permissions nécessaires pour supprimer ce candidat",
			})
			return
		}

		// Erreur générique
		c.JSON(http.StatusInternalServerError, models.BoondCandidateResponse{
			Success:          false,
			Message:          "Échec de la suppression du candidat",
			ErrorCode:        "DELETE_CANDIDATE_FAILED",
			Details:          "Une erreur inattendue s'est produite lors de la suppression du candidat dans Boond Manager",
			TechnicalDetails: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BoondCandidateResponse{
		Success: true,
		Message: "Candidat supprimé avec succès",
		Data: map[string]string{
			"boondCandidateId": req.BoondCandidateId,
		},
	})
}

// BuildBoondOrgChart construit l'organigramme et renvoie la structure + un graphe DOT
func BuildBoondOrgChart(c *gin.Context) {
	var req models.BoondOrgChartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BoondCandidateResponse{
			Success: false,
			Message: "Données de requête invalides: " + err.Error(),
		})
		return
	}

	if strings.TrimSpace(req.BoondJwt) == "" {
		c.JSON(http.StatusBadRequest, models.BoondCandidateResponse{
			Success: false,
			Message: "Le JWT Boond est requis",
		})
		return
	}

	client := boond.New(req.BoondJwt)
	max := 500
	if req.MaxResults != nil {
		if *req.MaxResults < 1 {
			max = 1
		} else if *req.MaxResults > 500 {
			max = 500
		} else {
			max = *req.MaxResults
		}
	}

	resources, err := client.GetAllResources(c.Request.Context(), max, nil, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.BoondCandidateResponse{
			Success:          false,
			Message:          "Échec de la récupération des ressources",
			ErrorCode:        "GET_RESOURCES_FAILED",
			Details:          "Impossible de générer l'organigramme sans ressources",
			TechnicalDetails: err.Error(),
		})
		return
	}

	includeHR := false
	if req.IncludeHREdges != nil {
		includeHR = *req.IncludeHREdges
	}

	roots, dot, orphanCount := orgchart.BuildOrgChartFromResources(resources, includeHR)

	c.JSON(http.StatusOK, models.BoondCandidateResponse{
		Success: true,
		Message: "Organigramme généré",
		Data: map[string]any{
			"roots":       roots,
			"dot":         dot,
			"orphanCount": orphanCount,
			"count":       len(resources),
		},
	})
}

// ModifyBoondCandidate gère la modification d'un candidat Boond
func ModifyBoondCandidate(c *gin.Context) {
	var req models.BoondCandidateModifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BoondCandidateResponse{
			Success: false,
			Message: "Données de requête invalides: " + err.Error(),
		})
		return
	}

	// Validation des champs requis
	if strings.TrimSpace(req.BoondCandidateId) == "" {
		c.JSON(http.StatusBadRequest, models.BoondCandidateResponse{
			Success: false,
			Message: "L'ID du candidat Boond est requis",
		})
		return
	}

	if strings.TrimSpace(req.BoondJwt) == "" {
		c.JSON(http.StatusBadRequest, models.BoondCandidateResponse{
			Success: false,
			Message: "Le JWT Boond est requis",
		})
		return
	}

	// Vérifier qu'au moins un champ à modifier est fourni
	if req.Email == nil && req.Phone == nil && req.Mobilite == nil && req.Disponibilite == nil && req.Status == nil {
		c.JSON(http.StatusBadRequest, models.BoondCandidateResponse{
			Success: false,
			Message: "Au moins un champ à modifier doit être fourni (email, phone, mobilite, disponibilite, status)",
		})
		return
	}

	// Créer le client Boond
	client := boond.New(req.BoondJwt)

	// Construire les attributs à mettre à jour
	attributes := make(map[string]any)

	if req.Email != nil && strings.TrimSpace(*req.Email) != "" {
		attributes["email1"] = strings.TrimSpace(*req.Email)
	}

	if req.Phone != nil && strings.TrimSpace(*req.Phone) != "" {
		attributes["phone1"] = strings.TrimSpace(*req.Phone)
	}

	if req.Mobilite != nil && strings.TrimSpace(*req.Mobilite) != "" {
		// Mobilité - récupérer les IDs depuis Boond
		mobilityMap, err := client.GetMobilityAreas(c.Request.Context())
		if err != nil {
			fmt.Printf("⚠️  [Boond] Erreur récupération zones mobilité: %v\n", err)
		} else {
			// Parser les zones demandées (séparées par virgule)
			requestedAreas := strings.Split(strings.TrimSpace(*req.Mobilite), ",")
			var mobilityIDs []string

			for _, area := range requestedAreas {
				area = strings.TrimSpace(area)
				if id, exists := mobilityMap[area]; exists {
					mobilityIDs = append(mobilityIDs, id)
					fmt.Printf("✅ [Boond] Zone trouvée: %s -> ID %s\n", area, id)
				} else {
					fmt.Printf("⚠️  [Boond] Zone non trouvée: %s\n", area)
				}
			}

			if len(mobilityIDs) > 0 {
				attributes["mobilityAreas"] = mobilityIDs
				fmt.Printf("✅ [Boond] Mobilité mise à jour: %v\n", mobilityIDs)
			}
		}
	}

	if req.Disponibilite != nil && strings.TrimSpace(*req.Disponibilite) != "" {
		// Disponibilité - récupérer les types depuis Boond ou traiter comme date
		dispo := strings.TrimSpace(*req.Disponibilite)

		// Vérifier si c'est une date au format YYYY-MM-DD
		if matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, dispo); matched {
			attributes["availability"] = dispo
			fmt.Printf("✅ [Boond] Disponibilité (date): %s\n", dispo)
		} else {
			// Essayer de trouver dans les types prédéfinis
			availabilityMap, err := client.GetAvailabilityTypes(c.Request.Context())
			if err != nil {
				fmt.Printf("⚠️  [Boond] Erreur récupération types disponibilité: %v\n", err)
				// Fallback: traiter comme ID numérique
				if matched, _ := regexp.MatchString(`^\d+$`, dispo); matched {
					attributes["availability"] = dispo
					fmt.Printf("✅ [Boond] Disponibilité (ID fallback): %s\n", dispo)
				}
			} else {
				// Chercher le type correspondant
				if id, exists := availabilityMap[dispo]; exists {
					attributes["availability"] = id
					fmt.Printf("✅ [Boond] Type trouvé: %s -> ID %s\n", dispo, id)
				} else {
					fmt.Printf("⚠️  [Boond] Type de disponibilité non trouvé: %s\n", dispo)
					// Fallback: traiter comme ID numérique
					if matched, _ := regexp.MatchString(`^\d+$`, dispo); matched {
						attributes["availability"] = dispo
						fmt.Printf("✅ [Boond] Disponibilité (ID fallback): %s\n", dispo)
					}
				}
			}
		}
	}

	if req.Status != nil && strings.TrimSpace(*req.Status) != "" {
		// Mapping des statuts vers les IDs Boond
		statusMap := map[string]int{
			"a_traiter":             1,
			"vivier":                2,
			"top_profil":            3,
			"ne_plus_contacter":     4,
			"converti_en_ressource": 5,
			"non_defini":            0,
		}

		status := strings.TrimSpace(*req.Status)
		if stateID, exists := statusMap[status]; exists {
			attributes["state"] = stateID
			fmt.Printf("✅ [Boond] Statut mappé: %s -> ID %d\n", status, stateID)
		} else {
			fmt.Printf("⚠️  [Boond] Statut non reconnu: %s (valeurs acceptées: a_traiter, vivier, top_profil, ne_plus_contacter, converti_en_ressource, non_defini)\n", status)
		}
	}

	// Mettre à jour le candidat
	response, err := client.UpdateCandidate(c.Request.Context(), req.BoondCandidateId, attributes)
	if err != nil {
		// Gestion d'erreurs spécifiques selon le type d'erreur
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, models.BoondCandidateResponse{
				Success:   false,
				Message:   "Candidat non trouvé",
				ErrorCode: "CANDIDATE_NOT_FOUND",
				Details:   "Le candidat avec l'ID '" + req.BoondCandidateId + "' n'existe pas dans Boond Manager",
			})
			return
		}

		if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusUnauthorized, models.BoondCandidateResponse{
				Success:   false,
				Message:   "Authentification échouée",
				ErrorCode: "AUTHENTICATION_FAILED",
				Details:   "Le token JWT fourni n'est pas valide ou a expiré",
			})
			return
		}

		if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "forbidden") {
			c.JSON(http.StatusForbidden, models.BoondCandidateResponse{
				Success:   false,
				Message:   "Accès refusé",
				ErrorCode: "INSUFFICIENT_PERMISSIONS",
				Details:   "Vous n'avez pas les permissions nécessaires pour modifier ce candidat",
			})
			return
		}

		if strings.Contains(err.Error(), "422") || strings.Contains(err.Error(), "validation") {
			c.JSON(http.StatusUnprocessableEntity, models.BoondCandidateResponse{
				Success:          false,
				Message:          "Données invalides",
				ErrorCode:        "VALIDATION_ERROR",
				Details:          "Les données fournies ne respectent pas le format attendu par Boond Manager",
				TechnicalDetails: err.Error(),
			})
			return
		}

		// Erreur générique
		c.JSON(http.StatusInternalServerError, models.BoondCandidateResponse{
			Success:          false,
			Message:          "Échec de la modification du candidat",
			ErrorCode:        "UPDATE_CANDIDATE_FAILED",
			Details:          "Une erreur inattendue s'est produite lors de la modification du candidat dans Boond Manager",
			TechnicalDetails: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BoondCandidateResponse{
		Success: true,
		Message: "Candidat modifié avec succès",
		Data: map[string]any{
			"boondCandidateId": req.BoondCandidateId,
			"updatedFields":    attributes,
			"boondResponse":    response,
		},
	})
}

// UploadBoondCandidateDC gère l'upload d'un dossier de compétences vers un candidat Boond
func UploadBoondCandidateDC(c *gin.Context) {
	var req models.BoondCandidateDCRequest

	// Parser le multipart form
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BoondCandidateResponse{
			Success: false,
			Message: "Erreur de parsing: " + err.Error(),
		})
		return
	}

	// Ouvrir le fichier
	file, err := req.File.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BoondCandidateResponse{
			Success: false,
			Message: "Impossible d'ouvrir le fichier: " + err.Error(),
		})
		return
	}
	defer file.Close()

	// Lire le contenu du fichier
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BoondCandidateResponse{
			Success: false,
			Message: "Impossible de lire le fichier: " + err.Error(),
		})
		return
	}

	// Créer le client Boond
	client := boond.New(req.BoondJwt)

	// Uploader le dossier de compétences
	docID, err := client.UploadDocument(c.Request.Context(), req.BoondCandidateId, fileBytes, req.Filename)
	if err != nil {
		// Gestion d'erreurs spécifiques selon le type d'erreur
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, models.BoondCandidateResponse{
				Success:   false,
				Message:   "Candidat non trouvé",
				ErrorCode: "CANDIDATE_NOT_FOUND",
				Details:   "Le candidat avec l'ID '" + req.BoondCandidateId + "' n'existe pas dans Boond Manager",
			})
			return
		}

		if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusUnauthorized, models.BoondCandidateResponse{
				Success:   false,
				Message:   "Authentification échouée",
				ErrorCode: "AUTHENTICATION_FAILED",
				Details:   "Le token JWT fourni n'est pas valide ou a expiré",
			})
			return
		}

		if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "forbidden") {
			c.JSON(http.StatusForbidden, models.BoondCandidateResponse{
				Success:   false,
				Message:   "Accès refusé",
				ErrorCode: "INSUFFICIENT_PERMISSIONS",
				Details:   "Vous n'avez pas les permissions nécessaires pour uploader des documents pour ce candidat",
			})
			return
		}

		if strings.Contains(err.Error(), "413") || strings.Contains(err.Error(), "too large") {
			c.JSON(http.StatusRequestEntityTooLarge, models.BoondCandidateResponse{
				Success:   false,
				Message:   "Fichier trop volumineux",
				ErrorCode: "FILE_TOO_LARGE",
				Details:   "Le fichier '" + req.Filename + "' dépasse la taille maximale autorisée par Boond Manager",
			})
			return
		}

		if strings.Contains(err.Error(), "415") || strings.Contains(err.Error(), "unsupported media type") {
			c.JSON(http.StatusUnsupportedMediaType, models.BoondCandidateResponse{
				Success:   false,
				Message:   "Type de fichier non supporté",
				ErrorCode: "UNSUPPORTED_FILE_TYPE",
				Details:   "Le type de fichier '" + req.Filename + "' n'est pas supporté par Boond Manager",
			})
			return
		}

		// Erreur générique
		c.JSON(http.StatusInternalServerError, models.BoondCandidateResponse{
			Success:          false,
			Message:          "Échec de l'upload du dossier de compétences",
			ErrorCode:        "UPLOAD_DOCUMENT_FAILED",
			Details:          "Une erreur inattendue s'est produite lors de l'upload du document dans Boond Manager",
			TechnicalDetails: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BoondCandidateResponse{
		Success: true,
		Message: "Dossier de compétences uploadé avec succès",
		Data: map[string]string{
			"boondCandidateId": req.BoondCandidateId,
			"documentId":       docID,
			"filename":         req.Filename,
		},
	})
}

// GetBoondAgencies gère la récupération des agences Boond
func GetBoondAgencies(c *gin.Context) {
	var req models.BoondAgenciesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BoondCandidateResponse{
			Success: false,
			Message: "Données de requête invalides: " + err.Error(),
		})
		return
	}

	// Validation des champs requis
	if strings.TrimSpace(req.BoondJwt) == "" {
		c.JSON(http.StatusBadRequest, models.BoondCandidateResponse{
			Success: false,
			Message: "Le JWT Boond est requis",
		})
		return
	}

	// Créer le client Boond
	client := boond.New(req.BoondJwt)

	// Récupérer les agences
	agencies, err := client.GetAgencies(c.Request.Context())
	if err != nil {
		// Gestion d'erreurs spécifiques selon le type d'erreur
		if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusUnauthorized, models.BoondCandidateResponse{
				Success:   false,
				Message:   "Authentification échouée",
				ErrorCode: "AUTHENTICATION_FAILED",
				Details:   "Le token JWT fourni n'est pas valide ou a expiré",
			})
			return
		}

		if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "forbidden") {
			c.JSON(http.StatusForbidden, models.BoondCandidateResponse{
				Success:   false,
				Message:   "Accès refusé",
				ErrorCode: "INSUFFICIENT_PERMISSIONS",
				Details:   "Vous n'avez pas les permissions nécessaires pour récupérer les agences",
			})
			return
		}

		// Erreur générique
		c.JSON(http.StatusInternalServerError, models.BoondCandidateResponse{
			Success:          false,
			Message:          "Échec de la récupération des agences",
			ErrorCode:        "GET_AGENCIES_FAILED",
			Details:          "Une erreur inattendue s'est produite lors de la récupération des agences depuis Boond Manager",
			TechnicalDetails: err.Error(),
		})
		return
	}

	// Convertir en format BoondAgency
	var boondAgencies []models.BoondAgency
	for _, agency := range agencies {
		boondAgencies = append(boondAgencies, models.BoondAgency{
			ID:   agency["id"],
			Name: agency["name"],
		})
	}

	c.JSON(http.StatusOK, models.BoondCandidateResponse{
		Success: true,
		Message: "Agences récupérées avec succès",
		Data: map[string]any{
			"agencies": boondAgencies,
			"count":    len(boondAgencies),
		},
	})
}

// GetBoondResources gère la récupération paginée de toutes les ressources Boond
func GetBoondResources(c *gin.Context) {
	var req models.BoondResourcesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BoondCandidateResponse{
			Success: false,
			Message: "Données de requête invalides: " + err.Error(),
		})
		return
	}

	if strings.TrimSpace(req.BoondJwt) == "" {
		c.JSON(http.StatusBadRequest, models.BoondCandidateResponse{
			Success: false,
			Message: "Le JWT Boond est requis",
		})
		return
	}

	client := boond.New(req.BoondJwt)

	max := 500
	if req.MaxResults != nil {
		if *req.MaxResults < 1 {
			max = 1
		} else if *req.MaxResults > 500 {
			max = 500
		} else {
			max = *req.MaxResults
		}
	}

	resources, err := client.GetAllResources(c.Request.Context(), max, req.TypeOf, req.IsVisible)
	if err != nil {
		if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusUnauthorized, models.BoondCandidateResponse{
				Success:   false,
				Message:   "Authentification échouée",
				ErrorCode: "AUTHENTICATION_FAILED",
				Details:   "Le token JWT fourni n'est pas valide ou a expiré",
			})
			return
		}

		if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "forbidden") {
			c.JSON(http.StatusForbidden, models.BoondCandidateResponse{
				Success:   false,
				Message:   "Accès refusé",
				ErrorCode: "INSUFFICIENT_PERMISSIONS",
				Details:   "Vous n'avez pas les permissions nécessaires pour récupérer les ressources",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.BoondCandidateResponse{
			Success:          false,
			Message:          "Échec de la récupération des ressources",
			ErrorCode:        "GET_RESOURCES_FAILED",
			Details:          "Une erreur inattendue s'est produite lors de la récupération des ressources depuis Boond Manager",
			TechnicalDetails: err.Error(),
		})
		return
	}

	// Créer une liste simplifiée pour les équipes internes (managers, direction, RH)
	var filteredResources []map[string]any
	for _, resource := range resources {
		attrs, ok := resource["attributes"].(map[string]any)
		if !ok {
			continue
		}

		// Extraire les infos essentielles
		simplified := map[string]any{
			"id":        resource["id"],
			"email":     attrs["email1"],
			"typeOf":    attrs["typeOf"],
			"firstName": attrs["firstName"],
			"lastName":  attrs["lastName"],
			"title":     attrs["title"],
		}
		filteredResources = append(filteredResources, simplified)
	}

	c.JSON(http.StatusOK, models.BoondCandidateResponse{
		Success: true,
		Message: "Ressources récupérées avec succès",
		Data: map[string]any{
			"resources":         resources,
			"count":             len(resources),
			"filteredResources": filteredResources,
		},
	})
}

// SyncBoondManagerID synchronise le manager ID d'un utilisateur depuis les ressources tester Boond
// Cette fonction:
// 1. Récupère les ressources "tester" depuis Boond
// 2. Trouve la ressource correspondant à l'email de l'utilisateur
// 3. Extrait le manager ID de cette ressource
// 4. Stocke le manager ID dans le profil utilisateur (boond_manager.boondManagerId)
func SyncBoondManagerID(c *gin.Context) {
	var req models.BoondSyncManagerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BoondCandidateResponse{
			Success: false,
			Message: "Données de requête invalides: " + err.Error(),
		})
		return
	}

	// Validation des champs requis
	if strings.TrimSpace(req.BoondJwt) == "" {
		c.JSON(http.StatusBadRequest, models.BoondCandidateResponse{
			Success: false,
			Message: "Le JWT Boond est requis",
		})
		return
	}

	if strings.TrimSpace(req.UserEmail) == "" {
		c.JSON(http.StatusBadRequest, models.BoondCandidateResponse{
			Success: false,
			Message: "L'email utilisateur est requis",
		})
		return
	}

	if strings.TrimSpace(req.UserID) == "" {
		c.JSON(http.StatusBadRequest, models.BoondCandidateResponse{
			Success: false,
			Message: "L'ID utilisateur est requis",
		})
		return
	}

	// Créer le client Boond
	client := boond.New(req.BoondJwt)

	// Trouver la ressource tester par email et extraire le manager ID
	managerID, err := client.FindTesterResourceByEmail(c.Request.Context(), req.UserEmail)
	if err != nil {
		log.Printf("❌ [SyncBoondManagerID] Erreur lors de la recherche de la ressource tester: %v", err)
		c.JSON(http.StatusInternalServerError, models.BoondCandidateResponse{
			Success:          false,
			Message:          "Échec de la recherche de la ressource tester",
			ErrorCode:        "FIND_TESTER_RESOURCE_FAILED",
			Details:          "Une erreur s'est produite lors de la recherche de la ressource tester dans Boond Manager",
			TechnicalDetails: err.Error(),
		})
		return
	}

	if managerID == "" {
		log.Printf("⚠️  [SyncBoondManagerID] Aucune ressource tester trouvée pour l'email: %s", req.UserEmail)
		c.JSON(http.StatusOK, models.BoondCandidateResponse{
			Success: true,
			Message: "Aucune ressource tester trouvée pour cet email",
			Data: map[string]any{
				"managerId": nil,
				"found":     false,
			},
		})
		return
	}

	log.Printf("✅ [SyncBoondManagerID] Manager ID trouvé: %s pour l'email: %s", managerID, req.UserEmail)

	// Récupérer le profil actuel pour préserver les autres champs de boond_manager
	profileData, _, err := supabase.Client.
		From("profiles").
		Select("boond_manager", "exact", false).
		Eq("user_id", req.UserID).
		Single().
		Execute()

	var boondManagerData map[string]any
	if err == nil && len(profileData) > 0 {
		// Parser le JSON existant
		var profile struct {
			BoondManager json.RawMessage `json:"boond_manager"`
		}
		if err := json.Unmarshal(profileData, &profile); err == nil && len(profile.BoondManager) > 0 {
			if err := json.Unmarshal(profile.BoondManager, &boondManagerData); err != nil {
				log.Printf("⚠️  [SyncBoondManagerID] Erreur parsing boond_manager existant: %v", err)
				boondManagerData = make(map[string]any)
			}
		} else {
			boondManagerData = make(map[string]any)
		}
	} else {
		boondManagerData = make(map[string]any)
	}

	// Mettre à jour le manager ID
	boondManagerData["boondManagerId"] = managerID

	// Convertir en JSON pour l'update
	boondManagerJSON, err := json.Marshal(boondManagerData)
	if err != nil {
		log.Printf("❌ [SyncBoondManagerID] Erreur marshalling boond_manager: %v", err)
		c.JSON(http.StatusInternalServerError, models.BoondCandidateResponse{
			Success:          false,
			Message:          "Échec de la préparation des données",
			ErrorCode:        "MARSHAL_ERROR",
			TechnicalDetails: err.Error(),
		})
		return
	}

	// Mettre à jour le profil dans Supabase
	updateData := map[string]interface{}{
		"boond_manager": string(boondManagerJSON),
	}

	_, _, err = supabase.Client.
		From("profiles").
		Update(updateData, "", "").
		Eq("user_id", req.UserID).
		Execute()

	if err != nil {
		log.Printf("❌ [SyncBoondManagerID] Erreur lors de la mise à jour du profil: %v", err)
		c.JSON(http.StatusInternalServerError, models.BoondCandidateResponse{
			Success:          false,
			Message:          "Échec de la mise à jour du profil",
			ErrorCode:        "UPDATE_PROFILE_FAILED",
			Details:          "Une erreur s'est produite lors de la mise à jour du profil utilisateur",
			TechnicalDetails: err.Error(),
		})
		return
	}

	log.Printf("✅ [SyncBoondManagerID] Profil mis à jour avec succès pour user_id: %s, manager_id: %s", req.UserID, managerID)

	c.JSON(http.StatusOK, models.BoondCandidateResponse{
		Success: true,
		Message: "Manager ID synchronisé avec succès",
		Data: map[string]any{
			"managerId": managerID,
			"found":     true,
			"userId":    req.UserID,
		},
	})
}
