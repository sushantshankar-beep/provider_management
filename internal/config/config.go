package config

import (
	"os"
	"strings"
)


type PayUConfig struct {
	Key     string
	Salt    string
	BaseURL string
}

type Config struct {
	MongoURI       string
	MongoDBName    string
	HTTPAddr       string
	AllowedOrigins []string
	PayU           PayUConfig 
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	origins := os.Getenv("ALLOWED_ORIGINS")
	var allowedOrigins []string
	if origins != "" {
		allowedOrigins = strings.Split(origins, ",")
	}
	
	return Config{
		MongoURI:       mustEnv("MONGO_URI"),
		MongoDBName:    mustEnv("MONGO_DB"),
		HTTPAddr:       ":" + port,
		AllowedOrigins: allowedOrigins,
		PayU: PayUConfig{
			Key:     mustEnv("PAYU_KEY"),
			Salt:    mustEnv("PAYU_SALT"),
			BaseURL: mustEnv("PAYU_BASE_URL"),
		},
	}
	
}

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic("Missing required env: " + key)
	}
	return val
}
