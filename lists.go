package main

import (
	"encoding/json"
	"strings"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"fmt"
)

// // Movie represents a basic movie info you want to show
// type Movie struct {
// 	Title  string  `json:"title"`
// 	Rating float64 `json:"rating,omitempty"`
// 	Poster string  `json:"poster,omitempty"`
// }

// // MovieList represents a named list of movies
// type MovieList struct {
// 	Name   string  `json:"name"`
// 	Movies []Movie `json:"movies"`
// }

var (
	listsDir = "./lists"
	listsMux sync.Mutex
)

// SaveMovieList saves a MovieList to disk as JSON
func SaveMovieList(list MovieList) error {
	listsMux.Lock()
	defer listsMux.Unlock()

	if err := os.MkdirAll(listsDir, 0755); err != nil {
		return err
	}

	filePath := filepath.Join(listsDir, list.Title+".json")
	bytes, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filePath, bytes, 0644)
}

// LoadMovieList loads a MovieList by name
func LoadMovieList(name string) (MovieList, error) {
	listsMux.Lock()
	defer listsMux.Unlock()

	filePath := filepath.Join(listsDir, name+".json")
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return MovieList{}, err
	}

	var list MovieList
	if err := json.Unmarshal(bytes, &list); err != nil {
		return MovieList{}, err
	}

	return list, nil
}

// ListMovieLists returns names of all saved lists (without .json)
func ListMovieLists() ([]string, error) {
	listsMux.Lock()
	defer listsMux.Unlock()

	files, err := ioutil.ReadDir(listsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil // no lists yet
		}
		return nil, err
	}

	var names []string
	for _, f := range files {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".json" {
			name := f.Name()
			names = append(names, name[:len(name)-len(".json")])
		}
	}

	return names, nil
}

// HTTP Handlers below — register these in main.go with http.HandleFunc
func sanitizeFilename(name string) string {
	// Remove everything except letters, numbers, dash, underscore, and space
	name = strings.ToLower(name)
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, " ", "_")
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') ||
			(r >= '0' && r <= '9') ||
			r == '_' || r == '-' {
			return r
		}
		return -1
	}, name)
}


// CreateListHandler accepts POST JSON {"name": "...", "movies": [...]}
func CreateListHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Name   string  `json:"name"`
		Movies []Movie `json:"movies"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if payload.Name == "" || len(payload.Movies) == 0 {
		http.Error(w, "Missing list name or movies", http.StatusBadRequest)
		return
	}

	filename := fmt.Sprintf("./lists/%s.json", sanitizeFilename(payload.Name))
	file, err := os.Create(filename)
	if err != nil {
		http.Error(w, "Failed to save list: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(payload.Movies); err != nil {
		http.Error(w, "Failed to write list: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"saved"}`))
	fmt.Printf("Saving list: %s with %d movies\n", payload.Name, len(payload.Movies))

}


// GetListHandler handles GET /api/lists/get?name=listname
func GetListHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, `{"error":"Missing list name"}`, http.StatusBadRequest)
		return
	}

	path := filepath.Join("./lists", sanitizeFilename(name)+".json")
	file, err := os.Open(path)
	if err != nil {
		http.Error(w, `{"error":"List not found or failed to load"}`, http.StatusNotFound)
		return
	}
	defer file.Close()

	var movies []Movie
	if err := json.NewDecoder(file).Decode(&movies); err != nil {
		http.Error(w, `{"error":"Failed to parse list file"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movies)
}


// ListListsHandler returns all saved list names (GET /api/lists)
func ListListsHandler(w http.ResponseWriter, r *http.Request) {
	names, err := ListMovieLists()
	if err != nil {
		http.Error(w, "Failed to list lists: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(names)
}
