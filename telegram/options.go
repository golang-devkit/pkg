package telegram

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Option mutates Config during client construction.
// Option thay đổi Config trong lúc khởi tạo client.
type Option func(*Config) error

// WithParseMode sets the default parse mode for text and captions.
// WithParseMode đặt parse mode mặc định cho text và caption.
func WithParseMode(mode ParseMode) Option {
	return func(cfg *Config) error {
		if err := mode.validate(); err != nil {
			return err
		}
		cfg.DefaultParseMode = mode
		return nil
	}
}

// WithDefaultDisableNotification sets the default silent-send behavior.
// WithDefaultDisableNotification đặt mặc định gửi silent/not silent.
func WithDefaultDisableNotification(disable bool) Option {
	return func(cfg *Config) error {
		cfg.DefaultDisableNotification = disable
		return nil
	}
}

// WithTimeout sets the HTTP client timeout.
// WithTimeout đặt timeout cho HTTP client.
func WithTimeout(timeout time.Duration) Option {
	return func(cfg *Config) error {
		if timeout <= 0 {
			return errors.New("telegram: timeout must be greater than zero")
		}
		cfg.Timeout = timeout
		return nil
	}
}

// WithHTTPClient injects a custom HTTP client.
// WithHTTPClient inject custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(cfg *Config) error {
		if client == nil {
			return errors.New("telegram: http client cannot be nil")
		}
		cfg.HTTPClient = client
		return nil
	}
}

// WithLogger injects a logger used for retry and transport diagnostics.
// WithLogger inject logger dùng cho retry và chẩn đoán transport.
func WithLogger(logger Logger) Option {
	return func(cfg *Config) error {
		cfg.Logger = logger
		return nil
	}
}

// WithProxy configures an outbound HTTP proxy.
// WithProxy cấu hình HTTP proxy outbound.
func WithProxy(rawURL string) Option {
	return func(cfg *Config) error {
		if strings.TrimSpace(rawURL) == "" {
			return errors.New("telegram: proxy URL cannot be empty")
		}
		parsed, err := url.Parse(rawURL)
		if err != nil {
			return fmt.Errorf("telegram: parse proxy URL: %w", err)
		}
		cfg.ProxyURL = parsed
		return nil
	}
}

// WithRetry configures automatic retry with exponential backoff.
// WithRetry cấu hình retry tự động với exponential backoff.
func WithRetry(retry RetryConfig) Option {
	return func(cfg *Config) error {
		cfg.Retry = retry.normalized()
		return nil
	}
}

// WithRateLimit configures client-side pacing between requests.
// WithRateLimit cấu hình pacing phía client giữa các request.
func WithRateLimit(rateLimit RateLimitConfig) Option {
	return func(cfg *Config) error {
		cfg.RateLimit = rateLimit.normalized()
		return nil
	}
}

// WithBatchConcurrency sets the worker count used by SendBatch.
// WithBatchConcurrency đặt số worker mà SendBatch sử dụng.
func WithBatchConcurrency(n int) Option {
	return func(cfg *Config) error {
		if n <= 0 {
			return errors.New("telegram: batch concurrency must be greater than zero")
		}
		cfg.BatchConcurrency = n
		return nil
	}
}

// WithBaseURL overrides the Telegram API base URL.
// WithBaseURL override base URL của Telegram API.
func WithBaseURL(rawURL string) Option {
	return func(cfg *Config) error {
		if strings.TrimSpace(rawURL) == "" {
			return errors.New("telegram: base URL cannot be empty")
		}
		cfg.BaseURL = strings.TrimRight(rawURL, "/")
		return nil
	}
}
