package helpers

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/lmittmann/tint"
)

// SetupLogger настраивает структурированный логгер с tint для цветного вывода
// level - уровень логирования (slog.LevelDebug, slog.LevelInfo и т.д.)
func SetupLogger(level slog.Level) {

	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		Level:      level,
		TimeFormat: time.StampMilli,
	}))
	slog.SetDefault(logger)
}

// ParseLogLevel парсит уровень логирования из строки
// Поддерживаемые значения: "debug", "info", "warn", "error"
// По умолчанию возвращает slog.LevelInfo
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
