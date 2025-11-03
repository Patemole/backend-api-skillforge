package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"backend-api-skillforge/internal/models"
	"backend-api-skillforge/internal/openai"

	"github.com/gin-gonic/gin"
)

// DossierVersionningRequest représente la requête pour créer une nouvelle version
type DossierVersionningRequest struct {
	NewTitle          string                        `json:"new_title"`
	Need              *string                       `json:"need,omitempty"`
	CandidateID       string                        `json:"candidate_id" binding:"required"`
	CompetenceDossier CompetenceDossierIntermediate `json:"competence_dossier" binding:"required"`
	Language          string                        `json:"language,omitempty"` // "fr" ou "en", défaut: "fr"
}

// CompetenceDossierIntermediate est une structure intermédiaire pour accepter les langues au format objet ou string
type CompetenceDossierIntermediate struct {
	Prenom                  string              `json:"prenom"`
	Nom                     string              `json:"nom"`
	Email                   string              `json:"email"`
	Phone                   string              `json:"phone"`
	Summary                 string              `json:"summary"`
	Age                     string              `json:"age"`
	Poste                   string              `json:"poste"`
	Diplome                 string              `json:"diplome"`
	Experience              string              `json:"expérience"`
	Mobilite                string              `json:"mobilité"`
	Disponibilite           string              `json:"disponibilité"`
	PermisB                 string              `json:"permis_B"`
	Hobbies                 []string            `json:"hobbies"`
	Languages               []interface{}       `json:"languages"` // Interface{} pour accepter strings et objets
	SecteursActivites       []string            `json:"secteurs_activites"`
	DomainesExpertise       []string            `json:"domaines_expertise"`
	Formations              []models.Formation  `json:"formations"`
	Experiences             []models.Experience `json:"expériences"`
	Logiciels               []models.Logiciel   `json:"logiciels"`
	Certifications          []string            `json:"certifications"`
	TechnicalSkills         []string            `json:"technical_skills"`
	CompetenceFonctionnelle []string            `json:"competence_fonctionnelle"`
}

// ToCompetenceDossier convertit la structure intermédiaire vers CompetenceDossier
func (c *CompetenceDossierIntermediate) ToCompetenceDossier() models.CompetenceDossier {
	// Convertir les langues
	var languages []string
	for _, lang := range c.Languages {
		switch v := lang.(type) {
		case string:
			// Langue simple (ex: "Français")
			languages = append(languages, v)
		case map[string]interface{}:
			// Langue avec niveau (ex: {"language": "Chinois", "level": "Intermediaire"})
			if langName, ok := v["language"].(string); ok {
				if level, ok := v["level"].(string); ok && level != "" {
					languages = append(languages, langName+" ("+level+")")
				} else {
					languages = append(languages, langName)
				}
			}
		}
	}

	return models.CompetenceDossier{
		Prenom:                  c.Prenom,
		Nom:                     c.Nom,
		Email:                   c.Email,
		Phone:                   c.Phone,
		Summary:                 c.Summary,
		Age:                     c.Age,
		Poste:                   c.Poste,
		Diplome:                 c.Diplome,
		Experience:              c.Experience,
		Mobilite:                c.Mobilite,
		Disponibilite:           c.Disponibilite,
		PermisB:                 c.PermisB,
		Hobbies:                 c.Hobbies,
		Languages:               languages,
		SecteursActivites:       c.SecteursActivites,
		DomainesExpertise:       c.DomainesExpertise,
		Formations:              c.Formations,
		Experiences:             c.Experiences,
		Logiciels:               c.Logiciels,
		Certifications:          c.Certifications,
		TechnicalSkills:         c.TechnicalSkills,
		CompetenceFonctionnelle: c.CompetenceFonctionnelle,
	}
}

// DossierVersionningResponse représente la réponse
type DossierVersionningResponse struct {
	CandidateID       string                   `json:"candidate_id"`
	CompetenceDossier models.CompetenceDossier `json:"competence_dossier"`
	Changelog         []models.ChangelogEntry  `json:"changelog"`
	Language          string                   `json:"language"` // "fr" ou "en"
}

