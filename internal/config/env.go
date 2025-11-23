package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL string
}

func InitConfig() *Config {
	databaseUrl, err := getEnv("DATABASE_URL")
	if err != nil {
		panic(err)
	}

	return &Config{
		DatabaseURL: databaseUrl,
	}
}

func getEnv(key string) (string, error) {
	if value, exists := os.LookupEnv(key); exists {
		return value, nil
	}

	return "", fmt.Errorf("cannot find %s in your environment", key)
}
