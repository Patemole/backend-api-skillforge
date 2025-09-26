package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"backend-api-skillforge/internal/nuextract"
)

func main() {
	fmt.Println("=== TEST D'EXTRACTION DU FICHIER PAGES (FINAL) ===")
	
	// Lire le fichier Pages
	fileData, err := ioutil.ReadFile("test_gregory.pages")
	if err != nil {
		fmt.Printf("❌ Erreur lecture fichier: %v\n", err)
		return
	}
	
	fmt.Printf("📁 Fichier: test_gregory.pages\n")
	fmt.Printf("📊 Taille: %d bytes\n", len(fileData))
	
	// Créer un client pour tester
	client := nuextract.New()
	
	// Tester la détection de type
	fileType := detectFileType("test_gregory.pages", fileData)
	fmt.Printf("🔍 Type détecté: %s\n", fileType)
	
	// Tester l'extraction
	fmt.Printf("\n🔧 Test extraction Pages...\n")
	result, err := client.ExtractAndEnrichWithFilename(fileData, "test_gregory.pages")
	if err != nil {
		fmt.Printf("❌ Erreur extraction: %v\n", err)
	} else {
		fmt.Printf("✅ Extraction réussie, taille résultat: %d bytes\n", len(result))
		fmt.Printf("📄 Début du JSON: %.200s...\n", string(result))
	}
}

func detectFileType(filename string, fileData []byte) string {
	filename = strings.ToLower(filename)
	
	if strings.HasSuffix(filename, ".pages") {
		if len(fileData) >= 4 && string(fileData[:4]) == "PK\x03\x04" {
			return "pages"
		}
	}
	return "unknown"
}
