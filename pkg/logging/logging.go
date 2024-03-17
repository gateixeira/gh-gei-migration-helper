package logging

import (
	"context"
	"log/slog"
	"os"
	"sync"
)

type ContextKey string

const IDKey ContextKey = "ID"

var (
	logKeys = map[ContextKey]bool{
		IDKey: true,
	}

	isDebug bool
	once    sync.Once
)

func NewLoggerFromContext(ctx context.Context, debug bool) *slog.Logger {
	once.Do(func() {
		isDebug = debug
	})

	lvl := new(slog.LevelVar)

	if isDebug {
		lvl.Set(slog.LevelDebug)
	} else {
		lvl.Set(slog.LevelInfo)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: lvl,
	}))
	slog.SetDefault(logger)

	if ctx != nil {
		for key := range logKeys {
			value := ctx.Value(key)
			if value != nil {
				logger = logger.With(string(key), value)
			}
		}
	}

	return logger
}
