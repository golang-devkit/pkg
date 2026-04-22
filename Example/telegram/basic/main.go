package main

import (
	"context"
	"log"
	"os"

	"github.com/golang-devkit/pkg/telegram"
)

func main() {
	// Load the bot token from the environment instead of hard-coding it.
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is required")
	}

	// 1. Initialize the Telegram client
	client, err := telegram.New(token)
	if err != nil {
		log.Fatalf("Failed to initialize telegram client: %v", err)
	}
	log.Println("Telegram client initialized successfully")

	// 2. Prepare the content to send
	messageContent := telegram.Content{
		Type: telegram.ContentText,
		Text: "Hello! This is a test message from my new Telegram bot.",
	}

	// 3. Send the message
	// Note: You must replace "YOUR_CHAT_ID" with your actual numeric Telegram Chat ID,
	// or use a public channel handle like "@my_channel_name" with SendMessage()
	// Make sure the bot is added to your chat or channel first!
	var chatID int64 = 5890042997

	log.Printf("Sending message to chat ID: %d", chatID)

	ctx := context.Background()
	message, err := client.SendChatMessage(ctx, chatID, messageContent)
	if err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}

	log.Printf("Message sent successfully! Message ID: %d", message.MessageID)
}
