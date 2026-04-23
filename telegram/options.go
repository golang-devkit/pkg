package telegram

import (
	"net/http"
	"time"
)

// Option mutates Config during client construction.
// Option thay đổi Config trong lúc khởi tạo client.
type Option = GatewayOption

// WithTimeout sets the HTTP client timeout.
// WithTimeout đặt timeout cho HTTP client.
func WithTimeout(timeout time.Duration) Option {
	return WithGatewayTimeout(timeout)
}

// WithHTTPClient injects a custom HTTP client.
// WithHTTPClient inject custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return WithGatewayHTTPClient(client)
}

// WithLogger injects a logger used for retry and transport diagnostics.
// WithLogger inject logger dùng cho retry và chẩn đoán transport.
func WithLogger(logger Logger) Option {
	return WithGatewayLogger(logger)
}

// WithProxy configures an outbound HTTP proxy.
// WithProxy cấu hình HTTP proxy outbound.
func WithProxy(rawURL string) Option {
	return WithGatewayProxy(rawURL)
}

// WithRetry configures automatic retry with exponential backoff.
// WithRetry cấu hình retry tự động với exponential backoff.
func WithRetry(retry RetryConfig) Option {
	return WithGatewayRetry(retry)
}

// WithRateLimit configures client-side pacing between requests.
// WithRateLimit cấu hình pacing phía client giữa các request.
func WithRateLimit(rateLimit RateLimitConfig) Option {
	return WithGatewayRateLimit(rateLimit)
}

// WithBaseURL overrides the Telegram Gateway API base URL.
// WithBaseURL override base URL của Telegram Gateway API.
func WithBaseURL(rawURL string) Option {
	return WithGatewayBaseURL(rawURL)
}
