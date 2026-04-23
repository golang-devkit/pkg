package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendVerificationMessage(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sendVerificationMessage" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer TOKEN" {
			t.Fatalf("unexpected authorization header: %s", auth)
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

		if payload["phone_number"] != "+84901234567" {
			t.Fatalf("unexpected phone number: %#v", payload["phone_number"])
		}
		if payload["code_length"] != float64(6) {
			t.Fatalf("unexpected code length: %#v", payload["code_length"])
		}
		if payload["ttl"] != float64(60) {
			t.Fatalf("unexpected ttl: %#v", payload["ttl"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"ok":true,"result":{"request_id":"req-1","phone_number":"+84901234567","request_cost":0.01,"delivery_status":{"status":"sent","updated_at":1710000000}}}`)
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

	status, err := client.SendVerificationMessage(context.Background(), SendVerificationMessageRequest{
		PhoneNumber: "+84901234567",
		CodeLength:  6,
		TTL:         60,
		Payload:     "login:user-1",
	})
	if err != nil {
		t.Fatal(err)
	}

	if status == nil || status.RequestID != "req-1" {
		t.Fatalf("unexpected status: %#v", status)
	}
	if status.DeliveryStatus == nil || status.DeliveryStatus.Status != "sent" {
		t.Fatalf("unexpected delivery status: %#v", status.DeliveryStatus)
	}
}

func TestCheckVerificationStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/checkVerificationStatus" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}

		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatal(err)
		}

		if payload["request_id"] != "req-1" {
			t.Fatalf("unexpected request ID: %#v", payload["request_id"])
		}
		if payload["code"] != "123456" {
			t.Fatalf("unexpected code: %#v", payload["code"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"ok":true,"result":{"request_id":"req-1","phone_number":"+84901234567","request_cost":0.01,"verification_status":{"status":"code_valid","updated_at":1710000001,"code_entered":"123456"}}}`)
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

	status, err := client.CheckVerificationStatus(context.Background(), "req-1", "123456")
	if err != nil {
		t.Fatal(err)
	}

	if status == nil || status.VerificationStatus == nil {
		t.Fatalf("unexpected status: %#v", status)
	}
	if status.VerificationStatus.Status != "code_valid" {
		t.Fatalf("unexpected verification status: %#v", status.VerificationStatus)
	}
}

func TestSendVerificationMessageValidation(t *testing.T) {
	t.Parallel()

	client, err := New("TOKEN", WithRateLimit(RateLimitConfig{}))
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.SendVerificationMessage(context.Background(), SendVerificationMessageRequest{
		PhoneNumber: "0901234567",
		TTL:         10,
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "E.164") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGatewayErrorResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"ok":false,"error":"ACCESS_TOKEN_INVALID"}`)
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

	_, err = client.CheckSendAbility(context.Background(), "+84901234567")
	if err == nil {
		t.Fatal("expected gateway error")
	}

	var gatewayErr *Error
	if !errors.As(err, &gatewayErr) {
		t.Fatalf("expected Error, got %T", err)
	}
	if gatewayErr.Message != "ACCESS_TOKEN_INVALID" {
		t.Fatalf("unexpected gateway error: %#v", gatewayErr)
	}
}
