package logger

import (
	"context"

	"go.uber.org/zap"
)

type ContextLoggerType string

const ContextLogger ContextLoggerType = "logger-on-context"

func GetLoggerFromContext(ctx context.Context) *zap.Logger {
	value := ctx.Value(ContextLogger)
	switch logger := value.(type) {
	case *zap.Logger:
		return logger
	default:
		return NewEntry()
	}
}

func SetLoggerToContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, ContextLogger, logger)
}
