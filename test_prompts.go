package main

import (
	"backend-api-skillforge/internal/nuextract"
	"fmt"
)

func main() {
	fmt.Println("🧪 Test des prompts de langue")
	fmt.Println("=============================")

	// Données de test
	sampleData := `{
		"text": "Jean Dupont\nIngénieur Informatique\nEmail: jean.dupont@email.com\nTéléphone: 01.23.45.67.89\n\nExpérience:\n- 2020-2023: Développeur Full Stack chez TechCorp\n- 2018-2020: Stagiaire Développeur chez StartupXYZ\n\nFormation:\n- 2016-2020: Diplôme d'Ingénieur en Informatique - École Polytechnique\n\nCompétences: Python, JavaScript, React, Node.js"
	}`

	fmt.Println("\n🇫🇷 Test du prompt français:")
	fmt.Println("----------------------------")
	frenchPrompt := nuextract.GetExtractionPromptWithLanguage(sampleData, "fr")
	fmt.Printf("Longueur du prompt: %d caractères\n", len(frenchPrompt))
	fmt.Printf("Début du prompt: %s...\n", frenchPrompt[:200])

	fmt.Println("\n🇬🇧 Test du prompt anglais:")
	fmt.Println("----------------------------")
	englishPrompt := nuextract.GetExtractionPromptWithLanguage(sampleData, "en")
	fmt.Printf("Longueur du prompt: %d caractères\n", len(englishPrompt))
	fmt.Printf("Début du prompt: %s...\n", englishPrompt[:200])

	fmt.Println("\n🔍 Vérification des différences:")
	fmt.Println("-------------------------------")
	if len(frenchPrompt) != len(englishPrompt) {
		fmt.Printf("✅ Les prompts ont des longueurs différentes (FR: %d, EN: %d)\n", len(frenchPrompt), len(englishPrompt))
	} else {
		fmt.Println("⚠️  Les prompts ont la même longueur - vérifiez le contenu")
	}

	// Vérifier que les prompts contiennent des mots-clés spécifiques à chaque langue
	if contains(frenchPrompt, "Tu es un expert RH") {
		fmt.Println("✅ Prompt français contient les mots-clés français")
	} else {
		fmt.Println("❌ Prompt français ne contient pas les mots-clés attendus")
	}

	if contains(englishPrompt, "You are an HR expert") {
		fmt.Println("✅ Prompt anglais contient les mots-clés anglais")
	} else {
		fmt.Println("❌ Prompt anglais ne contient pas les mots-clés attendus")
	}

	fmt.Println("\n✅ Tests terminés")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}
