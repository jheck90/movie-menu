package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

// Movie represents a movie with poster
type Movie struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	PosterURL   string `json:"posterUrl"`
	CachedPoster string `json:"cachedPoster,omitempty"`
	Year        int    `json:"year"`
	TMDBID      int    `json:"tmdbId"`
	IMDBID      string `json:"imdbId,omitempty"`
}

// MovieList represents a collection of movies
type MovieList struct {
	Name      string    `json:"name"`
	Movies    []Movie   `json:"movies"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Config holds API configuration
type Config struct {
	RadarrURL    string `json:"radarrUrl"`
	RadarrAPIKey string `json:"radarrApiKey"`
	TVDBAPIKey   string `json:"tvdbApiKey"`
}

// RadarrMovie represents movie data from Radarr
type RadarrMovie struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Year   int    `json:"year"`
	TMDBID int    `json:"tmdbId"`
	IMDBID string `json:"imdbId"`
	Images []struct {
		CoverType string `json:"coverType"`
		URL       string `json:"url"`
		RemoteURL string `json:"remoteUrl"`
	} `json:"images"`
}

// MovieMenu is the main app component
type MovieMenu struct {
	app.Compo
	lists          []MovieList
	selectedList   int
	config         Config
	availableMovies []Movie
	showCreateList bool
	showEditList   bool
	newListName    string
	selectedMovies map[int]bool
	loading        bool
	error          string
}

func (m *MovieMenu) OnMount(ctx app.Context) {
	m.selectedMovies = make(map[int]bool)
	m.selectedList = 0
	m.loadConfig()
	m.loadLists()
	m.fetchAvailableMovies()
}

func (m *MovieMenu) loadConfig() {
	data, err := ioutil.ReadFile("config/config.json")
	if err != nil {
		log.Printf("Error loading config: %v", err)
		m.error = "Configuration not found. Please set up config.json"
		return
	}
	json.Unmarshal(data, &m.config)
}

func (m *MovieMenu) loadLists() {
	m.lists = []MovieList{}
	files, err := ioutil.ReadDir("lists")
	if err != nil {
		log.Printf("Error reading lists directory: %v", err)
		return
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			data, err := ioutil.ReadFile(filepath.Join("lists", file.Name()))
			if err != nil {
				continue
			}
			var list MovieList
			if err := json.Unmarshal(data, &list); err == nil {
				m.lists = append(m.lists, list)
			}
		}
	}
}

func (m *MovieMenu) fetchAvailableMovies() {
	m.loading = true
	m.Update()
	
	go func() {
		client := &http.Client{Timeout: 30 * time.Second}
		req, _ := http.NewRequest("GET", m.config.RadarrURL+"/api/v3/movie", nil)
		req.Header.Add("X-Api-Key", m.config.RadarrAPIKey)
		
		resp, err := client.Do(req)
		if err != nil {
			m.error = fmt.Sprintf("Failed to connect to Radarr: %v", err)
			m.loading = false
			m.Update()
			return
		}
		defer resp.Body.Close()
		
		var radarrMovies []RadarrMovie
		if err := json.NewDecoder(resp.Body).Decode(&radarrMovies); err != nil {
			m.error = "Failed to parse Radarr response"
			m.loading = false
			m.Update()
			return
		}
		
		m.availableMovies = []Movie{}
		for _, rm := range radarrMovies {
			movie := Movie{
				ID:     rm.ID,
				Title:  rm.Title,
				Year:   rm.Year,
				TMDBID: rm.TMDBID,
				IMDBID: rm.IMDBID,
			}
			
			// Find poster URL
			for _, img := range rm.Images {
				if img.CoverType == "poster" {
					if img.RemoteURL != "" {
						movie.PosterURL = img.RemoteURL
					} else {
						movie.PosterURL = img.URL
					}
					break
				}
			}
			
			m.availableMovies = append(m.availableMovies, movie)
		}
		
		m.loading = false
		m.error = ""
		m.Update()
	}()
}

func (m *MovieMenu) Render() app.UI {
	// Ensure selectedList is valid
	if m.selectedList < 0 || m.selectedList >= len(m.lists) {
		m.selectedList = 0
	}
	
	return app.Div().Class("min-h-screen bg-gray-900 text-white p-4").Body(
		app.Div().Class("max-w-7xl mx-auto").Body(
			// Header
			app.H1().Class("text-4xl font-bold mb-8 text-center").Text("Movie Menu"),
			
			// Error display
			app.If(m.error != "",
				app.Div().Class("bg-red-600 text-white p-4 rounded-lg mb-4").Text(m.error),
			),
			
			// Controls
			app.Div().Class("mb-8 flex justify-between items-center").Body(
				// List selector
				app.If(len(m.lists) > 0,
					app.Select().
						Class("bg-gray-800 text-white px-4 py-2 rounded-lg text-lg").
						OnChange(m.onListChange).
						Body(
							app.Range(m.lists).Slice(func(i int) app.UI {
								return app.Option().
									Value(i).
									Text(m.lists[i].Name).
									Selected(i == m.selectedList)
							}),
						),
				).Else(
					app.Div().Class("text-gray-400").Text("No lists yet"),
				),
				
				// Action buttons
				app.Div().Class("flex gap-2").Body(
					app.Button().
						Class("bg-blue-600 hover:bg-blue-700 px-4 py-2 rounded-lg").
						Text("Create New List").
						OnClick(m.onCreateListClick),
					app.If(len(m.lists) > 0 && m.selectedList >= 0 && m.selectedList < len(m.lists),
						app.Button().
							Class("bg-green-600 hover:bg-green-700 px-4 py-2 rounded-lg").
							Text("Edit List").
							OnClick(m.onEditListClick),
					),
				),
			),
			
			// Create/Edit List Modal
			app.If(m.showCreateList || m.showEditList,
				m.renderListModal(),
			),
			
			// Movie grid
			app.If(!m.showCreateList && !m.showEditList,
				app.Div().Class("grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4").Body(
					app.If(len(m.lists) > 0 && m.selectedList >= 0 && m.selectedList < len(m.lists),
						app.Range(m.lists[m.selectedList].Movies).Slice(func(i int) app.UI {
							movie := m.lists[m.selectedList].Movies[i]
							return m.renderMovieCard(movie)
						}),
					).Else(
						app.Div().Class("col-span-full text-center text-gray-400").
							Text("No movies to display. Create a list to get started."),
					),
				),
			),
		),
	)
}

func (m *MovieMenu) renderListModal() app.UI {
	isEdit := m.showEditList && len(m.lists) > 0 && m.selectedList >= 0 && m.selectedList < len(m.lists)
	title := "Create New List"
	if isEdit {
		title = "Edit List"
		m.newListName = m.lists[m.selectedList].Name
	}
	
	return app.Div().Class("fixed inset-0 bg-black bg-opacity-75 flex items-center justify-center p-4 z-50").Body(
		app.Div().Class("bg-gray-800 rounded-lg max-w-6xl w-full max-h-[90vh] overflow-hidden flex flex-col").Body(
			// Modal header
			app.Div().Class("p-6 border-b border-gray-700").Body(
				app.H2().Class("text-2xl font-bold").Text(title),
				app.Input().
					Type("text").
					Class("w-full mt-4 px-4 py-2 bg-gray-700 rounded-lg").
					Placeholder("List name...").
					Value(m.newListName).
					OnChange(m.onListNameChange),
			),
			
			// Movie selection area
			app.Div().Class("flex-1 overflow-y-auto p-6").Body(
				app.If(m.loading,
					app.Div().Class("text-center py-8").Text("Loading movies from Radarr..."),
				).Else(
					app.Div().Class("grid grid-cols-3 sm:grid-cols-4 md:grid-cols-5 lg:grid-cols-6 gap-3").Body(
						app.Range(m.availableMovies).Slice(func(i int) app.UI {
							movie := m.availableMovies[i]
							return m.renderSelectableMovie(movie)
						}),
					),
				),
			),
			
			// Modal footer
			app.Div().Class("p-6 border-t border-gray-700 flex justify-end gap-2").Body(
				app.Button().
					Class("px-6 py-2 bg-gray-600 hover:bg-gray-700 rounded-lg").
					Text("Cancel").
					OnClick(m.onCancelModal),
				app.Button().
					Class("px-6 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg").
					Text("Save List").
					OnClick(m.onSaveList).
					Disabled(m.newListName == "" || len(m.selectedMovies) == 0),
			),
		),
	)
}

func (m *MovieMenu) renderSelectableMovie(movie Movie) app.UI {
	selected := m.selectedMovies[movie.ID]
	
	return app.Div().
		Class("relative cursor-pointer transform transition hover:scale-105").
		OnClick(func(ctx app.Context, e app.Event) {
			m.toggleMovieSelection(movie.ID)
		}).
		Body(
			app.Img().
				Class("w-full h-auto rounded-lg").
				Src(m.getPosterURL(movie)).
				Alt(movie.Title),
			app.If(selected,
				app.Div().Class("absolute inset-0 bg-blue-600 bg-opacity-50 rounded-lg flex items-center justify-center").Body(
					app.Div().Class("bg-white text-blue-600 rounded-full p-2").Body(
						app.Text("✓"),
					),
				),
			),
			app.Div().Class("mt-1 text-xs text-center truncate").Text(movie.Title),
		)
}

func (m *MovieMenu) renderMovieCard(movie Movie) app.UI {
	return app.Div().
		Class("bg-gray-800 rounded-lg overflow-hidden cursor-pointer transform transition hover:scale-105").
		OnClick(func(ctx app.Context, e app.Event) {
			m.onMovieClick(movie)
		}).
		Body(
			app.Img().
				Class("w-full h-auto").
				Src(m.getPosterURL(movie)).
				Alt(movie.Title),
			app.Div().Class("p-2").Body(
				app.H3().Class("text-sm font-semibold truncate").Text(movie.Title),
				app.If(movie.Year > 0,
					app.P().Class("text-xs text-gray-400").Text(fmt.Sprintf("(%d)", movie.Year)),
				),
			),
		)
}

func (m *MovieMenu) getPosterURL(movie Movie) string {
	if movie.CachedPoster != "" {
		return "/cache/" + movie.CachedPoster
	}
	if movie.PosterURL != "" {
		return movie.PosterURL
	}
	return "/web/placeholder.jpg"
}

func (m *MovieMenu) toggleMovieSelection(movieID int) {
	if m.selectedMovies[movieID] {
		delete(m.selectedMovies, movieID)
	} else {
		m.selectedMovies[movieID] = true
	}
	m.Update()
}

func (m *MovieMenu) onListChange(ctx app.Context, e app.Event) {
	m.selectedList = ctx.JSSrc().Get("selectedIndex").Int()
}

func (m *MovieMenu) onListNameChange(ctx app.Context, e app.Event) {
	m.newListName = ctx.JSSrc().Get("value").String()
}

func (m *MovieMenu) onCreateListClick(ctx app.Context, e app.Event) {
	m.showCreateList = true
	m.showEditList = false
	m.newListName = ""
	m.selectedMovies = make(map[int]bool)
	m.Update()
}

func (m *MovieMenu) onEditListClick(ctx app.Context, e app.Event) {
	if m.selectedList >= len(m.lists) || len(m.lists) == 0 {
		return
	}
	
	m.showEditList = true
	m.showCreateList = false
	m.newListName = m.lists[m.selectedList].Name
	m.selectedMovies = make(map[int]bool)
	
	// Pre-select existing movies
	for _, movie := range m.lists[m.selectedList].Movies {
		m.selectedMovies[movie.ID] = true
	}
	m.Update()
}

func (m *MovieMenu) onCancelModal(ctx app.Context, e app.Event) {
	m.showCreateList = false
	m.showEditList = false
	m.newListName = ""
	m.selectedMovies = make(map[int]bool)
	m.Update()
}

func (m *MovieMenu) onSaveList(ctx app.Context, e app.Event) {
	selectedMovies := []Movie{}
	for _, movie := range m.availableMovies {
		if m.selectedMovies[movie.ID] {
			selectedMovies = append(selectedMovies, movie)
		}
	}
	
	// Download and cache posters
	go m.cachePosters(selectedMovies)
	
	newList := MovieList{
		Name:      m.newListName,
		Movies:    selectedMovies,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	// Save to file
	filename := fmt.Sprintf("lists/%s.json", sanitizeFilename(m.newListName))
	data, _ := json.MarshalIndent(newList, "", "  ")
	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		m.error = "Failed to save list"
		m.Update()
		return
	}
	
	// Reload lists
	m.loadLists()
	m.showCreateList = false
	m.showEditList = false
	m.newListName = ""
	m.selectedMovies = make(map[int]bool)
	m.Update()
}

func (m *MovieMenu) cachePosters(movies []Movie) {
	for i, movie := range movies {
		if movie.PosterURL == "" {
			continue
		}
		
		// Generate cache filename
		hash := md5.Sum([]byte(movie.PosterURL))
		cacheFile := fmt.Sprintf("%x.jpg", hash)
		cachePath := filepath.Join("cache", cacheFile)
		
		// Check if already cached
		if _, err := os.Stat(cachePath); err == nil {
			movies[i].CachedPoster = cacheFile
			continue
		}
		
		// Download poster
		resp, err := http.Get(movie.PosterURL)
		if err != nil {
			continue
		}
		defer resp.Body.Close()
		
		// Save to cache
		out, err := os.Create(cachePath)
		if err != nil {
			continue
		}
		defer out.Close()
		
		io.Copy(out, resp.Body)
		movies[i].CachedPoster = cacheFile
	}
}

func (m *MovieMenu) onMovieClick(movie Movie) {
	// For now, just log. You can add play functionality later
	log.Printf("Selected movie: %s", movie.Title)
}

func sanitizeFilename(name string) string {
	// Remove invalid characters
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ToLower(name)
	// Keep only alphanumeric and hyphens
	result := ""
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result += string(r)
		}
	}
	return result
}

func main() {
	// Create directories if they don't exist
	os.MkdirAll("config", 0755)
	os.MkdirAll("lists", 0755)
	os.MkdirAll("cache", 0755)
	os.MkdirAll("web", 0755)
	
	// Create sample config if it doesn't exist
	if _, err := os.Stat("config/config.json"); os.IsNotExist(err) {
		sampleConfig := Config{
			RadarrURL:    "http://localhost:7878",
			RadarrAPIKey: "your-radarr-api-key",
			TVDBAPIKey:   "your-tvdb-api-key",
		}
		data, _ := json.MarshalIndent(sampleConfig, "", "  ")
		ioutil.WriteFile("config/config.json", data, 0644)
	}
	
	// Register the component
	app.Route("/", &MovieMenu{})
	
	// The app handler
	http.Handle("/", &app.Handler{
		Name:        "Movie Menu",
		Description: "A simple movie selection menu",
		Styles: []string{
			"https://cdn.jsdelivr.net/npm/tailwindcss@2.2.19/dist/tailwind.min.css",
		},
		Icon: app.Icon{
			Default: "/web/icon-192.png",
		},
	})
	
	// Static files
	http.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("web"))))
	http.Handle("/cache/", http.StripPrefix("/cache/", http.FileServer(http.Dir("cache"))))
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}