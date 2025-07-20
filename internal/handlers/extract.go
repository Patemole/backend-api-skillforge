package handlers

import (
	"io"
	"net/http"

	"backend-api-skillforge/internal/nuextract"

	"github.com/gin-gonic/gin"
)

// ExtractCV traite l’upload d’un CV et renvoie le JSON NuExtract.
func ExtractCV(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file not provided"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot read file"})
		return
	}

	client := nuextract.New()
	result, err := client.Extract(data)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	// NuExtract retourne déjà du JSON → on le transmet tel quel.
	c.Data(http.StatusOK, "application/json", result)
}
