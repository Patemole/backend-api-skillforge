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
	r.OPTIONS("/candidate-validation", func(c *gin.Context) {
		c.Status(200)
	})
	r.OPTIONS("/candidate-invite", func(c *gin.Context) {
		c.Status(200)
	})
	r.OPTIONS("/api/templates/generate", func(c *gin.Context) {
		c.Status(200)
	})
	r.OPTIONS("/api/email/generate-presentation-v2", func(c *gin.Context) {
		c.Status(200)
	})
	r.OPTIONS("/meeting/transcript", func(c *gin.Context) {
		c.Status(200)
	})
	r.OPTIONS("/api/dossier/versionning", func(c *gin.Context) {
		c.Status(200)
	})
	r.OPTIONS("/versionning", func(c *gin.Context) {
		c.Status(200)
	})
	r.OPTIONS("/boond/candidat/delete", func(c *gin.Context) {
		c.Status(200)
	})
	r.OPTIONS("/boond/candidat/modify", func(c *gin.Context) {
		c.Status(200)
	})
	r.OPTIONS("/boond/candidat/dc", func(c *gin.Context) {
		c.Status(200)
	})
	r.OPTIONS("/boond/agencies", func(c *gin.Context) {
		c.Status(200)
	})
	r.OPTIONS("/boond/resources", func(c *gin.Context) {
		c.Status(200)
	})
	r.OPTIONS("/boond/orgchart", func(c *gin.Context) {
		c.Status(200)
	})

	// ✅ Initialiser les handlers
	candidateValidationHandler := handlers.NewCandidateValidationHandler()
	candidateInviteHandler := handlers.NewCandidateInviteHandler()

	r.GET("/health", handlers.Health)
	r.POST("/extract", handlers.ExtractCV)
	r.POST("/jobs", handlers.CreateJob)
	r.GET("/jobs/:id/status", handlers.GetJobStatus)
	r.POST("/api/email/generate-presentation", handlers.GeneratePresentationEmail)
	r.POST("/api/email/generate-presentation-v2", handlers.GeneratePresentationEmailV2)
	r.POST("/api/templates/generate", handlers.GenerateTemplate)
	r.POST("/api/dossier/versionning", handlers.CreateDossierVersion)
	r.POST("/versionning", handlers.CreateDossierVersion)
	r.POST("/candidate-validation", candidateValidationHandler.HandleCandidateValidation)
	r.POST("/candidate-invite", candidateInviteHandler.HandleCandidateInvite)
	r.POST("/boond/candidat/delete", handlers.DeleteBoondCandidate)
	r.POST("/boond/candidat/modify", handlers.ModifyBoondCandidate)
	r.POST("/boond/candidat/dc", handlers.UploadBoondCandidateDC)
	r.POST("/meeting/transcript", handlers.MeetingTranscript)
	r.POST("/boond/agencies", handlers.GetBoondAgencies)
	r.POST("/boond/resources", handlers.GetBoondResources)
	r.POST("/boond/orgchart", handlers.BuildBoondOrgChart)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("⇨ listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
