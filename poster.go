package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GetTVDBPoster queries TVDB's v4 API for a series poster by name
func GetTVDBPoster(query string) (string, error) {
	token, err := getTVDBToken()
	if err != nil {
		return "", err
	}

	endpoint := fmt.Sprintf("https://api4.thetvdb.com/v4/search?query=%s&type=series", url.QueryEscape(query))
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("TVDB API error: %s - %s", resp.Status, string(body))
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	results, ok := data["data"].([]interface{})
	if !ok || len(results) == 0 {
		return "", fmt.Errorf("no TVDB results found")
	}

	first := results[0].(map[string]interface{})
	if image, ok := first["image"].(string); ok && image != "" {
		return image, nil
	}

	return "", fmt.Errorf("TVDB result found, but no image available")
}

// GetRadarrPoster queries Radarr for a movie poster by exact title match
func GetRadarrPoster(title string) (string, error) {
	cacheKey := "poster_radarr_" + title

	// Attempt to load poster URL from cache (valid for 24h)
	var cachedURL string
	found, err := LoadCache(cacheKey, &cachedURL, 24*time.Hour)
	if err == nil && found {
		// Cache hit, return cached URL
		return cachedURL, nil
	}

	// Cache miss or expired, fetch from Radarr API
	endpoint := fmt.Sprintf("%s/api/v3/movie", Config.RadarrURL)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-Api-Key", Config.RadarrAPIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Radarr API error: %s - %s", resp.Status, string(body))
	}

	var movies []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&movies); err != nil {
		return "", err
	}

	for _, movie := range movies {
		if strings.EqualFold(movie["title"].(string), title) {
			if images, ok := movie["images"].([]interface{}); ok {
				for _, img := range images {
					imgMap := img.(map[string]interface{})
					if imgMap["coverType"] == "poster" {
						if url, ok := imgMap["remoteUrl"].(string); ok {
							// Save to cache before returning
							_ = SaveCache(cacheKey, url)
							return url, nil
						}
					}
				}
			}
		}
	}

	return "", fmt.Errorf("Radarr: no matching movie/poster found")
}


func PosterHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "missing query parameter", http.StatusBadRequest)
		return
	}

	radarrPoster, radarrErr := GetRadarrPoster(query)

	resp := map[string]interface{}{
		"query": query,
	}

	if radarrPoster != "" {
		resp["radarrPoster"] = radarrPoster
	}
	if radarrErr != nil {
		resp["radarrError"] = radarrErr.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
