package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	LogLevel     string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	readTimeout, _ := strconv.Atoi(getEnv("READ_TIMEOUT", "5"))
	writeTimeout, _ := strconv.Atoi(getEnv("WRITE_TIMEOUT", "5"))
	idleTimeout, _ := strconv.Atoi(getEnv("IDLE_TIMEOUT", "30"))

	return &Config{
		Port:         getEnv("PORT", "4221"),
		ReadTimeout:  time.Duration(readTimeout) * time.Second,
		WriteTimeout: time.Duration(writeTimeout) * time.Second,
		IdleTimeout:  time.Duration(idleTimeout) * time.Second,
		LogLevel:     getEnv("LOG-LEVEL", "Info"),
	}
}

func getEnv(key, fallBack string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallBack
}
