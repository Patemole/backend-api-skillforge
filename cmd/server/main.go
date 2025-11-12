package main

import (
	"io"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"backend-api-skillforge/internal/handlers"
	"backend-api-skillforge/internal/middleware"
	"backend-api-skillforge/internal/supabase"
	"backend-api-skillforge/internal/worker"
)

func main() {
	_ = godotenv.Load() // charge .env (facultatif en prod)

	// Vérification des variables Apify au démarrage
	apifyToken := os.Getenv("APIFY_TOKEN")
	if apifyToken == "" {
		apifyToken = os.Getenv("APIFY_API_TOKEN")
	}
	apifyActorID := os.Getenv("APIFY_ACTOR_ID")

	if apifyToken != "" && apifyActorID != "" {
		log.Printf("✅ Apify configuré - Actor ID: '%s', Token: %s... (len=%d)",
			apifyActorID,
			func() string {
				if len(apifyToken) > 10 {
					return apifyToken[:10] + "..."
				}
				return "***"
			}(),
			len(apifyToken))
	} else {
		log.Printf("⚠️  Apify non configuré (APIFY_TOKEN: %t, APIFY_ACTOR_ID: %t)",
			apifyToken != "",
			apifyActorID != "")
		if apifyActorID != "" {
			log.Printf("   Actor ID trouvé: '%s'", apifyActorID)
		}
	}

	supabase.MustInit() // client global

	// 🔧 Logger → stdout + fichier server.log
	logFilePath := "server.log"
	if f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		mw := io.MultiWriter(os.Stdout, f)
		log.SetOutput(mw)
		log.SetFlags(log.LstdFlags | log.Lmicroseconds)
		log.Printf("📝 Logging initialisé → %s + stdout", logFilePath)
	} else {
		log.Printf("⚠️  Impossible d'ouvrir %s: %v (fallback stdout uniquement)", logFilePath, err)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())

	// 🚀 Lancer PLUSIEURS workers pour traiter en parallèle
	// Avec 50 workers: on peut traiter jusqu'à 50 CV en même temps (~1000 CV/minute)
	workerCount := 50
	for i := 1; i <= workerCount; i++ {
		worker.StartExtractCVWorker()
	}
	log.Printf("✅ Tous les workers ont démarré (%d workers extract_cv en parallèle)", workerCount)

	// ✅ Configuration CORS robuste - fonctionne toujours
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true, // Accepte toutes les origines
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD", "PATCH"},
		AllowHeaders:     []string{"*"},  // Accepte tous les headers
		ExposeHeaders:    []string{"*"},  // Expose tous les headers
		AllowCredentials: false,          // Pas de credentials = plus simple
		MaxAge:           24 * time.Hour, // Cache plus long
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
	r.OPTIONS("/generate-pdf", func(c *gin.Context) {
		c.Status(200)
	})
	r.OPTIONS("/billing/checkout-session", func(c *gin.Context) {
		c.Status(200)
	})
	r.OPTIONS("/billing/webhook", func(c *gin.Context) {
		c.Status(200)
	})

	// ✅ Initialiser les handlers
	candidateValidationHandler := handlers.NewCandidateValidationHandler()
	candidateInviteHandler := handlers.NewCandidateInviteHandler()

	r.GET("/health", handlers.Health)
	// Basculer /extract en asynchrone (création d'un job extract_cv)
	r.POST("/extract", handlers.ExtractCVAsync)
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
	r.POST("/generate-pdf", handlers.GeneratePDF)
	r.POST("/billing/checkout-session", handlers.CreateCheckoutSessionHandler)
	r.POST("/billing/webhook", handlers.StripeWebhook)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("⇨ listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
