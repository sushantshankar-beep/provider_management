package config

import (
	"os"
	"strings"
)

type Config struct {
	MongoURI       string
	MongoDBName    string
	HTTPAddr       string
	AllowedOrigins []string
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
	}
	
}

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic("Missing required env: " + key)
	}
	return val
}
