// Модуль logger позволяет реализовать продвинутое логирование
package logger

import (
	"log/slog"
	"os"
)

// Init - констуктор логгера
func Init(level string, format string) error {
	var logLevel slog.Level
	var logHandler slog.Handler

	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	switch format {
	case "json":
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	default:
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	}

	slog.SetDefault(slog.New(logHandler))

	return nil
}
