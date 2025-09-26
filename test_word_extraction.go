package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"code.sajari.com/docconv"
)

func main() {
	fmt.Println("=== TEST D'EXTRACTION DE FICHIERS WORD ===")
	
	// Test avec le fichier texte existant
	fmt.Println("\n1. Test avec fichier .txt:")
	testExtraction("test_cv_word.txt", "txt")
	
	// Test avec des données simulées DOCX
	fmt.Println("\n2. Test avec données DOCX simulées:")
	testSimulatedExtraction("test_cv.docx", "docx")
	
	// Test avec des données simulées DOC
	fmt.Println("\n3. Test avec données DOC simulées:")
	testSimulatedExtraction("test_cv.doc", "doc")
}

func testExtraction(filename, expectedType string) {
	// Lire le fichier
	fileData, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("  ❌ Erreur lecture fichier: %v\n", err)
		return
	}
	
	fmt.Printf("  Fichier: %s\n", filename)
	fmt.Printf("  Taille: %d bytes\n", len(fileData))
	
	// Tester la détection de type
	fileType := detectFileType(filename, fileData)
	fmt.Printf("  Type détecté: %s\n", fileType)
	
	// Tester l'extraction selon le type
	switch fileType {
	case "txt":
		fmt.Printf("  ✅ Fichier texte - traitement direct\n")
		content := string(fileData)
		fmt.Printf("  Contenu extrait: %d caractères\n", len(content))
		fmt.Printf("  Début du contenu: %.100s...\n", content)
		
	case "docx":
		fmt.Printf("  🔧 Test extraction DOCX...\n")
		content, err := extractTextFromDOCX(fileData)
		if err != nil {
			fmt.Printf("  ❌ Erreur extraction DOCX: %v\n", err)
		} else {
			fmt.Printf("  ✅ Extraction DOCX réussie: %d caractères\n", len(content))
			fmt.Printf("  Début du contenu: %.100s...\n", content)
		}
		
	case "doc":
		fmt.Printf("  🔧 Test extraction DOC...\n")
		content, err := extractTextFromDOC(fileData)
		if err != nil {
			fmt.Printf("  ❌ Erreur extraction DOC: %v\n", err)
		} else {
			fmt.Printf("  ✅ Extraction DOC réussie: %d caractères\n", len(content))
			fmt.Printf("  Début du contenu: %.100s...\n", content)
		}
		
	default:
		fmt.Printf("  ❓ Type non géré: %s\n", fileType)
	}
}

func testSimulatedExtraction(filename, fileType string) {
	var fileData []byte
	
	if fileType == "docx" {
		// Magic number DOCX (ZIP) + contenu simulé
		fileData = []byte("PK\x03\x04" + strings.Repeat("x", 100))
	} else if fileType == "doc" {
		// Magic number DOC (OLE) + contenu simulé
		fileData = []byte("\xd0\xcf\x11\xe0\xa1\xb1\x1a\xe1" + strings.Repeat("x", 100))
	}
	
	fmt.Printf("  Fichier: %s\n", filename)
	fmt.Printf("  Taille: %d bytes\n", len(fileData))
	
	detectedType := detectFileType(filename, fileData)
	fmt.Printf("  Type détecté: %s\n", detectedType)
	
	// Tester l'extraction
	if fileType == "docx" {
		content, err := extractTextFromDOCX(fileData)
		if err != nil {
			fmt.Printf("  ❌ Erreur extraction DOCX: %v\n", err)
		} else {
			fmt.Printf("  ✅ Extraction DOCX réussie: %d caractères\n", len(content))
		}
	} else if fileType == "doc" {
		content, err := extractTextFromDOC(fileData)
		if err != nil {
			fmt.Printf("  ❌ Erreur extraction DOC: %v\n", err)
		} else {
			fmt.Printf("  ✅ Extraction DOC réussie: %d caractères\n", len(content))
		}
	}
}

// Fonctions copiées du package nuextract pour les tests
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
