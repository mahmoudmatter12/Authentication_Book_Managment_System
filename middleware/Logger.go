package middleware

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger    *zap.Logger
	initOnce  sync.Once
	logFile   *os.File
	fileMutex sync.Mutex
)

// LoggerConfig defines configuration for the logger middleware
type LoggerConfig struct {
	LogDirectory  string
	LogFileName   string
	MaxSizeMB     int
	MaxBackups    int
	MaxAgeDays    int
	EnableConsole bool
	LogLevel      string
}

// DefaultLoggerConfig provides default configuration
func DefaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		LogDirectory:  "logs",
		LogFileName:   "application.log",
		MaxSizeMB:     10,
		MaxBackups:    5,
		MaxAgeDays:    30,
		EnableConsole: true,
		LogLevel:      "info",
	}
}

// InitLogger initializes the global logger instance
func InitLogger(config LoggerConfig) (*zap.Logger, error) {
	var initErr error
	initOnce.Do(func() {
		// Ensure log directory exists
		if err := os.MkdirAll(config.LogDirectory, 0755); err != nil {
			initErr = fmt.Errorf("failed to create log directory: %w", err)
			return
		}

		// Create or open log file
		logPath := filepath.Join(config.LogDirectory, config.LogFileName)
		logFile, initErr = os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if initErr != nil {
			initErr = fmt.Errorf("failed to open log file: %w", initErr)
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

	return logger, initErr
}

// GetLogger returns the initialized logger instance
func GetLogger() *zap.Logger {
	if logger == nil {
		// Fallback to a basic console logger if not initialized
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		logger, _ = config.Build()
	}
	return logger
}

// Logger returns a gin.HandlerFunc that logs requests
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get logger instance
		log := GetLogger()

		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		clientIP := c.ClientIP()
		method := c.Request.Method
		userAgent := c.Request.UserAgent()

		// Read request body
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Process request
		c.Next()

		// Collect response data
		end := time.Now()
		latency := end.Sub(start)
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// Log the request
		log.Info("HTTP Request",
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", clientIP),
			zap.String("user-agent", userAgent),
			zap.String("time", end.Format(time.RFC3339)),
			zap.Duration("latency", latency),
			zap.Int("status", statusCode),
			zap.String("error", errorMessage),
			zap.ByteString("body", requestBody),
		)

		// Additional slow request warning
		if latency > time.Second {
			log.Warn("Slow request detected",
				zap.String("path", path),
				zap.Duration("latency", latency),
			)
		}
	}
}

// RecoveryWithLogger returns a gin.HandlerFunc that recovers from panics and logs them
func RecoveryWithLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get logger instance
				log := GetLogger()

				// Get stack trace
				stack := make([]byte, 4096)
				length := runtime.Stack(stack, false)
				stackTrace := string(stack[:length])

				// Log the panic
				log.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("request", c.Request.Method+" "+c.Request.URL.Path),
					zap.String("stack", stackTrace),
				)

				// Respond with error
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "Internal Server Error",
				})
			}
		}()
		c.Next()
	}
}

// CloseLogger closes the log file and flushes any buffered logs
func CloseLogger() error {
	fileMutex.Lock()
	defer fileMutex.Unlock()

	if logger != nil {
		if err := logger.Sync(); err != nil {
			return err
		}
	}

	if logFile != nil {
		return logFile.Close()
	}

	return nil
}