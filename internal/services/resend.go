package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"backend-api-skillforge/internal/models"
)

// ResendService gère l'envoi d'emails via l'API Resend
type ResendService struct {
	APIKey string
	BaseURL string
	FromEmail string
}

// NewResendService crée une nouvelle instance du service Resend
func NewResendService() *ResendService {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		panic("RESEND_API_KEY environment variable is required")
	}

	fromEmail := os.Getenv("RESEND_FROM_EMAIL")
	if fromEmail == "" {
		fromEmail = "noreply@getskillforge.app" // Email par défaut
	}

	return &ResendService{
		APIKey:    apiKey,
		BaseURL:   "https://api.resend.com",
		FromEmail: fromEmail,
	}
}

// SendCandidateValidationEmail envoie un email de notification de validation candidat
func (r *ResendService) SendCandidateValidationEmail(data models.EmailTemplateData) (*models.ResendEmailResponse, error) {
	// Générer le template HTML
	htmlContent, err := r.generateCandidateValidationHTML(data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate HTML template: %w", err)
	}

	// Préparer la requête
	emailReq := models.ResendEmailRequest{
		From:    r.FromEmail,
		To:      data.InviterEmail, // L'email sera envoyé à l'inviteur, pas au candidat
		Subject: fmt.Sprintf("✅ Dossier de compétences validé - %s", data.CandidateName),
		HTML:    htmlContent,
	}

	// Envoyer l'email
	return r.sendEmail(emailReq)
}

// SendCandidateInviteEmail envoie un email d'invitation au candidat
func (r *ResendService) SendCandidateInviteEmail(data models.CandidateInviteEmailData) (*models.ResendEmailResponse, error) {
	// Générer le template HTML
	htmlContent, err := r.generateCandidateInviteHTML(data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate HTML template: %w", err)
	}

	// Préparer la requête
	emailReq := models.ResendEmailRequest{
		From:    r.FromEmail,
		To:      data.RecipientEmail, // L'email sera envoyé au candidat
		Subject: "📝 Votre dossier de compétences vous attend sur SkillForge",
		HTML:    htmlContent,
	}

	// Envoyer l'email
	return r.sendEmail(emailReq)
}

