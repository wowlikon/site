package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var config Config
var jwtKey = []byte("your_secret_key")

// User структура для хранения данных пользователя
type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Database - имитация базы данных
var usersDB = map[string]string{} // email -> hashedPassword

// Claims структура для JWT
type Claims struct {
	Email                string `json:"email"`
	jwt.RegisteredClaims        // Используем стандартные поля JWT
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	var (
		blockedPaths []*regexp.Regexp
		blockedUA    []*regexp.Regexp
		mu           sync.RWMutex
		err          error
	)

	blockedPaths, err = loadBlock("./data/blocked_paths.txt")
	if err != nil {
		panic(err)
	}

	blockedUA, err = loadBlock("./data/blocked_ua.txt")
	if err != nil {
		panic(err)
	}

	// Запуск автообновления списков
	go AutoUpdateBlockedLists("./data/blocked_paths.txt", "./data/blocked_ua.txt", &blockedPaths, &blockedUA, &mu)

	// Установка блокировки ботов
	rateLimiter := NewRateLimiter(60, time.Minute)
	go rateLimiter.CleanupExpiredBlocks()
	r.Use(func(c *gin.Context) {
		mu.RLock()
		defer mu.RUnlock()
		rateLimiter.Limit(blockedPaths, blockedUA)(c)
	})

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
			choices = []string{"Red", "Green", "Blue", "Black"}
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

	r.POST("/account/login", login)
	r.POST("/account/register", register)
	r.GET("/account/profile", authenticateJWT(), profile)

	// Запуск сервера
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

// register - обработчик регистрации
func register(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Проверка наличия пользователя
	if _, exists := usersDB[user.Email]; exists {
		c.JSON(http.StatusConflict, gin.H{"error": "User  already exists"})
		return
	}

	// Хэширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while hashing password"})
		return
	}

	// Сохранение пользователя
	usersDB[user.Email] = string(hashedPassword)
	c.JSON(http.StatusCreated, gin.H{"message": "User  registered successfully"})
}

// login - обработчик входа
func login(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Проверка пользователя
	storedPassword, exists := usersDB[user.Email]
	if !exists || bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(user.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Создание JWT
	expirationTime := time.Now().Add(15 * time.Minute)
	claims := &Claims{
		Email: user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while creating token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
	})
}

// profile - обработчик получения профиля
func profile(c *gin.Context) {
	claims := c.MustGet("claims").(*Claims)
	c.JSON(http.StatusOK, gin.H{
		"email":   claims.Email,
		"message": "Welcome to your profile!",
	})
}

// authenticateJWT - middleware для проверки токена
func authenticateJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(strings.TrimPrefix(tokenString, "Bearer "), claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		c.Set("claims", claims)
		c.Next()
	}
}
