package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Redis   RedisConfig
	Server  ServerConfig
	Limiter LimiterConfig
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type ServerConfig struct {
	Port string
}

type LimiterConfig struct {
	IPRateLimit    int
	IPBlockTime    time.Duration
	TokenRateLimit int
	TokenBlockTime time.Duration
	TokenConfigs   map[string]TokenConfig
}

type TokenConfig struct {
	RateLimit int
	BlockTime time.Duration
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	config := &Config{
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Limiter: LimiterConfig{
			IPRateLimit:    getEnvInt("IP_RATE_LIMIT", 10),
			IPBlockTime:    time.Duration(getEnvInt("IP_BLOCK_TIME", 300)) * time.Second,
			TokenRateLimit: getEnvInt("TOKEN_RATE_LIMIT", 100),
			TokenBlockTime: time.Duration(getEnvInt("TOKEN_BLOCK_TIME", 300)) * time.Second,
			TokenConfigs:   parseTokenConfigs(),
		},
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func parseTokenConfigs() map[string]TokenConfig {
	configs := make(map[string]TokenConfig)

	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "TOKEN_") {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) != 2 {
				continue
			}

			tokenName := strings.TrimPrefix(parts[0], "TOKEN_")
			configParts := strings.Split(parts[1], ":")

			if len(configParts) >= 2 {
				rateLimit, err1 := strconv.Atoi(configParts[0])
				blockTime, err2 := strconv.Atoi(configParts[1])

				if err1 == nil && err2 == nil {
					configs[tokenName] = TokenConfig{
						RateLimit: rateLimit,
						BlockTime: time.Duration(blockTime) * time.Second,
					}
				}
			}
		}
	}

	return configs
}
