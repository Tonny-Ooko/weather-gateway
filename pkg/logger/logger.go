package logger

import (
	"log/slog"
	"os"
)

var Log *slog.Logger

func InitializeLogger() {
	Log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(Log)
}

func Info(message string, keysAndValues ...any) {
	Log.Info(message, keysAndValues...)
}

func Error(message string, err error, keysAndValues ...any) {
	args := append([]any{"error", err.Error()}, keysAndValues...)
	Log.Error(message, args...)
}
