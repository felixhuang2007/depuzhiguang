package logger

import (
	"io"
	"log/slog"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

// New creates a structured JSON logger that writes to both stdout and a rotating
// log file via lumberjack. The service name is included in every log entry.
func New(service string) *slog.Logger {
	fileWriter := &lumberjack.Logger{
		Filename:   "/app/logs/service.log",
		MaxSize:    100, // megabytes
		MaxBackups: 10,
		MaxAge:     30, // days
		Compress:   true,
	}

	multi := io.MultiWriter(os.Stdout, fileWriter)

	handler := slog.NewJSONHandler(multi, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	return slog.New(handler).With("service", service)
}
