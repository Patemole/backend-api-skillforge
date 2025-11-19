package nuextract

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/otiai10/gosseract/v2"
)

// extractTextFromPDFWithOCR converts PDF pages to images using pdftoppm and runs OCR
// This is used as a fallback when text extraction fails (image-based/scanned PDFs)
func extractTextFromPDFWithOCR(fileData []byte) (string, error) {
	log.Printf("🔍 [OCR] Début extraction OCR pour PDF image-based")

	// 1. Create temporary PDF file
	tmpPDF, err := os.CreateTemp("", "ocr_pdf_*.pdf")
	if err != nil {
		return "", fmt.Errorf("erreur création fichier temporaire: %v", err)
	}
	defer os.Remove(tmpPDF.Name())
	defer tmpPDF.Close()

	if _, err := tmpPDF.Write(fileData); err != nil {
		return "", fmt.Errorf("erreur écriture PDF temporaire: %v", err)
	}
	tmpPDF.Close()

	// 2. Convert PDF pages to PNG images using pdftoppm
	tmpDir := filepath.Dir(tmpPDF.Name())
	outputPrefix := filepath.Join(tmpDir, "ocr_page_")

	cmd := exec.Command("pdftoppm", "-png", "-r", "300", tmpPDF.Name(), outputPrefix)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("erreur conversion PDF en images (pdftoppm non installé?): %v", err)
	}

	// 3. Find all generated PNG files
	pattern := outputPrefix + "*.png"
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return "", fmt.Errorf("aucune image générée depuis le PDF")
	}

	log.Printf("🔍 [OCR] %d pages converties en images", len(matches))

	// 4. Initialize Tesseract client
	client := gosseract.NewClient()
	defer client.Close()

	// Set language (French + English for CVs)
	ocrLang := os.Getenv("TESSERACT_LANG")
	if ocrLang == "" {
		ocrLang = "fra+eng" // French + English
	}
	if err := client.SetLanguage(ocrLang); err != nil {
		log.Printf("⚠️ [OCR] Erreur configuration langue OCR (%s), utilisation par défaut: %v", ocrLang, err)
		// Continue with default language
	}

	// 5. Process each image with OCR
	var allText strings.Builder
	for i, imgPath := range matches {
		pageNum := i + 1
		log.Printf("🔍 [OCR] Traitement page %d/%d: %s", pageNum, len(matches), filepath.Base(imgPath))

		// Clean up image file after processing
		defer os.Remove(imgPath)

		// Run OCR on image
		if err := client.SetImage(imgPath); err != nil {
			log.Printf("⚠️ [OCR] Erreur chargement image page %d: %v", pageNum, err)
			continue
		}

		text, err := client.Text()
		if err != nil {
			log.Printf("⚠️ [OCR] Erreur OCR page %d: %v", pageNum, err)
			continue
		}

		if len(strings.TrimSpace(text)) > 0 {
			allText.WriteString(text)
			allText.WriteString("\n\n") // Separate pages
			log.Printf("✅ [OCR] Page %d: %d caractères extraits", pageNum, len(text))
		} else {
			log.Printf("⚠️ [OCR] Page %d: aucun texte détecté", pageNum)
		}
	}

	result := strings.TrimSpace(allText.String())
	if len(result) < 50 {
		return "", fmt.Errorf("OCR extrait trop peu de texte (%d caractères)", len(result))
	}

	log.Printf("✅ [OCR] Extraction terminée: %d caractères au total", len(result))
	return result, nil
}
