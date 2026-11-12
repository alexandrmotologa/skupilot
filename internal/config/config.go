package config

import (
	"os"
)

// Config represents runtime settings for skupilot.
type Config struct {
	Port         string
	DBPath       string
	GeminiAPIKey string
	OpenAIAPIKey string
	ShopifyStore string
	ShopifyToken string
}

// Load reads settings from environment variables with sensible defaults.
func Load() *Config {
	port := os.Getenv("SKUPILOT_PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("SKUPILOT_DB")
	if dbPath == "" {
		dbPath = "./catalog.db"
	}

	return &Config{
		Port:         port,
		DBPath:       dbPath,
		GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
		OpenAIAPIKey: os.Getenv("OPENAI_API_KEY"),
		ShopifyStore: os.Getenv("SHOPIFY_STORE_DOMAIN"),
		ShopifyToken: os.Getenv("SHOPIFY_ADMIN_TOKEN"),
	}
}
