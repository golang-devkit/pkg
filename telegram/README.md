# Telegram Gateway API Client

A simple, thread-safe Telegram Gateway API client for Go. This package is focused on phone-number verification through Telegram Gateway and no longer includes the Telegram Bot API surface.

## Installation

```bash
go get github.com/golang-devkit/pkg/telegram
```

## Initialization

```go
package main

import (
	"log"
	"time"

	"github.com/golang-devkit/pkg/telegram"
)

func main() {
	client, err := telegram.New(
		"YOUR_GATEWAY_TOKEN_HERE",
		telegram.WithTimeout(15*time.Second),
		telegram.WithRetry(telegram.RetryConfig{
			MaxAttempts: 3,
			BaseDelay:   300 * time.Millisecond,
			MaxDelay:    3 * time.Second,
		}),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	_ = client
}
```

## Sending OTP

```go
status, err := client.SendVerificationMessage(ctx, telegram.SendVerificationMessageRequest{
	PhoneNumber: "+84901234567", // E.164 format
	CodeLength:  6,
	TTL:         60,
	Payload:     "login:user-123",
})
if err != nil {
	log.Fatal(err)
}

log.Printf("request_id=%s cost=%.2f", status.RequestID, status.RequestCost)
```

## Checking Delivery Ability

```go
status, err := client.CheckSendAbility(ctx, "+84901234567")
if err != nil {
	log.Fatal(err)
}

log.Printf("request_id=%s", status.RequestID)
```

## Checking OTP Status

```go
status, err := client.CheckVerificationStatus(ctx, "request-id-from-send", "123456")
if err != nil {
	log.Fatal(err)
}

if status.VerificationStatus != nil && status.VerificationStatus.Status == "code_valid" {
	log.Println("OTP is valid")
}
```

## Revoking an OTP Message

```go
ok, err := client.RevokeVerificationMessage(ctx, "request-id-from-send")
if err != nil {
	log.Fatal(err)
}

log.Printf("revocation accepted=%v", ok)
```

## Compatibility Aliases

The package still exposes `NewGateway`, `GatewayClient`, and the `Gateway*` request/response types as compatibility aliases, but the primary API is now `telegram.New(...)`.
