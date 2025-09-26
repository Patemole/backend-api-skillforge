package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"backend-api-skillforge/internal/nuextract"
)

func main() {
	fmt.Println("=== TEST DÉTECTION ET EXTRACTION DE TOUS LES FORMATS ===")
	
	// Test 1: RTF
	fmt.Println("\n1. Test RTF:")
	testFormat("test_cv.rtf", "rtf", "{\\rtf1\\ansi\\deff0 {\\fonttbl {\\f0 Times New Roman;}} \\f0\\fs24 CV Test RTF\\par }")
	
	// Test 2: ODT (simulé comme ZIP)
	fmt.Println("\n2. Test ODT:")
	testFormat("test_cv.odt", "odt", "PK\x03\x04"+strings.Repeat("a", 100))
	
	// Test 3: HTML
	fmt.Println("\n3. Test HTML:")
	htmlContent := `<!DOCTYPE html>
<html>
<head><title>CV Test</title></head>
<body>
<h1>Jean Dupont</h1>
<p>Développeur Full Stack</p>
<p>Email: jean@example.com</p>
</body>
</html>`
	testFormat("test_cv.html", "html", htmlContent)
	
	// Test 4: Markdown
	fmt.Println("\n4. Test Markdown:")
	markdownContent := `# CV - Jean Dupont

## Informations personnelles
- **Nom**: Jean Dupont
- **Email**: jean@example.com
- **Poste**: Développeur Full Stack

## Expérience
- 2020-2024: Développeur chez TechCorp
- 2018-2020: Stagiaire chez StartupX

## Compétences
- JavaScript (Expert)
- Python (Avancé)
- React (Expert)`
	testFormat("test_cv.md", "markdown", markdownContent)
	
	// Test 5: TXT
	fmt.Println("\n5. Test TXT:")
	txtContent := `CV - Jean Dupont
Développeur Full Stack
Email: jean@example.com
Tél: 01 23 45 67 89

EXPÉRIENCE:
2020-2024: Développeur Full Stack - TechCorp
2018-2020: Stagiaire - StartupX

COMPÉTENCES:
- JavaScript (Expert)
- Python (Avancé)
- React (Expert)`
	testFormat("test_cv.txt", "txt", txtContent)
}

func testFormat(filename, expectedType, content string) {
	// Créer le fichier de test
	err := ioutil.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		fmt.Printf("  ❌ Erreur création fichier: %v\n", err)
		return
	}
	defer func() {
		// Nettoyer le fichier de test
		_ = ioutil.WriteFile(filename, []byte{}, 0644)
	}()
	
	// Lire le fichier
	fileData, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("  ❌ Erreur lecture: %v\n", err)
		return
	}
	
	fmt.Printf("  📁 Fichier: %s\n", filename)
	fmt.Printf("  📊 Taille: %d bytes\n", len(fileData))
	
	// Tester l'extraction complète avec le client
	fmt.Printf("  🔧 Test extraction complète...\n")
	client := nuextract.New()
	
	// Tester l'extraction (sans OpenAI pour éviter l'erreur API)
	result, err := client.ExtractAndEnrichWithFilename(fileData, filename)
	if err != nil {
		fmt.Printf("  ❌ Erreur extraction: %v\n", err)
	} else {
		fmt.Printf("  ✅ Extraction réussie!\n")
		fmt.Printf("  📝 Taille résultat: %d bytes\n", len(result))
		
		// Afficher un extrait du JSON
		displayLength := 300
		if len(result) < displayLength {
			displayLength = len(result)
		}
		fmt.Printf("  📄 Extrait JSON: %.300s...\n", string(result))
	}
}
