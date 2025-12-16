package config

import "os"

type Config struct {
	MongoURI    string
	MongoDBName string
	RabbitURL   string
	HTTPAddr    string
}

func Load() Config {
	port := getEnv("PORT", "8080")

	return Config{
		MongoURI:    getEnv("MONGO_URI", ""),
		MongoDBName: getEnv("MONGO_DB", "test"),
		RabbitURL:   getEnv("RABBIT_URL", ""),
		HTTPAddr:    ":" + port,
	}
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}
