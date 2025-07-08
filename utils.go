package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"net/http"
)

const (
	configPath = "./config/config.json"
	listsDir   = "./lists"
)

// AppConfig holds API keys and endpoints
type AppConfig struct {
	RadarrAPIKey string `json:"radarr_api_key"`
	RadarrURL    string `json:"radarr_url"`
	TVDBAPIKey   string `json:"tvdb_api_key"`
}

// Movie represents a single movie entry
type Movie struct {
	Title     string `json:"title"`
	PosterURL string `json:"poster_url"`
	RadarrID  int    `json:"radarr_id"`
}

// MovieList represents a collection of movies
type MovieList struct {
	Title  string  `json:"title"`
	Movies []Movie `json:"movies"`
}

// Global variables for use in app
var (
	Config     AppConfig
	MovieMenus map[string]MovieList
)

// LoadConfig loads the config from file or creates default
func LoadConfig() error {
	data, err := ioutil.ReadFile(configPath)
	if os.IsNotExist(err) {
		// create default config.json if needed
	} else if err != nil {
		return err
	} else {
		if err := json.Unmarshal(data, &Config); err != nil {
			return err
		}
	}

	// Override with env vars if present
	if key := os.Getenv("RADARR_API_KEY"); key != "" {
		Config.RadarrAPIKey = key
	}
	if url := os.Getenv("RADARR_URL"); url != "" {
		Config.RadarrURL = url
	}
	if key := os.Getenv("TVDB_API_KEY"); key != "" {
		Config.TVDBAPIKey = key
	}

	return nil
}


// LoadMovieLists loads all movie lists from /root/lists
func LoadMovieLists() error {
	MovieMenus = make(map[string]MovieList)

	files, err := filepath.Glob(filepath.Join(listsDir, "*.json"))
	if err != nil {
		return fmt.Errorf("failed to list json files: %w", err)
	}

	for _, file := range files {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Printf("Skipping file %s: %v\n", file, err)
			continue
		}

		var list MovieList
		if err := json.Unmarshal(data, &list); err != nil {
			fmt.Printf("Skipping invalid JSON in %s: %v\n", file, err)
			continue
		}

		name := filepath.Base(file)
		name = name[:len(name)-len(filepath.Ext(name))] // remove .json
		MovieMenus[name] = list
	}

	return nil
}

// saveJSON is a helper to write any struct as indented JSON
func saveJSON(path string, v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	_ = ioutil.WriteFile(path, data, 0644)
}

func TestRadarrAPI() error {
    req, err := http.NewRequest("GET", Config.RadarrURL+"/api/v3/movie", nil)
    if err != nil {
        return err
    }
    req.Header.Set("X-Api-Key", Config.RadarrAPIKey)

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    body, _ := ioutil.ReadAll(resp.Body)
    fmt.Println("Radarr response:", string(body))
    return nil
}

func TestTVDBAPI() error {
	req, err := http.NewRequest("GET", "https://api.thetvdb.com/search/series?name=star%20wars", nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+Config.TVDBAPIKey)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("TVDB API error: %s - %s", resp.Status, string(body))
	}

	// Optional: parse response for debugging
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	fmt.Println("TVDB API response sample:", result)
	return nil
}