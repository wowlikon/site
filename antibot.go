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

func (rl *RateLimiter) Limit(blockedPaths, blockedUA []*regexp.Regexp) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		rl.mu.Lock()
		defer rl.mu.Unlock()

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
		if rl.requests[ip] > rl.limit {
			for key, value := range c.Request.Header {
				fmt.Printf("%s: %s\n", key, value)
			}
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}

		// Проверка на заблокированные пути
		requestPath := c.Request.URL.Path
		for _, regex := range blockedPaths {
			if regex.MatchString(requestPath) {
				log.Println("%s blocked by path %s\n", ip, requestPath)
				c.AbortWithStatus(http.StatusForbidden)
				rl.requests[ip] += rl.limit
				return
			}
		}

		// Проверка на заблокированные User-Agent
		userAgent := c.Request.UserAgent()
		for _, regex := range blockedUA {
			if regex.MatchString(userAgent) {
				log.Println("%s blocked by ua %s\n", ip, userAgent)
				c.AbortWithStatus(http.StatusForbidden)
				rl.requests[ip] += rl.limit
				return
			}
		}

		c.Next()
	}
}
