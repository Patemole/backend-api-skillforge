package main

import (
	"fmt"
	"io/ioutil"

	"backend-api-skillforge/internal/nuextract"
)

func main() {
	fmt.Println("=== TEST AVEC DE VRAIS FICHIERS ===")
	
	// Test 1: RTF réel
	fmt.Println("\n1. Test RTF réel:")
	rtfContent := `{\rtf1\ansi\deff0 {\fonttbl {\f0 Times New Roman;}} \f0\fs24 
CV - Jean Dupont
Développeur Full Stack
Email: jean@example.com
Tél: 01 23 45 67 89

EXPÉRIENCE:
2020-2024: Développeur Full Stack - TechCorp
2018-2020: Stagiaire - StartupX

COMPÉTENCES:
- JavaScript (Expert)
- Python (Avancé)
- React (Expert)
}`
	testRealFormat("test_real.rtf", rtfContent)
	
	// Test 2: HTML réel
	fmt.Println("\n2. Test HTML réel:")
	htmlContent := `<!DOCTYPE html>
<html>
<head>
    <title>CV - Jean Dupont</title>
    <meta charset="UTF-8">
</head>
<body>
    <h1>Jean Dupont</h1>
    <h2>Développeur Full Stack</h2>
    <p><strong>Email:</strong> jean@example.com</p>
    <p><strong>Tél:</strong> 01 23 45 67 89</p>
    
    <h3>Expérience</h3>
    <ul>
        <li>2020-2024: Développeur Full Stack - TechCorp</li>
        <li>2018-2020: Stagiaire - StartupX</li>
    </ul>
    
    <h3>Compétences</h3>
    <ul>
        <li>JavaScript (Expert)</li>
        <li>Python (Avancé)</li>
        <li>React (Expert)</li>
    </ul>
</body>
</html>`
	testRealFormat("test_real.html", htmlContent)
	
	// Test 3: Markdown réel
	fmt.Println("\n3. Test Markdown réel:")
	markdownContent := `# CV - Jean Dupont

## Informations personnelles
- **Nom**: Jean Dupont
- **Email**: jean@example.com
- **Tél**: 01 23 45 67 89
- **Poste**: Développeur Full Stack

## Expérience professionnelle
- **2020-2024**: Développeur Full Stack - TechCorp
  - Développement d'applications web avec React et Node.js
  - Gestion de bases de données PostgreSQL
  - Collaboration avec une équipe de 5 développeurs

- **2018-2020**: Stagiaire - StartupX
  - Création d'interfaces utilisateur avec Vue.js
  - Intégration de designs responsive
  - Optimisation des performances web

## Compétences techniques
- **Langages**: JavaScript (Expert), Python (Avancé), Go (Intermédiaire)
- **Frameworks**: React (Expert), Vue.js (Avancé), Express.js (Intermédiaire)
- **Bases de données**: PostgreSQL (Avancé), MongoDB (Intermédiaire)
- **Outils**: Git (Avancé), Docker (Intermédiaire), AWS (Intermédiaire)

## Formation
- **2016-2018**: Master en Informatique - Université de Paris
- **2014-2016**: Licence en Mathématiques - Université de Lyon

## Langues
- Français: Langue maternelle
- Anglais: Courant
- Espagnol: Intermédiaire

## Centres d'intérêt
- Photographie
- Randonnée
- Lecture
- Développement open source`
	testRealFormat("test_real.md", markdownContent)
}

func testRealFormat(filename, content string) {
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
		displayLength := 500
		if len(result) < displayLength {
			displayLength = len(result)
		}
		fmt.Printf("  📄 Extrait JSON: %.500s...\n", string(result))
	}
}
