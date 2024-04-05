package logger

import (
	"log/slog"
	"os"
)

func InitLogger(debugIsEnabled bool) *slog.Logger {
	loglevel := slog.LevelInfo
	if debugIsEnabled {
		loglevel = slog.LevelDebug
	}

	lgr := slog.New(
		slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{
				// AddSource: true,
				Level: loglevel,
			},
		),
	)

	slog.SetDefault(lgr)

	return lgr
}
