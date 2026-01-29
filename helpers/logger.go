package helpers

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/lmittmann/tint"
)

func SetupLogger(level slog.Level) {

	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		Level:      level,
		TimeFormat: time.StampMilli,
	}))
	slog.SetDefault(logger)
}

func ParseLogLevel(levelStr string) (level slog.Level) {

	switch strings.ToLower(levelStr) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	return
}
