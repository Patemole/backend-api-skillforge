package nuextract

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"baliance.com/gooxml/document"
)

// extractTextFromDOCX extrait le texte d'un fichier DOCX (Word)
func extractTextFromDOCX(fileData []byte) (string, error) {
	// Supprimer le bruit de logs de la lib gooxml pendant l'ouverture/lecture
	prevOut := log.Writer()
	prevFlags := log.Flags()
	prevPrefix := log.Prefix()
	log.SetOutput(io.Discard)
	log.SetFlags(0)
	log.SetPrefix("")
	defer func() {
		log.SetOutput(prevOut)
		log.SetFlags(prevFlags)
		log.SetPrefix(prevPrefix)
	}()

	// Écrire les données dans un fichier temporaire car gooxml lit depuis un chemin de fichier
	tmpFile, err := os.CreateTemp("", "cv-*.docx")
	if err != nil {
		return "", fmt.Errorf("erreur création fichier temporaire DOCX: %v", err)
	}
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	if _, err := tmpFile.Write(fileData); err != nil {
		_ = tmpFile.Close()
		return "", fmt.Errorf("erreur écriture données DOCX: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		return "", fmt.Errorf("erreur fermeture fichier temporaire DOCX: %v", err)
	}

	doc, err := document.Open(tmpFile.Name())
	if err != nil {
		return "", fmt.Errorf("erreur ouverture DOCX: %v", err)
	}

	var sb strings.Builder

	paragraphText := func(p document.Paragraph) string {
		var b strings.Builder
		for _, run := range p.Runs() {
			b.WriteString(run.Text())
		}
		return b.String()
	}

	// Texte principal
	for _, para := range doc.Paragraphs() {
		sb.WriteString(paragraphText(para))
		sb.WriteString("\n")
	}

	// En-têtes et pieds de page
	for _, hdr := range doc.Headers() {
		for _, para := range hdr.Paragraphs() {
			sb.WriteString(paragraphText(para))
			sb.WriteString("\n")
		}
	}
	for _, ftr := range doc.Footers() {
		for _, para := range ftr.Paragraphs() {
			sb.WriteString(paragraphText(para))
			sb.WriteString("\n")
		}
	}

	// Tableaux
	for _, tbl := range doc.Tables() {
		for _, row := range tbl.Rows() {
			for _, cell := range row.Cells() {
				for _, para := range cell.Paragraphs() {
					sb.WriteString(paragraphText(para))
					sb.WriteString("\t")
				}
			}
			sb.WriteString("\n")
		}
	}

	content := sb.String()
	if len(strings.TrimSpace(content)) == 0 {
		log.Printf("WARNING: Extraction DOCX vide")
	}
	return content, nil
}
