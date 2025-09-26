package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"backend-api-skillforge/internal/nuextract"
)

func main() {
	fmt.Println("=== TEST DE DÉTECTION DE TYPE DE FICHIER ===")
	
	// Test avec le fichier texte
	fmt.Println("\n1. Test avec fichier .txt:")
	testFileDetection("test_cv_word.txt")
	
	// Test avec un fichier PDF existant
	fmt.Println("\n2. Test avec fichier PDF:")
	testFileDetection("test_160k.bin")
	
	// Test avec un fichier imaginaire .docx
	fmt.Println("\n3. Test avec fichier .docx (simulé):")
	testFileDetection("test_cv.docx")
	
	// Test avec un fichier imaginaire .doc
	fmt.Println("\n4. Test avec fichier .doc (simulé):")
	testFileDetection("test_cv.doc")
}

func testFileDetection(filename string) {
	// Lire le fichier s'il existe
	var fileData []byte
	var err error
	
	if _, err := os.Stat(filename); err == nil {
		fileData, err = ioutil.ReadFile(filename)
		if err != nil {
			log.Printf("Erreur lecture fichier %s: %v", filename, err)
			return
		}
	} else {
		// Simuler des données pour les fichiers qui n'existent pas
		if strings.HasSuffix(filename, ".docx") {
			// Magic number DOCX (ZIP)
			fileData = []byte("PK\x03\x04" + strings.Repeat("x", 100))
		} else if strings.HasSuffix(filename, ".doc") {
			// Magic number DOC (OLE)
			fileData = []byte("\xd0\xcf\x11\xe0\xa1\xb1\x1a\xe1" + strings.Repeat("x", 100))
		} else {
			fileData = []byte("Contenu de test")
		}
	}
	
	// Créer un client pour tester la détection
	client := nuextract.New()
	
	// Tester la détection de type
	fileType := detectFileType(filename, fileData)
	fmt.Printf("  Fichier: %s\n", filename)
	fmt.Printf("  Taille: %d bytes\n", len(fileData))
	fmt.Printf("  Type détecté: %s\n", fileType)
	
	// Tester l'extraction
	fmt.Printf("  Test d'extraction...\n")
	result, err := client.ExtractAndEnrichWithFilename(fileData, filename)
	if err != nil {
		fmt.Printf("  ❌ Erreur extraction: %v\n", err)
	} else {
		fmt.Printf("  ✅ Extraction réussie, taille résultat: %d bytes\n", len(result))
	}
}

// Fonction de détection copiée du package nuextract pour les tests
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
