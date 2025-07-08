package main

import (
	"log"
	"net/http"
	"fmt"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, proceeding with environment variables or config.json")
	}

	if err := LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := LoadMovieLists(); err != nil {
		log.Fatalf("Failed to load movie lists: %v", err)
	}

	app.Route("/", &MenuPage{})
	app.RunWhenOnBrowser()

	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/web/", http.StripPrefix("/web/", fs))

	http.Handle("/", &app.Handler{
		Name:        "Movie Menu",
		Description: "Toddler-friendly movie selection UI",
		Styles:      []string{"/web/styles.css"},
	})

	http.HandleFunc("/app.wasm", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/wasm")
		http.ServeFile(w, r, "./app.wasm")
	})

	// Correctly separate handlers — not nested!
	http.HandleFunc("/api/test-radarr", func(w http.ResponseWriter, r *http.Request) {
		err := TestRadarrAPI()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"error":"%v"}`, err)))
			return
		}
		w.Write([]byte(`{"status":"Radarr API reachable"}`))
	})

	http.HandleFunc("/api/test-tvdb", func(w http.ResponseWriter, r *http.Request) {
		err := TestTVDBAPI()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"error":"%v"}`, err)))
			return
		}
		w.Write([]byte(`{"status":"TVDB API reachable"}`))
	})

	log.Println("Serving on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
