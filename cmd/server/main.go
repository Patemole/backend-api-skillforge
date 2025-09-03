package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"backend-api-skillforge/internal/handlers"
	"backend-api-skillforge/internal/middleware"
	"backend-api-skillforge/internal/supabase"
)

func main() {
	_ = godotenv.Load() // charge .env (facultatif en prod)
	supabase.MustInit() // client global

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())

	// ✅ Middleware CORS avancé
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // En prod, remplace par ton domaine front
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "apikey", "x-client-info"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ✅ Gère les requêtes OPTIONS (nécessaire pour le preflight)
	r.OPTIONS("/extract", func(c *gin.Context) {
		c.Status(200)
	})
	r.OPTIONS("/jobs", func(c *gin.Context) {
		c.Status(200)
	})
	r.OPTIONS("/api/email/generate-presentation", func(c *gin.Context) {
		c.Status(200)
	})

	r.GET("/health", handlers.Health)
	r.POST("/extract", handlers.ExtractCV)
	r.POST("/jobs", handlers.CreateJob)
	r.GET("/jobs/:id/status", handlers.GetJobStatus)
	r.POST("/api/email/generate-presentation", handlers.GeneratePresentationEmail)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("⇨ listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
