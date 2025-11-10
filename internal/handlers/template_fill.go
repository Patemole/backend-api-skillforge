package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"backend-api-skillforge/internal/services"
	"backend-api-skillforge/internal/supabase"
)

type FillTemplateRequest struct {
	OrganizationID string         `json:"organization_id" binding:"required"`
	CandidateID    string         `json:"candidate_id"`
	CandidateData  map[string]any `json:"candidate_data"`
}

type organizationTemplate struct {
	TemplateBasique map[string]any `json:"template_basique"`
}

func FillTemplateWithAnthropic(c *gin.Context) {
	var req FillTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload", "detail": err.Error()})
		return
	}

	// Load organization template_basique
	data, _, err := supabase.Client.
		From("organizations").
		Select("template_basique", "exact", false).
		Eq("id", req.OrganizationID).
		Limit(1, "").
		Execute()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch organization", "detail": err.Error()})
		return
	}

	var orgs []organizationTemplate
	if err := json.Unmarshal(data, &orgs); err != nil || len(orgs) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		return
	}

	templateBasique := orgs[0].TemplateBasique
	if templateBasique == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization has no template_basique"})
		return
	}

	sanitizedHtml, _ := templateBasique["customTemplateHtml"].(string)
	if strings.TrimSpace(sanitizedHtml) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "custom template HTML is empty"})
		return
	}

	variableDelimiter, _ := templateBasique["variableDelimiter"].(string)
	if strings.TrimSpace(variableDelimiter) == "" {
		variableDelimiter = "$var$"
	}

	// Load candidate data if not provided inline
	candidateData := req.CandidateData
	if candidateData == nil {
		candidateData, err = fetchCandidateSnapshot(req.CandidateID, req.OrganizationID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to load candidate data", "detail": err.Error()})
			return
		}
	}

	service, err := services.NewTemplateFillService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filledHTML, err := service.FillTemplate(c.Request.Context(), sanitizedHtml, variableDelimiter, candidateData)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "anthropic call failed", "detail": err.Error()})
		return
	}

	templateBasique["customTemplateHtml"] = filledHTML
	templateBasique["updatedAt"] = time.Now().UTC().Format(time.RFC3339)

	_, _, err = supabase.Client.
		From("organizations").
		Update(map[string]any{
			"template_basique": templateBasique,
		}, "representation", "").
		Eq("id", req.OrganizationID).
		Execute()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist filled template", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"html":    filledHTML,
	})
}

// fetchCandidateSnapshot loads candidate data from Supabase. Adjust the query to match your schema.
func fetchCandidateSnapshot(candidateID string, organizationID string) (map[string]any, error) {
	if strings.TrimSpace(candidateID) == "" {
		return nil, fmt.Errorf("candidate_id is required when candidate_data is not provided")
	}

	// Example query: adapt "competence_dossiers" to your actual table/view containing the DC data.
	data, _, err := supabase.Client.
		From("competence_dossiers").
		Select("*", "exact", false).
		Eq("candidate_id", candidateID).
		Limit(1, "").
		Execute()
	if err != nil {
		return nil, err
	}

	var rows []map[string]any
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("candidate data not found")
	}

	return rows[0], nil
}
