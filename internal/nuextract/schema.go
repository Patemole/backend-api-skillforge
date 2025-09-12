package nuextract

// CVExtractionSchema définit la structure JSON envoyée au front-end après extraction et enrichissement
type CVExtractionSchema struct {
	Prenom        string       `json:"prenom"`        // string
	Email         string       `json:"email"`         // string (laisser vide si non présent)
	Phone         string       `json:"phone"`         // string (numéro de téléphone)
	Summary       string       `json:"summary"`       // string (résumé professionnel)
	Age           string       `json:"age"`           // string (peut être "Non précisé(e)")
	Poste         string       `json:"poste"`         // string
	Diplome       string       `json:"diplome"`       // string
	Experience    string       `json:"expérience"`    // string
	Mobilite      string       `json:"mobilité"`      // string
	Disponibilite string       `json:"disponibilité"` // string
	PermisB       interface{}  `json:"permis_B"`      // bool ou string
	Hobbies       []string     `json:"hobbies"`       // []string
	Languages     []string     `json:"languages"`     // []string (langues parlées)
	Formations    []Formation  `json:"formations"`    // []Formation
	Experiences   []Experience `json:"expériences"`   // []Experience
	Logiciels     []Logiciel   `json:"logiciels"`     // []Logiciel
}

// Formation définit la structure d'une formation
type Formation struct {
	DateDebut   string `json:"date_debut"`   // string (format: "YYYY-MM")
	DateFin     string `json:"date_fin"`     // string (format: "YYYY-MM")
	Diplome     string `json:"diplome"`      // string
	EcoleCursus string `json:"ecole_cursus"` // string
}

// Experience définit la structure d'une expérience professionnelle
type Experience struct {
	DateDebut    string   `json:"date_debut"`   // string (format: "Février 2025")
	DateFin      string   `json:"date_fin"`     // string (format: "Décembre 2025")
	Entreprise   string   `json:"entreprise"`   // string
	Duree        string   `json:"durée"`        // string (calculée automatiquement)
	Poste        string   `json:"poste"`        // string
	Contexte     string   `json:"contexte"`     // string
	Projet       string   `json:"projet"`       // string
	Logiciels    []string `json:"logiciels"`    // []string
	Realisations []string `json:"réalisations"` // []string
	AISuggest    []string `json:"AI_suggest"`   // []string
}

// Logiciel définit la structure d'un logiciel
type Logiciel struct {
	Logiciel         string `json:"logiciel"`          // string
	Level            string `json:"level"`             // string ("Débutant", "Intermédiaire", "Avancé", "Expert")
	TempsUtilisation string `json:"temps_utilisation"` // string (en mois)
}

