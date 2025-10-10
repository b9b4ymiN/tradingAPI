package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// Server
	Port    string
	GinMode string

	// Security
	APIKey string

	// Binance
	BinanceAPIKey    string
	BinanceSecretKey string

	// Firebase
	FirebaseDBURL         string
	FirebaseCredentialsFile string
}

// Load loads configuration from environment variables
func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config := &Config{
		// Server
		Port:    getEnv("PORT", "8080"),
		GinMode: getEnv("GIN_MODE", "release"),

		// Security
		APIKey: getEnv("API_KEY", ""),

		// Binance
		BinanceAPIKey:    getEnv("BINANCE_API_KEY", ""),
		BinanceSecretKey: getEnv("BINANCE_SECRET_KEY", ""),

		// Firebase
		FirebaseDBURL:         getEnv("FIREBASE_DATABASE_URL", ""),
		FirebaseCredentialsFile: getEnv("FIREBASE_CREDENTIALS_FILE", ""),
	}

	// Validate required fields
	if config.APIKey == "" {
		log.Fatal("API_KEY environment variable is required")
	}

	if config.BinanceAPIKey == "" || config.BinanceSecretKey == "" {
		log.Fatal("BINANCE_API_KEY and BINANCE_SECRET_KEY environment variables are required")
	}

	if config.FirebaseDBURL == "" {
		log.Fatal("FIREBASE_DATABASE_URL environment variable is required")
	}

	return config
}

// getEnv retrieves an environment variable or returns a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
