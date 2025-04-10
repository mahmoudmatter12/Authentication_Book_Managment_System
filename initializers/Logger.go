package initializers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger     *zap.Logger
	once       sync.Once
	logFile    *os.File
	fileMutex  sync.Mutex
)

// LoggerConfig defines configuration for the logger middleware
type LoggerConfig struct {
	LogDirectory  string
	LogFileName   string
	MaxSizeMB     int    // Maximum size in MB before rotation
	MaxBackups    int    // Maximum number of old log files to retain
	MaxAgeDays    int    // Maximum number of days to retain old log files
	EnableConsole bool   // Whether to log to console
	LogLevel      string // debug, info, warn, error
}

// DefaultLoggerConfig provides default configuration
func DefaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		LogDirectory:  "logs",
		LogFileName:   "logs.log",
		MaxSizeMB:     10,
		MaxBackups:    5,
		MaxAgeDays:    30,
		EnableConsole: true,
		LogLevel:      "info",
	}
}



// InitLogger initializes the global logger instance
func InitLogger(config LoggerConfig) error {
	var err error
	once.Do(func() {
		// Ensure log directory exists
		if err := os.MkdirAll(config.LogDirectory, 0755); err != nil {
			fmt.Printf("Failed to create log directory: %v\n", err)
			return
		}

		// Create or open log file
		logPath := filepath.Join(config.LogDirectory, config.LogFileName)
		logFile, err = os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Failed to open log file: %v\n", err)
			return
		}

		// Configure zap logger
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

		// Set log level
		var level zapcore.Level
		switch strings.ToLower(config.LogLevel) {
		case "debug":
			level = zapcore.DebugLevel
		case "warn":
			level = zapcore.WarnLevel
		case "error":
			level = zapcore.ErrorLevel
		default:
			level = zapcore.InfoLevel
		}

		// Create cores
		cores := []zapcore.Core{
			zapcore.NewCore(
				zapcore.NewJSONEncoder(encoderConfig),
				zapcore.AddSync(logFile),
				level,
			),
		}

		if config.EnableConsole {
			cores = append(cores, zapcore.NewCore(
				zapcore.NewConsoleEncoder(encoderConfig),
				zapcore.AddSync(os.Stdout),
				level,
			))
		}

		core := zapcore.NewTee(cores...)
		logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	})

	return err
}