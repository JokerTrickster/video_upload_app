package logger

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/JokerTrickster/video-upload-backend/internal/config"
)

// Logger levels
const (
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
)

// Logger represents the application logger
type Logger struct {
	level string
}

// New creates a new logger instance
func New(cfg *config.Config) *Logger {
	return &Logger{
		level: cfg.Server.LogLevel,
	}
}

// Debug logs a debug message
func (l *Logger) Debug(message string, args ...interface{}) {
	if l.shouldLog(LevelDebug) {
		l.log("DEBUG", message, args...)
	}
}

// Info logs an info message
func (l *Logger) Info(message string, args ...interface{}) {
	if l.shouldLog(LevelInfo) {
		l.log("INFO", message, args...)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(message string, args ...interface{}) {
	if l.shouldLog(LevelWarn) {
		l.log("WARN", message, args...)
	}
}

// Error logs an error message
func (l *Logger) Error(message string, args ...interface{}) {
	if l.shouldLog(LevelError) {
		l.log("ERROR", message, args...)
	}
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(message string, args ...interface{}) {
	l.log("FATAL", message, args...)
	os.Exit(1)
}

// log formats and writes the log message
func (l *Logger) log(level, message string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}

	log.Printf("[%s] %s: %s\n", timestamp, level, message)
}

// shouldLog determines if a message should be logged based on the log level
func (l *Logger) shouldLog(level string) bool {
	levels := map[string]int{
		LevelDebug: 0,
		LevelInfo:  1,
		LevelWarn:  2,
		LevelError: 3,
	}

	currentLevel := levels[l.level]
	messageLevel := levels[level]

	return messageLevel >= currentLevel
}

// Global logger functions for convenience
var defaultLogger *Logger

// Init initializes the default logger
func Init(cfg *config.Config) {
	defaultLogger = New(cfg)
}

// Debug logs a debug message using the default logger
func Debug(message string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Debug(message, args...)
	}
}

// Info logs an info message using the default logger
func Info(message string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Info(message, args...)
	}
}

// Warn logs a warning message using the default logger
func Warn(message string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Warn(message, args...)
	}
}

// Error logs an error message using the default logger
func Error(message string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Error(message, args...)
	}
}

// Fatal logs a fatal message and exits using the default logger
func Fatal(message string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Fatal(message, args...)
	}
}
