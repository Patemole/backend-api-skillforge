package nuextract

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"code.sajari.com/docconv"
)

// extractTextFromDOCX extrait le texte d'un fichier DOCX en utilisant Docconv
func extractTextFromDOCX(fileData []byte) (string, error) {
	log.Printf("DEBUG: Début extraction DOCX avec Docconv")
	
	// Créer un reader à partir des données du fichier
	reader := bytes.NewReader(fileData)
	
	// Utiliser Docconv pour extraire le texte DOCX
	extractedText, _, err := docconv.ConvertDocx(reader)
	if err != nil {
		return "", fmt.Errorf("erreur extraction DOCX avec Docconv: %v", err)
	}
	
	if extractedText == "" {
		return "", fmt.Errorf("aucun texte extrait du fichier DOCX")
	}
	
	log.Printf("DEBUG: Extraction DOCX réussie, %d caractères extraits", len(extractedText))
	return extractedText, nil
}

// extractTextFromDOC extrait le texte d'un fichier DOC en utilisant Docconv
func extractTextFromDOC(fileData []byte) (string, error) {
	log.Printf("DEBUG: Début extraction DOC avec Docconv")
	
	// Créer un reader à partir des données du fichier
	reader := bytes.NewReader(fileData)
	
	// Utiliser Docconv pour extraire le texte DOC
	extractedText, _, err := docconv.ConvertDoc(reader)
	if err != nil {
		return "", fmt.Errorf("erreur extraction DOC avec Docconv: %v", err)
	}
	
	if extractedText == "" {
		return "", fmt.Errorf("aucun texte extrait du fichier DOC")
	}
	
	log.Printf("DEBUG: Extraction DOC réussie, %d caractères extraits", len(extractedText))
	return extractedText, nil
}

// convertPagesToDOCX convertit un fichier Pages en DOCX en utilisant LibreOffice
func convertPagesToDOCX(fileData []byte, filename string) ([]byte, error) {
	log.Printf("DEBUG: Conversion Pages → DOCX avec LibreOffice")
	
	// Créer un fichier temporaire pour le Pages
	tempPagesFile := fmt.Sprintf("/tmp/%s", filename)
	err := ioutil.WriteFile(tempPagesFile, fileData, 0644)
	if err != nil {
		return nil, fmt.Errorf("erreur création fichier temporaire Pages: %v", err)
	}
	defer os.Remove(tempPagesFile)
	
	// Créer un répertoire temporaire pour la conversion
	tempDir := fmt.Sprintf("/tmp/pages_conversion_%d", time.Now().Unix())
	err = os.MkdirAll(tempDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("erreur création répertoire temporaire: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Convertir Pages → DOCX avec LibreOffice
	cmd := exec.Command("soffice", "--headless", "--convert-to", "docx", "--outdir", tempDir, tempPagesFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("erreur conversion LibreOffice: %v, output: %s", err, string(output))
	}
	
	log.Printf("DEBUG: LibreOffice conversion output: %s", string(output))
	
	// Chercher le fichier DOCX généré
	docxFile := fmt.Sprintf("%s/%s.docx", tempDir, strings.TrimSuffix(filename, ".pages"))
	if _, err := os.Stat(docxFile); err != nil {
		// Essayer avec un nom différent
		files, _ := ioutil.ReadDir(tempDir)
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".docx") {
				docxFile = fmt.Sprintf("%s/%s", tempDir, file.Name())
				break
			}
		}
	}
	
	// Lire le fichier DOCX généré
	docxData, err := ioutil.ReadFile(docxFile)
	if err != nil {
		return nil, fmt.Errorf("erreur lecture fichier DOCX généré: %v", err)
	}
	
	log.Printf("DEBUG: Conversion Pages → DOCX réussie, %d bytes générés", len(docxData))
	return docxData, nil
}

