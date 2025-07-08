package main

import (
	// "os"
	// "path/filepath"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
	"os"
)

const (
	configPath = "./config/config.json"
	// listsDir   = "./lists"
)

type AppConfig struct {
	RadarrAPIKey string `json:"radarrApiKey"`
	RadarrURL    string `json:"radarrUrl"`
	TVDBAPIKey   string `json:"tvdbApiKey"`
}

// Movie represents a single movie entry
type Movie struct {
	Title     string  `json:"title"`
	PosterURL string  `json:"poster_url"`
	RadarrID  int     `json:"radarr_id"`
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

var (
	tvdbToken     string
	tvdbTokenMu   sync.Mutex
	tvdbTokenExp  time.Time
)

// Structs for JSON parsing
type loginRequest struct {
	ApiKey string `json:"apikey"`
	Pin    string `json:"pin,omitempty"`
}

type loginResponse struct {
	Data struct {
		Token string `json:"token"`
	} `json:"data"`
	Status string `json:"status"`
}

func getTVDBToken() (string, error) {
	tvdbTokenMu.Lock()
	defer tvdbTokenMu.Unlock()

	// Return cached token if valid
	if tvdbToken != "" && time.Now().Before(tvdbTokenExp) {
		return tvdbToken, nil
	}

	// Prepare login payload
	payload := loginRequest{
		ApiKey: Config.TVDBAPIKey,
		// Pin: Config.TVDBPin, // Optional: add if you have a PIN
	}
	body, _ := json.Marshal(payload)
	// fmt.Println("TVDB Login Payload:", string(body))

	req, err := http.NewRequest("POST", "https://api4.thetvdb.com/v4/login", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := ioutil.ReadAll(resp.Body)
		return "", errors.New(fmt.Sprintf("TVDB login failed: %s - %s", resp.Status, string(b)))
	}

	var lr loginResponse
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return "", err
	}

	tvdbToken = lr.Data.Token
	tvdbTokenExp = time.Now().Add(28 * 24 * time.Hour) // valid ~1 month

	return tvdbToken, nil
}

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
	// fmt.Println("TVDB API Key from config:", Config.TVDBAPIKey)

	return nil

}


// LoadMovieLists loads all movie lists from /root/lists
// func LoadMovieLists() error {
// 	MovieMenus = make(map[string]MovieList)

// 	files, err := filepath.Glob(filepath.Join(listsDir, "*.json"))
// 	if err != nil {
// 		return fmt.Errorf("failed to list json files: %w", err)
// 	}

// 	for _, file := range files {
// 		data, err := ioutil.ReadFile(file)
// 		if err != nil {
// 			fmt.Printf("Skipping file %s: %v\n", file, err)
// 			continue
// 		}

// 		var list MovieList
// 		if err := json.Unmarshal(data, &list); err != nil {
// 			fmt.Printf("Skipping invalid JSON in %s: %v\n", file, err)
// 			continue
// 		}

// 		name := filepath.Base(file)
// 		name = name[:len(name)-len(filepath.Ext(name))] // remove .json
// 		MovieMenus[name] = list
// 	}

// 	return nil
// }

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
	token, err := getTVDBToken()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("GET", "https://api4.thetvdb.com/v4/search?query=encanto&type=series", nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
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

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("TVDB search response:", string(body))

	return nil
}

