package config 

import (
	"os"
	"log"
)

type Config struct {
	Port 		string
	Environment	string
}

func Load() *Config {

	cfg := &Config{
		Port:		getEnv("PORT","8080"),
		Environment: getEnv("ENV","dev"),
	}
	
	log.Printf("Config Loaded: Port=%d, Environment=%s\n", cfg.Port, cfg.Environment)
	return cfg
}


func getEnv(key, defaultValue string) string{
	
	if value :=  os.Getenv(key); value != "" {
		return value
	}
	
	return defaultValue
}
