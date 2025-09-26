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
	fmt.Println("=== TEST D'EXTRACTION DU FICHIER GREGORY ===")
	
	// Lire le fichier Gregory
	fileData, err := ioutil.ReadFile("test_gregory.docx")
	if err != nil {
		fmt.Printf("❌ Erreur lecture fichier: %v\n", err)
		return
	}
	
	fmt.Printf("📁 Fichier: test_gregory.docx\n")
	fmt.Printf("📊 Taille: %d bytes\n", len(fileData))
	
	// Vérifier le type de fichier
	fileType := detectFileType("test_gregory.docx", fileData)
	fmt.Printf("🔍 Type détecté: %s\n", fileType)
	
	// Tester l'extraction DOCX
	fmt.Printf("\n🔧 Test extraction DOCX avec Docconv...\n")
	extractedText, err := extractTextFromDOCX(fileData)
	if err != nil {
		fmt.Printf("❌ Erreur extraction DOCX: %v\n", err)
		return
	}
	
	fmt.Printf("✅ Extraction DOCX réussie!\n")
	fmt.Printf("📝 Caractères extraits: %d\n", len(extractedText))
	fmt.Printf("\n📄 Début du contenu extrait:\n")
	fmt.Printf("---\n")
	fmt.Printf("%.500s...\n", extractedText)
	fmt.Printf("---\n")
	
	// Sauvegarder le texte extrait pour inspection
	err = ioutil.WriteFile("gregory_extracted.txt", []byte(extractedText), 0644)
	if err != nil {
		fmt.Printf("⚠️  Impossible de sauvegarder: %v\n", err)
	} else {
		fmt.Printf("💾 Texte sauvegardé dans: gregory_extracted.txt\n")
	}
}

func detectFileType(filename string, fileData []byte) string {
	filename = strings.ToLower(filename)
	
	if strings.HasSuffix(filename, ".docx") {
		if len(fileData) >= 4 && string(fileData[:4]) == "PK\x03\x04" {
			return "docx"
		}
	}
	return "unknown"
}

func extractTextFromDOCX(fileData []byte) (string, error) {
	log.Printf("DEBUG: Début extraction DOCX avec Docconv")
	
	reader := bytes.NewReader(fileData)
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
