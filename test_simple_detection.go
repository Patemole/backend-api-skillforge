package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	fmt.Println("=== TEST DE DÉTECTION DE TYPE DE FICHIER ===")
	
	// Test avec le fichier texte
	fmt.Println("\n1. Test avec fichier .txt:")
	testFileDetection("test_cv_word.txt")
	
	// Test avec un fichier PDF existant
	fmt.Println("\n2. Test avec fichier PDF:")
	testFileDetection("test_160k.bin")
	
	// Test avec des données simulées
	fmt.Println("\n3. Test avec données DOCX simulées:")
	testSimulatedFile("test_cv.docx", "docx")
	
	fmt.Println("\n4. Test avec données DOC simulées:")
	testSimulatedFile("test_cv.doc", "doc")
}

func testFileDetection(filename string) {
	// Lire le fichier s'il existe
	var fileData []byte
	
	if _, err := os.Stat(filename); err == nil {
		var err error
		fileData, err = ioutil.ReadFile(filename)
		if err != nil {
			fmt.Printf("  ❌ Erreur lecture fichier %s: %v\n", filename, err)
			return
		}
	} else {
		fmt.Printf("  ⚠️  Fichier %s n'existe pas\n", filename)
		return
	}
	
	// Tester la détection de type
	fileType := detectFileType(filename, fileData)
	fmt.Printf("  Fichier: %s\n", filename)
	fmt.Printf("  Taille: %d bytes\n", len(fileData))
	fmt.Printf("  Type détecté: %s\n", fileType)
	
	// Vérifier si la détection est correcte
	expectedType := getExpectedType(filename)
	if fileType == expectedType {
		fmt.Printf("  ✅ Détection correcte!\n")
	} else {
		fmt.Printf("  ❌ Détection incorrecte! Attendu: %s, Obtenu: %s\n", expectedType, fileType)
	}
}

func testSimulatedFile(filename, expectedType string) {
	var fileData []byte
	
	if expectedType == "docx" {
		// Magic number DOCX (ZIP)
		fileData = []byte("PK\x03\x04" + strings.Repeat("x", 100))
	} else if expectedType == "doc" {
		// Magic number DOC (OLE)
		fileData = []byte("\xd0\xcf\x11\xe0\xa1\xb1\x1a\xe1" + strings.Repeat("x", 100))
	}
	
	fileType := detectFileType(filename, fileData)
	fmt.Printf("  Fichier: %s\n", filename)
	fmt.Printf("  Taille: %d bytes\n", len(fileData))
	fmt.Printf("  Type détecté: %s\n", fileType)
	
	if fileType == expectedType {
		fmt.Printf("  ✅ Détection correcte!\n")
	} else {
		fmt.Printf("  ❌ Détection incorrecte! Attendu: %s, Obtenu: %s\n", expectedType, fileType)
	}
}

func getExpectedType(filename string) string {
	filename = strings.ToLower(filename)
	if strings.HasSuffix(filename, ".txt") {
		return "txt"
	} else if strings.HasSuffix(filename, ".pdf") {
		return "pdf"
	} else if strings.HasSuffix(filename, ".docx") {
		return "docx"
	} else if strings.HasSuffix(filename, ".doc") {
		return "doc"
	}
	return "unknown"
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
