package logger_test

import (
	"context"
	"testing"

	"github.com/golang-devkit/pkg/logger"
	"go.uber.org/zap"
)

type ContextLoggerType string

const ContextLogger ContextLoggerType = "logger-on-context"

func Test_SetLoggerFromContext(t *testing.T) {
	// Prepare a context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set logger to context
	outGoingCtx := logger.SetLoggerToContext(ctx, logger.NewEntry())

	// Get logger from context
	//
	// Act 1:
	if logger, ok := outGoingCtx.Value(ContextLogger).(*zap.Logger); !ok {
		t.Errorf("Act 1 | logger not found in context")
	} else {
		t.Logf("Act 1 | result: %#v\n", *logger)
	}

	// Get logger from context
	// Act 2:
	if logger, ok := outGoingCtx.Value(logger.ContextLogger).(*zap.Logger); !ok {
		t.Errorf("Act 2 | logger not found in context")
	} else {
		t.Logf("Act 2 | result: %#v\n", *logger)
	}
}

func Test_GetLoggerFromContext(t *testing.T) {
	// Prepare a context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Act 1:
	// Set logger to context
	ctxV1 := context.WithValue(ctx, "logger-on-context", logger.NewEntry())
	// Get logger from context
	v1 := ctxV1.Value(logger.ContextLogger)
	switch lg := v1.(type) {
	case *zap.Logger:
		t.Logf("Act 1 | result: %#v\n", *lg)
	default:
		t.Errorf("Act 1 | logger not found in context")
	}
	// if lg := ctxV1.Value(logger.ContextLogger).(*zap.Logger); lg == nil {
	// 	t.Errorf("Act 1 | logger not found in context")
	// } else {
	// 	t.Logf("Act 1 | result: %#v\n", *lg)
	// }

	// Act 2:
	// Set logger to context
	ctxV2 := context.WithValue(ctx, "logger-on-context", logger.NewEntry())
	// Get logger from context
	v2 := ctxV2.Value(logger.ContextLogger)
	switch lg := v2.(type) {
	case *zap.Logger:
		t.Logf("Act 2 | result: %#v\n", *lg)
	default:
		t.Errorf("Act 2 | logger not found in context")
	}
}
