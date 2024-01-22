package log

import (
	"log/slog"
	"os"
)

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Error(msg string, args ...any)
	With(args ...any) *slog.Logger
}

func NewLogger(level slog.Leveler) Logger {
	j := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
	})
	sl := slog.New(j)
	return sl
}
