package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
)

var Logger *zap.Logger
var Suggar *zap.SugaredLogger

func ParseLogLevel(levelStr string) zapcore.Level {
	// levelStr := strings.ToLower(os.Getenv("LOG_LEVEL"))
	var level zapcore.Level
	switch levelStr {
	case "debug":
		level = zapcore.DebugLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "fatal":
		level = zapcore.FatalLevel
	default:
		level = zapcore.InfoLevel
	}
	return level
}

// createLogger returns a zap.Logger that logs to both stderr and a log file.
func CreateLogger(level zapcore.Level, filename string) *zap.Logger {
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Outputs
	consoleSync := zapcore.Lock(os.Stderr)

	logPath := filepath.Join(os.TempDir(), filename)

	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		panic("Failed to open log file: " + err.Error())
	}
	fileSync := zapcore.AddSync(logFile)

	// Encoder
	encoder := zapcore.NewJSONEncoder(encoderCfg)

	// Combine cores: stderr + file
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, consoleSync, level),
		zapcore.NewCore(encoder, fileSync, zapcore.DebugLevel),
	)

	// Create logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.FatalLevel))
	return logger
}
