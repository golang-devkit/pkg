package telegram

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ErrorKind categorizes common Telegram failures.
// ErrorKind phân loại các lỗi Telegram phổ biến.
type ErrorKind string

const (
	// ErrorUnknown is the fallback category.
	// ErrorUnknown là category mặc định.
	ErrorUnknown ErrorKind = "unknown"
	// ErrorFlood indicates rate limiting / flood control.
	// ErrorFlood biểu thị rate limiting / flood control.
	ErrorFlood ErrorKind = "flood"
	// ErrorBlocked indicates the bot is blocked by the target.
	// ErrorBlocked biểu thị bot bị target chặn.
	ErrorBlocked ErrorKind = "blocked"
	// ErrorInvalidToken indicates an invalid bot token.
	// ErrorInvalidToken biểu thị bot token không hợp lệ.
	ErrorInvalidToken ErrorKind = "invalid_token"
	// ErrorChatNotFound indicates the target chat cannot be resolved.
	// ErrorChatNotFound biểu thị không tìm thấy target chat.
	ErrorChatNotFound ErrorKind = "chat_not_found"
	// ErrorForbidden indicates Telegram rejected the action as forbidden.
	// ErrorForbidden biểu thị Telegram từ chối do không đủ quyền.
	ErrorForbidden ErrorKind = "forbidden"
)

// Error wraps Telegram API and transport-level errors.
// Error wrap lỗi Telegram API và transport-level.
type Error struct {
	Method      string
	HTTPStatus  int
	Code        int
	Kind        ErrorKind
	Description string
	Parameters  *ResponseParameters
	Cause       error
}

// Error returns a compact formatted error message.
// Error trả về thông điệp lỗi ngắn gọn.
func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	parts := []string{"telegram"}
	if e.Method != "" {
		parts = append(parts, e.Method)
	}
	if e.Kind != "" && e.Kind != ErrorUnknown {
		parts = append(parts, string(e.Kind))
	}
	if e.Code != 0 {
		parts = append(parts, fmt.Sprintf("code=%d", e.Code))
	}
	if e.Description != "" {
		parts = append(parts, e.Description)
	}
	if e.Cause != nil {
		parts = append(parts, e.Cause.Error())
	}
	return strings.Join(parts, ": ")
}

// Unwrap exposes the underlying transport or parsing error.
// Unwrap expose lỗi transport hoặc parsing gốc.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// RetryAfter returns Telegram retry_after as time.Duration when present.
// RetryAfter trả về retry_after dưới dạng time.Duration nếu có.
func (e *Error) RetryAfter() time.Duration {
	if e == nil || e.Parameters == nil || e.Parameters.RetryAfter <= 0 {
		return 0
	}
	return time.Duration(e.Parameters.RetryAfter) * time.Second
}

// IsRetryable reports whether the error is safe to retry automatically.
// IsRetryable cho biết lỗi có thể retry tự động hay không.
func (e *Error) IsRetryable() bool {
	if e == nil {
		return false
	}
	if e.Kind == ErrorFlood {
		return true
	}
	return e.HTTPStatus >= http.StatusInternalServerError
}

// IsFlood reports whether err is a Telegram flood / rate-limit error.
// IsFlood cho biết err có phải lỗi flood / rate-limit hay không.
func IsFlood(err error) bool {
	var telegramErr *Error
	return errors.As(err, &telegramErr) && telegramErr.Kind == ErrorFlood
}

// IsBlocked reports whether err indicates that the bot is blocked.
// IsBlocked cho biết err có biểu thị bot bị chặn hay không.
func IsBlocked(err error) bool {
	var telegramErr *Error
	return errors.As(err, &telegramErr) && telegramErr.Kind == ErrorBlocked
}

// IsInvalidToken reports whether err indicates an invalid bot token.
// IsInvalidToken cho biết err có biểu thị bot token không hợp lệ hay không.
func IsInvalidToken(err error) bool {
	var telegramErr *Error
	return errors.As(err, &telegramErr) && telegramErr.Kind == ErrorInvalidToken
}

// IsChatNotFound reports whether err indicates that the target chat was not found.
// IsChatNotFound cho biết err có biểu thị target chat không tồn tại hay không.
func IsChatNotFound(err error) bool {
	var telegramErr *Error
	return errors.As(err, &telegramErr) && telegramErr.Kind == ErrorChatNotFound
}

func newTelegramError(method string, httpStatus, code int, description string, params *ResponseParameters, cause error) *Error {
	return &Error{
		Method:      method,
		HTTPStatus:  httpStatus,
		Code:        code,
		Kind:        classifyErrorKind(httpStatus, code, description),
		Description: description,
		Parameters:  params,
		Cause:       cause,
	}
}

func classifyErrorKind(httpStatus, code int, description string) ErrorKind {
	desc := strings.ToLower(description)

	switch {
	case httpStatus == http.StatusUnauthorized || code == http.StatusUnauthorized:
		return ErrorInvalidToken
	case httpStatus == http.StatusTooManyRequests || code == http.StatusTooManyRequests || strings.Contains(desc, "too many requests"):
		return ErrorFlood
	case strings.Contains(desc, "bot was blocked by the user"), strings.Contains(desc, "user is deactivated"):
		return ErrorBlocked
	case strings.Contains(desc, "chat not found"):
		return ErrorChatNotFound
	case httpStatus == http.StatusForbidden || code == http.StatusForbidden || strings.Contains(desc, "forbidden"), strings.Contains(desc, "bot can't initiate conversation"):
		return ErrorForbidden
	default:
		return ErrorUnknown
	}
}
