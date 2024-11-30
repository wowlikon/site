package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Cache struct {
	sync.RWMutex
	data      map[string]interface{}
	timestamp map[string]time.Time
}

var cache = Cache{
	data:      make(map[string]interface{}),
	timestamp: make(map[string]time.Time),
}

const cacheDuration = 30 * time.Minute

func fetchFromGitHub(url string) (interface{}, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch data from GitHub: %s", resp.Status)
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func getCachedData(username, repo, endpoint string) (interface{}, error) {
	key := fmt.Sprintf("%s/%s/%s", username, repo, endpoint)

	cache.RLock()
	data, exists := cache.data[key]
	timestamp := cache.timestamp[key]
	cache.RUnlock()

	if exists && time.Since(timestamp) < cacheDuration {
		return data, nil
	}

	cache.Lock()
	defer cache.Unlock()

	data, exists = cache.data[key]
	timestamp = cache.timestamp[key]
	if exists && time.Since(timestamp) < cacheDuration {
		return data, nil
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s%s", username, repo, endpoint)
	data, err := fetchFromGitHub(url)
	if err != nil {
		return nil, err
	}

	cache.data[key] = data
	cache.timestamp[key] = time.Now()

	return data, nil
}

func ghCache(c *gin.Context) {
	username := c.Param("username")
	repo := c.Param("repo")

	repoData, err := getCachedData(username, repo, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	languagesData, err := getCachedData(username, repo, "/languages")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"repository": repoData,
		"languages":  languagesData,
	})
}
