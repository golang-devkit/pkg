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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-devkit/pkg/telegram/internal/utils"
)

// Client is a thread-safe Telegram Bot API client.
// Client là Telegram Bot API client an toàn cho concurrent use.
type Client struct {
	token                      string
	baseURL                    string
	httpClient                 *http.Client
	logger                     Logger
	defaultParseMode           ParseMode
	defaultDisableNotification bool
	retry                      RetryConfig
	rateLimiter                *rateLimiter
	batchConcurrency           int
}

type rateLimiter struct {
	enabled     bool
	minInterval time.Duration

	mu   sync.Mutex
	next time.Time
}

func (r *rateLimiter) Wait(ctx context.Context) error {
	if r == nil || !r.enabled || r.minInterval <= 0 {
		return nil
	}

	r.mu.Lock()
	now := time.Now()
	waitFor := time.Duration(0)
	if now.Before(r.next) {
		waitFor = r.next.Sub(now)
	}
	r.next = now.Add(waitFor).Add(r.minInterval)
	r.mu.Unlock()

	if waitFor <= 0 {
		return nil
	}
	return sleepContext(ctx, waitFor)
}

type apiRequest struct {
	method     string
	jsonBody   any
	formFields map[string]string
	formFiles  []utils.MultipartFile
}

// New creates a Telegram Bot API client.
// New tạo Telegram Bot API client.
func New(token string, opts ...Option) (*Client, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("telegram: token is required")
	}

	cfg := defaultConfig(token)
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}

	if err := cfg.DefaultParseMode.validate(); err != nil {
		return nil, err
	}

	httpClient, err := buildHTTPClient(cfg)
	if err != nil {
		return nil, err
	}

	return &Client{
		token:                      cfg.Token,
		baseURL:                    strings.TrimRight(cfg.BaseURL, "/"),
		httpClient:                 httpClient,
		logger:                     cfg.Logger,
		defaultParseMode:           cfg.DefaultParseMode,
		defaultDisableNotification: cfg.DefaultDisableNotification,
		retry:                      cfg.Retry.normalized(),
		rateLimiter: &rateLimiter{
			enabled:     cfg.RateLimit.Enabled,
			minInterval: cfg.RateLimit.MinInterval,
		},
		batchConcurrency: cfg.BatchConcurrency,
	}, nil
}

// Send sends content to a Telegram username/channel handle or numeric chat string.
// Send gửi content đến username/channel handle hoặc chat string dạng số.
func (c *Client) Send(ctx context.Context, to string, content Content) error {
	_, err := c.SendMessage(ctx, to, content)
	return err
}

// SendChat sends content to a Telegram chat ID.
// SendChat gửi content đến Telegram chat ID.
func (c *Client) SendChat(ctx context.Context, chatID int64, content Content) error {
	_, err := c.SendChatMessage(ctx, chatID, content)
	return err
}

// SendMessage sends content and returns the first resulting message.
// SendMessage gửi content và trả về message đầu tiên nhận được.
//
// For media groups, the first message in the returned album is exposed here.
// Với media group, method này trả về message đầu tiên trong album.
func (c *Client) SendMessage(ctx context.Context, to string, content Content) (*Message, error) {
	if strings.TrimSpace(to) == "" {
		return nil, errors.New("telegram: target username/chat string is required")
	}
	return c.sendResolved(ctx, to, content)
}

// SendChatMessage sends content to a numeric chat ID and returns the first message.
// SendChatMessage gửi content đến chat ID dạng số và trả về message đầu tiên.
func (c *Client) SendChatMessage(ctx context.Context, chatID int64, content Content) (*Message, error) {
	if chatID == 0 {
		return nil, errors.New("telegram: chat ID is required")
	}
	return c.sendResolved(ctx, chatID, content)
}

// SendMediaGroup sends a media group and returns all messages in the album.
// SendMediaGroup gửi media group và trả về toàn bộ messages trong album.
func (c *Client) SendMediaGroup(ctx context.Context, to string, content Content) ([]Message, error) {
	if strings.TrimSpace(to) == "" {
		return nil, errors.New("telegram: target username/chat string is required")
	}
	return c.sendMediaGroupResolved(ctx, to, content)
}

// SendChatMediaGroup sends a media group to a numeric chat ID.
// SendChatMediaGroup gửi media group đến chat ID dạng số.
func (c *Client) SendChatMediaGroup(ctx context.Context, chatID int64, content Content) ([]Message, error) {
	if chatID == 0 {
		return nil, errors.New("telegram: chat ID is required")
	}
	return c.sendMediaGroupResolved(ctx, chatID, content)
}

