package main

import (
	"log"
	"os"

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

	r.GET("/health", handlers.Health)
	r.POST("/extract", handlers.ExtractCV)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("⇨ listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