// generateCandidateValidationHTML génère le contenu HTML de l'email
func (r *ResendService) generateCandidateValidationHTML(data models.EmailTemplateData) (string, error) {
	// Template HTML pour l'email de validation candidat
	htmlTemplate := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Dossier de compétences validé</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    
    <div style="text-align: center; margin-bottom: 30px;">
        <h1 style="color: #2563eb; margin-bottom: 10px;">🎉 Dossier de compétences validé !</h1>
        <p style="color: #666; font-size: 16px;">Un candidat a terminé et validé son dossier</p>
    </div>

    <div style="background: #f8fafc; padding: 20px; border-radius: 8px; margin-bottom: 20px;">
        <h2 style="color: #1e40af; margin-top: 0;">📋 Informations du candidat</h2>
        <p><strong>Nom :</strong> %s</p>
        <p><strong>Email :</strong> %s</p>
        <p><strong>Organisation :</strong> %s</p>
        <p><strong>Date de validation :</strong> %s</p>
    </div>

    <div style="background: #ecfdf5; padding: 20px; border-radius: 8px; margin-bottom: 20px; border-left: 4px solid #10b981;">
        <h3 style="color: #059669; margin-top: 0;">✅ Statut</h3>
        <p style="margin: 0;">Le dossier de compétences est maintenant <strong>complet à 100%%</strong> et prêt à être consulté.</p>
    </div>

    <div style="text-align: center; margin: 30px 0;">
        <a href="%s" 
           style="background: #2563eb; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; font-weight: bold; display: inline-block;">
            👁️ Consulter le dossier
        </a>
    </div>

    <div style="background: #fef3c7; padding: 15px; border-radius: 6px; margin-top: 20px;">
        <p style="margin: 0; font-size: 14px; color: #92400e;">
            <strong>💡 Note :</strong> Ce lien vous mènera directement au dossier du candidat. Assurez-vous d'être connecté à votre compte pour y accéder.
        </p>
    </div>

    <hr style="border: none; border-top: 1px solid #e5e7eb; margin: 30px 0;">
    
    <div style="text-align: center; color: #6b7280; font-size: 12px;">
        <p>Cet email a été envoyé automatiquement par SkillForge</p>
        <p>Si vous n'avez pas invité ce candidat, vous pouvez ignorer cet email.</p>
    </div>

</body>
</html>`

	// Remplacer les variables dans le template
	htmlContent := fmt.Sprintf(htmlTemplate,
		data.CandidateName,
		data.CandidateEmail,
		data.OrganizationName,
		data.ValidationDate,
		data.DossierURL,
	)

	return htmlContent, nil
}

// generateCandidateInviteHTML génère le contenu HTML de l'email d'invitation candidat
func (r *ResendService) generateCandidateInviteHTML(data models.CandidateInviteEmailData) (string, error) {
	// Template HTML pour l'email d'invitation candidat
	htmlTemplate := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Votre dossier de compétences vous attend</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    
    <div style="text-align: center; margin-bottom: 30px;">
        <h1 style="color: #2563eb; margin-bottom: 10px;">📝 Votre dossier de compétences vous attend</h1>
        <p style="color: #666; font-size: 16px;">Complétez et vérifiez votre profil professionnel</p>
    </div>

    <div style="background: #f8fafc; padding: 20px; border-radius: 8px; margin-bottom: 20px;">
        <h2 style="color: #1e40af; margin-top: 0;">🎯 Prochaines étapes</h2>
        <p>Vous avez été invité à compléter votre dossier de compétences sur SkillForge. Ce dossier vous permettra de :</p>
        <ul style="margin: 10px 0; padding-left: 20px;">
            <li>Mettre en valeur vos compétences et expériences</li>
            <li>Créer un profil professionnel attractif</li>
            <li>Faciliter votre recherche d'opportunités</li>
            <li>Être visible par les recruteurs</li>
        </ul>
    </div>

    <div style="background: #ecfdf5; padding: 20px; border-radius: 8px; margin-bottom: 20px; border-left: 4px solid #10b981;">
        <h3 style="color: #059669; margin-top: 0;">✅ Action requise</h3>
        <p style="margin: 0;">Cliquez sur le bouton ci-dessous pour accéder à votre dossier et commencer à le compléter.</p>
    </div>

    <div style="text-align: center; margin: 30px 0;">
        <a href="%s" 
           style="background: #2563eb; color: white; padding: 15px 30px; text-decoration: none; border-radius: 8px; font-weight: bold; display: inline-block; font-size: 16px;">
            🚀 Accéder à mon dossier
        </a>
    </div>

    <div style="background: #fef3c7; padding: 15px; border-radius: 6px; margin-top: 20px;">
        <p style="margin: 0; font-size: 14px; color: #92400e;">
            <strong>💡 Conseil :</strong> Prenez le temps de bien remplir toutes les sections pour maximiser votre visibilité. Vous pourrez modifier votre dossier à tout moment.
        </p>
    </div>

    <div style="background: #f3f4f6; padding: 15px; border-radius: 6px; margin-top: 20px;">
        <p style="margin: 0; font-size: 14px; color: #6b7280;">
            <strong>📧 Contact :</strong> Si vous avez des questions, n'hésitez pas à répondre à cet email ou contacter %s.
        </p>
    </div>

    <hr style="border: none; border-top: 1px solid #e5e7eb; margin: 30px 0;">
    
    <div style="text-align: center; color: #6b7280; font-size: 12px;">
        <p>Cet email a été envoyé par SkillForge</p>
        <p>Si vous n'avez pas demandé à recevoir cet email, vous pouvez l'ignorer en toute sécurité.</p>
    </div>

</body>
</html>`

	// Remplacer les variables dans le template
	htmlContent := fmt.Sprintf(htmlTemplate,
		data.CandidateLink,
		data.InviterEmail,
	)

	return htmlContent, nil
}

// sendEmail envoie un email via l'API Resend
func (r *ResendService) sendEmail(emailReq models.ResendEmailRequest) (*models.ResendEmailResponse, error) {
	// Sérialiser la requête
	jsonData, err := json.Marshal(emailReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal email request: %w", err)
	}

	// Créer la requête HTTP
	req, err := http.NewRequest("POST", r.BaseURL+"/emails", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Ajouter les headers
	req.Header.Set("Authorization", "Bearer "+r.APIKey)
	req.Header.Set("Content-Type", "application/json")

	// Envoyer la requête
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Lire la réponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Vérifier le statut de la réponse
	if resp.StatusCode >= 400 {
		var errorResp models.ResendErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return nil, fmt.Errorf("resend API error %d: %s", resp.StatusCode, string(body))
		}
		// Debug: afficher la réponse complète
		fmt.Printf("DEBUG: Resend API Error Response: %+v\n", errorResp)
		fmt.Printf("DEBUG: Raw response body: %s\n", string(body))
		return nil, fmt.Errorf("resend API error: %s", errorResp.Message)
	}

	// Parser la réponse de succès
	var emailResp models.ResendEmailResponse
	if err := json.Unmarshal(body, &emailResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &emailResp, nil
}

// ValidateEmailAddress valide une adresse email (validation basique)
func (r *ResendService) ValidateEmailAddress(email string) bool {
	// Validation basique - on pourrait utiliser une regex plus complexe
	return len(email) > 0 && 
		   len(email) < 254 && 
		   contains(email, "@") && 
		   contains(email, ".")
}

// contains vérifie si une chaîne contient une sous-chaîne
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    (len(s) > len(substr) && 
		     (s[:len(substr)] == substr || 
		      s[len(s)-len(substr):] == substr || 
		      containsSubstring(s, substr))))
}

// containsSubstring vérifie si une chaîne contient une sous-chaîne (implémentation simple)
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
