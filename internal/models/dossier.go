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
	Entreprise   string   `json:"entreprise"`
	Duree        string   `json:"durée"`
	Poste        string   `json:"poste"`
	Contexte     string   `json:"contexte"`
	Projet       string   `json:"projet"`
	Logiciels    []string `json:"logiciels"`
	Realisations []string `json:"réalisations"`
	AISuggest    []string `json:"AI_suggest"`
}

// Logiciel defines the structure for a software skill record.
type Logiciel struct {
	Logiciel         string `json:"logiciel"`
	Level            string `json:"level"`
	TempsUtilisation string `json:"temps_utilisation"`
}

// CompetenceDossier defines the final, structured competence portfolio.
type CompetenceDossier struct {
	Prenom        string       `json:"prenom"`
	Age           string       `json:"age"`
	Poste         string       `json:"poste"`
	Diplome       string       `json:"diplome"`
	Experience    string       `json:"expérience"`
	Mobilite      string       `json:"mobilité"`
	Disponibilite string       `json:"disponibilité"`
	PermisB       string       `json:"permis_B"`
	Hobbies       []string     `json:"hobbies"`
	Formations    []Formation  `json:"formations"`
	Experiences   []Experience `json:"expériences"`
	Logiciels     []Logiciel   `json:"logiciels"`
}
