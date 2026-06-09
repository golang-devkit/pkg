package telegram

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestSendChatMessageJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/botTOKEN/sendMessage" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if ct := r.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
			t.Fatalf("unexpected content type: %s", ct)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}

		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatal(err)
		}

		if payload["text"] != "<b>Hello</b>" {
			t.Fatalf("unexpected text: %#v", payload["text"])
		}
		if payload["parse_mode"] != "HTML" {
			t.Fatalf("unexpected parse_mode: %#v", payload["parse_mode"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"ok":true,"result":{"message_id":1,"date":1710000000,"chat":{"id":123,"type":"private"},"text":"<b>Hello</b>"}}`)
	}))
	defer server.Close()

	client, err := New(
		"TOKEN",
		WithBaseURL(server.URL),
		WithRateLimit(RateLimitConfig{}),
	)
	if err != nil {
		t.Fatal(err)
	}

	message, err := client.SendChatMessage(context.Background(), 123, Content{
		Type: ContentHTML,
		Text: "<b>Hello</b>",
	})
	if err != nil {
		t.Fatal(err)
	}

	if message == nil || message.MessageID != 1 {
		t.Fatalf("unexpected message: %#v", message)
	}
}

func TestSendPhotoMultipartUpload(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "image.txt")
	if err := os.WriteFile(filePath, []byte("image-bytes"), 0o600); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/botTOKEN/sendPhoto" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		reader, err := r.MultipartReader()
		if err != nil {
			t.Fatal(err)
		}

		seenCaption := false
		seenFile := false

		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}

			data, err := io.ReadAll(part)
			if err != nil {
				t.Fatal(err)
			}

			switch part.FormName() {
			case "caption":
				seenCaption = string(data) == "hello"
			case "photo":
				seenFile = string(data) == "image-bytes"
			}
		}

		if !seenCaption || !seenFile {
			t.Fatalf("missing expected multipart fields: caption=%v file=%v", seenCaption, seenFile)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"ok":true,"result":{"message_id":7,"date":1710000000,"chat":{"id":123,"type":"private"},"caption":"hello"}}`)
	}))
	defer server.Close()

	client, err := New(
		"TOKEN",
		WithBaseURL(server.URL),
		WithRateLimit(RateLimitConfig{}),
	)
	if err != nil {
		t.Fatal(err)
	}

	message, err := client.SendChatMessage(context.Background(), 123, Content{
		Type:    ContentPhoto,
		File:    FileFromPath(filePath),
		Caption: "hello",
	})
	if err != nil {
		t.Fatal(err)
	}

	if message == nil || message.MessageID != 7 {
		t.Fatalf("unexpected message: %#v", message)
	}
}

func TestRetryOnServerError(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) == 1 {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = io.WriteString(w, `{"ok":false,"error_code":502,"description":"bad gateway"}`)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"ok":true,"result":{"message_id":3,"date":1710000000,"chat":{"id":123,"type":"private"},"text":"retry ok"}}`)
	}))
	defer server.Close()

	client, err := New(
		"TOKEN",
		WithBaseURL(server.URL),
		WithRateLimit(RateLimitConfig{}),
		WithRetry(RetryConfig{
			MaxAttempts: 2,
			BaseDelay:   10 * time.Millisecond,
			MaxDelay:    20 * time.Millisecond,
		}),
	)
	if err != nil {
		t.Fatal(err)
	}

	message, err := client.SendChatMessage(context.Background(), 123, Content{
		Type: ContentText,
		Text: "retry ok",
	})
	if err != nil {
		t.Fatal(err)
	}
	if message.MessageID != 3 {
		t.Fatalf("unexpected message: %#v", message)
	}
	if attempts.Load() != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts.Load())
	}
}

func TestErrorClassification(t *testing.T) {
	t.Parallel()

	err := newTelegramError("sendMessage", http.StatusTooManyRequests, 429, "Too Many Requests: retry later", &ResponseParameters{RetryAfter: 1}, nil)
	if !IsFlood(err) {
		t.Fatal("expected flood error")
	}
	if err.RetryAfter() != time.Second {
		t.Fatalf("unexpected retry after: %s", err.RetryAfter())
	}
}

func TestSendBatch(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"ok":true,"result":{"message_id":1,"date":1710000000,"chat":{"id":123,"type":"private"},"text":"ok"}}`)
	}))
	defer server.Close()

	client, err := New(
		"TOKEN",
		WithBaseURL(server.URL),
		WithRateLimit(RateLimitConfig{}),
		WithBatchConcurrency(2),
	)
	if err != nil {
		t.Fatal(err)
	}

	results, err := client.SendBatch(context.Background(), []BatchItem{
		{ChatID: 123, Content: Content{Type: ContentText, Text: "a"}},
		{To: "@channel", Content: Content{Type: ContentText, Text: "b"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("unexpected result count: %d", len(results))
	}
	for _, result := range results {
		if result.Error != nil {
			t.Fatalf("unexpected batch error: %v", result.Error)
		}
		if result.Message == nil {
			t.Fatal("expected message")
		}
	}
}

func TestMultipartBuilderCompatibility(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "file.txt")
	if err := os.WriteFile(filePath, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reader, err := r.MultipartReader()
		if err != nil {
			t.Fatal(err)
		}
		count := 0
		for {
			_, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			count++
		}
		if count == 0 {
			t.Fatal("expected multipart parts")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"ok":true,"result":{"message_id":9,"date":1710000000,"chat":{"id":1,"type":"private"},"caption":"ok"}}`)
	}))
	defer server.Close()

	client, err := New("TOKEN", WithBaseURL(server.URL), WithRateLimit(RateLimitConfig{}))
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.SendChatMessage(context.Background(), 1, Content{
		Type:    ContentDocument,
		File:    FileFromPath(filePath),
		Caption: "ok",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestMultipartEncodingIsValid(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		form, err := r.MultipartReader()
		if err != nil {
			t.Fatal(err)
		}
		part, err := form.NextPart()
		if err != nil {
			t.Fatal(err)
		}
		if part.FormName() == "" {
			t.Fatal("expected multipart file")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"ok":true,"result":{"message_id":1,"date":1710000000,"chat":{"id":1,"type":"private"}}}`)
	}))
	defer server.Close()

	client, err := New("TOKEN", WithBaseURL(server.URL), WithRateLimit(RateLimitConfig{}))
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.SendChatMessage(context.Background(), 1, Content{
		Type: ContentPhoto,
		File: FileFromBytes("img.txt", []byte("abc")),
	})
	if err != nil {
		t.Fatal(err)
	}
}