// CreateDossierVersion gère la création d'une nouvelle version de dossier de compétences.
// Payload attendu:
//
//	{
//	  "new_title": "Dossier v3 - ...",
//	  "need": "Mettre en avant ...", // optionnel
//	  "candidate_id": "uuid",
//	  "competence_dossier": { /* dossier source complet */ }
//	}
//
// Réponse:
//
//	{
//	  "candidate_id": "uuid",
//	  "competence_dossier": { /* dossier versionné */ }
//	}
func CreateDossierVersion(c *gin.Context) {
	log.Printf("🔄 DOSSIER VERSIONNING - Début de la requête")

	var req DossierVersionningRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ DOSSIER VERSIONNING - Erreur validation payload: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload: " + err.Error()})
		return
	}

	// Normaliser la langue (fr par défaut, seulement fr ou en supportés)
	language := strings.TrimSpace(req.Language)
	if language == "" {
		language = "fr"
	}
	if language != "en" && language != "fr" {
		language = "fr" // Fallback sur fr si langue non supportée
	}

	// Convertir la structure intermédiaire vers CompetenceDossier
	competenceDossier := req.CompetenceDossier.ToCompetenceDossier()

	log.Printf("✅ DOSSIER VERSIONNING - Payload validé:")
	log.Printf("   - Candidate ID: %s", req.CandidateID)
	log.Printf("   - New Title: %s", req.NewTitle)
	log.Printf("   - Need: %v", req.Need)
	log.Printf("   - Language: %s", language)
	log.Printf("   - Dossier source reçu: %d expériences, %d formations",
		len(competenceDossier.Experiences), len(competenceDossier.Formations))

	service := openai.NewDossierVersionningServiceAnthropic()
	log.Printf("🤖 DOSSIER VERSIONNING - Appel Anthropic en cours...")

	newDossier, err := service.GenerateVersionnedDossier(req.CandidateID, competenceDossier, req.Need, language)
	if err != nil {
		log.Printf("❌ DOSSIER VERSIONNING - Erreur Anthropic: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	// Sécurité: restaurer les tableaux et réalisations si manquants
	for i := range newDossier.Experiences {
		// Assurer non-nil
		if newDossier.Experiences[i].Logiciels == nil {
			newDossier.Experiences[i].Logiciels = []string{}
		}
		if newDossier.Experiences[i].Realisations == nil {
			newDossier.Experiences[i].Realisations = []string{}
		}
		if newDossier.Experiences[i].AISuggest == nil {
			newDossier.Experiences[i].AISuggest = []string{}
		}

		// Log des réalisations source pour debug
		if i < len(competenceDossier.Experiences) {
			log.Printf("🔍 DOSSIER VERSIONNING - Expérience[%d] source: %d réalisations", i, len(competenceDossier.Experiences[i].Realisations))
			if len(competenceDossier.Experiences[i].Realisations) > 0 {
				for j, real := range competenceDossier.Experiences[i].Realisations {
					log.Printf("   - Réalisation[%d]: %s", j, real)
				}
			}
		}

		// Si réalisations vides alors que le dossier source en avait, restaurer à l'identique
		if len(newDossier.Experiences[i].Realisations) == 0 && i < len(competenceDossier.Experiences) && len(competenceDossier.Experiences[i].Realisations) > 0 {
			log.Printf("⚠️ DOSSIER VERSIONNING - Réalisations manquantes pour expérience[%d], restauration des réalisations source (%d items)", i, len(competenceDossier.Experiences[i].Realisations))
			newDossier.Experiences[i].Realisations = append([]string{}, competenceDossier.Experiences[i].Realisations...)
		}
	}

	log.Printf("✅ DOSSIER VERSIONNING - Réponse Anthropic reçue:")
	log.Printf("   - Dossier versionné: %d expériences, %d formations",
		len(newDossier.Experiences), len(newDossier.Formations))
	log.Printf("   - Poste: %s", newDossier.Poste)
	log.Printf("   - Nombre de réalisations total: %d",
		len(newDossier.Experiences)+len(newDossier.Logiciels))

	// Générer le changelog des modifications
	changelog := generateChangelog(competenceDossier, *newDossier, req.Need)
	log.Printf("📝 DOSSIER VERSIONNING - Changelog généré: %d modifications détectées", len(changelog))

	// S'assurer que le changelog n'est jamais null
	if changelog == nil {
		changelog = []models.ChangelogEntry{}
	}

	// Log détaillé de ce qui est renvoyé au frontend
	log.Printf("📤 DOSSIER VERSIONNING - Données renvoyées au frontend:")
	log.Printf("   - Candidate ID: %s", req.CandidateID)
	log.Printf("   - Nom: %s %s", newDossier.Prenom, newDossier.Nom)
	log.Printf("   - Email: %s", newDossier.Email)
	log.Printf("   - Phone: %s", newDossier.Phone)
	log.Printf("   - Poste: %s", newDossier.Poste)
	log.Printf("   - Expériences: %d", len(newDossier.Experiences))
	log.Printf("   - Formations: %d", len(newDossier.Formations))
	log.Printf("   - Logiciels: %d", len(newDossier.Logiciels))
	log.Printf("   - Hobbies: %d", len(newDossier.Hobbies))
	log.Printf("   - Languages: %d", len(newDossier.Languages))
	log.Printf("   - Secteurs: %d", len(newDossier.SecteursActivites))
	log.Printf("   - Domaines: %d", len(newDossier.DomainesExpertise))
	log.Printf("   - Changelog: %d modifications", len(changelog))

	// Log spécifique pour les réalisations
	totalRealisations := 0
	for i, exp := range newDossier.Experiences {
		totalRealisations += len(exp.Realisations)
		log.Printf("   - Expérience[%d] réalisations: %d", i, len(exp.Realisations))
	}
	log.Printf("   - Total réalisations: %d", totalRealisations)

	resp := DossierVersionningResponse{
		CandidateID:       req.CandidateID,
		CompetenceDossier: *newDossier,
		Changelog:         changelog,
		Language:          language,
	}

	// Afficher le payload JSON complet envoyé au frontend
	respJSON, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		log.Printf("❌ DOSSIER VERSIONNING - Erreur sérialisation JSON: %v", err)
	} else {
		log.Printf("📄 DOSSIER VERSIONNING - Payload JSON complet envoyé au frontend:")
		log.Printf("```json")
		log.Printf("%s", string(respJSON))
		log.Printf("```")
	}

	c.JSON(http.StatusOK, resp)
}