// SendBatch sends multiple items concurrently and preserves result order.
// SendBatch gửi nhiều item đồng thời và giữ nguyên thứ tự kết quả.
func (c *Client) SendBatch(ctx context.Context, items []BatchItem) ([]BatchResult, error) {
	results := make([]BatchResult, len(items))
	if len(items) == 0 {
		return results, nil
	}

	workers := c.batchConcurrency
	if workers <= 0 {
		workers = 1
	}
	if workers > len(items) {
		workers = len(items)
	}

	jobs := make(chan int)
	var wg sync.WaitGroup

	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for index := range jobs {
				item := items[index]
				result := BatchResult{
					Index:  index,
					To:     item.To,
					ChatID: item.ChatID,
				}

				if ctx.Err() != nil {
					result.Error = ctx.Err()
					results[index] = result
					continue
				}

				if item.ChatID != 0 {
					if item.Content.Type == ContentMediaGroup {
						messages, err := c.sendMediaGroupResolved(ctx, item.ChatID, item.Content)
						if len(messages) > 0 {
							result.Message = &messages[0]
						}
						result.Messages = messages
						result.Error = err
					} else {
						result.Message, result.Error = c.sendResolved(ctx, item.ChatID, item.Content)
					}
				} else {
					target := strings.TrimSpace(item.To)
					if target == "" {
						result.Error = errors.New("telegram: batch item requires To or ChatID")
						results[index] = result
						continue
					}
					if item.Content.Type == ContentMediaGroup {
						messages, err := c.sendMediaGroupResolved(ctx, target, item.Content)
						if len(messages) > 0 {
							result.Message = &messages[0]
						}
						result.Messages = messages
						result.Error = err
					} else {
						result.Message, result.Error = c.sendResolved(ctx, target, item.Content)
					}
				}

				results[index] = result
			}
		}()
	}

	for index := range items {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return results, ctx.Err()
		case jobs <- index:
		}
	}

	close(jobs)
	wg.Wait()

	if ctx.Err() != nil {
		return results, ctx.Err()
	}
	return results, nil
}

