package models

// GenerateEmailRequest définit la structure de la requête pour générer un email de présentation
type GenerateEmailRequest struct {
	CandidateData CandidateData `json:"candidateData" binding:"required"`
	Need          *string       `json:"need,omitempty"` // Optionnel
}

// CandidateData contient toutes les données du candidat depuis le store editorData
type CandidateData struct {
	Title           string       `json:"title"`
	ExperienceYears int          `json:"experience_years"`
	Prenom          string       `json:"prenom"`
	Nom             string       `json:"nom"`
	Age             int          `json:"age"`
	Languages       string       `json:"languages"`
	Mobility        string       `json:"mobility"`
	Availability    string       `json:"availability"`
	PermisB         string       `json:"permis_b"`
	SecteursActivites []string   `json:"secteurs_activites"`
	DomainesExpertise []string   `json:"domaines_expertise"`
	Formations      []Formation  `json:"formations"`
	Experiences     []Experience `json:"experiences"`
	Logiciels       []Logiciel   `json:"logiciels"`
	Hobbies         []string     `json:"hobbies"`
}

// FormationEmail définit la structure pour une formation dans l'email
type FormationEmail struct {
	Degree      string `json:"degree"`
	Institution string `json:"institution"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
}

// ExperienceEmail définit la structure pour une expérience dans l'email
type ExperienceEmail struct {
	Company           string   `json:"company"`
	DetailEntreprise  string   `json:"detail_entreprise"`
	Title             string   `json:"title"`
	Project           string   `json:"project"`
	Poste             string   `json:"poste"`
	Entreprise        string   `json:"entreprise"`
	Projet            string   `json:"projet"`
	Realisations      []string `json:"realisations"`
	Logiciels         string   `json:"logiciels"`
}

// LogicielEmail définit la structure pour un logiciel dans l'email
type LogicielEmail struct {
	Logiciel         string `json:"logiciel"`
	Level            string `json:"level"`
	TempsUtilisation int    `json:"temps_utilisation"`
}

// GenerateEmailResponse définit la structure de la réponse pour la génération d'email
type GenerateEmailResponse struct {
	EmailContent string `json:"emailContent"`
	Success      bool   `json:"success"`
	Error        string `json:"error,omitempty"`
}

// TemplateGenerateRequest définit la structure de la requête pour générer un template d'email
type TemplateGenerateRequest struct {
	OrganizationID     string           `json:"organization_id" binding:"required"`
	TemplateName       string           `json:"template_name" binding:"required"`
	EmailExamples      []EmailExample   `json:"email_examples" binding:"required,min=3"`
	AvailableVariables []string         `json:"available_variables" binding:"required"`
}

// EmailExample définit la structure d'un exemple d'email
type EmailExample struct {
	Subject string `json:"subject" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// GeneratedTemplate définit la structure du template généré
type GeneratedTemplate struct {
	Subject string `json:"subject"`
	Content string `json:"content"`
}

// TemplateGenerateResponse définit la structure de la réponse pour la génération de template
type TemplateGenerateResponse struct {
	Success           bool              `json:"success"`
	TemplateID        string            `json:"template_id"`
	GeneratedTemplate GeneratedTemplate `json:"generated_template"`
	DetectedVariables []string          `json:"detected_variables"`
	ConfidenceScore   float64           `json:"confidence_score"`
	Error             string            `json:"error,omitempty"`
}

// PresentationEmailRequest définit la structure de la requête pour générer un email de présentation amélioré
type PresentationEmailRequest struct {
	CandidateData PresentationCandidateData `json:"candidateData" binding:"required"`
	Need          *string                   `json:"need,omitempty"`     // Optionnel - description du besoin
	TemplateID    *string                   `json:"templateId,omitempty"` // Optionnel
	Template      *PresentationTemplate     `json:"template,omitempty"`   // Optionnel
}

// PresentationCandidateData contient les données du candidat pour la présentation
type PresentationCandidateData struct {
	Prenom                string   `json:"prenom" binding:"required"`
	TitrePoste            string   `json:"titre_poste" binding:"required"`
	NombreExperience      int      `json:"nombre_experience" binding:"required"`
	Disponibilite         string   `json:"disponibilite" binding:"required"`
	Mobilite              string   `json:"mobilite" binding:"required"`
	Diplome               string   `json:"diplome" binding:"required"`
	Langues               []string `json:"langues"`
	Certifications        string   `json:"certifications"`
	Hobbies               string   `json:"hobbies"`
	Experience            string   `json:"experience"`            // Toutes les expériences formatées
	ExperienceCount       int      `json:"experience_count"`
	Logiciel              string   `json:"logiciel"`              // Logiciel principal
	Logiciels             string   `json:"logiciels"`             // Liste des logiciels
	LogicielsCount        int      `json:"logiciels_count"`
	CompetencesTechniques string   `json:"competences_techniques"`
	CompetencesFonctionnelles string `json:"competences_fonctionnelles"`
	Projets               string   `json:"projets"`
	ProjetsCount          int      `json:"projets_count"`
}

// PresentationTemplate définit la structure du template pour la présentation
type PresentationTemplate struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Subject   string `json:"subject"`   // Avec variables remplacées
	Content   string `json:"content"`   // Avec variables remplacées
	IsDefault bool   `json:"isDefault"`
}

// PresentationEmailResponse définit la structure de la réponse pour l'email de présentation
type PresentationEmailResponse struct {
	EmailContent string `json:"emailContent"`
	Success      bool   `json:"success"`
	Error        string `json:"error,omitempty"`
}
