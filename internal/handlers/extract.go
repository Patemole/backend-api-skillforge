package handlers

import (
	"io"
	"log"
	"net/http"

	"backend-api-skillforge/internal/nuextract"

	"github.com/gin-gonic/gin"
)

// ExtractCV traite l'upload d'un CV et renvoie le JSON NuExtract.
func ExtractCV(c *gin.Context) {
	log.Printf("DEBUG: Début de l'extraction CV")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		log.Printf("ERROR: Erreur récupération fichier: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "file not provided"})
		return
	}
	defer file.Close()

	log.Printf("DEBUG: Fichier reçu - Nom: %s, Taille: %d", header.Filename, header.Size)

	data, err := io.ReadAll(file)
	if err != nil {
		log.Printf("ERROR: Erreur lecture fichier: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot read file"})
		return
	}

	log.Printf("DEBUG: Données lues - Taille: %d bytes", len(data))

	client := nuextract.New()
	log.Printf("DEBUG: Client NuExtract créé, début de l'extraction...")

	result, err := client.ExtractAndEnrichWithFilename(data, header.Filename)
	if err != nil {
		log.Printf("ERROR: Erreur extraction NuExtract: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	log.Printf("DEBUG: Extraction réussie, taille résultat: %d bytes", len(result))
	// NuExtract retourne déjà du JSON → on le transmet tel quel.
	c.Data(http.StatusOK, "application/json", result)
}
