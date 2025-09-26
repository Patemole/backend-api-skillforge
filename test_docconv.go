package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"

	"code.sajari.com/docconv"
)

func main() {
	fmt.Println("=== TEST DOCCONV DIRECT ===")
	
	// Lire le fichier DOCX
	fileData, err := ioutil.ReadFile("/tmp/CV_test_backend.docx")
	if err != nil {
		log.Fatalf("Erreur lecture fichier: %v", err)
	}
	
	fmt.Printf("📁 Fichier: /tmp/CV_test_backend.docx\n")
	fmt.Printf("📊 Taille: %d bytes\n", len(fileData))
	
	// Tester l'extraction avec Docconv
	reader := bytes.NewReader(fileData)
	text, metadata, err := docconv.ConvertDocx(reader)
	if err != nil {
		fmt.Printf("❌ Erreur Docconv: %v\n", err)
	} else {
		fmt.Printf("✅ Extraction réussie!\n")
		fmt.Printf("📝 Caractères extraits: %d\n", len(text))
		fmt.Printf("📋 Métadonnées: %+v\n", metadata)
		fmt.Printf("📄 Début du texte:\n%.500s...\n", text)
	}
}
