package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/douglastaylorb/pos-ratelimiter/internal/config"
	"github.com/douglastaylorb/pos-ratelimiter/internal/limiter"
	"github.com/douglastaylorb/pos-ratelimiter/internal/middleware"
	"github.com/douglastaylorb/pos-ratelimiter/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	store := storage.NewMemoryStorage()

	cfg := &config.LimiterConfig{
		IPRateLimit:    2,
		IPBlockTime:    5 * time.Second,
		TokenRateLimit: 5,
		TokenBlockTime: 5 * time.Second,
		TokenConfigs: map[string]config.TokenConfig{
			"test-token": {
				RateLimit: 3,
				BlockTime: 5 * time.Second,
			},
		},
	}

	rateLimiter := limiter.NewRateLimiter(store)
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(rateLimiter, cfg)

	router := gin.New()
	router.Use(rateLimitMiddleware.Handler())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	return router
}

func TestIPRateLimit(t *testing.T) {
	router := setupTestRouter()

	for i := 0; i < 2; i++ {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.1")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestTokenRateLimit(t *testing.T) {
	router := setupTestRouter()

	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("API_KEY", "test-token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("API_KEY", "test-token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestTokenPrecedenceOverIP(t *testing.T) {
	router := setupTestRouter()

	for i := 0; i < 2; i++ {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.2")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.2")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	req, _ = http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.2")
	req.Header.Set("API_KEY", "test-token")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLimiterLogic(t *testing.T) {
	store := storage.NewMemoryStorage()
	rateLimiter := limiter.NewRateLimiter(store)

	limitInfo := limiter.LimitInfo{
		Key:       "test-key",
		RateLimit: 2,
		BlockTime: 5 * time.Second,
	}

	ctx := context.Background()

	for i := 0; i < 2; i++ {
		allowed, err := rateLimiter.Allow(ctx, limitInfo)
		assert.NoError(t, err)
		assert.True(t, allowed)
	}

	allowed, err := rateLimiter.Allow(ctx, limitInfo)
	assert.NoError(t, err)
	assert.False(t, allowed)

	allowed, err = rateLimiter.Allow(ctx, limitInfo)
	assert.NoError(t, err)
	assert.False(t, allowed)
}
