package config

import (
	"os"
	"strconv"
	"time"

	"github.com/Abb133Se/httpServer/internal/utils"
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
	if err := godotenv.Load(); err != nil {
		utils.Warn("No .env file found, using default environment values")
	}

	readTimeout, err := strconv.Atoi(getEnv("READ_TIMEOUT", "5"))
	if err != nil {
		utils.Warn("Invalid READ_TIMEOUT value, using default 5s")
		readTimeout = 5
	}

	writeTimeout, err := strconv.Atoi(getEnv("WRITE_TIMEOUT", "5"))
	if err != nil {
		utils.Warn("Invalid WRITE_TIMEOUT value, using default 5s")
		writeTimeout = 5
	}

	idleTimeout, err := strconv.Atoi(getEnv("IDLE_TIMEOUT", "30"))
	if err != nil {
		utils.Warn("Invalid IDLE_TIMEOUT value, using default 30s")
		idleTimeout = 30
	}

	cfg := &Config{
		Port:         getEnv("PORT", "4221"),
		ReadTimeout:  time.Duration(readTimeout) * time.Second,
		WriteTimeout: time.Duration(writeTimeout) * time.Second,
		IdleTimeout:  time.Duration(idleTimeout) * time.Second,
		LogLevel:     getEnv("LOG_LEVEL", "Info"),
	}

	utils.Info("Configuration loaded: Port=%s, ReadTimeout=%ds, WriteTimeout=%ds, IdleTimeout=%ds, LogLevel=%s",
		cfg.Port, readTimeout, writeTimeout, idleTimeout, cfg.LogLevel)
	return cfg
}

func getEnv(key, fallBack string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallBack
}
