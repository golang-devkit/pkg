package telegram

import (
	"net/http"
	"net/url"
	"time"
)

const (
	defaultBaseURL         = "https://api.telegram.org"
	defaultTimeout         = 10 * time.Second
	defaultMinRequestDelay = 35 * time.Millisecond
	defaultBatchWorkers    = 4
)

// Logger is the minimal logger contract used by the client.
// Logger la interface logger tối thiểu mà client sử dụng.
type Logger interface {
	Printf(format string, args ...any)
}

// RetryConfig controls exponential backoff retry behavior.
// RetryConfig cấu hình retry theo exponential backoff.
type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Jitter      time.Duration
}

// normalized returns a safe retry config for runtime use.
func (r RetryConfig) normalized() RetryConfig {
	if r.MaxAttempts <= 0 {
		r.MaxAttempts = 3
	}
	if r.BaseDelay <= 0 {
		r.BaseDelay = 300 * time.Millisecond
	}
	if r.MaxDelay <= 0 {
		r.MaxDelay = 3 * time.Second
	}
	if r.MaxDelay < r.BaseDelay {
		r.MaxDelay = r.BaseDelay
	}
	if r.Jitter < 0 {
		r.Jitter = 0
	}
	return r
}

// RateLimitConfig adds client-side pacing before every Telegram API call.
// RateLimitConfig thêm pacing phía client trước mỗi lần gọi Telegram API.
type RateLimitConfig struct {
	Enabled     bool
	MinInterval time.Duration
}

// normalized returns a safe rate-limit config for runtime use.
func (r RateLimitConfig) normalized() RateLimitConfig {
	if !r.Enabled {
		return RateLimitConfig{}
	}
	if r.MinInterval <= 0 {
		r.MinInterval = defaultMinRequestDelay
	}
	return r
}

// Config holds the final client configuration after options are applied.
// Config giữ cấu hình cuối cùng của client sau khi apply options.
type Config struct {
	Token                      string
	BaseURL                    string
	DefaultParseMode           ParseMode
	DefaultDisableNotification bool
	Timeout                    time.Duration
	HTTPClient                 *http.Client
	Logger                     Logger
	ProxyURL                   *url.URL
	Retry                      RetryConfig
	RateLimit                  RateLimitConfig
	BatchConcurrency           int
}

func defaultConfig(token string) Config {
	return Config{
		Token:            token,
		BaseURL:          defaultBaseURL,
		Timeout:          defaultTimeout,
		Retry:            (RetryConfig{}).normalized(),
		RateLimit:        (RateLimitConfig{Enabled: true}).normalized(),
		BatchConcurrency: defaultBatchWorkers,
	}
}
