package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)


func GetAllRadarrMovies() ([]Movie, error) {
    endpoint := fmt.Sprintf("%s/api/v3/movie", Config.RadarrURL)

    req, err := http.NewRequest("GET", endpoint, nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("X-Api-Key", Config.RadarrAPIKey)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("Radarr API error: %s - %s", resp.Status, string(body))
    }

    var radarrMovies []struct {
        Title   string `json:"title"`
        HasFile bool   `json:"hasFile"`
        Images  []struct {
            CoverType string `json:"coverType"`
            RemoteURL string `json:"remoteUrl"`
        } `json:"images"`
        // add other fields you need here
    }

    if err := json.NewDecoder(resp.Body).Decode(&radarrMovies); err != nil {
        return nil, err
    }

    // Filter to only movies with downloaded files
    var filtered []Movie
    for _, m := range radarrMovies {
        if m.HasFile {
            posterURL := ""
            for _, img := range m.Images {
                if img.CoverType == "poster" {
                    posterURL = img.RemoteURL
                    break
                }
            }
            filtered = append(filtered, Movie{
                Title:     m.Title,
                PosterURL: posterURL,
                // other fields as needed
            })
        }
    }

    return filtered, nil
}


func extractPosterURL(images interface{}) string {
	list, ok := images.([]interface{})
	if !ok {
		return ""
	}

	for _, i := range list {
		img, ok := i.(map[string]interface{})
		if !ok {
			continue
		}
		coverType, _ := img["coverType"].(string)
		if strings.ToLower(coverType) == "poster" {
			if url, ok := img["remoteUrl"].(string); ok && url != "" {
				return url
			}
			if url, ok := img["url"].(string); ok && url != "" {
				return url
			}
		}
	}

	return ""
}


// MoviesHandler - GET /api/movies
func MoviesHandler(w http.ResponseWriter, r *http.Request) {
	movies, err := GetAllRadarrMovies()
	if err != nil {
		http.Error(w, "Failed to fetch movies: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movies)
}
