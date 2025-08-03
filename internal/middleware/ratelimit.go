package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/douglastaylorb/pos-ratelimiter/internal/config"
	"github.com/douglastaylorb/pos-ratelimiter/internal/limiter"
	"github.com/gin-gonic/gin"
)

type RateLimitMiddleware struct {
	limiter *limiter.RateLimiter
	config  *config.LimiterConfig
}

func NewRateLimitMiddleware(rateLimiter *limiter.RateLimiter, cfg *config.LimiterConfig) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter: rateLimiter,
		config:  cfg,
	}
}

func (m *RateLimitMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()

		// Check API primeiro
		apiKey := c.GetHeader("API_KEY")

		var limitInfo limiter.LimitInfo

		if apiKey != "" {
			// Usa o token
			limitInfo = m.getTokenLimitInfo(apiKey)
		} else {
			// Usa o IP
			limitInfo = m.getIPLimitInfo(c.ClientIP())
		}

		allowed, err := m.limiter.Allow(ctx, limitInfo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "you have reached the maximum number of requests or actions allowed within a certain time frame",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (m *RateLimitMiddleware) getTokenLimitInfo(token string) limiter.LimitInfo {
	// Limpa o token (remove qualquer espaço em branco)
	token = strings.TrimSpace(token)

	// Verifica se existe uma configuração específica para este token
	if tokenConfig, exists := m.config.TokenConfigs[token]; exists {
		return limiter.LimitInfo{
			Key:       "token:" + token,
			RateLimit: tokenConfig.RateLimit,
			BlockTime: tokenConfig.BlockTime,
		}
	}

	return limiter.LimitInfo{
		Key:       "token:" + token,
		RateLimit: m.config.TokenRateLimit,
		BlockTime: m.config.TokenBlockTime,
	}
}

func (m *RateLimitMiddleware) getIPLimitInfo(ip string) limiter.LimitInfo {
	return limiter.LimitInfo{
		Key:       "ip:" + ip,
		RateLimit: m.config.IPRateLimit,
		BlockTime: m.config.IPBlockTime,
	}
}