func (c *Client) sendResolved(ctx context.Context, chatID any, content Content) (*Message, error) {
	if content.Type == ContentMediaGroup {
		messages, err := c.sendMediaGroupResolved(ctx, chatID, content)
		if err != nil {
			return nil, err
		}
		if len(messages) == 0 {
			return nil, errors.New("telegram: empty media group response")
		}
		return &messages[0], nil
	}

	req, err := c.buildRequest(chatID, content)
	if err != nil {
		return nil, err
	}

	message, err := doAPI[Message](ctx, c, req)
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (c *Client) sendMediaGroupResolved(ctx context.Context, chatID any, content Content) ([]Message, error) {
	if content.Type != ContentMediaGroup {
		return nil, errors.New("telegram: content type must be media_group")
	}

	req, err := c.buildRequest(chatID, content)
	if err != nil {
		return nil, err
	}

	return doAPI[[]Message](ctx, c, req)
}

func (c *Client) buildRequest(chatID any, content Content) (apiRequest, error) {
	if err := content.validate(c.defaultParseMode); err != nil {
		return apiRequest{}, err
	}

	switch content.Type {
	case "", ContentText, ContentHTML, ContentMarkdown, ContentMarkdownV2:
		return c.buildTextRequest(chatID, content)
	case ContentPhoto:
		return c.buildMediaRequest("sendPhoto", "photo", chatID, content, true)
	case ContentDocument:
		return c.buildMediaRequest("sendDocument", "document", chatID, content, true)
	case ContentVideo:
		return c.buildMediaRequest("sendVideo", "video", chatID, content, true)
	case ContentAudio:
		return c.buildMediaRequest("sendAudio", "audio", chatID, content, true)
	case ContentVoice:
		return c.buildMediaRequest("sendVoice", "voice", chatID, content, true)
	case ContentSticker:
		return c.buildMediaRequest("sendSticker", "sticker", chatID, content, false)
	case ContentAnimation:
		return c.buildMediaRequest("sendAnimation", "animation", chatID, content, true)
	case ContentMediaGroup:
		return c.buildMediaGroupRequest(chatID, content)
	default:
		return apiRequest{}, fmt.Errorf("telegram: unsupported content type %q", content.Type)
	}
}

func (c *Client) buildTextRequest(chatID any, content Content) (apiRequest, error) {
	body := map[string]any{
		"chat_id": chatID,
		"text":    content.Text,
	}

	if disable := c.resolveDisableNotification(content); disable != nil {
		body["disable_notification"] = *disable
	}
	if preview := resolveOptionalBool(content.DisableWebPagePreviewSet, content.DisableWebPagePreview); preview != nil {
		body["disable_web_page_preview"] = *preview
	}
	if protect := resolveOptionalBool(content.ProtectContentSet, content.ProtectContent); protect != nil {
		body["protect_content"] = *protect
	}
	if content.ReplyMarkup != nil {
		body["reply_markup"] = content.ReplyMarkup
	}
	if parseMode := content.effectiveParseMode(c.defaultParseMode); parseMode != "" {
		body["parse_mode"] = string(parseMode)
	}

	return apiRequest{
		method:   "sendMessage",
		jsonBody: body,
	}, nil
}

func (c *Client) buildMediaRequest(apiMethod, fileField string, chatID any, content Content, withCaption bool) (apiRequest, error) {
	fields := map[string]any{
		"chat_id": chatID,
	}
	if disable := c.resolveDisableNotification(content); disable != nil {
		fields["disable_notification"] = *disable
	}
	if protect := resolveOptionalBool(content.ProtectContentSet, content.ProtectContent); protect != nil {
		fields["protect_content"] = *protect
	}
	if content.ReplyMarkup != nil {
		fields["reply_markup"] = content.ReplyMarkup
	}

	attachName := fileField
	fields[fileField] = content.File.valueOrAttach(attachName)

	if withCaption && strings.TrimSpace(content.Caption) != "" {
		fields["caption"] = content.Caption
	}
	if withCaption {
		if parseMode := content.effectiveParseMode(c.defaultParseMode); parseMode != "" {
			fields["parse_mode"] = string(parseMode)
		}
	}

	if content.File.isUpload() {
		formFields, err := stringifyFormFields(fields)
		if err != nil {
			return apiRequest{}, err
		}
		return apiRequest{
			method:     apiMethod,
			formFields: formFields,
			formFiles: []utils.MultipartFile{
				c.toMultipartFile(fileField, content.File),
			},
		}, nil
	}

	return apiRequest{
		method:   apiMethod,
		jsonBody: fields,
	}, nil
}

func (c *Client) buildMediaGroupRequest(chatID any, content Content) (apiRequest, error) {
	fields := map[string]any{
		"chat_id": chatID,
	}
	if disable := c.resolveDisableNotification(content); disable != nil {
		fields["disable_notification"] = *disable
	}
	if protect := resolveOptionalBool(content.ProtectContentSet, content.ProtectContent); protect != nil {
		fields["protect_content"] = *protect
	}

	mediaPayload := make([]map[string]any, 0, len(content.Media))
	files := make([]utils.MultipartFile, 0)
	hasUpload := false

	for index, item := range content.Media {
		entry := map[string]any{
			"type": string(item.Type),
		}

		attachName := "media_" + strconv.Itoa(index)
		entry["media"] = item.File.valueOrAttach(attachName)
		if item.Caption != "" {
			entry["caption"] = item.Caption
		}
		if parseMode := item.effectiveParseMode(c.defaultParseMode); parseMode != "" {
			entry["parse_mode"] = string(parseMode)
		}
		if item.File.isUpload() {
			hasUpload = true
			files = append(files, c.toMultipartFile(attachName, item.File))
		}
		mediaPayload = append(mediaPayload, entry)
	}

	fields["media"] = mediaPayload

	if hasUpload {
		formFields, err := stringifyFormFields(fields)
		if err != nil {
			return apiRequest{}, err
		}
		return apiRequest{
			method:     "sendMediaGroup",
			formFields: formFields,
			formFiles:  files,
		}, nil
	}

	return apiRequest{
		method:   "sendMediaGroup",
		jsonBody: fields,
	}, nil
}

func doAPI[T any](ctx context.Context, c *Client, req apiRequest) (T, error) {
	var zero T
	retry := c.retry.normalized()

	var lastErr error
	for attempt := 1; attempt <= retry.MaxAttempts; attempt++ {
		if err := c.rateLimiter.Wait(ctx); err != nil {
			return zero, err
		}

		result, err := doOnce[T](ctx, c, req)
		if err == nil {
			return result, nil
		}

		lastErr = err
		if !c.shouldRetry(err) || attempt == retry.MaxAttempts {
			return zero, err
		}

		delay := c.retryDelay(err, retry, attempt)
		if c.logger != nil {
			c.logger.Printf("telegram retry attempt=%d method=%s delay=%s err=%v", attempt+1, req.method, delay, err)
		}
		if err := sleepContext(ctx, delay); err != nil {
			return zero, err
		}
	}

	if lastErr == nil {
		lastErr = errors.New("telegram: request failed")
	}
	return zero, lastErr
}

func doOnce[T any](ctx context.Context, c *Client, req apiRequest) (T, error) {
	var zero T
	httpReq, err := c.newHTTPRequest(ctx, req)
	if err != nil {
		return zero, err
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return zero, normalizeTransportError(req.method, err)
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return zero, newTelegramError(req.method, resp.StatusCode, resp.StatusCode, "read response body failed", nil, readErr)
	}

	var envelope apiResponse[T]
	if unmarshalErr := json.Unmarshal(body, &envelope); unmarshalErr != nil {
		if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
			return zero, newTelegramError(req.method, resp.StatusCode, resp.StatusCode, "decode response failed", nil, unmarshalErr)
		}
		return zero, newTelegramError(req.method, resp.StatusCode, resp.StatusCode, http.StatusText(resp.StatusCode), nil, unmarshalErr)
	}

	if !envelope.Ok {
		return zero, newTelegramError(req.method, resp.StatusCode, envelope.ErrorCode, envelope.Description, envelope.Parameters, nil)
	}

	return envelope.Result, nil
}

