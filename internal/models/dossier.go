package models

// Formation defines the structure for an educational record.
type Formation struct {
	DateDebut   string `json:"date_debut"`
	DateFin     string `json:"date_fin"`
	Diplome     string `json:"diplome"`
	EcoleCursus string `json:"ecole_cursus"`
}

// Experience defines the structure for a professional experience record.
type Experience struct {
	DateDebut        string   `json:"date_debut"`
	DateFin          string   `json:"date_fin"`
	Entreprise       string   `json:"entreprise"`
	DetailEntreprise string   `json:"detail_entreprise"`
	Duree            string   `json:"durée"`
	Poste            string   `json:"poste"`
	Contexte         string   `json:"contexte"`
	Projet           string   `json:"projet"`
	ProjetSummary    string   `json:"projet_summary"`
	ProjetsName      []string `json:"projets_name"`
	Result           string   `json:"result"`
	Logiciels        []string `json:"logiciels"`
	Realisations     []string `json:"réalisations"`
	AISuggest        []string `json:"AI_suggest"`
}

// Logiciel defines the structure for a software skill record.
type Logiciel struct {
	Logiciel         string `json:"logiciel"`
	Level            string `json:"level"`
	TempsUtilisation string `json:"temps_utilisation"`
}

// CompetenceDossier defines the final, structured competence portfolio.
type CompetenceDossier struct {
	Prenom            string       `json:"prenom"`
	Nom               string       `json:"nom"`
	Email             string       `json:"email"`
	Phone             string       `json:"phone"`
	Summary           string       `json:"summary"`
	Age               string       `json:"age"`
	Poste             string       `json:"poste"`
	Diplome           string       `json:"diplome"`
	Experience        string       `json:"expérience"`
	Mobilite          string       `json:"mobilité"`
	Disponibilite     string       `json:"disponibilité"`
	PermisB           string       `json:"permis_B"`
	Hobbies           []string     `json:"hobbies"`
	Languages         []string     `json:"languages"`
	SecteursActivites []string     `json:"secteurs_activites"`
	DomainesExpertise []string     `json:"domaines_expertise"`
	Formations        []Formation  `json:"formations"`
	Experiences       []Experience `json:"expériences"`
	Logiciels              []Logiciel   `json:"logiciels"`
	Certifications         []string     `json:"certifications"`
	TechnicalSkills        []string     `json:"technical_skills"`
	CompetenceFonctionnelle []string     `json:"competence_fonctionnelle"`
}

// ChangelogEntry représente une modification dans le changelog
type ChangelogEntry struct {
	Field    string `json:"field"`
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`
	Reason   string `json:"reason"`
}
