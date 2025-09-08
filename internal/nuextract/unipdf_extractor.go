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

// UniPDFExtractor utilise UniPDF pour extraire le texte des PDFs
type UniPDFExtractor struct {
	apiKey string
}

// Variable globale pour éviter de reconfigurer la licence à chaque fois
var licenseConfigured bool

// NewUniPDFExtractor crée un nouvel extracteur UniPDF
func NewUniPDFExtractor() *UniPDFExtractor {
	// Initialiser la licence UniPDF (une seule fois)
	apiKey := os.Getenv("UNIPDF_API_KEY")
	if apiKey != "" && !licenseConfigured {
		// Définir la licence (à faire une seule fois au démarrage de l'app)
		err := license.SetMeteredKey(apiKey)
		if err != nil {
			log.Printf("WARNING: UniPDF license error: %v", err)
		} else {
			log.Printf("DEBUG: Licence UniPDF configurée avec succès")
			licenseConfigured = true
		}
	} else if apiKey == "" {
		log.Printf("WARNING: UNIPDF_API_KEY non définie, utilisation en mode essai (14 jours)")
	} else {
		log.Printf("DEBUG: Licence UniPDF déjà configurée")
	}
	
	return &UniPDFExtractor{
		apiKey: apiKey,
	}
}

// ExtractTextFromPDF extrait le texte d'un PDF en utilisant UniPDF
func (e *UniPDFExtractor) ExtractTextFromPDF(fileData []byte) (string, error) {
	log.Printf("DEBUG: Début extraction UniPDF")
	
	// Créer un reader à partir des données du fichier
	reader := bytes.NewReader(fileData)
	
	// Ouvrir le PDF
	pdfReader, err := model.NewPdfReader(reader)
	if err != nil {
		return "", fmt.Errorf("erreur ouverture PDF: %v", err)
	}
	
	// Vérifier si le PDF est crypté
	isEncrypted, err := pdfReader.IsEncrypted()
	if err != nil {
		return "", fmt.Errorf("erreur vérification cryptage: %v", err)
	}
	
	if isEncrypted {
		// Essayer de décrypter avec un mot de passe vide
		auth, err := pdfReader.Decrypt([]byte(""))
		if err != nil {
			return "", fmt.Errorf("PDF crypté, impossible de décrypter: %v", err)
		}
		if !auth {
			return "", fmt.Errorf("PDF crypté, mot de passe requis")
		}
	}
	
	// Obtenir le nombre de pages
	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return "", fmt.Errorf("erreur obtention nombre de pages: %v", err)
	}
	
	log.Printf("DEBUG: PDF ouvert, %d pages détectées", numPages)
	
	var extractedText strings.Builder
	
	// Extraire le texte de chaque page
	for i := 1; i <= numPages; i++ {
		log.Printf("DEBUG: Extraction page %d/%d", i, numPages)
		
		page, err := pdfReader.GetPage(i)
		if err != nil {
			log.Printf("WARNING: Erreur page %d: %v", i, err)
			continue
		}
		
		// Créer un extracteur pour cette page
		extractor, err := extractor.New(page)
		if err != nil {
			log.Printf("WARNING: Erreur création extracteur page %d: %v", i, err)
			continue
		}
		
		// Extraire le texte de la page
		text, err := extractor.ExtractText()
		if err != nil {
			log.Printf("WARNING: Erreur extraction texte page %d: %v", i, err)
			continue
		}
		
		// Ajouter le texte extrait
		extractedText.WriteString(text)
		extractedText.WriteString("\n")
		
		log.Printf("DEBUG: Page %d extraite, %d caractères", i, len(text))
	}
	
	finalText := extractedText.String()
	log.Printf("DEBUG: Extraction UniPDF terminée, %d caractères au total", len(finalText))
	
	return finalText, nil
}

// ExtractTextFromPDFWithTables extrait le texte ET les tableaux d'un PDF
func (e *UniPDFExtractor) ExtractTextFromPDFWithTables(fileData []byte) (string, error) {
	log.Printf("DEBUG: Début extraction UniPDF avec tableaux")
	
	// Créer un reader à partir des données du fichier
	reader := bytes.NewReader(fileData)
	
	// Ouvrir le PDF
	pdfReader, err := model.NewPdfReader(reader)
	if err != nil {
		return "", fmt.Errorf("erreur ouverture PDF: %v", err)
	}
	
	// Vérifier si le PDF est crypté
	isEncrypted, err := pdfReader.IsEncrypted()
	if err != nil {
		return "", fmt.Errorf("erreur vérification cryptage: %v", err)
	}
	
	if isEncrypted {
		// Essayer de décrypter avec un mot de passe vide
		auth, err := pdfReader.Decrypt([]byte(""))
		if err != nil {
			return "", fmt.Errorf("PDF crypté, impossible de décrypter: %v", err)
		}
		if !auth {
			return "", fmt.Errorf("PDF crypté, mot de passe requis")
		}
	}
	
	// Obtenir le nombre de pages
	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return "", fmt.Errorf("erreur obtention nombre de pages: %v", err)
	}
	
	log.Printf("DEBUG: PDF ouvert, %d pages détectées", numPages)
	
	var extractedText strings.Builder
	
	// Extraire le texte de chaque page
	for i := 1; i <= numPages; i++ {
		log.Printf("DEBUG: Extraction page %d/%d avec tableaux", i, numPages)
		
		page, err := pdfReader.GetPage(i)
		if err != nil {
			log.Printf("WARNING: Erreur page %d: %v", i, err)
			continue
		}
		
		// Créer un extracteur pour cette page
		extractor, err := extractor.New(page)
		if err != nil {
			log.Printf("WARNING: Erreur création extracteur page %d: %v", i, err)
			continue
		}
		
		// Extraire le texte de la page
		text, err := extractor.ExtractText()
		if err != nil {
			log.Printf("WARNING: Erreur extraction texte page %d: %v", i, err)
			continue
		}
		
		// Ajouter le texte extrait
		extractedText.WriteString(text)
		extractedText.WriteString("\n")
		
		// Note: L'extraction de tableaux nécessite une licence complète UniPDF
		// Pour l'instant, on se contente de l'extraction de texte standard
		
		log.Printf("DEBUG: Page %d extraite, %d caractères", i, len(text))
	}
	
	finalText := extractedText.String()
	log.Printf("DEBUG: Extraction UniPDF avec tableaux terminée, %d caractères au total", len(finalText))
	
	return finalText, nil
}