// extractTextFromPages extrait le texte d'un fichier Pages en le convertissant d'abord en DOCX
func extractTextFromPages(fileData []byte) (string, error) {
	log.Printf("DEBUG: Début extraction Pages avec conversion Pages → DOCX")
	
	// Convertir Pages → DOCX
	docxData, err := convertPagesToDOCX(fileData, "temp_pages_file.pages")
	if err != nil {
		return "", fmt.Errorf("erreur conversion Pages → DOCX: %v", err)
	}
	
	// Utiliser notre système d'extraction DOCX existant
	log.Printf("DEBUG: Extraction du DOCX généré avec Docconv")
	extractedText, _, err := docconv.ConvertDocx(bytes.NewReader(docxData))
	if err != nil {
		return "", fmt.Errorf("erreur extraction DOCX généré: %v", err)
	}
	
	if extractedText == "" {
		return "", fmt.Errorf("aucun texte extrait du DOCX généré")
	}
	
	log.Printf("DEBUG: Extraction Pages réussie, %d caractères extraits", len(extractedText))
	return extractedText, nil
}

// extractTextFromRTF extrait le texte d'un fichier RTF en utilisant Docconv
func extractTextFromRTF(fileData []byte) (string, error) {
	log.Printf("DEBUG: Début extraction RTF avec Docconv")
	
	reader := bytes.NewReader(fileData)
	extractedText, _, err := docconv.ConvertRTF(reader)
	if err != nil {
		return "", fmt.Errorf("erreur extraction RTF avec Docconv: %v", err)
	}
	
	if extractedText == "" {
		return "", fmt.Errorf("aucun texte extrait du fichier RTF")
	}
	
	log.Printf("DEBUG: Extraction RTF réussie, %d caractères extraits", len(extractedText))
	return extractedText, nil
}

// extractTextFromODT extrait le texte d'un fichier ODT en utilisant Docconv
func extractTextFromODT(fileData []byte) (string, error) {
	log.Printf("DEBUG: Début extraction ODT avec Docconv")
	
	reader := bytes.NewReader(fileData)
	extractedText, _, err := docconv.ConvertODT(reader)
	if err != nil {
		return "", fmt.Errorf("erreur extraction ODT avec Docconv: %v", err)
	}
	
	if extractedText == "" {
		return "", fmt.Errorf("aucun texte extrait du fichier ODT")
	}
	
	log.Printf("DEBUG: Extraction ODT réussie, %d caractères extraits", len(extractedText))
	return extractedText, nil
}

// extractTextFromHTML extrait le texte d'un fichier HTML
func extractTextFromHTML(fileData []byte) (string, error) {
	log.Printf("DEBUG: Début extraction HTML")
	
	// Essayer d'abord Docconv
	reader := bytes.NewReader(fileData)
	extractedText, _, err := docconv.ConvertHTML(reader, true) // true = readability
	if err != nil {
		log.Printf("WARNING: Docconv HTML échoué: %v", err)
		// Fallback: extraction manuelle simple
		return extractTextFromHTMLSimple(fileData), nil
	}
	
	if extractedText == "" {
		log.Printf("WARNING: Docconv HTML vide, essai extraction manuelle")
		// Fallback: extraction manuelle simple
		return extractTextFromHTMLSimple(fileData), nil
	}
	
	log.Printf("DEBUG: Extraction HTML réussie, %d caractères extraits", len(extractedText))
	return extractedText, nil
}

