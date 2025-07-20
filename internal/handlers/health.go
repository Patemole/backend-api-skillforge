package handlers

import "github.com/gin-gonic/gin"

// Health renvoie 200 OK pour les probes Koyeb.
func Health(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}
