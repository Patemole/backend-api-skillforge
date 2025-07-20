package supabase

import (
	"log"
	"os"

	supabase "github.com/supabase-community/supabase-go"
)

var Client *supabase.Client

// MustInit crée le client et panique si les variables manquent.
func MustInit() {
	url := os.Getenv("SUPABASE_URL")
	key := os.Getenv("SUPABASE_KEY")
	if url == "" || key == "" {
		log.Fatal("SUPABASE_URL et SUPABASE_KEY doivent être définis")
	}
	var err error
	Client, err = supabase.NewClient(url, key, nil)
	if err != nil {
		log.Fatalf("Erreur d'initialisation Supabase: %v", err)
	}
	log.Println("Supabase initialisé")
}
