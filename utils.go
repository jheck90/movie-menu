package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// RadarrMovie represents a movie from Radarr API
type RadarrMovie struct {
	Title  string `json:"title"`
	Year   int    `json:"year"`
	TMDBID int    `json:"tmdbId"`
	Images []struct {
		CoverType string `json:"coverType"`
		URL       string `json:"url"`
	} `json:"images"`
}

// FetchMoviesFromRadarr fetches movies from Radarr and creates a list
func FetchMoviesFromRadarr(configPath, outputPath, listName string) error {
	// Load config
	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("error reading config: %v", err)
	}
	
	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		return fmt.Errorf("error parsing config: %v", err)
	}
	
	// Fetch from Radarr
	client := &http.Client{}
	req, err := http.NewRequest("GET", config.RadarrURL+"/api/v3/movie", nil)
	if err != nil {
		return err
	}
	req.Header.Add("X-Api-Key", config.RadarrAPIKey)
	
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error fetching from Radarr: %v", err)
	}
	defer resp.Body.Close()
	
	var radarrMovies []RadarrMovie
	if err := json.NewDecoder(resp.Body).Decode(&radarrMovies); err != nil {
		return fmt.Errorf("error decoding response: %v", err)
	}
	
	// Convert to our format
	movieList := MovieList{
		Name:   listName,
		Movies: make([]Movie, 0),
	}
	
	for _, rm := range radarrMovies {
		movie := Movie{
			ID:    rm.TMDBID,
			Title: rm.Title,
			Year:  rm.Year,
		}
		
		// Find poster URL
		for _, img := range rm.Images {
			if img.CoverType == "poster" {
				movie.PosterURL = img.URL
				break
			}
		}
		
		movieList.Movies = append(movieList.Movies, movie)
	}
	
	// Save to file
	data, err := json.MarshalIndent(movieList, "", "  ")
	if err != nil {
		return err
	}
	
	return ioutil.WriteFile(outputPath, data, 0644)
}

// FilterMoviesByTag filters movies suitable for kids based on common patterns
func FilterMoviesByTag(inputPath, outputPath string, keywords []string) error {
	data, err := ioutil.ReadFile(inputPath)
	if err != nil {
		return err
	}
	
	var list MovieList
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}
	
	filtered := MovieList{
		Name:   list.Name + " (Filtered)",
		Movies: make([]Movie, 0),
	}
	
	for _, movie := range list.Movies {
		include := false
		titleLower := strings.ToLower(movie.Title)
		
		for _, keyword := range keywords {
			if strings.Contains(titleLower, strings.ToLower(keyword)) {
				include = true
				break
			}
		}
		
		if include {
			filtered.Movies = append(filtered.Movies, movie)
		}
	}
	
	data, err = json.MarshalIndent(filtered, "", "  ")
	if err != nil {
		return err
	}
	
	return ioutil.WriteFile(outputPath, data, 0644)
}

// Command line utility
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  fetch-radarr <config.json> <output.json> <list-name>")
		fmt.Println("  filter <input.json> <output.json> <keyword1> <keyword2> ...")
		return
	}
	
	command := os.Args[1]
	
	switch command {
	case "fetch-radarr":
		if len(os.Args) < 5 {
			fmt.Println("Usage: fetch-radarr <config.json> <output.json> <list-name>")
			return
		}
		err := FetchMoviesFromRadarr(os.Args[2], os.Args[3], os.Args[4])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Movies fetched successfully!")
		
	case "filter":
		if len(os.Args) < 4 {
			fmt.Println("Usage: filter <input.json> <output.json> <keyword1> <keyword2> ...")
			return
		}
		keywords := os.Args[4:]
		err := FilterMoviesByTag(os.Args[2], os.Args[3], keywords)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Filtered movies with keywords: %v\n", keywords)
		
	default:
		fmt.Printf("Unknown command: %s\n", command)
	}
}