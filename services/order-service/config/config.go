package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBSource     string
	Port         string
	JWTSecretKey string
	StripeAPIKey string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, reading configuration from environment variables")
	}

	config := &Config{
		DBSource:     getEnv("DB_SOURCE", "postgres://postgres:postgres@localhost:5435/order_service_db?sslmode=disable"),
		Port:         getEnv("PORT", "50054"),
		JWTSecretKey: getEnv("JWT_SECRET_KEY", "secret"),
		StripeAPIKey: getEnv("STRIPE_API_KEY", ""),
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
