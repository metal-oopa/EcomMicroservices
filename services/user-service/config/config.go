package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBSource      string
	Port          string
	JWTSecretKey  string
	TokenDuration string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, reading configuration from environment variables")
	}

	config := &Config{
		DBSource:      getEnv("DB_SOURCE", "postgres://postgres:postgres@localhost:5432/user_service_db?sslmode=disable"),
		Port:          getEnv("PORT", "50051"),
		JWTSecretKey:  getEnv("JWT_SECRET_KEY", "your-secret-key"),
		TokenDuration: getEnv("TOKEN_DURATION", "24h"),
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
