package main

import (
	"fmt"
	"log"
	"net/http"
	"github.com/joho/godotenv"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, continuing with environment variables or config.json")
	}

	// Then populate Config from env, for example:
	Config.RadarrAPIKey = os.Getenv("RADARR_API_KEY")
	Config.RadarrURL = os.Getenv("RADARR_URL")
	Config.TVDBAPIKey = os.Getenv("TVDB_API_KEY")

	// Or fallback to loading from config.json
	if Config.RadarrAPIKey == "" || Config.RadarrURL == "" {
		if err := LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
		}
	}

	// err = LoadMovieLists()
	if err != nil {
		log.Printf("Warning: failed to load movie lists: %v", err)
	}
	http.HandleFunc("/api/lists", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			ListListsHandler(w, r)
			return
		}
		if r.Method == http.MethodPost {
			CreateListHandler(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	http.HandleFunc("/api/lists/get", GetListHandler)

	http.HandleFunc("/", serveIndex)
	http.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("./web"))))

	http.HandleFunc("/api/test-radarr", func(w http.ResponseWriter, r *http.Request) {
		err := TestRadarrAPI()
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(`{"status":"Radarr API reachable"}`))
	})
	http.HandleFunc("/api/movies", MoviesHandler)

	http.HandleFunc("/api/test-tvdb", func(w http.ResponseWriter, r *http.Request) {
		err := TestTVDBAPI()
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(`{"status":"TVDB API reachable"}`))
	})

	http.HandleFunc("/api/poster", PosterHandler) // implement PosterHandler using your MovieMenus and API logic

	fmt.Println("Serving on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/index.html")
}
