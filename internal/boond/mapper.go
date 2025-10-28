package boond

import (
	"strings"

	"backend-api-skillforge/internal/nuextract"
)

// BuildCandidateAttributesFromCV construit un map d'attributs Boond à partir
// du schéma d'extraction interne. Ne renseigne que les champs disponibles.
func BuildCandidateAttributesFromCV(cv nuextract.CVExtractionSchema) map[string]any {
	attributes := make(map[string]any)

	if s := strings.TrimSpace(cv.Prenom); s != "" {
		attributes["firstName"] = s
	}
	if s := strings.TrimSpace(cv.Nom); s != "" {
		attributes["lastName"] = s
	}
	if s := strings.TrimSpace(cv.Email); s != "" {
		attributes["email1"] = s
	}
	if s := strings.TrimSpace(cv.Phone); s != "" {
		attributes["phone1"] = s
	}
	if s := strings.TrimSpace(cv.Poste); s != "" {
		attributes["title"] = s
	}
	// Ne pas envoyer "availability" pour éviter les erreurs 422 côté Boond
	if s := strings.TrimSpace(cv.Mobilite); s != "" {
		// Mobilité - pour l'instant on stocke le texte, plus tard on pourra mapper vers les IDs Boond
		attributes["mobilityAreas"] = []string{s}
	}
	if s := strings.TrimSpace(cv.Diplome); s != "" {
		attributes["education"] = s
	}
	if s := strings.TrimSpace(cv.Experience); s != "" {
		attributes["experience"] = s
	}

	// Concaténer les domaines, secteurs, langues, certifications, compétences techniques et logiciels en un champ "skills" textuel
	var skillsParts []string
	if len(cv.DomainesExpertise) > 0 {
		skillsParts = append(skillsParts, "Domaines d'expertise: "+strings.Join(cv.DomainesExpertise, ", "))
	}
	if len(cv.SecteursActivites) > 0 {
		skillsParts = append(skillsParts, "Secteurs: "+strings.Join(cv.SecteursActivites, ", "))
	}
	if len(cv.Languages) > 0 {
		var languages []string
		for _, lang := range cv.Languages {
			if name := strings.TrimSpace(lang.Language); name != "" {
				level := strings.TrimSpace(lang.Level)
				if level != "" {
					languages = append(languages, name+" ("+level+")")
				} else {
					languages = append(languages, name)
				}
			}
		}
		if len(languages) > 0 {
			skillsParts = append(skillsParts, "Langues: "+strings.Join(languages, ", "))
		}
	}
	if len(cv.Certifications) > 0 {
		skillsParts = append(skillsParts, "Certifications: "+strings.Join(cv.Certifications, ", "))
	}
	if len(cv.TechnicalSkills) > 0 {
		skillsParts = append(skillsParts, "Compétences techniques: "+strings.Join(cv.TechnicalSkills, ", "))
	}
	if len(cv.Logiciels) > 0 {
		var tools []string
		for _, l := range cv.Logiciels {
			if name := strings.TrimSpace(l.Logiciel); name != "" {
				tools = append(tools, name)
			}
		}
		if len(tools) > 0 {
			skillsParts = append(skillsParts, "Logiciels: "+strings.Join(tools, ", "))
		}
	}
	if len(skillsParts) > 0 {
		attributes["skills"] = strings.Join(skillsParts, " | ")
	}

	// Le résumé peut alimenter un commentaire business par défaut
	if s := strings.TrimSpace(cv.Summary); s != "" {
		attributes["PARAM_COMMENTAIRE"] = s
	}

	// Source : SkillForge
	attributes["source"] = map[string]any{
		"typeOf": 1, // À ajuster selon les valeurs Boond
		"detail": "SkillForge",
	}

	return attributes
}
