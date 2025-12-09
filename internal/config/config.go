package config

import "os"

type Config struct {
	MongoURI    string
	MongoDBName string
	RabbitURL   string
	HTTPAddr    string
}

func Load() Config {
	return Config{
		MongoURI:    getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName: getEnv("MONGO_DB", "provider_db"),
		RabbitURL:   getEnv("RABBIT_URL", "amqp://guest:guest@localhost:5672/"),
		HTTPAddr:    getEnv("HTTP_ADDR", ":8080"),
	}
}

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
