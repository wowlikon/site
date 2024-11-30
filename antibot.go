package main

import (
	"bufio"
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

func loadBlockedPaths(filename string) ([]*regexp.Regexp, error) {
	var blockedPaths []*regexp.Regexp
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		path := strings.TrimSpace(scanner.Text())
		if path != "" {
			regex, err := regexp.Compile(path)
			if err != nil {
				return nil, err
			}
			blockedPaths = append(blockedPaths, regex)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return blockedPaths, nil
}

func (rl *RateLimiter) Limit(blockedPaths []*regexp.Regexp) gin.HandlerFunc {
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
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}

		requestPath := c.Request.URL.Path

		for _, regex := range blockedPaths {
			if regex.MatchString(requestPath) {
				c.AbortWithStatus(http.StatusForbidden)
				rl.requests[ip] += rl.limit
				return
			}
		}

		c.Next()
	}
}
