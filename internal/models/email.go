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
	Age             int          `json:"age"`
	Languages       string       `json:"languages"`
	Mobility        string       `json:"mobility"`
	Availability    string       `json:"availability"`
	PermisB         string       `json:"permis_b"`
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
	Company      string   `json:"company"`
	Title        string   `json:"title"`
	Project      string   `json:"project"`
	Poste        string   `json:"poste"`
	Entreprise   string   `json:"entreprise"`
	Projet       string   `json:"projet"`
	Realisations []string `json:"realisations"`
	Logiciels    string   `json:"logiciels"`
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