// generateChangelog génère un changelog des modifications apportées
func generateChangelog(original, modified models.CompetenceDossier, need *string) []models.ChangelogEntry {
	var changelog []models.ChangelogEntry

	// Comparer les champs principaux
	if original.Poste != modified.Poste {
		changelog = append(changelog, models.ChangelogEntry{
			Field:    "poste",
			OldValue: original.Poste,
			NewValue: modified.Poste,
			Reason:   "Adaptation du poste selon le besoin client",
		})
	}

	if original.Summary != modified.Summary {
		changelog = append(changelog, models.ChangelogEntry{
			Field:    "summary",
			OldValue: original.Summary,
			NewValue: modified.Summary,
			Reason:   "Reformulation du résumé professionnel",
		})
	}

	// Comparer les expériences
	changelog = append(changelog, compareExperiences(original.Experiences, modified.Experiences)...)

	// Comparer les formations
	changelog = append(changelog, compareFormations(original.Formations, modified.Formations)...)

	// Comparer les logiciels
	changelog = append(changelog, compareLogiciels(original.Logiciels, modified.Logiciels)...)

	// Comparer les secteurs d'activités
	changelog = append(changelog, compareStringArrays("secteurs_activites", original.SecteursActivites, modified.SecteursActivites)...)

	// Comparer les domaines d'expertise
	changelog = append(changelog, compareStringArrays("domaines_expertise", original.DomainesExpertise, modified.DomainesExpertise)...)

	// Comparer les hobbies
	changelog = append(changelog, compareStringArrays("hobbies", original.Hobbies, modified.Hobbies)...)

	// Comparer les langues
	changelog = append(changelog, compareStringArrays("languages", original.Languages, modified.Languages)...)

	return changelog
}

