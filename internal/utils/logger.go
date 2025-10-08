package utils

import (
	"fmt"
	"log"
	"os"
	"time"
)

var (
	logLevel string
	logger   *log.Logger
)

func InitLogger(level string) {
	logLevel = level
	logger = log.New(os.Stdout, "", 0)
}

func logMessage(level, message string, args ...any) {
	timestamp := time.Now().Format(time.RFC3339)
	fmt.Fprintf(os.Stdout, "[%s] %s %s\n", level, timestamp, fmt.Sprintf(message, args...))
}

func Info(message string, args ...any) {
	if logLevel == "info" || logLevel == "debug" {
		logMessage("INFO", message, args...)
	}
}

func Debug(message string, args ...any) {
	if logLevel == "debug" {
		logMessage("DEBUG", message, args...)
	}
}

func Warn(message string, args ...any) {
	if logLevel == "warn" || logLevel == "debug" {
		logMessage("WARN", message, args...)
	}
}

func Error(message string, args ...any) {
	logMessage("ERROR", message, args...)
}
