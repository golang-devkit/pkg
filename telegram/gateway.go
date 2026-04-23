package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/golang-devkit/pkg/telegram/internal/utils"
)

const defaultGatewayBaseURL = "https://gatewayapi.telegram.org"

var gatewayPhonePattern = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

// GatewayConfig holds the final Telegram Gateway client configuration.
// GatewayConfig giữ cấu hình cuối cùng của Telegram Gateway client.
type GatewayConfig struct {
	Token      string
	BaseURL    string
	Timeout    time.Duration
	HTTPClient *http.Client
	Logger     Logger
	ProxyURL   *url.URL
	Retry      RetryConfig
	RateLimit  RateLimitConfig
}

// GatewayOption mutates GatewayConfig during client construction.
// GatewayOption thay đổi GatewayConfig trong lúc khởi tạo client.
type GatewayOption func(*GatewayConfig) error

// GatewayClient is a thread-safe Telegram Gateway API client.
// GatewayClient là Telegram Gateway API client an toàn cho concurrent use.
type GatewayClient struct {
	token       string
	baseURL     string
	httpClient  *http.Client
	logger      Logger
	retry       RetryConfig
	rateLimiter *rateLimiter
	randSource  *rand.Rand
}

// GatewaySendVerificationMessageRequest is the payload for sendVerificationMessage.
// GatewaySendVerificationMessageRequest là payload cho sendVerificationMessage.
type GatewaySendVerificationMessageRequest struct {
	PhoneNumber    string `json:"phone_number"`
	RequestID      string `json:"request_id,omitempty"`
	SenderUsername string `json:"sender_username,omitempty"`
	Code           string `json:"code,omitempty"`
	CodeLength     int    `json:"code_length,omitempty"`
	CallbackURL    string `json:"callback_url,omitempty"`
	Payload        string `json:"payload,omitempty"`
	TTL            int    `json:"ttl,omitempty"`
}

// GatewayRequestStatus is the Telegram Gateway request status object.
// GatewayRequestStatus là object trạng thái request của Telegram Gateway.
type GatewayRequestStatus struct {
	RequestID          string                     `json:"request_id"`
	PhoneNumber        string                     `json:"phone_number"`
	RequestCost        float64                    `json:"request_cost"`
	IsRefunded         bool                       `json:"is_refunded,omitempty"`
	RemainingBalance   float64                    `json:"remaining_balance,omitempty"`
	DeliveryStatus     *GatewayDeliveryStatus     `json:"delivery_status,omitempty"`
	VerificationStatus *GatewayVerificationStatus `json:"verification_status,omitempty"`
	Payload            string                     `json:"payload,omitempty"`
}

// GatewayDeliveryStatus is the delivery status returned by Telegram Gateway.
// GatewayDeliveryStatus là trạng thái giao message của Telegram Gateway.
type GatewayDeliveryStatus struct {
	Status    string `json:"status"`
	UpdatedAt int64  `json:"updated_at"`
}

// GatewayVerificationStatus is the verification status returned by Telegram Gateway.
// GatewayVerificationStatus là trạng thái xác thực của Telegram Gateway.
type GatewayVerificationStatus struct {
	Status      string `json:"status"`
	UpdatedAt   int64  `json:"updated_at"`
	CodeEntered string `json:"code_entered,omitempty"`
}

// GatewayError wraps Telegram Gateway API and transport-level errors.
// GatewayError wrap lỗi Telegram Gateway API và transport-level.
type GatewayError struct {
	Method     string
	HTTPStatus int
	Message    string
	Cause      error
}

type gatewayAPIResponse[T any] struct {
	Ok     bool   `json:"ok"`
	Result T      `json:"result"`
	Error  string `json:"error,omitempty"`
}

type gatewayPhoneNumberRequest struct {
	PhoneNumber string `json:"phone_number"`
}

type gatewayVerificationStatusRequest struct {
	RequestID string `json:"request_id"`
	Code      string `json:"code,omitempty"`
}

type gatewayRevokeRequest struct {
	RequestID string `json:"request_id"`
}

func defaultGatewayConfig(token string) GatewayConfig {
	return GatewayConfig{
		Token:     token,
		BaseURL:   defaultGatewayBaseURL,
		Timeout:   defaultTimeout,
		Retry:     (RetryConfig{}).normalized(),
		RateLimit: (RateLimitConfig{Enabled: true}).normalized(),
	}
}

