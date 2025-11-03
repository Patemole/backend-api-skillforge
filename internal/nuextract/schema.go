package nuextract

// CVExtractionSchema représente la structure d'un CV extrait
type CVExtractionSchema struct {
	Prenom                  string         `json:"prenom"`
	Nom                     string         `json:"nom"`
	Email                   string         `json:"email"`
	Phone                   string         `json:"phone"`
	Age                     string         `json:"age"`
	Summary                 string         `json:"summary"`
	Poste                   string         `json:"poste"`
	Diplome                 string         `json:"diplome"`
	Experience              string         `json:"expérience"`
	Mobilite                string         `json:"mobilité"`
	Disponibilite           string         `json:"disponibilité"`
	PermisB                 bool           `json:"permis_B"`
	Hobbies                 []string       `json:"hobbies"`
	Languages               []LanguageInfo `json:"languages"`
	Certifications          []string       `json:"certifications"`
	TechnicalSkills         []string       `json:"technical_skills"`
	CompetenceFonctionnelle []string       `json:"competence_fonctionnelle"`
	SecteursActivites       []string       `json:"secteurs_activites"`
	DomainesExpertise       []string       `json:"domaines_expertise"`
	Formations              []Formation    `json:"formations"`
	Experiences             []Experience   `json:"expériences"`
	Logiciels               []LogicielInfo `json:"logiciels"`
}

// LanguageInfo représente une langue parlée
type LanguageInfo struct {
	Language string `json:"language"`
	Level    string `json:"level"`
}

// Formation représente une formation
type Formation struct {
	DateDebut   string `json:"date_debut"`
	DateFin     string `json:"date_fin"`
	Diplome     string `json:"diplome"`
	EcoleCursus string `json:"ecole_cursus"`
}

// Experience représente une expérience professionnelle
type Experience struct {
	DateDebut        string   `json:"date_debut"`
	DateFin          string   `json:"date_fin"`
	Entreprise       string   `json:"entreprise"`
	DetailEntreprise string   `json:"detail_entreprise"`
	Duree            string   `json:"durée"`
	Poste            string   `json:"poste"`
	Secteur          string   `json:"secteur"`
	Contexte         string   `json:"contexte"`
	Projet           string   `json:"projet"`
	ProjetsName      []string `json:"projets_name"`
	Result           string   `json:"result"`
	Logiciels        []string `json:"logiciels"`
	Realisations     []string `json:"réalisations"`
	AISuggest        []string `json:"AI_suggest"`
}

// LogicielInfo représente un logiciel utilisé
type LogicielInfo struct {
	Logiciel         string `json:"logiciel"`
	Level            string `json:"level"`
	TempsUtilisation string `json:"temps_utilisation"`
}

// GetCVExtractionSchema retourne le schéma JSON pour le mode structured outputs
func GetCVExtractionSchema() map[string]interface{} {
	return map[string]interface{}{
		"prenom": map[string]interface{}{
			"type":        "string",
			"description": "Prénom du candidat",
		},
		"nom": map[string]interface{}{
			"type":        "string",
			"description": "Nom du candidat",
		},
		"email": map[string]interface{}{
			"type":        "string",
			"description": "Email du candidat",
		},
		"phone": map[string]interface{}{
			"type":        "string",
			"description": "Téléphone du candidat",
		},
		"age": map[string]interface{}{
			"type":        "string",
			"description": "Âge du candidat",
		},
		"summary": map[string]interface{}{
			"type":        "string",
			"description": "Résumé professionnel",
		},
		"poste": map[string]interface{}{
			"type":        "string",
			"description": "Poste recherché",
		},
		"diplome": map[string]interface{}{
			"type":        "string",
			"description": "Diplôme principal",
		},
		"expérience": map[string]interface{}{
			"type":        "string",
			"description": "Années d'expérience",
		},
		"mobilité": map[string]interface{}{
			"type":        "string",
			"description": "Mobilité géographique",
		},
		"disponibilité": map[string]interface{}{
			"type":        "string",
			"description": "Disponibilité",
		},
		"permis_B": map[string]interface{}{
			"type":        "boolean",
			"description": "Permis B",
		},
		"hobbies": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "string",
			},
			"description": "Hobbies du candidat",
		},
		"languages": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"language": map[string]interface{}{
						"type": "string",
					},
					"level": map[string]interface{}{
						"type": "string",
					},
				},
				"required":             []string{"language", "level"},
				"additionalProperties": false,
			},
			"description": "Langues parlées",
		},
		"certifications": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "string",
			},
			"description": "Certifications",
		},
		"technical_skills": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "string",
			},
			"description": "Compétences techniques",
		},
		"competence_fonctionnelle": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "string",
			},
			"description": "Compétences fonctionnelles",
		},
		"secteurs_activites": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "string",
			},
			"description": "Secteurs d'activité",
		},
		"domaines_expertise": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "string",
			},
			"description": "Domaines d'expertise",
		},
		"formations": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"date_debut": map[string]interface{}{
						"type": "string",
					},
					"date_fin": map[string]interface{}{
						"type": "string",
					},
					"diplome": map[string]interface{}{
						"type": "string",
					},
					"ecole_cursus": map[string]interface{}{
						"type": "string",
					},
				},
				"required":             []string{"date_debut", "date_fin", "diplome", "ecole_cursus"},
				"additionalProperties": false,
			},
			"description": "Formations",
		},
		"expériences": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"date_debut": map[string]interface{}{
						"type": "string",
					},
					"date_fin": map[string]interface{}{
						"type": "string",
					},
					"entreprise": map[string]interface{}{
						"type": "string",
					},
					"detail_entreprise": map[string]interface{}{
						"type": "string",
					},
					"durée": map[string]interface{}{
						"type": "string",
					},
					"poste": map[string]interface{}{
						"type": "string",
					},
					"secteur": map[string]interface{}{
						"type":        "string",
						"description": "Secteur d'activité de l'expérience (1-2 mots, ex: Agroalimentaire, Automobile, Banque, Bâtiments, Biomédical, Chimie, Conseil, Défense, Énergie, Environnement, Ferroviaire, Grande distribution, Infrastructure, Logistique, Métallurgie / Sidérurgie, Naval, Nucléaire, Oil & Gas, Pétrochimie, Pharmaceutique, Santé, Secteur public, Télécommunications, IRVE, Photovoltaïque, Traitement des eaux, Revalorisation énergétique, Hydroélectricité, ENR (Énergies renouvelables), Énergie éolienne, Biogaz, Education, Ressources Humaines)",
					},
					"contexte": map[string]interface{}{
						"type": "string",
					},
					"projet": map[string]interface{}{
						"type": "string",
					},
					"projets_name": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
						},
						"description": "Noms des projets mentionnés pour cette expérience",
					},
					"result": map[string]interface{}{
						"type":        "string",
						"description": "Phrase courte expliquant le résultat de l'expérience",
					},
					"logiciels": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"réalisations": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"AI_suggest": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
				},
				"required":             []string{"date_debut", "date_fin", "entreprise", "poste"},
				"additionalProperties": false,
			},
			"description": "Expériences professionnelles",
		},
		"logiciels": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"logiciel": map[string]interface{}{
						"type": "string",
					},
					"level": map[string]interface{}{
						"type": "string",
					},
					"temps_utilisation": map[string]interface{}{
						"type": "string",
					},
				},
				"required":             []string{"logiciel"},
				"additionalProperties": false,
			},
			"description": "Logiciels et outils",
		},
	}
}
