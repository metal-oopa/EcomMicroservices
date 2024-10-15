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
	ConsulAddress string
	ServiceName   string
	ServiceHost   string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, reading configuration from environment variables")
	}

	config := &Config{
		DBSource:      getEnv("DB_SOURCE", "postgres://postgres:postgres@localhost:5434/cart_service_db?sslmode=disable"),
		Port:          getEnv("PORT", "50053"),
		JWTSecretKey:  getEnv("JWT_SECRET_KEY", "secret"),
		ConsulAddress: getEnv("CONSUL_ADDRESS", "consul:8500"),
		ServiceName:   getEnv("SERVICE_NAME", "user-service"),
		ServiceHost:   getEnv("SERVICE_HOST", "user-service"),
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
