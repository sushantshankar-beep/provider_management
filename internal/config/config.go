package config

import "os"

type Config struct {
	MongoURI    string
	MongoDBName string
	HTTPAddr    string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return Config{
		MongoURI:    mustEnv("MONGO_URI"),
		MongoDBName: mustEnv("MONGO_DB"),
		HTTPAddr:    ":" + port,
	}
}

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic("Missing required env: " + key)
	}
	return val
}
