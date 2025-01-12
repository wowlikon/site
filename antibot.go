package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	mu         sync.Mutex
	requests   map[string]int
	timestamps map[string]time.Time
	limit      int
	interval   time.Duration
}

func NewRateLimiter(limit int, interval time.Duration) *RateLimiter {
	return &RateLimiter{
		requests:   make(map[string]int),
		timestamps: make(map[string]time.Time),
		limit:      limit,
		interval:   interval,
	}
}

func loadBlock(filename string) ([]*regexp.Regexp, error) {
	var blocked []*regexp.Regexp
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	//Загрузка списка regex
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		filter := strings.TrimSpace(scanner.Text())
		if filter != "" {
			regex, err := regexp.Compile(filter)
			if err != nil {
				return nil, err
			}
			blocked = append(blocked, regex)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return blocked, nil
}

func (rl *RateLimiter) CleanupExpiredBlocks() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, lastRequestTime := range rl.timestamps {
			if now.Sub(lastRequestTime) > rl.interval {
				delete(rl.requests, ip)
				delete(rl.timestamps, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func AutoUpdateBlockedLists(pathsFile, uaFile string, paths, ua *[]*regexp.Regexp, mu *sync.RWMutex) {
	ticker := time.NewTicker(2 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		updatedPaths, err := loadBlock(pathsFile)
		if err != nil {
			log.Printf("Error updating blocked paths: %v", err)
			continue
		}

		updatedUA, err := loadBlock(uaFile)
		if err != nil {
			log.Printf("Error updating blocked UA: %v", err)
			continue
		}

		mu.Lock()
		*paths = updatedPaths
		*ua = updatedUA
		mu.Unlock()
	}
}

func (rl *RateLimiter) Limit(blockedPaths, blockedUA []*regexp.Regexp) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		rl.mu.Lock()
		defer rl.mu.Unlock()

		// Проверка на сброс блокировки по времени
		now := time.Now()
		if lastRequestTime, exists := rl.timestamps[ip]; exists {
			if now.Sub(lastRequestTime) > rl.interval {
				rl.requests[ip] = max(rl.requests[ip]-rl.limit, 0)
				rl.timestamps[ip] = now
			}
		} else {
			rl.timestamps[ip] = now
		}

		rl.requests[ip]++

		// Проверка на подозрительные типы запросов
		acceptHeader := c.Request.Header.Get("Accept")
		if acceptHeader == "*/*" {
			rl.requests[ip]++
		}

		// Проверка на превышение лимита
		if rl.requests[ip] > rl.limit {
			for key, value := range c.Request.Header {
				fmt.Printf("%s: %s\n", key, value)
			}
			time.Sleep(3 * time.Second)
			c.JSON(http.StatusTooManyRequests, gin.H{"message": "You have exceeded the number of allowed requests. Please wait before trying again."})
			c.Abort()
			return
		}

		// Проверка на заблокированные пути
		requestPath := c.Request.URL.Path
		for _, regex := range blockedPaths {
			if regex.MatchString(requestPath) {
				log.Printf("%s blocked by path %s\n", ip, requestPath)
				c.JSON(http.StatusForbidden, gin.H{"message": "Access forbidden: This path is restricted."})
				c.Abort()
				rl.requests[ip] += rl.limit
				return
			}
		}

		// Проверка на заблокированные User-Agent
		userAgent := c.Request.UserAgent()
		for _, regex := range blockedUA {
			if regex.MatchString(userAgent) {
				log.Printf("%s blocked by ua %s\n", ip, userAgent)
				c.JSON(http.StatusForbidden, gin.H{"message": "Access temporarily blocked due to User-Agent restrictions."})
				c.Abort()
				rl.requests[ip] += rl.limit
				return
			}
		}

		c.Next()
	}
}
