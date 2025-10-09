// Package utils provides shared utility functions for logging and diagnostics.
// It includes a lightweight logger with configurable verbosity levels.
package utils

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

var (
	logLevel string
	logger   *log.Logger
)

// InitLogger initializes the global logger with the specified verbosity level.
//
// Supported log levels:
//   - "debug": logs DEBUG, INFO, WARN, and ERROR messages
//   - "info":  logs INFO, WARN, and ERROR messages
//   - "warn":  logs WARN and ERROR messages
//   - any other value: logs only ERROR messages
//
// Example:
//
//	cfg := config.LoadConfig()
//	utils.InitLogger(cfg.LogLevel)
func InitLogger(level string) {
	logLevel = strings.ToLower(level)
	logger = log.New(os.Stdout, "", 0)
}

// logMessage writes a formatted log entry to standard output.
// It includes a timestamp and the log level prefix.
//
// This is an internal helper used by all public log methods.
func logMessage(level, message string, args ...any) {
	timestamp := time.Now().Format(time.RFC3339)
	fmt.Fprintf(os.Stdout, "[%s] %s %s\n", level, timestamp, fmt.Sprintf(message, args...))
}

// Info logs informational messages that describe normal server operations.
// It is active when the log level is set to "info" or "debug".
func Info(message string, args ...any) {
	if logLevel == "info" || logLevel == "debug" {
		logMessage("INFO", message, args...)
	}
}

// Debug logs detailed diagnostic information useful for debugging.
// It is active only when the log level is set to "debug".
func Debug(message string, args ...any) {
	if logLevel == "debug" {
		logMessage("DEBUG", message, args...)
	}
}

// Warn logs non-critical issues that may require attention but do not
// prevent the program from running. It is active for "warn" and "debug" levels.
func Warn(message string, args ...any) {
	if logLevel == "warn" || logLevel == "debug" {
		logMessage("WARN", message, args...)
	}
}

// Error logs errors and critical issues that indicate a failure
// in execution or configuration. This method always logs regardless of log level.
func Error(message string, args ...any) {
	logMessage("ERROR", message, args...)
}