// NewGateway creates a Telegram Gateway API client.
// NewGateway tạo Telegram Gateway API client.
func NewGateway(token string, opts ...GatewayOption) (*GatewayClient, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("telegram gateway: token is required")
	}

	cfg := defaultGatewayConfig(token)
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}

	httpClient, err := buildHTTPClient(Config{
		Timeout:    cfg.Timeout,
		HTTPClient: cfg.HTTPClient,
		ProxyURL:   cfg.ProxyURL,
	})
	if err != nil {
		return nil, err
	}

	return &GatewayClient{
		token:      cfg.Token,
		baseURL:    strings.TrimRight(cfg.BaseURL, "/"),
		httpClient: httpClient,
		logger:     cfg.Logger,
		retry:      cfg.Retry.normalized(),
		rateLimiter: &rateLimiter{
			enabled:     cfg.RateLimit.Enabled,
			minInterval: cfg.RateLimit.MinInterval,
		},
		randSource: rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

// WithGatewayTimeout sets the HTTP client timeout.
// WithGatewayTimeout đặt timeout cho HTTP client.
func WithGatewayTimeout(timeout time.Duration) GatewayOption {
	return func(cfg *GatewayConfig) error {
		if timeout <= 0 {
			return errors.New("telegram gateway: timeout must be greater than zero")
		}
		cfg.Timeout = timeout
		return nil
	}
}

// WithGatewayHTTPClient injects a custom HTTP client.
// WithGatewayHTTPClient inject custom HTTP client.
func WithGatewayHTTPClient(client *http.Client) GatewayOption {
	return func(cfg *GatewayConfig) error {
		if client == nil {
			return errors.New("telegram gateway: http client cannot be nil")
		}
		cfg.HTTPClient = client
		return nil
	}
}

// WithGatewayLogger injects a logger used for retry and transport diagnostics.
// WithGatewayLogger inject logger dùng cho retry và chẩn đoán transport.
func WithGatewayLogger(logger Logger) GatewayOption {
	return func(cfg *GatewayConfig) error {
		cfg.Logger = logger
		return nil
	}
}

// WithGatewayProxy configures an outbound HTTP proxy.
// WithGatewayProxy cấu hình HTTP proxy outbound.
func WithGatewayProxy(rawURL string) GatewayOption {
	return func(cfg *GatewayConfig) error {
		if strings.TrimSpace(rawURL) == "" {
			return errors.New("telegram gateway: proxy URL cannot be empty")
		}
		parsed, err := url.Parse(rawURL)
		if err != nil {
			return fmt.Errorf("telegram gateway: parse proxy URL: %w", err)
		}
		cfg.ProxyURL = parsed
		return nil
	}
}

// WithGatewayRetry configures automatic retry with exponential backoff.
// WithGatewayRetry cấu hình retry tự động với exponential backoff.
func WithGatewayRetry(retry RetryConfig) GatewayOption {
	return func(cfg *GatewayConfig) error {
		cfg.Retry = retry.normalized()
		return nil
	}
}

// WithGatewayRateLimit configures client-side pacing between requests.
// WithGatewayRateLimit cấu hình pacing phía client giữa các request.
func WithGatewayRateLimit(rateLimit RateLimitConfig) GatewayOption {
	return func(cfg *GatewayConfig) error {
		cfg.RateLimit = rateLimit.normalized()
		return nil
	}
}

// WithGatewayBaseURL overrides the Telegram Gateway API base URL.
// WithGatewayBaseURL override base URL của Telegram Gateway API.
func WithGatewayBaseURL(rawURL string) GatewayOption {
	return func(cfg *GatewayConfig) error {
		if strings.TrimSpace(rawURL) == "" {
			return errors.New("telegram gateway: base URL cannot be empty")
		}
		cfg.BaseURL = strings.TrimRight(rawURL, "/")
		return nil
	}
}

// SendVerificationMessage sends an OTP or verification code through Telegram Gateway.
// SendVerificationMessage gửi OTP hoặc mã xác thực qua Telegram Gateway.
func (c *GatewayClient) SendVerificationMessage(ctx context.Context, req GatewaySendVerificationMessageRequest) (*GatewayRequestStatus, error) {
	req = req.normalized()
	if err := req.validate(); err != nil {
		return nil, err
	}
	if req.Code != "" {
		req.CodeLength = 0
	}

	result, err := doGatewayAPI[GatewayRequestStatus](ctx, c, "sendVerificationMessage", req)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// CheckSendAbility checks whether Telegram Gateway can deliver a verification message.
// CheckSendAbility kiểm tra Telegram Gateway có thể giao verification message hay không.
func (c *GatewayClient) CheckSendAbility(ctx context.Context, phoneNumber string) (*GatewayRequestStatus, error) {
	phoneNumber = strings.TrimSpace(phoneNumber)
	if err := validateGatewayPhoneNumber(phoneNumber); err != nil {
		return nil, err
	}

	result, err := doGatewayAPI[GatewayRequestStatus](ctx, c, "checkSendAbility", gatewayPhoneNumberRequest{
		PhoneNumber: phoneNumber,
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// CheckVerificationStatus checks request status and optionally validates a code entered by the user.
// CheckVerificationStatus kiểm tra trạng thái request và có thể validate code người dùng nhập.
func (c *GatewayClient) CheckVerificationStatus(ctx context.Context, requestID, code string) (*GatewayRequestStatus, error) {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return nil, errors.New("telegram gateway: request ID is required")
	}

	code = strings.TrimSpace(code)
	if code != "" {
		if err := validateGatewayCode(code); err != nil {
			return nil, err
		}
	}

	result, err := doGatewayAPI[GatewayRequestStatus](ctx, c, "checkVerificationStatus", gatewayVerificationStatusRequest{
		RequestID: requestID,
		Code:      code,
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// RevokeVerificationMessage revokes a previously sent verification message.
// RevokeVerificationMessage thu hồi verification message đã gửi trước đó.
func (c *GatewayClient) RevokeVerificationMessage(ctx context.Context, requestID string) (bool, error) {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return false, errors.New("telegram gateway: request ID is required")
	}

	result, err := doGatewayAPI[bool](ctx, c, "revokeVerificationMessage", gatewayRevokeRequest{
		RequestID: requestID,
	})
	if err != nil {
		return false, err
	}
	return result, nil
}

// Error returns a compact formatted error message.
// Error trả về thông điệp lỗi ngắn gọn.
func (e *GatewayError) Error() string {
	if e == nil {
		return ""
	}

	parts := []string{"telegram gateway"}
	if e.Method != "" {
		parts = append(parts, e.Method)
	}
	if e.Message != "" {
		parts = append(parts, e.Message)
	}
	if e.Cause != nil {
		parts = append(parts, e.Cause.Error())
	}
	return strings.Join(parts, ": ")
}

// Unwrap exposes the underlying transport or parsing error.
// Unwrap expose lỗi transport hoặc parsing gốc.
func (e *GatewayError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// IsRetryable reports whether the error is safe to retry automatically.
// IsRetryable cho biết lỗi có thể retry tự động hay không.
func (e *GatewayError) IsRetryable() bool {
	if e == nil {
		return false
	}
	return e.HTTPStatus >= http.StatusInternalServerError
}

func (r GatewaySendVerificationMessageRequest) normalized() GatewaySendVerificationMessageRequest {
	r.PhoneNumber = strings.TrimSpace(r.PhoneNumber)
	r.RequestID = strings.TrimSpace(r.RequestID)
	r.SenderUsername = strings.TrimSpace(r.SenderUsername)
	r.Code = strings.TrimSpace(r.Code)
	r.CallbackURL = strings.TrimSpace(r.CallbackURL)
	return r
}

func (r GatewaySendVerificationMessageRequest) validate() error {
	if err := validateGatewayPhoneNumber(r.PhoneNumber); err != nil {
		return err
	}
	if r.Code != "" {
		if err := validateGatewayCode(r.Code); err != nil {
			return err
		}
	}
	if r.Code == "" && r.CodeLength != 0 {
		if r.CodeLength < 4 || r.CodeLength > 8 {
			return errors.New("telegram gateway: code length must be between 4 and 8")
		}
	}
	if r.TTL != 0 && (r.TTL < 30 || r.TTL > 3600) {
		return errors.New("telegram gateway: ttl must be between 30 and 3600 seconds")
	}
	if r.CallbackURL != "" {
		parsed, err := url.Parse(r.CallbackURL)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return errors.New("telegram gateway: callback URL must be a valid HTTPS URL")
		}
	}
	if len([]byte(r.Payload)) > 128 {
		return errors.New("telegram gateway: payload must not exceed 128 bytes")
	}
	return nil
}

func validateGatewayPhoneNumber(phoneNumber string) error {
	if phoneNumber == "" {
		return errors.New("telegram gateway: phone number is required")
	}
	if !gatewayPhonePattern.MatchString(phoneNumber) {
		return errors.New("telegram gateway: phone number must be in E.164 format")
	}
	return nil
}

func validateGatewayCode(code string) error {
	if len(code) < 4 || len(code) > 8 {
		return errors.New("telegram gateway: code must be a numeric string between 4 and 8 digits")
	}
	for _, r := range code {
		if r < '0' || r > '9' {
			return errors.New("telegram gateway: code must be a numeric string between 4 and 8 digits")
		}
	}
	return nil
}

func doGatewayAPI[T any](ctx context.Context, c *GatewayClient, method string, body any) (T, error) {
	var zero T
	retry := c.retry.normalized()

	var lastErr error
	for attempt := 1; attempt <= retry.MaxAttempts; attempt++ {
		if err := c.rateLimiter.Wait(ctx); err != nil {
			return zero, err
		}

		result, err := doGatewayOnce[T](ctx, c, method, body)
		if err == nil {
			return result, nil
		}

		lastErr = err
		if !c.shouldRetry(err) || attempt == retry.MaxAttempts {
			return zero, err
		}

		delay := c.retryDelay(retry, attempt)
		if c.logger != nil {
			c.logger.Printf("telegram gateway retry attempt=%d method=%s delay=%s err=%v", attempt+1, method, delay, err)
		}
		if err := sleepContext(ctx, delay); err != nil {
			return zero, err
		}
	}

	if lastErr == nil {
		lastErr = errors.New("telegram gateway: request failed")
	}
	return zero, lastErr
}

func doGatewayOnce[T any](ctx context.Context, c *GatewayClient, method string, body any) (T, error) {
	var zero T

	httpReq, err := utils.NewJSONRequest(ctx, c.endpoint(method), body)
	if err != nil {
		return zero, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return zero, normalizeGatewayTransportError(method, err)
	}
	defer resp.Body.Close()

	payload, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return zero, newGatewayError(method, resp.StatusCode, "read response body failed", readErr)
	}

	var envelope gatewayAPIResponse[T]
	if unmarshalErr := json.Unmarshal(payload, &envelope); unmarshalErr != nil {
		message := "decode response failed"
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			message = http.StatusText(resp.StatusCode)
		}
		return zero, newGatewayError(method, resp.StatusCode, message, unmarshalErr)
	}

	if !envelope.Ok {
		if envelope.Error == "" {
			envelope.Error = http.StatusText(resp.StatusCode)
		}
		return zero, newGatewayError(method, resp.StatusCode, envelope.Error, nil)
	}

	return envelope.Result, nil
}

func (c *GatewayClient) endpoint(method string) string {
	return fmt.Sprintf("%s/%s", c.baseURL, method)
}

func (c *GatewayClient) shouldRetry(err error) bool {
	var gatewayErr *GatewayError
	if errors.As(err, &gatewayErr) && gatewayErr.IsRetryable() {
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	return false
}

func (c *GatewayClient) retryDelay(retry RetryConfig, attempt int) time.Duration {
	delay := retry.BaseDelay
	for i := 1; i < attempt; i++ {
		delay *= 2
		if delay >= retry.MaxDelay {
			delay = retry.MaxDelay
			break
		}
	}

	if retry.Jitter > 0 {
		delay += time.Duration(c.randSource.Int63n(int64(retry.Jitter) + 1))
	}

	if delay > retry.MaxDelay {
		return retry.MaxDelay
	}
	return delay
}

func newGatewayError(method string, httpStatus int, message string, cause error) *GatewayError {
	return &GatewayError{
		Method:     method,
		HTTPStatus: httpStatus,
		Message:    message,
		Cause:      cause,
	}
}

func normalizeGatewayTransportError(method string, err error) error {
	if err == nil {
		return nil
	}
	return newGatewayError(method, 0, "transport error", err)
}
