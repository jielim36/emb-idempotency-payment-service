package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
}

type ServerConfig struct {
	Port    string
	GinMode string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

type AppConfig struct {
	Name    string
	Version string
}

func LoadConfig() *Config {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	return &Config{
		Server: ServerConfig{
			Port:    getEnvOrPanic("PORT"),       // required
			GinMode: getEnv("GIN_MODE", "debug"), // optional
		},
		Database: DatabaseConfig{
			Host:     getEnvOrPanic("DB_HOST"),
			Port:     getEnvOrPanic("DB_PORT"),
			User:     getEnvOrPanic("DB_USER"),
			Password: getEnvOrPanic("DB_PASSWORD"),
			DBName:   getEnvOrPanic("DB_NAME"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"), // optional
		},
		Redis: RedisConfig{
			Host:     getEnvOrPanic("REDIS_HOST"),
			Port:     getEnvOrPanic("REDIS_PORT"),
			Password: getEnvOrPanic("REDIS_PASSWORD"),
		},
		App: AppConfig{
			Name:    getEnvOrPanic("APP_NAME"),
			Version: getEnvOrPanic("APP_VERSION"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvOrPanic(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic("Missing required environment variable: " + key)
	}
	fmt.Printf("Environment: [%s]=%s", key, value)
	return value
}