// extractTextFromHTMLSimple extraction HTML manuelle simple
func extractTextFromHTMLSimple(fileData []byte) string {
	content := string(fileData)
	
	// Méthode plus robuste: utiliser regex pour supprimer les balises
	// Supprimer les balises HTML avec une approche simple
	lines := strings.Split(content, "\n")
	var result []string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Supprimer les balises HTML simples
		line = strings.ReplaceAll(line, "<html>", "")
		line = strings.ReplaceAll(line, "</html>", "")
		line = strings.ReplaceAll(line, "<head>", "")
		line = strings.ReplaceAll(line, "</head>", "")
		line = strings.ReplaceAll(line, "<body>", "")
		line = strings.ReplaceAll(line, "</body>", "")
		line = strings.ReplaceAll(line, "<title>", "")
		line = strings.ReplaceAll(line, "</title>", "")
		line = strings.ReplaceAll(line, "<h1>", "")
		line = strings.ReplaceAll(line, "</h1>", "")
		line = strings.ReplaceAll(line, "<h2>", "")
		line = strings.ReplaceAll(line, "</h2>", "")
		line = strings.ReplaceAll(line, "<h3>", "")
		line = strings.ReplaceAll(line, "</h3>", "")
		line = strings.ReplaceAll(line, "<p>", "")
		line = strings.ReplaceAll(line, "</p>", "")
		line = strings.ReplaceAll(line, "<ul>", "")
		line = strings.ReplaceAll(line, "</ul>", "")
		line = strings.ReplaceAll(line, "<li>", "- ")
		line = strings.ReplaceAll(line, "</li>", "")
		line = strings.ReplaceAll(line, "<strong>", "")
		line = strings.ReplaceAll(line, "</strong>", "")
		line = strings.ReplaceAll(line, "<meta", "")
		line = strings.ReplaceAll(line, "<!DOCTYPE", "")
		
		// Nettoyer les espaces multiples
		line = strings.TrimSpace(line)
		
		// Garder seulement les lignes avec du contenu
		if line != "" && !strings.HasPrefix(line, "<") && len(line) > 2 {
			result = append(result, line)
		}
	}
	
	extractedText := strings.Join(result, "\n")
	log.Printf("DEBUG: Extraction HTML manuelle réussie, %d caractères extraits", len(extractedText))
	return extractedText
}

// detectFileType détecte le type de fichier basé sur l'extension et le contenu
func detectFileType(filename string, fileData []byte) string {
	filename = strings.ToLower(filename)
	
	// Détection par extension
	if strings.HasSuffix(filename, ".pdf") {
		// Vérifier le magic number PDF
		if len(fileData) >= 4 && string(fileData[:4]) == "%PDF" {
			return "pdf"
		}
	} else if strings.HasSuffix(filename, ".docx") {
		// Vérifier le magic number DOCX (ZIP-based)
		if len(fileData) >= 4 && string(fileData[:4]) == "PK\x03\x04" {
			return "docx"
		}
	} else if strings.HasSuffix(filename, ".doc") {
		// Les fichiers .doc ont un magic number différent
		if len(fileData) >= 8 {
			// Magic number pour les fichiers .doc (OLE compound document)
			if string(fileData[:8]) == "\xd0\xcf\x11\xe0\xa1\xb1\x1a\xe1" {
				return "doc"
			}
		}
	} else if strings.HasSuffix(filename, ".pages") {
		// Vérifier le magic number Pages (ZIP-based)
		if len(fileData) >= 4 && string(fileData[:4]) == "PK\x03\x04" {
			return "pages"
		}
	} else if strings.HasSuffix(filename, ".rtf") {
		// RTF commence par {\rtf
		if len(fileData) >= 5 && string(fileData[:5]) == "{\\rtf" {
			return "rtf"
		}
	} else if strings.HasSuffix(filename, ".odt") {
		// ODT est un ZIP
		if len(fileData) >= 4 && string(fileData[:4]) == "PK\x03\x04" {
			return "odt"
		}
	} else if strings.HasSuffix(filename, ".html") || strings.HasSuffix(filename, ".htm") {
		// HTML commence par <html ou <!DOCTYPE
		content := strings.ToLower(string(fileData[:min(100, len(fileData))]))
		if strings.Contains(content, "<html") || strings.Contains(content, "<!doctype") {
			return "html"
		}
	} else if strings.HasSuffix(filename, ".md") || strings.HasSuffix(filename, ".markdown") {
		return "markdown"
	} else if strings.HasSuffix(filename, ".txt") {
		return "txt"
	}
	
	// Détection par taille et contenu pour les fichiers sans extension
	if len(fileData) < 1000 {
		return "txt"
	}
	
	// Par défaut, essayer de traiter comme PDF
	return "pdf"
}

// Fonction utilitaire pour min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}