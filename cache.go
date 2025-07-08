package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// CacheDir is the directory where cache files are stored
const CacheDir = "./cache"

// CacheItem represents a cached object with timestamp
type CacheItem struct {
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

// SaveCache saves the given data under the cache key (filename)
func SaveCache(key string, data interface{}) error {
	if err := os.MkdirAll(CacheDir, 0755); err != nil {
		return err
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	item := CacheItem{
		Timestamp: time.Now(),
		Data:      bytes,
	}

	itemBytes, err := json.Marshal(item)
	if err != nil {
		return err
	}

	filePath := filepath.Join(CacheDir, key+".json")
	return ioutil.WriteFile(filePath, itemBytes, 0644)
}

// LoadCache loads cached data by key into the provided destination pointer.
// It returns true if the cache was loaded and is valid (not expired), false otherwise.
func LoadCache(key string, dest interface{}, maxAge time.Duration) (bool, error) {
	filePath := filepath.Join(CacheDir, key+".json")

	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		// cache miss or file error
		return false, err
	}

	var item CacheItem
	if err := json.Unmarshal(bytes, &item); err != nil {
		return false, err
	}

	if time.Since(item.Timestamp) > maxAge {
		// cache expired
		return false, nil
	}

	if err := json.Unmarshal(item.Data, dest); err != nil {
		return false, err
	}

	return true, nil
}
