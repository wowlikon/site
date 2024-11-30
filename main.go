package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
)

var config Config

func main() {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	blockedPaths, err := loadBlockedPaths("./data/blocked_paths.txt")
	if err != nil {
		panic(err)
	}

	rateLimiter := NewRateLimiter(60, time.Minute)
	r.Use(rateLimiter.Limit(blockedPaths))
	r.SetFuncMap(template.FuncMap{
		"lower": strings.ToLower,
	})
	r.LoadHTMLGlob("./static/pages/*")

	r.GET("/", func(c *gin.Context) {
		repositories, err := loadRepositories("./data/repositories.json")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить репозитории"})
			return
		}

		certificates, err := loadCertificates("./data/certificates.json")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить сертификаты"})
			return
		}

		data := MainPageData{
			Certificates: certificates,
			Repos:        repositories,
		}

		c.HTML(http.StatusOK, "index.html", data)
	})

	r.GET("/question", func(c *gin.Context) {
		question := c.Query("question")
		choices := c.QueryArray("choices")

		if question == "" || len(choices) == 0 {
			question = "What is your favorite color?"
			choices = []string{"Red", "Green", "Blue", " Black "}
		}

		q := Question{
			Question: question,
			Choices:  choices,
		}
		c.HTML(http.StatusOK, "question.html", q)
	})

	r.GET("/api/repos/:username/:repo", ghCache)
	r.GET("/api/stats", func(c *gin.Context) {
		getSystemStats(c)
	})

	r.Static("/static", "./static")
	r.StaticFile("/favicon.ico", "./static/images/favicons/site.webmanifest")
	fmt.Printf("Starting server http://%s:%d\n", config.Server.Host, config.Server.HttpPort)
	if config.Server.EnableHTTPS {
		fmt.Printf("Starting server https://%s:%d\n", config.Server.Host, config.Server.HttpsPort)
	}

	// Запускаем HTTPS
	if config.Server.EnableHTTPS {
		go func() {
			if err := r.Run(fmt.Sprintf("%s:%d", config.Server.Host, config.Server.HttpPort)); err != nil {
				log.Fatalf("HTTP server failed: %v", err)
			}
		}()

		if err := r.RunTLS(
			fmt.Sprintf("%s:%d", config.Server.Host, config.Server.HttpsPort),
			fmt.Sprintf("ssl/%s.crt", config.Server.Host),
			fmt.Sprintf("ssl/%s.key", config.Server.Host),
		); err != nil {
			log.Fatalf("HTTPS server failed: %v", err)
		}
	} else {
		// Запускаем HTTP
		if err := r.Run(fmt.Sprintf("%s:%d", config.Server.Host, config.Server.HttpPort)); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}
}
