package boond

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"backend-api-skillforge/internal/nuextract"
)

// splitFullName intelligently splits a full name into first name and last name.
// Handles cases like "Manuel de OLIVEIRA" where "de OLIVEIRA" should be the surname.
func splitFullName(fullName string) (firstName, lastName string) {
	fullName = strings.TrimSpace(fullName)
	if fullName == "" {
		return "", ""
	}

	// Common surname prefixes that should be included with the surname
	surnamePrefixes := []string{"de", "du", "des", "le", "la", "les", "van", "von", "der", "del", "da", "dos", "das"}

	parts := strings.Fields(fullName)
	if len(parts) == 0 {
		return fullName, ""
	}
	if len(parts) == 1 {
		return fullName, ""
	}

	// Check if the last word is all uppercase (common pattern for surnames in CVs)
	lastWord := parts[len(parts)-1]
	isLastWordUppercase := lastWord == strings.ToUpper(lastWord) && len(lastWord) > 1

	// If last word is uppercase, it's likely the surname
	// Include any preceding prefix (like "de", "van", etc.)
	if isLastWordUppercase {
		// Check if the word before the last is a surname prefix
		if len(parts) >= 2 {
			secondLast := strings.ToLower(parts[len(parts)-2])
			for _, prefix := range surnamePrefixes {
				if secondLast == prefix {
					// Include both prefix and last word as surname
					lastName = strings.Join(parts[len(parts)-2:], " ")
					firstName = strings.Join(parts[:len(parts)-2], " ")
					return firstName, lastName
				}
			}
		}
		// Just the last word is the surname
		lastName = lastWord
		firstName = strings.Join(parts[:len(parts)-1], " ")
		return firstName, lastName
	}

	// Check if there's a surname prefix before the last word
	if len(parts) >= 2 {
		secondLast := strings.ToLower(parts[len(parts)-2])
		for _, prefix := range surnamePrefixes {
			if secondLast == prefix {
				// Include both prefix and last word as surname
				lastName = strings.Join(parts[len(parts)-2:], " ")
				firstName = strings.Join(parts[:len(parts)-2], " ")
				return firstName, lastName
			}
		}
	}

	// Default: last word is surname, everything else is first name
	lastName = parts[len(parts)-1]
	firstName = strings.Join(parts[:len(parts)-1], " ")
	return firstName, lastName
}