func (c *Client) newHTTPRequest(ctx context.Context, req apiRequest) (*http.Request, error) {
	endpoint := c.endpoint(req.method)
	if len(req.formFiles) > 0 {
		return utils.NewMultipartRequest(ctx, endpoint, req.formFields, req.formFiles)
	}
	return utils.NewJSONRequest(ctx, endpoint, req.jsonBody)
}

func (c *Client) endpoint(method string) string {
	return fmt.Sprintf("%s/bot%s/%s", c.baseURL, c.token, method)
}

func (c *Client) toMultipartFile(fieldName string, file *InputFile) utils.MultipartFile {
	return utils.MultipartFile{
		FieldName:   fieldName,
		FileName:    file.fileName(),
		ContentType: file.ContentType,
		Open:        file.open,
	}
}

func (c *Client) resolveDisableNotification(content Content) *bool {
	if content.DisableNotificationSet {
		return boolPtr(content.DisableNotification)
	}
	if c.defaultDisableNotification {
		return boolPtr(true)
	}
	return nil
}

func (c *Client) shouldRetry(err error) bool {
	var telegramErr *Error
	if errors.As(err, &telegramErr) && telegramErr.IsRetryable() {
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	return false
}

func (c *Client) retryDelay(err error, retry RetryConfig, attempt int) time.Duration {
	var telegramErr *Error
	if errors.As(err, &telegramErr) {
		if after := telegramErr.RetryAfter(); after > 0 {
			return after
		}
	}

	delay := retry.BaseDelay
	for i := 1; i < attempt; i++ {
		delay *= 2
		if delay >= retry.MaxDelay {
			delay = retry.MaxDelay
			break
		}
	}

	if retry.Jitter > 0 {
		delay += time.Duration(rand.Int63n(int64(retry.Jitter) + 1))
	}

	if delay > retry.MaxDelay {
		return retry.MaxDelay
	}
	return delay
}

func buildHTTPClient(cfg Config) (*http.Client, error) {
	if cfg.HTTPClient != nil {
		cloned := *cfg.HTTPClient
		if cfg.Timeout > 0 {
			cloned.Timeout = cfg.Timeout
		}
		if cfg.ProxyURL != nil {
			transport, err := cloneTransport(cloned.Transport)
			if err != nil {
				return nil, err
			}
			transport.Proxy = http.ProxyURL(cfg.ProxyURL)
			cloned.Transport = transport
		}
		return &cloned, nil
	}

	transport, err := cloneTransport(nil)
	if err != nil {
		return nil, err
	}
	if cfg.ProxyURL != nil {
		transport.Proxy = http.ProxyURL(cfg.ProxyURL)
	}

	return &http.Client{
		Timeout:   cfg.Timeout,
		Transport: transport,
	}, nil
}

func cloneTransport(rt http.RoundTripper) (*http.Transport, error) {
	if rt == nil {
		if base, ok := http.DefaultTransport.(*http.Transport); ok {
			return base.Clone(), nil
		}
		return &http.Transport{}, nil
	}

	transport, ok := rt.(*http.Transport)
	if !ok {
		return nil, errors.New("telegram: custom transport must be *http.Transport when using proxy")
	}
	return transport.Clone(), nil
}

func normalizeTransportError(method string, err error) error {
	if err == nil {
		return nil
	}
	return newTelegramError(method, 0, 0, "transport error", nil, err)
}

func stringifyFormFields(fields map[string]any) (map[string]string, error) {
	result := make(map[string]string, len(fields))
	for key, value := range fields {
		stringValue, err := stringifyFormValue(value)
		if err != nil {
			return nil, fmt.Errorf("telegram: encode field %q: %w", key, err)
		}
		result[key] = stringValue
	}
	return result, nil
}

func stringifyFormValue(value any) (string, error) {
	switch typed := value.(type) {
	case string:
		return typed, nil
	case bool:
		return strconv.FormatBool(typed), nil
	case int:
		return strconv.Itoa(typed), nil
	case int64:
		return strconv.FormatInt(typed, 10), nil
	case ParseMode:
		return string(typed), nil
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
}

func resolveOptionalBool(set bool, value bool) *bool {
	if !set {
		return nil
	}
	return boolPtr(value)
}

func boolPtr(value bool) *bool {
	return &value
}

func sleepContext(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
