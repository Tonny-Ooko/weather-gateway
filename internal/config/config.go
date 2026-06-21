package config

import (
	"log"
	"os"
	"time"
)

type Config struct {
	WeatherAPIKey string
	ServerPort    string
	CacheTTL      time.Duration
	MaxCacheSize  int
}

func LoadConfig() *Config {
	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		log.Fatal("CRITICAL CONFIGURATION ERROR: WEATHER_API_KEY environment variable is required")
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		WeatherAPIKey: apiKey,
		ServerPort:    port,
		CacheTTL:      5 * time.Minute,
		MaxCacheSize:  50000, // Caps memory footprint
	}
}