// BuildCandidateAttributesFromCV construit un map d'attributs Boond à partir
// du schéma d'extraction interne. Ne renseigne que les champs disponibles.
func BuildCandidateAttributesFromCV(cv nuextract.CVExtractionSchema) map[string]any {
	attributes := make(map[string]any)

	// Handle name: use cv.Nom if provided, otherwise split cv.Prenom
	if s := strings.TrimSpace(cv.Prenom); s != "" {
		if nom := strings.TrimSpace(cv.Nom); nom != "" {
			// If Nom is provided, use Prenom as firstName and Nom as lastName
			attributes["firstName"] = s
			attributes["lastName"] = nom
		} else {
			// If Nom is empty, split Prenom (which contains full name) into firstName and lastName
			firstName, lastName := splitFullName(s)
			if firstName != "" {
				attributes["firstName"] = firstName
			}
			if lastName != "" {
				attributes["lastName"] = lastName
			}
			// If splitting failed, fallback to using full name as firstName
			if firstName == "" && lastName == "" {
				attributes["firstName"] = s
			}
		}
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
	// Stocker mobilite et disponibilite pour enrichissement ultérieur avec les IDs Boond
	if s := strings.TrimSpace(cv.Mobilite); s != "" {
		// Stocker temporairement le texte brut - sera mappé vers les IDs Boond par EnrichAttributesWithBoondIDs
		attributes["_mobilite_raw"] = s
	}
	if s := strings.TrimSpace(cv.Disponibilite); s != "" {
		// Stocker temporairement le texte brut - sera mappé vers les IDs Boond par EnrichAttributesWithBoondIDs
		attributes["_disponibilite_raw"] = s
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

// EnrichAttributesWithBoondIDs enrichit les attributs avec les IDs Boond pour mobilite et disponibilite.
// Cette fonction doit être appelée après BuildCandidateAttributesFromCV et avant CreateCandidate.
// Elle mappe les valeurs textuelles vers les IDs Boond appropriés.
func EnrichAttributesWithBoondIDs(ctx context.Context, client *Client, attributes map[string]any) error {
	// Traiter la mobilité
	if mobiliteRaw, ok := attributes["_mobilite_raw"].(string); ok && strings.TrimSpace(mobiliteRaw) != "" {
		delete(attributes, "_mobilite_raw") // Supprimer la clé temporaire

		// Récupérer les zones de mobilité depuis Boond
		mobilityMap, err := client.GetMobilityAreas(ctx)
		if err != nil {
			fmt.Printf("⚠️  [Boond] Erreur récupération zones mobilité: %v\n", err)
			// En cas d'erreur, on ne met pas de mobilité plutôt que d'envoyer une valeur invalide
		} else {
			// Parser les zones demandées (séparées par virgule)
			requestedAreas := strings.Split(strings.TrimSpace(mobiliteRaw), ",")
			var mobilityIDs []string

			for _, area := range requestedAreas {
				area = strings.TrimSpace(area)
				// Essayer une correspondance exacte d'abord
				if id, exists := mobilityMap[area]; exists {
					mobilityIDs = append(mobilityIDs, id)
					fmt.Printf("✅ [Boond] Zone trouvée: %s -> ID %s\n", area, id)
				} else {
					// Essayer une correspondance insensible à la casse
					found := false
					for name, id := range mobilityMap {
						if strings.EqualFold(name, area) {
							mobilityIDs = append(mobilityIDs, id)
							fmt.Printf("✅ [Boond] Zone trouvée (insensible casse): %s -> ID %s\n", area, id)
							found = true
							break
						}
					}
					if !found {
						fmt.Printf("⚠️  [Boond] Zone non trouvée: %s\n", area)
					}
				}
			}

			if len(mobilityIDs) > 0 {
				attributes["mobilityAreas"] = mobilityIDs
				fmt.Printf("✅ [Boond] Mobilité enrichie: %v\n", mobilityIDs)
			} else {
				fmt.Printf("⚠️  [Boond] Aucune zone de mobilité valide trouvée pour: %s\n", mobiliteRaw)
			}
		}
	}

	// Traiter la disponibilité
	if disponibiliteRaw, ok := attributes["_disponibilite_raw"].(string); ok && strings.TrimSpace(disponibiliteRaw) != "" {
		delete(attributes, "_disponibilite_raw") // Supprimer la clé temporaire

		dispo := strings.TrimSpace(disponibiliteRaw)

		// Vérifier si c'est une date au format YYYY-MM-DD
		if matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, dispo); matched {
			attributes["availability"] = dispo
			fmt.Printf("✅ [Boond] Disponibilité (date): %s\n", dispo)
		} else {
			// Essayer de trouver dans les types prédéfinis
			availabilityMap, err := client.GetAvailabilityTypes(ctx)
			if err != nil {
				fmt.Printf("⚠️  [Boond] Erreur récupération types disponibilité: %v\n", err)
				// Fallback: traiter comme ID numérique si c'est un nombre
				if matched, _ := regexp.MatchString(`^\d+$`, dispo); matched {
					attributes["availability"] = dispo
					fmt.Printf("✅ [Boond] Disponibilité (ID fallback): %s\n", dispo)
				}
			} else {
				// Chercher le type correspondant (correspondance exacte d'abord)
				if id, exists := availabilityMap[dispo]; exists {
					attributes["availability"] = id
					fmt.Printf("✅ [Boond] Type trouvé: %s -> ID %s\n", dispo, id)
				} else {
					// Essayer une correspondance insensible à la casse
					found := false
					for name, id := range availabilityMap {
						if strings.EqualFold(name, dispo) {
							attributes["availability"] = id
							fmt.Printf("✅ [Boond] Type trouvé (insensible casse): %s -> ID %s\n", dispo, id)
							found = true
							break
						}
					}
					if !found {
						fmt.Printf("⚠️  [Boond] Type de disponibilité non trouvé: %s\n", dispo)
						// Fallback: traiter comme ID numérique si c'est un nombre
						if matched, _ := regexp.MatchString(`^\d+$`, dispo); matched {
							attributes["availability"] = dispo
							fmt.Printf("✅ [Boond] Disponibilité (ID fallback): %s\n", dispo)
						}
					}
				}
			}
		}
	}

	return nil
}
