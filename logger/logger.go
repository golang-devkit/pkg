package logger

import (
	"fmt"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
)

var (
	logOnce sync.Once
	logger  *zap.Logger
)

func SetLogEntry(zlg *zap.Logger) {
	if zlg == nil {
		fmt.Println("entry logger set is unavailable, skipping logger setup")
		return
	}

	// cache the logger
	buffered := logger
	// Set the logger by the given zap.Logger
	logger = zlg
	// Flush any buffered entry entries
	if buffered != nil {
		if err := buffered.Sync(); err != nil {
			logger.Debug("Flush the buffered entry entries", zap.Error(err))
		}
	}
}

func NewEntry() *zap.Logger {
	logOnce.Do(func() {
		if logger != nil {
			// If logger is already initialized, skip re-initialization
			return
		}
		// Renew the logger default
		val := os.Getenv(EnvDeploymentKey)
		switch val {
		case "development", "dev":
			// Initialize logger
			config := zap.NewDevelopmentConfig()
			config.Encoding = "json"
			dev, err := config.Build()
			if err != nil {
				logger = zap.NewExample()
				logger.Debug("failed to initialize production logger", zap.Error(err))
			} else {

				logger = dev
			}
		default:
			// Consider production environment
			// Initialize logger
			prod, err := zap.NewProduction()
			if err != nil {
				logger = zap.NewExample()
				logger.Debug("failed to initialize production logger", zap.Error(err))
			} else {
				logger = prod
			}
		}
	})

	return logger.With(
		zap.Time(KeyTimestamp, time.Now()),
		zap.String(KeyEnvironment, os.Getenv(EnvDeploymentKey)))
}
