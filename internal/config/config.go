// Package config provides centralized configuration management for the HTTP server.
// It loads settings from environment variables (with support for .env files),
// applies sensible defaults, and exposes strongly typed timeouts and log levels.
package config

import (
	"os"
	"strconv"
	"time"

	"github.com/Abb133Se/httpServer/internal/utils"
	"github.com/joho/godotenv"
)

// Config holds the server's runtime configuration parameters.
//
// All timeout values are expressed in seconds and are converted into time.Duration.
// It supports configuration through environment variables or a .env file.
//
// Environment variables:
//   - PORT:          Server listening port (default: "4221")
//   - READ_TIMEOUT:  Maximum duration for reading a request (default: 5 seconds)
//   - WRITE_TIMEOUT: Maximum duration for writing a response (default: 5 seconds)
//   - IDLE_TIMEOUT:  Maximum time to keep an idle connection open (default: 30 seconds)
//   - LOG_LEVEL:     Logging verbosity level ("debug", "info", "warn", default: "info")

type Config struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	LogLevel     string
}

// LoadConfig loads configuration settings from environment variables or a .env file.
//
// If no .env file is found, defaults are applied and a warning is logged.
// Invalid numeric values for timeout fields are replaced with default durations.
// Returns a pointer to a fully populated Config struct.
//
// Example:
//
//	cfg := config.LoadConfig()
//	utils.InitLogger(cfg.LogLevel)
//	server.Start(cfg.Port, cfg)
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

// getEnv returns the value of the specified environment variable.
// If the variable is not set, it returns the provided fallback value.
//
// This helper ensures that missing environment variables do not cause runtime errors.
func getEnv(key, fallBack string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallBack
}
