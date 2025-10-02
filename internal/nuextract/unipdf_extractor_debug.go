package nuextract

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/unidoc/unipdf/v3/common/license"
	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
)

// UniPDFExtractorDebug version avec logs détaillés pour déboguer
type UniPDFExtractorDebug struct {
	apiKey string
}

// NewUniPDFExtractorDebug crée un nouvel extracteur UniPDF avec logs détaillés
func NewUniPDFExtractorDebug() *UniPDFExtractorDebug {
	// Initialiser la licence UniPDF (une seule fois)
	apiKey := os.Getenv("UNIPDF_API_KEY")
	log.Printf("DEBUG: [UniPDF] Clé API récupérée: %s", apiKey[:10]+"...")
	
	if apiKey != "" && !licenseConfigured {
		// Définir la licence (à faire une seule fois au démarrage de l'app)
		log.Printf("DEBUG: [UniPDF] Configuration de la licence...")
		err := license.SetMeteredKey(apiKey)
		if err != nil {
			log.Printf("ERROR: [UniPDF] Erreur licence: %v", err)
		} else {
			log.Printf("DEBUG: [UniPDF] Licence configurée avec succès")
			licenseConfigured = true
		}
	} else if apiKey == "" {
		log.Printf("WARNING: [UniPDF] UNIPDF_API_KEY non définie, utilisation en mode essai (14 jours)")
	} else {
		log.Printf("DEBUG: [UniPDF] Licence déjà configurée")
	}
	
	return &UniPDFExtractorDebug{
		apiKey: apiKey,
	}
}

// ExtractTextFromPDFWithTablesDebug version avec logs détaillés
func (e *UniPDFExtractorDebug) ExtractTextFromPDFWithTablesDebug(fileData []byte) (string, error) {
	log.Printf("DEBUG: [UniPDF] === DÉBUT EXTRACTION ===")
	log.Printf("DEBUG: [UniPDF] Taille des données: %d bytes", len(fileData))
	log.Printf("DEBUG: [UniPDF] Clé API présente: %t", e.apiKey != "")
	log.Printf("DEBUG: [UniPDF] Licence configurée: %t", licenseConfigured)
	
	// Vérifier que c'est bien un PDF
	if len(fileData) < 4 || string(fileData[:4]) != "%PDF" {
		log.Printf("ERROR: [UniPDF] Fichier ne semble pas être un PDF valide")
		return "", fmt.Errorf("fichier ne semble pas être un PDF valide")
	}
	log.Printf("DEBUG: [UniPDF] Signature PDF détectée: %s", string(fileData[:4]))
	
	// Créer un reader à partir des données du fichier
	reader := bytes.NewReader(fileData)
	log.Printf("DEBUG: [UniPDF] Reader créé, taille: %d", reader.Size())
	
	// Ouvrir le PDF
	log.Printf("DEBUG: [UniPDF] Création du reader PDF...")
	pdfReader, err := model.NewPdfReader(reader)
	if err != nil {
		log.Printf("ERROR: [UniPDF] Erreur ouverture PDF: %v", err)
		return "", fmt.Errorf("erreur ouverture PDF: %v", err)
	}
	log.Printf("DEBUG: [UniPDF] PDF ouvert avec succès")
	
	// Vérifier si le PDF est crypté
	log.Printf("DEBUG: [UniPDF] Vérification du cryptage...")
	isEncrypted, err := pdfReader.IsEncrypted()
	if err != nil {
		log.Printf("ERROR: [UniPDF] Erreur vérification cryptage: %v", err)
		return "", fmt.Errorf("erreur vérification cryptage: %v", err)
	}
	log.Printf("DEBUG: [UniPDF] PDF crypté: %t", isEncrypted)
	
	if isEncrypted {
		log.Printf("DEBUG: [UniPDF] Tentative de décryptage...")
		// Essayer de décrypter avec un mot de passe vide
		auth, err := pdfReader.Decrypt([]byte(""))
		if err != nil {
			log.Printf("ERROR: [UniPDF] Erreur décryptage: %v", err)
			return "", fmt.Errorf("PDF crypté, impossible de décrypter: %v", err)
		}
		if !auth {
			log.Printf("ERROR: [UniPDF] PDF crypté, mot de passe requis")
			return "", fmt.Errorf("PDF crypté, mot de passe requis")
		}
		log.Printf("DEBUG: [UniPDF] Décryptage réussi")
	}
	
	// Obtenir le nombre de pages
	log.Printf("DEBUG: [UniPDF] Obtention du nombre de pages...")
	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		log.Printf("ERROR: [UniPDF] Erreur obtention nombre de pages: %v", err)
		return "", fmt.Errorf("erreur obtention nombre de pages: %v", err)
	}
	
	log.Printf("DEBUG: [UniPDF] PDF ouvert, %d pages détectées", numPages)
	
	if numPages == 0 {
		log.Printf("ERROR: [UniPDF] PDF ne contient aucune page")
		return "", fmt.Errorf("PDF ne contient aucune page")
	}
	
	var extractedText strings.Builder
	
	// Extraire le texte de chaque page
	for i := 1; i <= numPages; i++ {
		log.Printf("DEBUG: [UniPDF] === EXTRACTION PAGE %d/%d ===", i, numPages)
		
		page, err := pdfReader.GetPage(i)
		if err != nil {
			log.Printf("ERROR: [UniPDF] Erreur page %d: %v", i, err)
			continue
		}
		log.Printf("DEBUG: [UniPDF] Page %d récupérée", i)
		
		// Créer un extracteur pour cette page
		extractor, err := extractor.New(page)
		if err != nil {
			log.Printf("ERROR: [UniPDF] Erreur création extracteur page %d: %v", i, err)
			continue
		}
		log.Printf("DEBUG: [UniPDF] Extracteur page %d créé", i)
		
		// Extraire le texte de la page
		text, err := extractor.ExtractText()
		if err != nil {
			log.Printf("ERROR: [UniPDF] Erreur extraction texte page %d: %v", i, err)
			continue
		}
		
		// Ajouter le texte extrait
		extractedText.WriteString(text)
		extractedText.WriteString("\n")
		
		log.Printf("DEBUG: [UniPDF] Page %d extraite, %d caractères", i, len(text))
		if len(text) > 0 {
			log.Printf("DEBUG: [UniPDF] Aperçu page %d: %.100s...", i, text)
		}
	}
	
	finalText := extractedText.String()
	log.Printf("DEBUG: [UniPDF] === EXTRACTION TERMINÉE ===")
	log.Printf("DEBUG: [UniPDF] Total caractères extraits: %d", len(finalText))
	if len(finalText) > 0 {
		log.Printf("DEBUG: [UniPDF] Aperçu final: %.200s...", finalText)
	}
	
	return finalText, nil
}
