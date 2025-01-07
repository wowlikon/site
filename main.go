package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"image"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var config Config
var canvas *image.RGBA

func main() {
	gin.SetMode(gin.ReleaseMode)

	// Создание canvas
	canvas = createCanvas(64)

	r := gin.Default()
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	blockedPaths, err := loadBlock("./data/blocked_paths.txt")
	if err != nil {
		panic(err)
	}

	blockedUA, err := loadBlock("./data/blocked_ua.txt")
	if err != nil {
		panic(err)
	}

	// Установка блокировки ботов
	rateLimiter := NewRateLimiter(60, time.Minute)
	r.Use(rateLimiter.Limit(blockedPaths, blockedUA))
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

	r.GET("/canvas", func(c *gin.Context) {
		handleCanvas(c, canvas)
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

	// Запуск сервера HTTP
	if (config.Server.Protocols & int(HTTP)) != 0 {
		fmt.Printf("Starting server http://%s:%d\n", config.Server.Host, config.Server.HttpPort)
		go func() {
			if err := r.Run(fmt.Sprintf("%s:%d", config.Server.Host, config.Server.HttpPort)); err != nil {
				log.Fatalf("HTTP server failed: %v", err)
			}
		}()
	}

	// Запуск сервера HTTPS
	if (config.Server.Protocols & int(HTTPS)) != 0 {
		fmt.Printf("Starting server https://%s:%d\n", config.Server.Host, config.Server.HttpsPort)
		go func() {
			if err := r.RunTLS(
				fmt.Sprintf("%s:%d", config.Server.Host, config.Server.HttpsPort),
				fmt.Sprintf("ssl/%s.crt", config.Server.Host),
				fmt.Sprintf("ssl/%s.key", config.Server.Host),
			); err != nil {
				log.Fatalf("HTTPS server failed: %v", err)
			}
		}()
	}

	// Запуск ботов
	bot, err := tgbotapi.NewBotAPI(config.Tokens.Telegram)
	if err != nil {
		log.Panic(err)
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		s, _ := json.MarshalIndent(update, "", " ")
		fmt.Println(string(s))
		if update.Message != nil {
			if strings.HasPrefix(update.Message.Text, "/pixel") {
				handlePixelCommand(update.Message.Text, canvas)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Pixel updated!")
				bot.Send(msg)
			}
			if strings.HasPrefix(update.Message.Text, "/test") {
				TestStar(bot, update)
			}
			if update.Message.SuccessfulPayment != nil {
				msg := tgbotapi.NewMessage(
					update.Message.Chat.ID,
					fmt.Sprintf(
						"Спасибо за ваш платеж!\n%d %s\n%s",
						update.Message.SuccessfulPayment.TotalAmount,
						update.Message.SuccessfulPayment.Currency,
						update.Message.SuccessfulPayment.InvoicePayload,
					),
				)
				bot.Send(msg)
			}
		}

		if update.PreCheckoutQuery != nil {
			// Подтверждение предоплаты
			pca := tgbotapi.PreCheckoutConfig{
				PreCheckoutQueryID: update.PreCheckoutQuery.ID,
				OK:                 true,
				ErrorMessage:       "Ok",
			}
			if _, err := bot.Request(pca); err != nil {
				msg := tgbotapi.NewMessage(
					update.PreCheckoutQuery.From.ID, fmt.Sprintf("Ошибка при подтверждении предзаказа: \n%s", err),
				)
				bot.Send(msg)
				continue
			}
		}
	}
}

func TestStar(b *tgbotapi.BotAPI, u tgbotapi.Update) {
	price := []tgbotapi.LabeledPrice{
		{
			Label:  "XTR",
			Amount: 1,
		},
	}

	invoice := tgbotapi.NewInvoice(
		u.Message.Chat.ID,
		"Star Pay",
		"Purchase Stars",
		"test",
		"",
		"start_param_unique_v1",
		"XTR",
		price,
	)
	invoice.SuggestedTipAmounts = []int{}
	_, err := b.Send(invoice)
	if err != nil {
		log.Printf("Error sending invoice: %v", err)
		return
	}
}
