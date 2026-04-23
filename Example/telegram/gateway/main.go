package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/golang-devkit/pkg/telegram"
)

func main() {
	token := os.Getenv("TELEGRAM_GATEWAY_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_GATEWAY_TOKEN is required")
	}

	phoneNumber := "+84987654321"
	//if phoneNumber == "" {
	//	log.Fatal("TELEGRAM_GATEWAY_PHONE is required and must be in E.164 format")
	//}

	client, err := telegram.New(
		token,
		telegram.WithTimeout(10*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	status, err := client.SendVerificationMessage(ctx, telegram.SendVerificationMessageRequest{
		PhoneNumber: phoneNumber,
		CodeLength:  6,
		TTL:         60,
		Payload:     "login",
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Gateway request created successfully: request_id=%s cost=%.2f", status.RequestID, status.RequestCost)
}
