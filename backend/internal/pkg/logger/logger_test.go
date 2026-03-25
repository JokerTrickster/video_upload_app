package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/JokerTrickster/video-upload-backend/internal/config"
)

func TestNew(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{LogLevel: "info"},
	}

	l := New(cfg)
	assert.NotNil(t, l)
	assert.Equal(t, "info", l.level)
}

func TestLogger_shouldLog(t *testing.T) {
	tests := []struct {
		name         string
		loggerLevel  string
		messageLevel string
		want         bool
	}{
		// Debug logger logs everything
		{name: "debug logger - debug msg", loggerLevel: LevelDebug, messageLevel: LevelDebug, want: true},
		{name: "debug logger - info msg", loggerLevel: LevelDebug, messageLevel: LevelInfo, want: true},
		{name: "debug logger - warn msg", loggerLevel: LevelDebug, messageLevel: LevelWarn, want: true},
		{name: "debug logger - error msg", loggerLevel: LevelDebug, messageLevel: LevelError, want: true},

		// Info logger skips debug
		{name: "info logger - debug msg", loggerLevel: LevelInfo, messageLevel: LevelDebug, want: false},
		{name: "info logger - info msg", loggerLevel: LevelInfo, messageLevel: LevelInfo, want: true},
		{name: "info logger - warn msg", loggerLevel: LevelInfo, messageLevel: LevelWarn, want: true},
		{name: "info logger - error msg", loggerLevel: LevelInfo, messageLevel: LevelError, want: true},

		// Warn logger skips debug and info
		{name: "warn logger - debug msg", loggerLevel: LevelWarn, messageLevel: LevelDebug, want: false},
		{name: "warn logger - info msg", loggerLevel: LevelWarn, messageLevel: LevelInfo, want: false},
		{name: "warn logger - warn msg", loggerLevel: LevelWarn, messageLevel: LevelWarn, want: true},
		{name: "warn logger - error msg", loggerLevel: LevelWarn, messageLevel: LevelError, want: true},

		// Error logger only logs errors
		{name: "error logger - debug msg", loggerLevel: LevelError, messageLevel: LevelDebug, want: false},
		{name: "error logger - info msg", loggerLevel: LevelError, messageLevel: LevelInfo, want: false},
		{name: "error logger - warn msg", loggerLevel: LevelError, messageLevel: LevelWarn, want: false},
		{name: "error logger - error msg", loggerLevel: LevelError, messageLevel: LevelError, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Logger{level: tt.loggerLevel}
			got := l.shouldLog(tt.messageLevel)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestInit(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{LogLevel: "warn"},
	}

	Init(cfg)
	assert.NotNil(t, defaultLogger, "defaultLogger should be set after Init")
	assert.Equal(t, "warn", defaultLogger.level)
}

func TestGlobalLogFunctions_NilLogger(t *testing.T) {
	// Save and restore defaultLogger
	saved := defaultLogger
	defaultLogger = nil
	defer func() { defaultLogger = saved }()

	// These should not panic when defaultLogger is nil
	assert.NotPanics(t, func() { Debug("test %s", "msg") })
	assert.NotPanics(t, func() { Info("test %s", "msg") })
	assert.NotPanics(t, func() { Warn("test %s", "msg") })
	assert.NotPanics(t, func() { Error("test %s", "msg") })
}

func TestGlobalLogFunctions_WithLogger(t *testing.T) {
	saved := defaultLogger
	defer func() { defaultLogger = saved }()

	cfg := &config.Config{
		Server: config.ServerConfig{LogLevel: "debug"},
	}
	Init(cfg)

	// These should not panic
	assert.NotPanics(t, func() { Debug("debug message %s", "test") })
	assert.NotPanics(t, func() { Info("info message %s", "test") })
	assert.NotPanics(t, func() { Warn("warn message %s", "test") })
	assert.NotPanics(t, func() { Error("error message %s", "test") })
}

func TestLoggerMethods_NoPanic(t *testing.T) {
	l := &Logger{level: "debug"}

	assert.NotPanics(t, func() { l.Debug("test message") })
	assert.NotPanics(t, func() { l.Info("test %s", "with args") })
	assert.NotPanics(t, func() { l.Warn("warning: %d items", 5) })
	assert.NotPanics(t, func() { l.Error("error: %v", "details") })
}
