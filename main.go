package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
)

var config Config

type Question struct {
	Question string
	Choices  []string
}

func main() {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	r.LoadHTMLGlob("./static/pages/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
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
		// Запускаем HTTP сервер
		if err := r.Run(fmt.Sprintf("%s:%d", config.Server.Host, config.Server.HttpPort)); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}
}
