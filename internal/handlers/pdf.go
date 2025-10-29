package handlers

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

// GeneratePDF reçoit du HTML et renvoie un PDF généré via WeasyPrint
func GeneratePDF(c *gin.Context) {
	contentType := c.GetHeader("Content-Type")

	var htmlBytes []byte
	if strings.Contains(contentType, "application/json") {
		var payload struct {
			HTML string `json:"html"`
		}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Payload JSON invalide: " + err.Error()})
			return
		}
		htmlBytes = []byte(payload.HTML)
	} else {
		data, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Impossible de lire le corps de la requête: " + err.Error()})
			return
		}
		defer c.Request.Body.Close()
		htmlBytes = data
	}

	if len(bytes.TrimSpace(htmlBytes)) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "HTML manquant ou vide"})
		return
	}

	// Vérifier que WeasyPrint est installé
	if _, err := exec.LookPath("weasyprint"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "WeasyPrint introuvable sur le serveur",
			"details": err.Error(),
			"help":    "Installez WeasyPrint avec: pip install weasyprint",
		})
		return
	}

	// Commande WeasyPrint : lit depuis stdin ("-") et écrit sur stdout ("-")
	cmd := exec.Command("weasyprint", "-", "-")
	cmd.Stdin = bytes.NewReader(htmlBytes)
	// Passer DYLD_LIBRARY_PATH pour macOS (nécessaire pour trouver les bibliothèques Pango/Cairo)
	if libPath := os.Getenv("DYLD_LIBRARY_PATH"); libPath != "" {
		cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+libPath)
	} else {
		// Valeur par défaut pour Homebrew sur macOS
		cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH=/opt/homebrew/lib")
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Printf("weasyprint error: %v | stderr: %s", err, stderr.String())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Erreur génération PDF",
			"details": stderr.String(),
		})
		return
	}

	filename := c.DefaultQuery("filename", "document.pdf")
	if !strings.HasSuffix(strings.ToLower(filename), ".pdf") {
		filename = filename + ".pdf"
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%q", filename))
	c.Data(http.StatusOK, "application/pdf", stdout.Bytes())
}
