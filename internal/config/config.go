package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	Environment string
	S3Config    S3Config
}

type S3Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	Region    string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENV", "dev"),
		S3Config: S3Config{
			Endpoint:  getEnv("S3_ENDPOINT", ""),
			AccessKey: getEnv("S3_ACCESS_KEY", ""),
			SecretKey: getEnv("S3_SECRET_KEY", ""),
			Bucket:    getEnv("S3_BUCKET", ""),
			Region:    getEnv("S3_REGION", "us-east-1"),
		},
	}

	log.Printf("Config Loaded: Port=%d, Environment=%s\n", cfg.Port, cfg.Environment)
	return cfg
}

func getEnv(key, defaultValue string) string {

	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}