// compareExperiences compare les expériences et détecte les modifications
func compareExperiences(original, modified []models.Experience) []models.ChangelogEntry {
	var changelog []models.ChangelogEntry

	// Comparer le nombre d'expériences
	if len(original) != len(modified) {
		changelog = append(changelog, models.ChangelogEntry{
			Field:    "expériences",
			OldValue: fmt.Sprintf("%d expériences", len(original)),
			NewValue: fmt.Sprintf("%d expériences", len(modified)),
			Reason:   "Réorganisation des expériences par pertinence",
		})
	}

	// Comparer chaque expérience
	maxLen := len(original)
	if len(modified) > maxLen {
		maxLen = len(modified)
	}

	for i := 0; i < maxLen; i++ {
		if i >= len(original) || i >= len(modified) {
			continue
		}

		orig := original[i]
		mod := modified[i]

		// Comparer le contexte
		if orig.Contexte != mod.Contexte {
			changelog = append(changelog, models.ChangelogEntry{
				Field:    fmt.Sprintf("expériences[%d].contexte", i),
				OldValue: orig.Contexte,
				NewValue: mod.Contexte,
				Reason:   "Reformulation du contexte de l'expérience",
			})
		}

		// Comparer le projet
		if orig.Projet != mod.Projet {
			changelog = append(changelog, models.ChangelogEntry{
				Field:    fmt.Sprintf("expériences[%d].projet", i),
				OldValue: orig.Projet,
				NewValue: mod.Projet,
				Reason:   "Adaptation de la description du projet",
			})
		}

		// Comparer les réalisations
		if len(orig.Realisations) != len(mod.Realisations) {
			changelog = append(changelog, models.ChangelogEntry{
				Field:    fmt.Sprintf("expériences[%d].réalisations", i),
				OldValue: fmt.Sprintf("%d réalisations", len(orig.Realisations)),
				NewValue: fmt.Sprintf("%d réalisations", len(mod.Realisations)),
				Reason:   "Réorganisation des réalisations par pertinence",
			})
		}

		// Vérifier si les réalisations ont été supprimées (cas critique)
		if len(orig.Realisations) > 0 && len(mod.Realisations) == 0 {
			changelog = append(changelog, models.ChangelogEntry{
				Field:    fmt.Sprintf("expériences[%d].réalisations", i),
				OldValue: fmt.Sprintf("%d réalisations présentes", len(orig.Realisations)),
				NewValue: "0 réalisations (SUPPRIMÉES)",
				Reason:   "⚠️ ERREUR : Les réalisations ont été supprimées - elles doivent être conservées",
			})
		}
	}

	return changelog
}

// compareFormations compare les formations et détecte les modifications
func compareFormations(original, modified []models.Formation) []models.ChangelogEntry {
	var changelog []models.ChangelogEntry

	if len(original) != len(modified) {
		changelog = append(changelog, models.ChangelogEntry{
			Field:    "formations",
			OldValue: fmt.Sprintf("%d formations", len(original)),
			NewValue: fmt.Sprintf("%d formations", len(modified)),
			Reason:   "Réorganisation des formations par pertinence",
		})
	}

	return changelog
}

// compareLogiciels compare les logiciels et détecte les modifications
func compareLogiciels(original, modified []models.Logiciel) []models.ChangelogEntry {
	var changelog []models.ChangelogEntry

	if len(original) != len(modified) {
		changelog = append(changelog, models.ChangelogEntry{
			Field:    "logiciels",
			OldValue: fmt.Sprintf("%d logiciels", len(original)),
			NewValue: fmt.Sprintf("%d logiciels", len(modified)),
			Reason:   "Réorganisation des logiciels par pertinence",
		})
	}

	return changelog
}

// compareStringArrays compare deux tableaux de strings et détecte les modifications
func compareStringArrays(fieldName string, original, modified []string) []models.ChangelogEntry {
	var changelog []models.ChangelogEntry

	if len(original) != len(modified) {
		changelog = append(changelog, models.ChangelogEntry{
			Field:    fieldName,
			OldValue: fmt.Sprintf("%d éléments", len(original)),
			NewValue: fmt.Sprintf("%d éléments", len(modified)),
			Reason:   "Réorganisation par pertinence",
		})
	}

	return changelog
}
