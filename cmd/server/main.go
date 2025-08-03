package main

import (
	"log"

	"github.com/douglastaylorb/pos-ratelimiter/internal/config"
	"github.com/douglastaylorb/pos-ratelimiter/internal/limiter"
	"github.com/douglastaylorb/pos-ratelimiter/internal/middleware"
	"github.com/douglastaylorb/pos-ratelimiter/internal/storage"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize storage (Redis with fallback to memory)
	var store limiter.Storage

	redisStorage, err := storage.NewRedisStorage(
		cfg.Redis.Host,
		cfg.Redis.Port,
		cfg.Redis.Password,
		cfg.Redis.DB,
	)

	if err != nil {
		log.Printf("Failed to connect to Redis, using memory storage: %v", err)
		store = storage.NewMemoryStorage()
	} else {
		log.Println("Connected to Redis successfully")
		store = redisStorage
	}

	// Initialize rate limiter
	rateLimiter := limiter.NewRateLimiter(store)
	defer rateLimiter.Close()

	// Initialize middleware
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(rateLimiter, &cfg.Limiter)

	// Setup Gin router
	router := gin.Default()

	// Apply rate limiting middleware
	router.Use(rateLimitMiddleware.Handler())

	// Define routes
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World!",
			"ip":      c.ClientIP(),
			"token":   c.GetHeader("API_KEY"),
		})
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Test endpoint",
			"time":    "2025-01-29T12:48:34Z",
		})
	})

	// Start server
	port := cfg.Server.Port
	log.Printf("Server starting on port %s", port)
	log.Fatal(router.Run(":" + port))
}
